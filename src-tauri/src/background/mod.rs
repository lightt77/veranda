//! Background service for managing timers and stopwatches
//! 
//! This module provides a daemon that continues running timers and stopwatches
//! even when the GUI window is closed. It communicates with the GUI via IPC.
//! 
//! Features:
//! - Lazy spawning: Only starts when timers/stopwatches are active
//! - Self-termination: Stops after 10 seconds of inactivity when GUI is closed
//! - Daemon mode: Runs indefinitely when in daemon mode

use crate::db::Database;
use crate::models::timer::TimerStatus;
use crate::models::stopwatch::StopwatchStatus;
use crate::services::SoundService;
use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{RwLock, watch};
use tokio::time::interval;

/// Idle timeout before self-termination (seconds)
const IDLE_TIMEOUT_SECONDS: u64 = 10;
/// Tick interval for the main loop (milliseconds)
const TICK_INTERVAL_MS: u64 = 100;

/// Background service state
pub struct BackgroundService {
    db: Arc<Database>,
    sound_service: SoundService,
    /// Active timer calculations (timer_id -> last_update_time)
    timer_states: RwLock<HashMap<String, Instant>>,
    /// Active stopwatch calculations (stopwatch_id -> last_update_time)
    stopwatch_states: RwLock<HashMap<String, Instant>>,
    /// Service running flag
    running: RwLock<bool>,
    /// Whether GUI window is open
    gui_open: RwLock<bool>,
    /// Whether in daemon mode (run forever)
    daemon_mode: RwLock<bool>,
    /// Channel to signal shutdown
    shutdown_tx: watch::Sender<()>,
    shutdown_rx: RwLock<watch::Receiver<()>>,
}

impl BackgroundService {
    /// Create a new background service (doesn't start the async loop yet)
    pub fn new(db: Arc<Database>) -> Self {
        let sound_service = SoundService::new().expect("Failed to initialize sound service");
        let (shutdown_tx, shutdown_rx) = watch::channel(());
        Self {
            db,
            sound_service,
            timer_states: RwLock::new(HashMap::new()),
            stopwatch_states: RwLock::new(HashMap::new()),
            running: RwLock::new(false),
            gui_open: RwLock::new(false),
            daemon_mode: RwLock::new(false),
            shutdown_tx,
            shutdown_rx: RwLock::new(shutdown_rx),
        }
    }

    /// Check if the service is currently running
    pub async fn is_running(&self) -> bool {
        *self.running.read().await
    }

    /// Set whether the GUI window is open
    pub async fn set_gui_open(&self, open: bool) {
        *self.gui_open.write().await = open;
        println!("BackgroundService: GUI open = {}", open);
    }

    /// Set daemon mode (run forever without idle timeout)
    pub async fn set_daemon_mode(&self, daemon: bool) {
        *self.daemon_mode.write().await = daemon;
        println!("BackgroundService: Daemon mode = {}", daemon);
    }

    /// Ensure the background service is running
    /// This spawns the async task if not already running
    pub async fn ensure_running(&self) -> anyhow::Result<()> {
        if !*self.running.read().await {
            // Mark as running
            *self.running.write().await = true;
            
            // Spawn the async task
            let service_clone = Arc::new(self.clone_service());
            tokio::spawn(async move {
                if let Err(e) = service_clone.run_loop().await {
                    eprintln!("Background service error: {}", e);
                }
                // Mark as stopped when loop exits
                *service_clone.running.write().await = false;
                println!("Background service stopped");
            });
            
            println!("Background service started");
        }
        Ok(())
    }

    /// Clone the service state for spawning (needed for Arc)
    fn clone_service(&self) -> BackgroundService {
        let (shutdown_tx, shutdown_rx) = watch::channel(());
        BackgroundService {
            db: Arc::clone(&self.db),
            sound_service: SoundService::new().expect("Failed to initialize sound service"),
            timer_states: RwLock::new(HashMap::new()),
            stopwatch_states: RwLock::new(HashMap::new()),
            running: RwLock::new(true),
            gui_open: RwLock::new(*self.gui_open.try_read().map(|g| *g).unwrap_or(false)),
            daemon_mode: RwLock::new(*self.daemon_mode.try_read().map(|d| *d).unwrap_or(false)),
            shutdown_tx,
            shutdown_rx: RwLock::new(shutdown_rx),
        }
    }

