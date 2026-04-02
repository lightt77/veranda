// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod config;
mod db;

use db::Database;
use std::sync::Arc;

/// Application state shared across commands
pub struct AppState {
    db: Arc<Database>,
}

// Learn more about Tauri commands at https://tauri.app/v1/guides/features/command
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

/// Health check command to verify database is working
#[tauri::command]
fn health_check(state: tauri::State<AppState>) -> Result<String, String> {
    let conn = state.db.conn();
    match conn.query_row("SELECT COUNT(*) FROM _migrations", [], |row| {
        let count: i64 = row.get(0)?;
        Ok(count)
    }) {
        Ok(count) => Ok(format!("Database healthy. {} migrations applied.", count)),
        Err(e) => Err(format!("Database error: {}", e)),
    }
}

fn main() {
    // Initialize database
    let database = match db::init_db() {
        Ok(db) => {
            println!("Database initialized successfully");
            Arc::new(db)
        }
        Err(e) => {
            eprintln!("Failed to initialize database: {}", e);
            std::process::exit(1);
        }
    };

    let app_state = AppState { db: database };

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(app_state)
        .invoke_handler(tauri::generate_handler![greet, health_check])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
