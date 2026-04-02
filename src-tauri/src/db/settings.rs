//! Database operations for settings

use crate::db::Database;
use chrono::Utc;
use rusqlite::{params, Result, Row};
use serde::{Deserialize, Serialize};

/// Setting key constants
pub const SETTING_NOTIFICATION_SOUND_ENABLED: &str = "notification_sound_enabled";
pub const SETTING_NOTIFICATION_SOUND_NAME: &str = "notification_sound_name";
pub const SETTING_BACKGROUND_OPACITY: &str = "background_opacity";
pub const SETTING_THEME: &str = "theme";

/// Setting structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Setting {
    pub key: String,
    pub value: String,
    pub updated_at: chrono::DateTime<Utc>,
}

/// Helper to parse required datetime from string
fn parse_datetime(s: String) -> chrono::DateTime<Utc> {
    chrono::DateTime::parse_from_rfc3339(&s)
        .map(|dt| dt.with_timezone(&Utc))
        .unwrap_or_else(|_| Utc::now())
}

fn row_to_setting(row: &Row) -> Result<Setting> {
    Ok(Setting {
        key: row.get(0)?,
        value: row.get(1)?,
        updated_at: parse_datetime(row.get(2)?),
    })
}

impl Database {
    /// Get a setting value by key
    pub fn get_setting(&self, key: &str) -> Result<Option<String>> {
        let value: Option<String> = self
            .conn()
            .query_row("SELECT value FROM settings WHERE key = ?1", [key], |row| {
                row.get(0)
            })
            .ok();
        Ok(value)
    }

    /// Get a setting with default value
    pub fn get_setting_or_default(&self, key: &str, default: &str) -> String {
        self.get_setting(key)
            .ok()
            .flatten()
            .unwrap_or_else(|| default.to_string())
    }

    /// Set a setting value
    pub fn set_setting(&self, key: &str, value: &str) -> Result<()> {
        self.conn().execute(
            "INSERT INTO settings (key, value, updated_at) VALUES (?1, ?2, ?3)
             ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at",
            params![key, value, Utc::now().to_rfc3339()],
        )?;
        Ok(())
    }

    /// Get all settings
    pub fn get_all_settings(&self) -> Result<Vec<Setting>> {
        let conn = self.conn();
        let mut stmt = conn.prepare("SELECT key, value, updated_at FROM settings ORDER BY key")?;

        let rows = stmt.query_map([], row_to_setting)?;
        rows.collect()
    }

    /// Delete a setting
    pub fn delete_setting(&self, key: &str) -> Result<bool> {
        let rows = self
            .conn()
            .execute("DELETE FROM settings WHERE key = ?1", [key])?;
        Ok(rows > 0)
    }

    /// Check if notification sound is enabled
    pub fn is_notification_sound_enabled(&self) -> bool {
        self.get_setting_or_default(SETTING_NOTIFICATION_SOUND_ENABLED, "true") == "true"
    }

    /// Get notification sound name
    pub fn get_notification_sound_name(&self) -> String {
        self.get_setting_or_default(SETTING_NOTIFICATION_SOUND_NAME, "default")
    }

    /// Get background opacity (0.0 - 1.0)
    pub fn get_background_opacity(&self) -> f64 {
        self.get_setting_or_default(SETTING_BACKGROUND_OPACITY, "0.3")
            .parse()
            .unwrap_or(0.3)
    }

    /// Get theme
    pub fn get_theme(&self) -> String {
        self.get_setting_or_default(SETTING_THEME, "dark")
    }
}
