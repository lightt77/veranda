use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Timer status enumeration
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "UPPERCASE")]
pub enum TimerStatus {
    Created,
    Running,
    Paused,
    Done,
}

impl TimerStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            TimerStatus::Created => "CREATED",
            TimerStatus::Running => "RUNNING",
            TimerStatus::Paused => "PAUSED",
            TimerStatus::Done => "DONE",
        }
    }
}

impl TryFrom<&str> for TimerStatus {
    type Error = String;

    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value {
            "CREATED" => Ok(TimerStatus::Created),
            "RUNNING" => Ok(TimerStatus::Running),
            "PAUSED" => Ok(TimerStatus::Paused),
            "DONE" => Ok(TimerStatus::Done),
            _ => Err(format!("Invalid timer status: {}", value)),
        }
    }
}

/// Timer data structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Timer {
    pub id: String,
    pub label: String,
    pub duration_seconds: i64,
    pub remaining_seconds: i64,
    pub status: TimerStatus,
    pub note: Option<String>,
    pub started_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

impl Timer {
    /// Create a new timer
    pub fn new(label: impl Into<String>, duration_seconds: i64) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            label: label.into(),
            duration_seconds,
            remaining_seconds: duration_seconds,
            status: TimerStatus::Created,
            note: None,
            started_at: None,
            created_at: now,
        }
    }

    /// Start the timer
    pub fn start(&mut self) {
        if self.status == TimerStatus::Created || self.status == TimerStatus::Paused {
            self.status = TimerStatus::Running;
            self.started_at = Some(Utc::now());
        }
    }

    /// Pause the timer
    pub fn pause(&mut self) {
        if self.status == TimerStatus::Running {
            self.status = TimerStatus::Paused;
            self.started_at = None;
        }
    }

    /// Mark timer as done
    pub fn complete(&mut self) {
        self.status = TimerStatus::Done;
        self.remaining_seconds = 0;
        self.started_at = None;
    }

    /// Reset timer to initial state
    pub fn reset(&mut self) {
        self.status = TimerStatus::Created;
        self.remaining_seconds = self.duration_seconds;
        self.started_at = None;
    }

    /// Check if timer is completed
    pub fn is_done(&self) -> bool {
        self.status == TimerStatus::Done || self.remaining_seconds <= 0
    }

    /// Get elapsed time in seconds
    pub fn elapsed_seconds(&self) -> i64 {
        self.duration_seconds - self.remaining_seconds
    }

    /// Get progress percentage (0.0 to 1.0)
    pub fn progress(&self) -> f64 {
        if self.duration_seconds == 0 {
            return 1.0;
        }
        let elapsed = self.elapsed_seconds() as f64;
        let total = self.duration_seconds as f64;
        (elapsed / total).clamp(0.0, 1.0)
    }
}

/// Request to create a new timer
#[derive(Debug, Deserialize)]
pub struct CreateTimerRequest {
    pub label: String,
    pub duration_seconds: i64,
    pub note: Option<String>,
}

/// Request to update a timer
#[derive(Debug, Deserialize)]
pub struct UpdateTimerRequest {
    pub label: Option<String>,
    pub note: Option<String>,
}

/// Timer with UI state for frontend
#[derive(Debug, Serialize)]
pub struct TimerView {
    #[serde(flatten)]
    pub timer: Timer,
    pub formatted_duration: String,
    pub formatted_remaining: String,
    pub progress_percent: i32,
}

impl From<Timer> for TimerView {
    fn from(timer: Timer) -> Self {
        let progress_percent = (timer.progress() * 100.0) as i32;
        Self {
            formatted_duration: format_duration(timer.duration_seconds),
            formatted_remaining: format_duration(timer.remaining_seconds),
            progress_percent,
            timer,
        }
    }
}

/// Format seconds as HH:MM:SS
pub fn format_duration(seconds: i64) -> String {
    let hours = seconds / 3600;
    let minutes = (seconds % 3600) / 60;
    let secs = seconds % 60;

    if hours > 0 {
        format!("{:02}:{:02}:{:02}", hours, minutes, secs)
    } else {
        format!("{:02}:{:02}", minutes, secs)
    }
}
