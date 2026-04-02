//! Database operations for stopwatches

use crate::db::Database;
use crate::models::stopwatch::{
    CreateStopwatchRequest, Stopwatch, StopwatchLap, StopwatchStatus, UpdateStopwatchRequest,
};
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

/// Helper to convert a database row to Stopwatch
fn row_to_stopwatch(row: &Row) -> Result<Stopwatch> {
    let status_str: String = row.get(3)?;
    let status = StopwatchStatus::try_from(status_str.as_str()).map_err(|_| {
        rusqlite::Error::InvalidColumnType(
            3,
            "StopwatchStatus".to_string(),
            rusqlite::types::Type::Text,
        )
    })?;

    Ok(Stopwatch {
        id: row.get(0)?,
        label: row.get(1)?,
        elapsed_seconds: row.get(2)?,
        status,
        started_at: parse_datetime_opt(row.get(4)?),
        created_at: parse_datetime(row.get(5)?),
        laps: None, // Loaded separately
    })
}

impl Database {
    /// Create a new stopwatch in the database
    pub fn create_stopwatch(&self, req: CreateStopwatchRequest) -> Result<Stopwatch> {
        let stopwatch = Stopwatch::new(req.label);

        self.conn().execute(
            "INSERT INTO stopwatches (id, label, elapsed_seconds, status, started_at, created_at)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            params![
                stopwatch.id,
                stopwatch.label,
                stopwatch.elapsed_seconds,
                stopwatch.status.as_str(),
                stopwatch.started_at.map(|t| t.to_rfc3339()),
                stopwatch.created_at.to_rfc3339()
            ],
        )?;

        Ok(stopwatch)
    }

    /// Get a stopwatch by ID
    pub fn get_stopwatch(&self, id: &str) -> Result<Option<Stopwatch>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, label, elapsed_seconds, status, started_at, created_at
             FROM stopwatches WHERE id = ?1",
        )?;

        let stopwatch = stmt.query_row([id], row_to_stopwatch);

        match stopwatch {
            Ok(mut sw) => {
                // Load laps
                sw.laps = Some(self.get_stopwatch_laps(id)?);
                Ok(Some(sw))
            }
            Err(rusqlite::Error::QueryReturnedNoRows) => Ok(None),
            Err(e) => Err(e),
        }
    }

    /// Get all stopwatches
    pub fn get_stopwatches(&self) -> Result<Vec<Stopwatch>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, label, elapsed_seconds, status, started_at, created_at
             FROM stopwatches ORDER BY created_at DESC",
        )?;

        let rows = stmt.query_map([], row_to_stopwatch)?;
        rows.collect()
    }

    /// Update stopwatch state
    pub fn update_stopwatch_state(
        &self,
        id: &str,
        status: StopwatchStatus,
        elapsed_seconds: i64,
        started_at: Option<chrono::DateTime<Utc>>,
    ) -> Result<bool> {
        let rows = self.conn().execute(
            "UPDATE stopwatches SET status = ?1, elapsed_seconds = ?2, started_at = ?3 WHERE id = ?4",
            params![
                status.as_str(),
                elapsed_seconds,
                started_at.map(|t| t.to_rfc3339()),
                id
            ],
        )?;

        Ok(rows > 0)
    }

    /// Update stopwatch label
    pub fn update_stopwatch(&self, id: &str, req: UpdateStopwatchRequest) -> Result<bool> {
        if let Some(label) = req.label {
            let rows = self.conn().execute(
                "UPDATE stopwatches SET label = ?1 WHERE id = ?2",
                params![label, id],
            )?;
            return Ok(rows > 0);
        }
        Ok(false)
    }

    /// Delete a stopwatch and its laps
    pub fn delete_stopwatch(&self, id: &str) -> Result<bool> {
        let mut conn = self.conn();
        let tx = conn.transaction()?;

        // Delete laps first (cascade should handle this, but be explicit)
        tx.execute("DELETE FROM stopwatch_laps WHERE stopwatch_id = ?1", [id])?;

        // Delete stopwatch
        let rows = tx.execute("DELETE FROM stopwatches WHERE id = ?1", [id])?;

        tx.commit()?;

        Ok(rows > 0)
    }

    /// Record a lap
    pub fn record_lap(&self, lap: &StopwatchLap) -> Result<()> {
        self.conn().execute(
            "INSERT INTO stopwatch_laps (id, stopwatch_id, lap_number, lap_duration_ms, note, created_at)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            params![
                lap.id,
                lap.stopwatch_id,
                lap.lap_number,
                lap.lap_duration_ms,
                lap.note,
                lap.created_at.to_rfc3339()
            ],
        )?;

        Ok(())
    }

    /// Get laps for a stopwatch
    pub fn get_stopwatch_laps(&self, stopwatch_id: &str) -> Result<Vec<StopwatchLap>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, stopwatch_id, lap_number, lap_duration_ms, note, created_at
             FROM stopwatch_laps WHERE stopwatch_id = ?1 ORDER BY lap_number ASC",
        )?;

        let rows = stmt.query_map([stopwatch_id], |row| {
            Ok(StopwatchLap {
                id: row.get(0)?,
                stopwatch_id: row.get(1)?,
                lap_number: row.get(2)?,
                lap_duration_ms: row.get(3)?,
                note: row.get(4)?,
                created_at: parse_datetime(row.get(5)?),
            })
        })?;

        rows.collect()
    }

    /// Delete a specific lap
    pub fn delete_lap(&self, lap_id: &str) -> Result<bool> {
        let rows = self
            .conn()
            .execute("DELETE FROM stopwatch_laps WHERE id = ?1", [lap_id])?;

        Ok(rows > 0)
    }

    /// Get running stopwatches (for background service)
    pub fn get_running_stopwatches(&self) -> Result<Vec<Stopwatch>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, label, elapsed_seconds, status, started_at, created_at
             FROM stopwatches WHERE status = 'RUNNING'",
        )?;

        let rows = stmt.query_map([], row_to_stopwatch)?;
        rows.collect()
    }
}
