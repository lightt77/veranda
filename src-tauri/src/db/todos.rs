//! Database operations for todos

use crate::db::Database;
use crate::models::todo::{
    CreateTodoRequest, LabelSummary, ProjectSummary, Todo, TodoFilter, TodoStatus,
    UpdateTodoRequest,
};
use chrono::Utc;
use rusqlite::{params, Result, Row};
use serde_json;
use std::collections::HashMap;

/// Helper to parse optional datetime from string
fn parse_datetime_opt(s: Option<String>) -> Option<chrono::DateTime<Utc>> {
    s.and_then(|str| chrono::DateTime::parse_from_rfc3339(&str).ok())
        .map(|dt| dt.with_timezone(&Utc))
}

/// Helper to parse required datetime from string
fn parse_datetime(s: String) -> chrono::DateTime<Utc> {
    chrono::DateTime::parse_from_rfc3339(&s)
        .map(|dt| dt.with_timezone(&Utc))
        .unwrap_or_else(|_| Utc::now())
}

/// Helper to convert a database row to Todo
fn row_to_todo(row: &Row) -> Result<Todo> {
    let status_str: String = row.get(5)?;
    let status = TodoStatus::try_from(status_str.as_str()).map_err(|_| {
        rusqlite::Error::InvalidColumnType(5, "TodoStatus".to_string(), rusqlite::types::Type::Text)
    })?;

    let labels_json: String = row.get(4)?;
    let labels = serde_json::from_str(&labels_json).unwrap_or_default();

    Ok(Todo {
        id: row.get(0)?,
        task_name: row.get(1)?,
        task_desc: row.get(2)?,
        project: row.get(3)?,
        labels,
        status,
        last_status_update_ts: parse_datetime_opt(row.get(6)?),
        created_at: parse_datetime(row.get(7)?),
    })
}

impl Database {
    /// Create a new todo in the database
    pub fn create_todo(&self, req: CreateTodoRequest) -> Result<Todo> {
        let mut todo = Todo::new(req.task_name);
        todo.task_desc = req.task_desc;
        todo.project = req.project;
        todo.labels = req.labels.unwrap_or_default();

        let labels_json = serde_json::to_string(&todo.labels).unwrap_or_else(|_| "[]".to_string());

        self.conn().execute(
            "INSERT INTO todos (id, task_name, task_desc, project, labels, status, last_status_update_ts, created_at)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
            params![
                todo.id,
                todo.task_name,
                todo.task_desc,
                todo.project,
                labels_json,
                todo.status.as_str(),
                todo.last_status_update_ts.map(|t| t.to_rfc3339()),
                todo.created_at.to_rfc3339()
            ],
        )?;

        Ok(todo)
    }