    /// Main async loop - runs until shutdown or idle timeout
    async fn run_loop(&self) -> anyhow::Result<()> {
        // Load active timers and stopwatches from database
        self.load_active_timers().await?;
        self.load_active_stopwatches().await?;

        // Start the main loop
        let mut ticker = interval(Duration::from_millis(TICK_INTERVAL_MS));
        let mut idle_ticks: u64 = 0;
        let idle_timeout_ticks = (IDLE_TIMEOUT_SECONDS * 1000) / TICK_INTERVAL_MS;

        loop {
            ticker.tick().await;
            
            // Check for shutdown signal
            if self.shutdown_rx.write().await.has_changed().unwrap_or(false) {
                println!("Background service received shutdown signal");
                break;
            }

            // Process timers
            let timer_count = self.process_timers().await?;

            // Process stopwatches
            let stopwatch_count = self.process_stopwatches().await?;

            // Check if we did any work
            let had_work = timer_count > 0 || stopwatch_count > 0;
            let has_active_items = self.has_active_items().await;
            let gui_open = *self.gui_open.read().await;
            let daemon_mode = *self.daemon_mode.read().await;

            if had_work {
                idle_ticks = 0;
            } else if !has_active_items {
                // No active timers/stopwatches - start idle counter
                idle_ticks += 1;

                // Only terminate if GUI is closed AND not in daemon mode
                if !gui_open && !daemon_mode && idle_ticks >= idle_timeout_ticks {
                    println!("Background service idle for {} seconds, stopping...", IDLE_TIMEOUT_SECONDS);
                    break;
                }
            }
        }

        Ok(())
    }

    /// Stop the background service
    pub async fn stop(&self) {
        let _ = self.shutdown_tx.send(());
        *self.running.write().await = false;
    }

    /// Check if there are any active items being tracked
    async fn has_active_items(&self) -> bool {
        let timers = self.timer_states.read().await;
        let stopwatches = self.stopwatch_states.read().await;
        !timers.is_empty() || !stopwatches.is_empty()
    }

    /// Load active timers from database
    async fn load_active_timers(&self) -> anyhow::Result<()> {
        let timers = self.db.get_running_timers()?;
        let mut states = self.timer_states.write().await;
        
        for timer in timers {
            states.insert(timer.id.clone(), Instant::now());
            println!("Loaded active timer: {} ({}s remaining)", timer.label, timer.remaining_seconds);
        }

        Ok(())
    }

    /// Load active stopwatches from database
    async fn load_active_stopwatches(&self) -> anyhow::Result<()> {
        let stopwatches = self.db.get_running_stopwatches()?;
        let mut states = self.stopwatch_states.write().await;

        for stopwatch in stopwatches {
            states.insert(stopwatch.id.clone(), Instant::now());
            println!("Loaded active stopwatch: {} ({}s elapsed)", stopwatch.label, stopwatch.elapsed_seconds);
        }

        Ok(())
    }

    /// Process all active timers - returns count of timers processed
    async fn process_timers(&self) -> anyhow::Result<usize> {
        let timer_ids: Vec<String> = {
            let states = self.timer_states.read().await;
            states.keys().cloned().collect()
        };

        for timer_id in &timer_ids {
            self.update_timer(timer_id).await?;
        }

        Ok(timer_ids.len())
    }

    /// Process all active stopwatches - returns count of stopwatches processed
    async fn process_stopwatches(&self) -> anyhow::Result<usize> {
        let stopwatch_ids: Vec<String> = {
            let states = self.stopwatch_states.read().await;
            states.keys().cloned().collect()
        };

        for stopwatch_id in &stopwatch_ids {
            self.update_stopwatch(stopwatch_id).await?;
        }

        Ok(stopwatch_ids.len())
    }

