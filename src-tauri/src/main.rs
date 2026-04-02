// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod background;
mod config;
mod db;
mod models;
mod services;

use background::BackgroundService;
use db::Database;
use models::*;
use services::{ImageService, SoundService};
use std::sync::Arc;
use tauri_plugin_cli::CliExt;


/// Application state shared across commands
pub struct AppState {
    db: Arc<Database>,
    background: Arc<BackgroundService>,
}

// ==================== Timer Commands ====================

#[tauri::command]
async fn create_timer(
    req: timer::CreateTimerRequest,
    state: tauri::State<'_, AppState>,
) -> Result<timer::Timer, String> {
    state
        .db
        .create_timer(req)
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_timers(
    status: Option<String>,
    state: tauri::State<'_, AppState>,
) -> Result<Vec<timer::Timer>, String> {
    let status_filter = status.and_then(|s| timer::TimerStatus::try_from(s.as_str()).ok());
    state
        .db
        .get_timers(status_filter)
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_timer(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<Option<timer::Timer>, String> {
    state.db.get_timer(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn update_timer(
    id: String,
    req: timer::UpdateTimerRequest,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state.db.update_timer(&id, req).map_err(|e| e.to_string())
}

#[tauri::command]
async fn delete_timer(id: String, state: tauri::State<'_, AppState>) -> Result<bool, String> {
    state.db.delete_timer(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn start_timer(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    // Get timer and update state
    if let Some(timer) = state.db.get_timer(&id).map_err(|e| e.to_string())? {
        let now = chrono::Utc::now();
        state
            .db
            .update_timer_state(&id, timer::TimerStatus::Running, timer.remaining_seconds, Some(now))
            .map_err(|e| e.to_string())?;
        
        // Notify background service
        state.background.start_timer(&id).await.map_err(|e| e.to_string())?;
        Ok(true)
    } else {
        Ok(false)
    }
}

#[tauri::command]
async fn pause_timer(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    // Get current state and update
    if let Some(timer) = state.db.get_timer(&id).map_err(|e| e.to_string())? {
        state
            .db
            .update_timer_state(&id, timer::TimerStatus::Paused, timer.remaining_seconds, None)
            .map_err(|e| e.to_string())?;
        
        // Notify background service
        state.background.pause_timer(&id).await;
        Ok(true)
    } else {
        Ok(false)
    }
}

#[tauri::command]
async fn reset_timer(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    if let Some(timer) = state.db.get_timer(&id).map_err(|e| e.to_string())? {
        state
            .db
            .update_timer_state(&id, timer::TimerStatus::Created, timer.duration_seconds, None)
            .map_err(|e| e.to_string())?;
        
        state.background.pause_timer(&id).await;
        Ok(true)
    } else {
        Ok(false)
    }
}

// ==================== Stopwatch Commands ====================

#[tauri::command]
async fn create_stopwatch(
    req: stopwatch::CreateStopwatchRequest,
    state: tauri::State<'_, AppState>,
) -> Result<stopwatch::Stopwatch, String> {
    state
        .db
        .create_stopwatch(req)
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_stopwatches(
    state: tauri::State<'_, AppState>,
) -> Result<Vec<stopwatch::Stopwatch>, String> {
    state.db.get_stopwatches().map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_stopwatch(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<Option<stopwatch::Stopwatch>, String> {
    state.db.get_stopwatch(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn update_stopwatch(
    id: String,
    req: stopwatch::UpdateStopwatchRequest,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state
        .db
        .update_stopwatch(&id, req)
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn delete_stopwatch(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state.db.delete_stopwatch(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn start_stopwatch(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    if let Some(stopwatch) = state.db.get_stopwatch(&id).map_err(|e| e.to_string())? {
        let now = chrono::Utc::now();
        state
            .db
            .update_stopwatch_state(&id, stopwatch::StopwatchStatus::Running, stopwatch.elapsed_seconds, Some(now))
            .map_err(|e| e.to_string())?;
        
        state.background.start_stopwatch(&id).await.map_err(|e| e.to_string())?;
        Ok(true)
    } else {
        Ok(false)
    }
}

#[tauri::command]
async fn pause_stopwatch(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    if let Some(stopwatch) = state.db.get_stopwatch(&id).map_err(|e| e.to_string())? {
        state
            .db
            .update_stopwatch_state(&id, stopwatch::StopwatchStatus::Paused, stopwatch.elapsed_seconds, None)
            .map_err(|e| e.to_string())?;
        
        state.background.pause_stopwatch(&id).await;
        Ok(true)
    } else {
        Ok(false)
    }
}

#[tauri::command]
async fn reset_stopwatch(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state
        .db
        .update_stopwatch_state(&id, stopwatch::StopwatchStatus::Created, 0, None)
        .map_err(|e| e.to_string())?;
    
    state.background.pause_stopwatch(&id).await;
    Ok(true)
}

#[tauri::command]
async fn record_lap(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<Option<stopwatch::StopwatchLap>, String> {
    if let Some(mut stopwatch) = state.db.get_stopwatch(&id).map_err(|e| e.to_string())? {
        let lap = stopwatch.lap();
        if let Some(ref lap) = lap {
            state.db.record_lap(lap).map_err(|e| e.to_string())?;
        }
        Ok(lap)
    } else {
        Ok(None)
    }
}

// ==================== Todo Commands ====================

#[tauri::command]
async fn create_todo(
    req: todo::CreateTodoRequest,
    state: tauri::State<'_, AppState>,
) -> Result<todo::Todo, String> {
    state.db.create_todo(req).map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_todos(
    filter: Option<todo::TodoFilter>,
    state: tauri::State<'_, AppState>,
) -> Result<Vec<todo::Todo>, String> {
    state.db.get_todos(filter).map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_todo(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<Option<todo::Todo>, String> {
    state.db.get_todo(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn update_todo(
    id: String,
    req: todo::UpdateTodoRequest,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state.db.update_todo(&id, req).map_err(|e| e.to_string())
}

#[tauri::command]
async fn delete_todo(id: String, state: tauri::State<'_, AppState>) -> Result<bool, String> {
    state.db.delete_todo(&id).map_err(|e| e.to_string())
}

#[tauri::command]
async fn move_todo(
    id: String,
    status: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    let status = todo::TodoStatus::try_from(status.as_str())
        .map_err(|e| e.to_string())?;
    state.db.update_todo_status(&id, status).map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_projects(state: tauri::State<'_, AppState>) -> Result<Vec<todo::ProjectSummary>, String> {
    state.db.get_projects().map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_labels(state: tauri::State<'_, AppState>) -> Result<Vec<todo::LabelSummary>, String> {
    state.db.get_labels().map_err(|e| e.to_string())
}

// ==================== Notification Commands ====================

#[tauri::command]
async fn show_notification(title: String, body: String) -> Result<(), String> {
    // Use macOS notification system
    #[cfg(target_os = "macos")]
    {
        use mac_notification_sys::{Notification, Sound};
        Notification::new()
            .title(&title)
            .message(&body)
            .sound(Sound::Default)
            .send()
            .map_err(|e| e.to_string())?;
    }
    
    Ok(())
}

#[tauri::command]
async fn play_sound(sound_type: Option<String>) -> Result<(), String> {
    let sound_service = SoundService::new().map_err(|e| e.to_string())?;
    
    match sound_type.as_deref() {
        Some("notification") => sound_service.play_notification(),
        _ => sound_service.play_timer_complete(),
    }
    
    Ok(())
}

// ==================== Image/Background Commands ====================

#[tauri::command]
async fn download_image(
    url: String,
    state: tauri::State<'_, AppState>,
) -> Result<String, String> {
    let image_service = ImageService::new(state.db.clone());
    image_service
        .download_image(&url)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn get_images(
    only_active: bool,
    state: tauri::State<'_, AppState>,
) -> Result<Vec<db::images::Image>, String> {
    state.db.get_images(only_active).map_err(|e| e.to_string())
}

#[tauri::command]
fn set_image_active(
    id: String,
    active: bool,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    state.db.set_image_active(&id, active).map_err(|e| e.to_string())
}

#[tauri::command]
async fn delete_image(
    id: String,
    state: tauri::State<'_, AppState>,
) -> Result<bool, String> {
    let image_service = ImageService::new(state.db.clone());
    image_service
        .delete_image(&id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn get_random_background(state: tauri::State<'_, AppState>) -> Result<Option<String>, String> {
    let image_service = ImageService::new(state.db.clone());
    image_service
        .get_random_background()
        .map_err(|e| e.to_string())
}

// ==================== Health & Settings Commands ====================

#[tauri::command]
fn health_check(state: tauri::State<'_, AppState>) -> Result<String, String> {
    let conn = state.db.conn();
    match conn.query_row("SELECT COUNT(*) FROM _migrations", [], |row| {
        let count: i64 = row.get(0)?;
        Ok(count)
    }) {
        Ok(count) => Ok(format!("Database healthy. {} migrations applied.", count)),
        Err(e) => Err(format!("Database error: {}", e)),
    }
}

#[tauri::command]
async fn get_setting(
    key: String,
    state: tauri::State<'_, AppState>,
) -> Result<Option<String>, String> {
    state.db.get_setting(&key).map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_setting(
    key: String,
    value: String,
    state: tauri::State<'_, AppState>,
) -> Result<(), String> {
    state.db.set_setting(&key, &value).map_err(|e| e.to_string())
}

// ==================== CLI Handlers ====================

fn handle_cli_matches(
    matches: &tauri_plugin_cli::Matches,
    db: &Arc<Database>,
    background: &Arc<BackgroundService>,
) -> Result<(), String> {
    if let Some(command) = &matches.subcommand {
        match command.name.as_str() {
            "timer" => handle_timer_command(&command.matches, db, background),
            "stopwatch" => handle_stopwatch_command(&command.matches, db, background),
            "todo" => handle_todo_command(&command.matches, db),
            "daemon" => {
                println!("Background daemon is running...");
                println!("Press Ctrl+C to exit.");
                loop {
                    std::thread::sleep(std::time::Duration::from_secs(1));
                }
            }
            _ => Err("Unknown command".to_string()),
        }
    } else {
        Err("No command specified. Use --help for usage information.".to_string())
    }
}

fn handle_timer_command(
    matches: &tauri_plugin_cli::Matches,
    db: &Arc<Database>,
    background: &Arc<BackgroundService>,
) -> Result<(), String> {
    let action = matches
        .args
        .get("action")
        .and_then(|v| v.value.as_str())
        .unwrap_or("list");

    match action {
        "list" => {
            let timers = db.get_timers(None).map_err(|e| e.to_string())?;
            if timers.is_empty() {
                println!("No timers found.");
            } else {
                println!("Timers:");
                for timer in timers {
                    let status_icon = match timer.status {
                        crate::models::timer::TimerStatus::Running => "▶",
                        crate::models::timer::TimerStatus::Paused => "⏸",
                        crate::models::timer::TimerStatus::Done => "✓",
                        _ => "○",
                    };
                    println!(
                        "  {} {} - {}s remaining ({:?})",
                        status_icon, timer.label, timer.remaining_seconds, timer.status
                    );
                }
            }
            Ok(())
        }
        "create" => {
            let label = matches
                .args
                .get("label")
                .and_then(|v| v.value.as_str())
                .unwrap_or("Untitled Timer")
                .to_string();
            let duration = matches
                .args
                .get("duration")
                .and_then(|v| v.value.as_str())
                .unwrap_or("300")
                .parse::<i64>()
                .map_err(|_| "Invalid duration")?;

            let req = timer::CreateTimerRequest {
                label,
                duration_seconds: duration,
                note: None,
            };
            let timer = db.create_timer(req).map_err(|e| e.to_string())?;
            println!("Created timer: {} (ID: {})", timer.label, timer.id);
            Ok(())
        }
        "start" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Timer ID required")?;
            if let Some(timer) = db.get_timer(id).map_err(|e| e.to_string())? {
                let now = chrono::Utc::now();
                db.update_timer_state(id, timer::TimerStatus::Running, timer.remaining_seconds, Some(now))
                    .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async {
                    background.start_timer(id).await.map_err(|e| e.to_string())
                })?;
                println!("Started timer: {}", timer.label);
                Ok(())
            } else {
                Err("Timer not found".to_string())
            }
        }
        "pause" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Timer ID required")?;
            if let Some(timer) = db.get_timer(id).map_err(|e| e.to_string())? {
                db.update_timer_state(id, timer::TimerStatus::Paused, timer.remaining_seconds, None)
                    .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async { background.pause_timer(id).await });
                println!("Paused timer: {}", timer.label);
                Ok(())
            } else {
                Err("Timer not found".to_string())
            }
        }
        "reset" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Timer ID required")?;
            if let Some(timer) = db.get_timer(id).map_err(|e| e.to_string())? {
                db.update_timer_state(
                    id,
                    timer::TimerStatus::Created,
                    timer.duration_seconds,
                    None,
                )
                .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async { background.pause_timer(id).await });
                println!("Reset timer: {}", timer.label);
                Ok(())
            } else {
                Err("Timer not found".to_string())
            }
        }
        _ => Err("Unknown timer action".to_string()),
    }
}

fn handle_stopwatch_command(
    matches: &tauri_plugin_cli::Matches,
    db: &Arc<Database>,
    background: &Arc<BackgroundService>,
) -> Result<(), String> {
    let action = matches
        .args
        .get("action")
        .and_then(|v| v.value.as_str())
        .unwrap_or("list");

    match action {
        "list" => {
            let stopwatches = db.get_stopwatches().map_err(|e| e.to_string())?;
            if stopwatches.is_empty() {
                println!("No stopwatches found.");
            } else {
                println!("Stopwatches:");
                for sw in stopwatches {
                    let status_icon = match sw.status {
                        crate::models::stopwatch::StopwatchStatus::Running => "▶",
                        crate::models::stopwatch::StopwatchStatus::Paused => "⏸",
                        _ => "○",
                    };
                    println!(
                        "  {} {} - {}s elapsed ({:?})",
                        status_icon, sw.label, sw.elapsed_seconds, sw.status
                    );
                }
            }
            Ok(())
        }
        "create" => {
            let label = matches
                .args
                .get("label")
                .and_then(|v| v.value.as_str())
                .unwrap_or("Untitled Stopwatch")
                .to_string();

            let req = stopwatch::CreateStopwatchRequest {
                label,
            };
            let sw = db.create_stopwatch(req).map_err(|e| e.to_string())?;
            println!("Created stopwatch: {} (ID: {})", sw.label, sw.id);
            Ok(())
        }
        "start" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Stopwatch ID required")?;
            if let Some(sw) = db.get_stopwatch(id).map_err(|e| e.to_string())? {
                let now = chrono::Utc::now();
                db.update_stopwatch_state(id, stopwatch::StopwatchStatus::Running, sw.elapsed_seconds, Some(now))
                    .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async {
                    background.start_stopwatch(id).await.map_err(|e| e.to_string())
                })?;
                println!("Started stopwatch: {}", sw.label);
                Ok(())
            } else {
                Err("Stopwatch not found".to_string())
            }
        }
        "pause" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Stopwatch ID required")?;
            if let Some(sw) = db.get_stopwatch(id).map_err(|e| e.to_string())? {
                db.update_stopwatch_state(id, stopwatch::StopwatchStatus::Paused, sw.elapsed_seconds, None)
                    .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async { background.pause_stopwatch(id).await });
                println!("Paused stopwatch: {}", sw.label);
                Ok(())
            } else {
                Err("Stopwatch not found".to_string())
            }
        }
        "reset" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Stopwatch ID required")?;
            if let Some(sw) = db.get_stopwatch(id).map_err(|e| e.to_string())? {
                db.update_stopwatch_state(id, stopwatch::StopwatchStatus::Created, 0, None)
                    .map_err(|e| e.to_string())?;
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async { background.pause_stopwatch(id).await });
                println!("Reset stopwatch: {}", sw.label);
                Ok(())
            } else {
                Err("Stopwatch not found".to_string())
            }
        }
        _ => Err("Unknown stopwatch action".to_string()),
    }
}

fn handle_todo_command(
    matches: &tauri_plugin_cli::Matches,
    db: &Arc<Database>,
) -> Result<(), String> {
    let action = matches
        .args
        .get("action")
        .and_then(|v| v.value.as_str())
        .unwrap_or("list");

    match action {
        "list" => {
            let todos = db
                .get_todos(None)
                .map_err(|e| e.to_string())?;
            if todos.is_empty() {
                println!("No todos found.");
            } else {
                println!("Todos:");
                for todo in todos {
                    let status_icon = match todo.status {
                        crate::models::todo::TodoStatus::Done => "✓",
                        crate::models::todo::TodoStatus::Archived => "✗",
                        _ => "○",
                    };
                    println!(
                        "  {} {} - ({:?})",
                        status_icon, todo.task_name, todo.status
                    );
                }
            }
            Ok(())
        }
        "add" => {
            let task = matches
                .args
                .get("task")
                .and_then(|v| v.value.as_str())
                .ok_or("Task name required")?
                .to_string();

            let req = todo::CreateTodoRequest {
                task_name: task,
                task_desc: None,
                project: None,
                labels: None,
            };
            let todo = db.create_todo(req).map_err(|e| e.to_string())?;
            println!("Added todo: {} (ID: {})", todo.task_name, todo.id);
            Ok(())
        }
        "done" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Todo ID required")?;
            db.update_todo_status(id, todo::TodoStatus::Done)
                .map_err(|e| e.to_string())?;
            println!("Marked todo as done.");
            Ok(())
        }
        "archive" => {
            let id = matches
                .args
                .get("id")
                .and_then(|v| v.value.as_str())
                .ok_or("Todo ID required")?;
            db.update_todo_status(id, todo::TodoStatus::Archived)
                .map_err(|e| e.to_string())?;
            println!("Archived todo.");
            Ok(())
        }
        _ => Err("Unknown todo action".to_string()),
    }
}

// ==================== App Lifecycle ====================

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

    // Spawn background service
    let background = background::spawn_background_service(database.clone());

    let app_state = AppState {
        db: database.clone(),
        background: background.clone(),
    };

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_cli::init())
        .manage(app_state)
        .invoke_handler(tauri::generate_handler![
            // Health
            health_check,
            // Notifications
            show_notification,
            play_sound,
            // Timers
            create_timer,
            get_timers,
            get_timer,
            update_timer,
            delete_timer,
            start_timer,
            pause_timer,
            reset_timer,
            // Stopwatches
            create_stopwatch,
            get_stopwatches,
            get_stopwatch,
            update_stopwatch,
            delete_stopwatch,
            start_stopwatch,
            pause_stopwatch,
            reset_stopwatch,
            record_lap,
            // Todos
            create_todo,
            get_todos,
            get_todo,
            update_todo,
            delete_todo,
            move_todo,
            get_projects,
            get_labels,
            // Settings
            get_setting,
            set_setting,
            // Images/Backgrounds
            download_image,
            get_images,
            set_image_active,
            delete_image,
            get_random_background,
        ])
        .setup(move |app| {
            // Check for CLI matches
            let cli = app.cli();
            match cli.matches() {
                Ok(matches) => {
                    if matches.subcommand.is_some() {
                        // CLI mode - handle command and exit
                        if let Err(e) = handle_cli_matches(&matches, &database, &background) {
                            eprintln!("Error: {}", e);
                            std::process::exit(1);
                        }
                        std::process::exit(0);
                    }
                    // No CLI command - continue with GUI
                    Ok(())
                }
                Err(e) => {
                    eprintln!("CLI error: {}", e);
                    std::process::exit(1);
                }
            }
        })
        .on_window_event(|_window, event| {
            // Handle window close - background service continues running
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                println!("Window closing, but background service continues...");
                // Don't prevent close - let the window close
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