    /// Get a todo by ID
    pub fn get_todo(&self, id: &str) -> Result<Option<Todo>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, task_name, task_desc, project, labels, status, last_status_update_ts, created_at
             FROM todos WHERE id = ?1"
        )?;

        let todo = stmt.query_row([id], row_to_todo);

        match todo {
            Ok(t) => Ok(Some(t)),
            Err(rusqlite::Error::QueryReturnedNoRows) => Ok(None),
            Err(e) => Err(e),
        }
    }

    /// Get todos with optional filtering
    pub fn get_todos(&self, filter: Option<TodoFilter>) -> Result<Vec<Todo>> {
        let mut where_clauses = vec![];
        let mut query_params: Vec<Box<dyn rusqlite::ToSql>> = vec![];

        if let Some(f) = filter {
            if let Some(status) = f.status {
                where_clauses.push("status = ?".to_string());
                query_params.push(Box::new(status.as_str().to_string()));
            }

            if let Some(project) = f.project {
                where_clauses.push("project = ?".to_string());
                query_params.push(Box::new(project));
            }

            if let Some(label) = f.label {
                where_clauses.push("labels LIKE ?".to_string());
                query_params.push(Box::new(format!("%{}%", label)));
            }

            if let Some(search) = f.search {
                where_clauses.push("(task_name LIKE ? OR task_desc LIKE ?)".to_string());
                let pattern = format!("%{}%", search);
                query_params.push(Box::new(pattern.clone()));
                query_params.push(Box::new(pattern));
            }
        }

        let sql = if where_clauses.is_empty() {
            "SELECT id, task_name, task_desc, project, labels, status, last_status_update_ts, created_at
             FROM todos ORDER BY created_at DESC".to_string()
        } else {
            format!(
                "SELECT id, task_name, task_desc, project, labels, status, last_status_update_ts, created_at
                 FROM todos WHERE {} ORDER BY created_at DESC",
                where_clauses.join(" AND ")
            )
        };

        let conn = self.conn();
        let mut stmt = conn.prepare(&sql)?;
        let param_refs: Vec<&dyn rusqlite::ToSql> =
            query_params.iter().map(|p| p.as_ref()).collect();

        let rows = stmt.query_map(param_refs.as_slice(), row_to_todo)?;
        rows.collect()
    }

    /// Update a todo
    pub fn update_todo(&self, id: &str, req: UpdateTodoRequest) -> Result<bool> {
        let todo = match self.get_todo(id)? {
            Some(t) => t,
            None => return Ok(false),
        };

        let task_name = req.task_name.unwrap_or(todo.task_name);
        let task_desc = req.task_desc.or(todo.task_desc);
        let project = req.project.or(todo.project);
        let labels = req.labels.unwrap_or(todo.labels);
        let labels_json = serde_json::to_string(&labels).unwrap_or_else(|_| "[]".to_string());

        let rows = self.conn().execute(
            "UPDATE todos SET task_name = ?1, task_desc = ?2, project = ?3, labels = ?4 WHERE id = ?5",
            params![task_name, task_desc, project, labels_json, id],
        )?;

        Ok(rows > 0)
    }

    /// Update todo status
    pub fn update_todo_status(&self, id: &str, status: TodoStatus) -> Result<bool> {
        let now = Utc::now().to_rfc3339();

        let rows = self.conn().execute(
            "UPDATE todos SET status = ?1, last_status_update_ts = ?2 WHERE id = ?3",
            params![status.as_str(), now, id],
        )?;

        Ok(rows > 0)
    }

    /// Delete a todo
    pub fn delete_todo(&self, id: &str) -> Result<bool> {
        let rows = self
            .conn()
            .execute("DELETE FROM todos WHERE id = ?1", [id])?;

        Ok(rows > 0)
    }

    /// Get all projects with counts
    pub fn get_projects(&self) -> Result<Vec<ProjectSummary>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT 
                COALESCE(project, 'No Project') as name,
                COUNT(*) as total,
                SUM(CASE WHEN status = 'IN_PROGRESS' THEN 1 ELSE 0 END) as in_progress,
                SUM(CASE WHEN status = 'DONE' THEN 1 ELSE 0 END) as done
             FROM todos
             WHERE status != 'ARCHIVED'
             GROUP BY project
             ORDER BY name",
        )?;

        let rows = stmt.query_map([], |row| {
            Ok(ProjectSummary {
                name: row.get(0)?,
                todo_count: row.get(1)?,
                in_progress_count: row.get::<_, Option<i64>>(2)?.unwrap_or(0),
                done_count: row.get::<_, Option<i64>>(3)?.unwrap_or(0),
            })
        })?;

        rows.collect()
    }

    /// Get all labels with usage counts
    pub fn get_labels(&self) -> Result<Vec<LabelSummary>> {
        let conn = self.conn();
        let mut stmt =
            conn.prepare("SELECT labels FROM todos WHERE labels IS NOT NULL AND labels != '[]'")?;

        let rows = stmt.query_map([], |row| -> Result<String> { row.get(0) })?;

        let mut counts: HashMap<String, i64> = HashMap::new();

        for labels_json in rows {
            if let Ok(json) = labels_json {
                if let Ok(labels) = serde_json::from_str::<Vec<String>>(&json) {
                    for label in labels {
                        *counts.entry(label).or_insert(0) += 1;
                    }
                }
            }
        }

        let mut summaries: Vec<LabelSummary> = counts
            .into_iter()
            .map(|(name, usage_count)| LabelSummary { name, usage_count })
            .collect();

        summaries.sort_by(|a, b| b.usage_count.cmp(&a.usage_count));

        Ok(summaries)
    }

    /// Get todos by project
    pub fn get_todos_by_project(&self, project: &str) -> Result<Vec<Todo>> {
        let filter = TodoFilter {
            project: Some(project.to_string()),
            ..Default::default()
        };
        self.get_todos(Some(filter))
    }

    /// Get active todos (not done or archived)
    pub fn get_active_todos(&self) -> Result<Vec<Todo>> {
        let conn = self.conn();
        let sql = "SELECT id, task_name, task_desc, project, labels, status, last_status_update_ts, created_at
                   FROM todos WHERE status IN ('TODO', 'IN_PROGRESS') ORDER BY created_at DESC";

        let mut stmt = conn.prepare(sql)?;
        let rows = stmt.query_map([], row_to_todo)?;
        rows.collect()
    }
}
