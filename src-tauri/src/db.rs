use rusqlite::{Connection, Result};
use std::fs;
use std::path::PathBuf;
use std::sync::Mutex;

// Include operation modules
pub mod images;
pub mod settings;
pub mod stopwatches;
pub mod timers;
pub mod todos;

/// Database connection wrapper with migration support
pub struct Database {
    conn: Mutex<Connection>,
}

impl Database {
    /// Initialize the database, creating it if it doesn't exist
    pub fn new() -> Result<Self> {
        let data_dir = crate::config::get_data_dir();

        // Ensure data directory exists
        fs::create_dir_all(&data_dir).map_err(|e| {
            eprintln!(
                "Failed to create data directory '{}': {}",
                data_dir.display(),
                e
            );
            rusqlite::Error::InvalidPath(data_dir.clone())
        })?;

        let db_path = data_dir.join("veranda.db");
        let conn = Connection::open(&db_path)?;

        let db = Database {
            conn: Mutex::new(conn),
        };

        // Run migrations
        db.run_migrations()?;

        Ok(db)
    }

    /// Get a reference to the connection
    pub fn conn(&self) -> std::sync::MutexGuard<'_, Connection> {
        self.conn
            .lock()
            .expect("Database connection mutex poisoned")
    }

    /// Run all pending migrations
    fn run_migrations(&self) -> Result<()> {
        let mut conn = self.conn();

        // Create migrations table to track applied migrations
        conn.execute(
            "CREATE TABLE IF NOT EXISTS _migrations (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL UNIQUE,
                applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )",
            [],
        )?;

        // Define all migrations
        let migrations = vec![
            (
                "001_create_timers",
                include_str!("migrations/001_create_timers.sql"),
            ),
            (
                "002_create_stopwatches",
                include_str!("migrations/002_create_stopwatches.sql"),
            ),
            (
                "003_create_todos",
                include_str!("migrations/003_create_todos.sql"),
            ),
            (
                "004_create_images",
                include_str!("migrations/004_create_images.sql"),
            ),
            (
                "005_create_settings",
                include_str!("migrations/005_create_settings.sql"),
            ),
        ];

        // Apply each migration if not already applied
        for (name, sql) in migrations {
            let already_applied: bool = conn
                .query_row("SELECT 1 FROM _migrations WHERE name = ?1", [name], |_| {
                    Ok(true)
                })
                .unwrap_or(false);

            if !already_applied {
                // Execute migration in a transaction
                let tx = conn.transaction()?;
                tx.execute_batch(sql)?;
                tx.execute("INSERT INTO _migrations (name) VALUES (?1)", [name])?;
                tx.commit()?;
            }
        }

        Ok(())
    }
}

/// Initialize the database and return a singleton instance
pub fn init_db() -> Result<Database> {
    Database::new()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_database_creation() {
        let db = Database::new();
        assert!(db.is_ok());
    }
}
