use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Todo status enumeration
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "UPPERCASE")]
pub enum TodoStatus {
    Todo,
    InProgress,
    Done,
    Archived,
}

impl TodoStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            TodoStatus::Todo => "TODO",
            TodoStatus::InProgress => "IN_PROGRESS",
            TodoStatus::Done => "DONE",
            TodoStatus::Archived => "ARCHIVED",
        }
    }

    pub fn display_name(&self) -> &'static str {
        match self {
            TodoStatus::Todo => "To Do",
            TodoStatus::InProgress => "In Progress",
            TodoStatus::Done => "Done",
            TodoStatus::Archived => "Archived",
        }
    }
}

impl TryFrom<&str> for TodoStatus {
    type Error = String;

    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value {
            "TODO" => Ok(TodoStatus::Todo),
            "IN_PROGRESS" => Ok(TodoStatus::InProgress),
            "DONE" => Ok(TodoStatus::Done),
            "ARCHIVED" => Ok(TodoStatus::Archived),
            _ => Err(format!("Invalid todo status: {}", value)),
        }
    }
}

/// Todo data structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Todo {
    pub id: String,
    pub task_name: String,
    pub task_desc: Option<String>,
    pub project: Option<String>,
    pub labels: Vec<String>,
    pub status: TodoStatus,
    pub last_status_update_ts: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

impl Todo {
    /// Create a new todo
    pub fn new(task_name: impl Into<String>) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            task_name: task_name.into(),
            task_desc: None,
            project: None,
            labels: vec![],
            status: TodoStatus::Todo,
            last_status_update_ts: Some(now),
            created_at: now,
        }
    }

    /// Set description
    pub fn with_description(mut self, desc: impl Into<String>) -> Self {
        self.task_desc = Some(desc.into());
        self
    }

    /// Set project
    pub fn with_project(mut self, project: impl Into<String>) -> Self {
        self.project = Some(project.into());
        self
    }

    /// Add labels
    pub fn with_labels(mut self, labels: Vec<String>) -> Self {
        self.labels = labels;
        self
    }

    /// Move to a new status
    pub fn move_to(&mut self, new_status: TodoStatus) {
        self.status = new_status;
        self.last_status_update_ts = Some(Utc::now());
    }

    /// Mark as in progress
    pub fn start(&mut self) {
        self.move_to(TodoStatus::InProgress);
    }

    /// Mark as done
    pub fn complete(&mut self) {
        self.move_to(TodoStatus::Done);
    }

    /// Archive the todo
    pub fn archive(&mut self) {
        self.move_to(TodoStatus::Archived);
    }

    /// Check if todo is done
    pub fn is_done(&self) -> bool {
        self.status == TodoStatus::Done
    }

    /// Check if todo is active (not done or archived)
    pub fn is_active(&self) -> bool {
        matches!(self.status, TodoStatus::Todo | TodoStatus::InProgress)
    }

    /// Add a label
    pub fn add_label(&mut self, label: impl Into<String>) {
        let label = label.into();
        if !self.labels.contains(&label) {
            self.labels.push(label);
        }
    }

    /// Remove a label
    pub fn remove_label(&mut self, label: &str) {
        self.labels.retain(|l| l != label);
    }

    /// Get labels as comma-separated string
    pub fn labels_string(&self) -> String {
        self.labels.join(", ")
    }
}

/// Request to create a new todo
#[derive(Debug, Deserialize)]
pub struct CreateTodoRequest {
    pub task_name: String,
    pub task_desc: Option<String>,
    pub project: Option<String>,
    pub labels: Option<Vec<String>>,
}

/// Request to update a todo
#[derive(Debug, Deserialize)]
pub struct UpdateTodoRequest {
    pub task_name: Option<String>,
    pub task_desc: Option<String>,
    pub project: Option<String>,
    pub labels: Option<Vec<String>>,
}

/// Filter options for todo queries
#[derive(Debug, Deserialize, Default)]
pub struct TodoFilter {
    pub status: Option<TodoStatus>,
    pub project: Option<String>,
    pub label: Option<String>,
    pub search: Option<String>,
}

/// Todo view model for UI
#[derive(Debug, Serialize)]
pub struct TodoView {
    #[serde(flatten)]
    pub todo: Todo,
    pub status_display: String,
    pub is_overdue: bool,
    pub days_since_created: i64,
}

impl From<Todo> for TodoView {
    fn from(todo: Todo) -> Self {
        let days_since_created = Utc::now().signed_duration_since(todo.created_at).num_days();

        Self {
            status_display: todo.status.display_name().to_string(),
            is_overdue: false, // Could be enhanced with due dates
            days_since_created,
            todo,
        }
    }
}

/// Project summary for sidebar
#[derive(Debug, Serialize)]
pub struct ProjectSummary {
    pub name: String,
    pub todo_count: i64,
    pub in_progress_count: i64,
    pub done_count: i64,
}

/// Label summary
#[derive(Debug, Serialize)]
pub struct LabelSummary {
    pub name: String,
    pub usage_count: i64,
}
