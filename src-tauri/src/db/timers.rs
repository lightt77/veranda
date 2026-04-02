//! Database operations for timers

use crate::db::Database;
use crate::models::timer::{CreateTimerRequest, Timer, TimerStatus, UpdateTimerRequest};
use chrono::Utc;
use rusqlite::{params, Result, Row};

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

/// Helper to convert a database row to Timer
fn row_to_timer(row: &Row) -> Result<Timer> {
    let status_str: String = row.get(4)?;
    let status = TimerStatus::try_from(status_str.as_str()).map_err(|_| {
        rusqlite::Error::InvalidColumnType(
            4,
            "TimerStatus".to_string(),
            rusqlite::types::Type::Text,
        )
    })?;

    Ok(Timer {
        id: row.get(0)?,
        label: row.get(1)?,
        duration_seconds: row.get(2)?,
        remaining_seconds: row.get(3)?,
        status,
        note: row.get(5)?,
        started_at: parse_datetime_opt(row.get(6)?),
        created_at: parse_datetime(row.get(7)?),
    })
}

impl Database {
    /// Create a new timer in the database
    pub fn create_timer(&self, req: CreateTimerRequest) -> Result<Timer> {
        let mut timer = Timer::new(req.label, req.duration_seconds);
        timer.note = req.note;

        self.conn().execute(
            "INSERT INTO timers (id, label, duration_seconds, remaining_seconds, status, note, started_at, created_at)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
            params![
                timer.id,
                timer.label,
                timer.duration_seconds,
                timer.remaining_seconds,
                timer.status.as_str(),
                timer.note,
                timer.started_at.map(|t| t.to_rfc3339()),
                timer.created_at.to_rfc3339()
            ],
        )?;

        Ok(timer)
    }

    /// Get a timer by ID
    pub fn get_timer(&self, id: &str) -> Result<Option<Timer>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, label, duration_seconds, remaining_seconds, status, note, started_at, created_at
             FROM timers WHERE id = ?1"
        )?;

        let timer = stmt.query_row([id], row_to_timer);

        match timer {
            Ok(t) => Ok(Some(t)),
            Err(rusqlite::Error::QueryReturnedNoRows) => Ok(None),
            Err(e) => Err(e),
        }
    }

    /// Get all timers, optionally filtered by status
    pub fn get_timers(&self, status_filter: Option<TimerStatus>) -> Result<Vec<Timer>> {
        let conn = self.conn();

        let sql = if status_filter.is_some() {
            "SELECT id, label, duration_seconds, remaining_seconds, status, note, started_at, created_at
             FROM timers WHERE status = ?1 ORDER BY created_at DESC"
        } else {
            "SELECT id, label, duration_seconds, remaining_seconds, status, note, started_at, created_at
             FROM timers ORDER BY created_at DESC"
        };

        let mut stmt = conn.prepare(sql)?;

        let rows = if let Some(status) = status_filter {
            stmt.query_map([status.as_str()], row_to_timer)?
        } else {
            stmt.query_map([], row_to_timer)?
        };

        rows.collect()
    }

    /// Update a timer
    pub fn update_timer(&self, id: &str, req: UpdateTimerRequest) -> Result<bool> {
        let mut updates = vec![];
        let mut params: Vec<Box<dyn rusqlite::ToSql>> = vec![];

        if let Some(label) = req.label {
            updates.push("label = ?");
            params.push(Box::new(label));
        }

        if let Some(note) = req.note {
            updates.push("note = ?");
            params.push(Box::new(note));
        }

        if updates.is_empty() {
            return Ok(false);
        }

        let sql = format!("UPDATE timers SET {} WHERE id = ?", updates.join(", "));
        params.push(Box::new(id.to_string()));

        let param_refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|p| p.as_ref()).collect();
        let rows = self.conn().execute(&sql, param_refs.as_slice())?;

        Ok(rows > 0)
    }

    /// Update timer status and remaining time
    pub fn update_timer_state(
        &self,
        id: &str,
        status: TimerStatus,
        remaining_seconds: i64,
        started_at: Option<chrono::DateTime<Utc>>,
    ) -> Result<bool> {
        let rows = self.conn().execute(
            "UPDATE timers SET status = ?1, remaining_seconds = ?2, started_at = ?3 WHERE id = ?4",
            params![
                status.as_str(),
                remaining_seconds,
                started_at.map(|t| t.to_rfc3339()),
                id
            ],
        )?;

        Ok(rows > 0)
    }

    /// Delete a timer
    pub fn delete_timer(&self, id: &str) -> Result<bool> {
        let rows = self
            .conn()
            .execute("DELETE FROM timers WHERE id = ?1", [id])?;

        Ok(rows > 0)
    }

    /// Get all running timers (for background service)
    pub fn get_running_timers(&self) -> Result<Vec<Timer>> {
        self.get_timers(Some(TimerStatus::Running))
    }

    /// Get active timers (CREATED, RUNNING, PAUSED)
    pub fn get_active_timers(&self) -> Result<Vec<Timer>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, label, duration_seconds, remaining_seconds, status, note, started_at, created_at
             FROM timers WHERE status IN ('CREATED', 'RUNNING', 'PAUSED') ORDER BY created_at DESC"
        )?;

        let rows = stmt.query_map([], row_to_timer)?;
        rows.collect()
    }
}
