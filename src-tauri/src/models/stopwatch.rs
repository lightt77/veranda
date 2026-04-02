use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Stopwatch status enumeration
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "UPPERCASE")]
pub enum StopwatchStatus {
    Created,
    Running,
    Paused,
    Stopped,
}

impl StopwatchStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            StopwatchStatus::Created => "CREATED",
            StopwatchStatus::Running => "RUNNING",
            StopwatchStatus::Paused => "PAUSED",
            StopwatchStatus::Stopped => "STOPPED",
        }
    }
}

impl TryFrom<&str> for StopwatchStatus {
    type Error = String;

    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value {
            "CREATED" => Ok(StopwatchStatus::Created),
            "RUNNING" => Ok(StopwatchStatus::Running),
            "PAUSED" => Ok(StopwatchStatus::Paused),
            "STOPPED" => Ok(StopwatchStatus::Stopped),
            _ => Err(format!("Invalid stopwatch status: {}", value)),
        }
    }
}

/// Stopwatch lap data
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StopwatchLap {
    pub id: String,
    pub stopwatch_id: String,
    pub lap_number: i32,
    pub lap_duration_ms: i64,
    pub note: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl StopwatchLap {
    pub fn new(stopwatch_id: impl Into<String>, lap_number: i32, lap_duration_ms: i64) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            stopwatch_id: stopwatch_id.into(),
            lap_number,
            lap_duration_ms,
            note: None,
            created_at: Utc::now(),
        }
    }

    /// Format lap duration as MM:SS.mmm
    pub fn formatted_duration(&self) -> String {
        format_milliseconds(self.lap_duration_ms)
    }
}

/// Stopwatch data structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Stopwatch {
    pub id: String,
    pub label: String,
    pub elapsed_seconds: i64,
    pub status: StopwatchStatus,
    pub started_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub laps: Option<Vec<StopwatchLap>>,
}

impl Stopwatch {
    /// Create a new stopwatch
    pub fn new(label: impl Into<String>) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            label: label.into(),
            elapsed_seconds: 0,
            status: StopwatchStatus::Created,
            started_at: None,
            created_at: now,
            laps: Some(vec![]),
        }
    }

    /// Start the stopwatch
    pub fn start(&mut self) {
        if self.status == StopwatchStatus::Created || self.status == StopwatchStatus::Paused {
            self.status = StopwatchStatus::Running;
            self.started_at = Some(Utc::now());
        }
    }

    /// Pause the stopwatch
    pub fn pause(&mut self) {
        if self.status == StopwatchStatus::Running {
            // Calculate elapsed time since start
            if let Some(started) = self.started_at {
                let elapsed = Utc::now().signed_duration_since(started).num_seconds();
                self.elapsed_seconds += elapsed;
            }
            self.status = StopwatchStatus::Paused;
            self.started_at = None;
        }
    }

    /// Stop the stopwatch
    pub fn stop(&mut self) {
        if self.status == StopwatchStatus::Running {
            if let Some(started) = self.started_at {
                let elapsed = Utc::now().signed_duration_since(started).num_seconds();
                self.elapsed_seconds += elapsed;
            }
        }
        self.status = StopwatchStatus::Stopped;
        self.started_at = None;
    }

    /// Reset the stopwatch
    pub fn reset(&mut self) {
        self.elapsed_seconds = 0;
        self.status = StopwatchStatus::Created;
        self.started_at = None;
        if let Some(ref mut laps) = self.laps {
            laps.clear();
        }
    }

    /// Record a lap
    pub fn lap(&mut self) -> Option<StopwatchLap> {
        if self.status != StopwatchStatus::Running {
            return None;
        }

        let current_elapsed = self.current_elapsed_ms();
        let lap_number = self.laps.as_ref().map(|l| l.len() as i32 + 1).unwrap_or(1);

        let lap = StopwatchLap::new(&self.id, lap_number, current_elapsed);

        if let Some(ref mut laps) = self.laps {
            laps.push(lap.clone());
        }

        Some(lap)
    }

    /// Get current elapsed time in milliseconds
    pub fn current_elapsed_ms(&self) -> i64 {
        let base_ms = self.elapsed_seconds * 1000;

        if self.status == StopwatchStatus::Running {
            if let Some(started) = self.started_at {
                let additional = Utc::now().signed_duration_since(started).num_milliseconds();
                return base_ms + additional;
            }
        }

        base_ms
    }

    /// Get current elapsed time in seconds
    pub fn current_elapsed_seconds(&self) -> i64 {
        self.current_elapsed_ms() / 1000
    }

    /// Format elapsed time as HH:MM:SS.mmm
    pub fn formatted_elapsed(&self) -> String {
        format_milliseconds(self.current_elapsed_ms())
    }

    /// Check if stopwatch is running
    pub fn is_running(&self) -> bool {
        self.status == StopwatchStatus::Running
    }
}

/// Request to create a new stopwatch
#[derive(Debug, Deserialize)]
pub struct CreateStopwatchRequest {
    pub label: String,
}

/// Request to update a stopwatch
#[derive(Debug, Deserialize)]
pub struct UpdateStopwatchRequest {
    pub label: Option<String>,
}

/// Stopwatch with UI state for frontend
#[derive(Debug, Serialize)]
pub struct StopwatchView {
    #[serde(flatten)]
    pub stopwatch: Stopwatch,
    pub formatted_elapsed: String,
    pub lap_count: usize,
}

impl From<Stopwatch> for StopwatchView {
    fn from(stopwatch: Stopwatch) -> Self {
        let lap_count = stopwatch.laps.as_ref().map(|l| l.len()).unwrap_or(0);
        Self {
            formatted_elapsed: stopwatch.formatted_elapsed(),
            lap_count,
            stopwatch,
        }
    }
}

/// Format milliseconds as HH:MM:SS.mmm or MM:SS.mmm
pub fn format_milliseconds(ms: i64) -> String {
    let total_seconds = ms / 1000;
    let hours = total_seconds / 3600;
    let minutes = (total_seconds % 3600) / 60;
    let seconds = total_seconds % 60;
    let millis = ms % 1000;

    if hours > 0 {
        format!("{:02}:{:02}:{:02}.{:03}", hours, minutes, seconds, millis)
    } else {
        format!("{:02}:{:02}.{:03}", minutes, seconds, millis)
    }
}
