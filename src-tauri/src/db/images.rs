//! Database operations for background images

use crate::db::Database;
use chrono::Utc;
use rusqlite::{params, Result, Row};
use serde::Serialize;

/// Image metadata structure
#[derive(Debug, Clone, Serialize)]
pub struct Image {
    pub id: String,
    pub url: String,
    pub local_path: String,
    pub downloaded_at: chrono::DateTime<Utc>,
    pub is_active: bool,
    pub metadata: Option<String>, // JSON string
}

/// Helper to parse required datetime from string
fn parse_datetime(s: String) -> chrono::DateTime<Utc> {
    chrono::DateTime::parse_from_rfc3339(&s)
        .map(|dt| dt.with_timezone(&Utc))
        .unwrap_or_else(|_| Utc::now())
}

fn row_to_image(row: &Row) -> Result<Image> {
    Ok(Image {
        id: row.get(0)?,
        url: row.get(1)?,
        local_path: row.get(2)?,
        downloaded_at: parse_datetime(row.get(3)?),
        is_active: row.get(4)?,
        metadata: row.get(5)?,
    })
}

impl Database {
    /// Save image record to database
    pub fn save_image(
        &self,
        id: &str,
        url: &str,
        local_path: &str,
        metadata: Option<&str>,
    ) -> Result<()> {
        self.conn().execute(
            "INSERT OR REPLACE INTO images (id, url, local_path, downloaded_at, is_active, metadata)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            params![
                id,
                url,
                local_path,
                Utc::now().to_rfc3339(),
                true,
                metadata
            ],
        )?;
        Ok(())
    }

    /// Get all images
    pub fn get_images(&self, only_active: bool) -> Result<Vec<Image>> {
        let sql = if only_active {
            "SELECT id, url, local_path, downloaded_at, is_active, metadata FROM images WHERE is_active = 1 ORDER BY downloaded_at DESC"
        } else {
            "SELECT id, url, local_path, downloaded_at, is_active, metadata FROM images ORDER BY downloaded_at DESC"
        };

        let conn = self.conn();
        let mut stmt = conn.prepare(sql)?;

        let rows = stmt.query_map([], row_to_image)?;
        rows.collect()
    }

    /// Get image by ID
    pub fn get_image(&self, id: &str) -> Result<Option<Image>> {
        let conn = self.conn();
        let mut stmt = conn.prepare(
            "SELECT id, url, local_path, downloaded_at, is_active, metadata FROM images WHERE id = ?1"
        )?;

        let image = stmt.query_row([id], row_to_image);

        match image {
            Ok(img) => Ok(Some(img)),
            Err(rusqlite::Error::QueryReturnedNoRows) => Ok(None),
            Err(e) => Err(e),
        }
    }

    /// Set image active status
    pub fn set_image_active(&self, id: &str, is_active: bool) -> Result<bool> {
        let rows = self.conn().execute(
            "UPDATE images SET is_active = ?1 WHERE id = ?2",
            params![is_active, id],
        )?;
        Ok(rows > 0)
    }

    /// Delete image record
    pub fn delete_image(&self, id: &str) -> Result<bool> {
        let rows = self
            .conn()
            .execute("DELETE FROM images WHERE id = ?1", [id])?;
        Ok(rows > 0)
    }

    /// Get active images count
    pub fn get_active_image_count(&self) -> Result<i64> {
        let count: i64 = self.conn().query_row(
            "SELECT COUNT(*) FROM images WHERE is_active = 1",
            [],
            |row| row.get(0),
        )?;
        Ok(count)
    }

    /// Check if URL already exists
    pub fn image_url_exists(&self, url: &str) -> Result<bool> {
        let exists: bool = self
            .conn()
            .query_row("SELECT 1 FROM images WHERE url = ?1", [url], |_| Ok(true))
            .unwrap_or(false);
        Ok(exists)
    }
}