    /// Update a specific timer's state
    async fn update_timer(&self, timer_id: &str) -> anyhow::Result<()> {
        let now = Instant::now();
        let mut states = self.timer_states.write().await;

        if let Some(last_update) = states.get(timer_id) {
            let elapsed = now.duration_since(*last_update).as_secs() as i64;

            if elapsed > 0 {
                // Get current timer state from DB
                if let Some(timer) = self.db.get_timer(timer_id)? {
                    if timer.status == TimerStatus::Running {
                        let new_remaining = (timer.remaining_seconds - elapsed).max(0);

                        if new_remaining == 0 {
                            // Timer completed!
                            self.db.update_timer_state(
                                timer_id,
                                TimerStatus::Done,
                                0,
                                None,
                            )?;
                            states.remove(timer_id);
                            
                            // Play notification sound and send notification
                            println!("Timer completed: {}", timer.label);
                            self.sound_service.play_timer_complete();
                        } else {
                            // Update remaining time
                            self.db.update_timer_state(
                                timer_id,
                                TimerStatus::Running,
                                new_remaining,
                                timer.started_at,
                            )?;
                            states.insert(timer_id.to_string(), now);
                        }
                    } else {
                        // Timer no longer running, remove from tracking
                        states.remove(timer_id);
                    }
                } else {
                    // Timer doesn't exist anymore
                    states.remove(timer_id);
                }
            }
        }

        Ok(())
    }

    /// Update a specific stopwatch's state
    async fn update_stopwatch(&self, stopwatch_id: &str) -> anyhow::Result<()> {
        let now = Instant::now();
        let mut states = self.stopwatch_states.write().await;

        if let Some(last_update) = states.get(stopwatch_id) {
            let elapsed = now.duration_since(*last_update).as_secs() as i64;

            if elapsed > 0 {
                // Get current stopwatch state from DB
                if let Some(stopwatch) = self.db.get_stopwatch(stopwatch_id)? {
                    if stopwatch.status == StopwatchStatus::Running {
                        let new_elapsed = stopwatch.elapsed_seconds + elapsed;

                        self.db.update_stopwatch_state(
                            stopwatch_id,
                            StopwatchStatus::Running,
                            new_elapsed,
                            stopwatch.started_at,
                        )?;
                        states.insert(stopwatch_id.to_string(), now);
                    } else {
                        // Stopwatch no longer running, remove from tracking
                        states.remove(stopwatch_id);
                    }
                } else {
                    // Stopwatch doesn't exist anymore
                    states.remove(stopwatch_id);
                }
            }
        }

        Ok(())
    }

    /// Start tracking a new timer
    pub async fn start_timer(&self, timer_id: &str) -> anyhow::Result<()> {
        let mut states = self.timer_states.write().await;
        states.insert(timer_id.to_string(), Instant::now());
        Ok(())
    }

    /// Start tracking a new stopwatch
    pub async fn start_stopwatch(&self, stopwatch_id: &str) -> anyhow::Result<()> {
        let mut states = self.stopwatch_states.write().await;
        states.insert(stopwatch_id.to_string(), Instant::now());
        Ok(())
    }

    /// Stop tracking a timer
    pub async fn pause_timer(&self, timer_id: &str) {
        let mut states = self.timer_states.write().await;
        states.remove(timer_id);
    }

    /// Stop tracking a stopwatch
    pub async fn pause_stopwatch(&self, stopwatch_id: &str) {
        let mut states = self.stopwatch_states.write().await;
        states.remove(stopwatch_id);
    }

    /// Get count of active timers
    pub async fn active_timer_count(&self) -> usize {
        self.timer_states.read().await.len()
    }

    /// Get count of active stopwatches
    pub async fn active_stopwatch_count(&self) -> usize {
        self.stopwatch_states.read().await.len()
    }
}

/// Create a background service without starting it
/// The service will lazily spawn when timers/stopwatches are started
pub fn create_background_service(db: Arc<Database>) -> Arc<BackgroundService> {
    Arc::new(BackgroundService::new(db))
}

/// Legacy function - kept for compatibility but now just calls create_background_service
#[deprecated(note = "Use create_background_service instead")]
pub fn spawn_background_service(db: Arc<Database>) -> Arc<BackgroundService> {
    create_background_service(db)
}
