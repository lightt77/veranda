//! Background service for managing timers and stopwatches
//! 
//! This module provides a daemon that continues running timers and stopwatches
//! even when the GUI window is closed. It communicates with the GUI via IPC.

use crate::db::Database;
use crate::models::timer::TimerStatus;
use crate::models::stopwatch::StopwatchStatus;
use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use tokio::time::interval;

/// Background service state
pub struct BackgroundService {
    db: Arc<Database>,
    /// Active timer calculations (timer_id -> last_update_time)
    timer_states: RwLock<HashMap<String, Instant>>,
    /// Active stopwatch calculations (stopwatch_id -> last_update_time)
    stopwatch_states: RwLock<HashMap<String, Instant>>,
    /// Service running flag
    running: RwLock<bool>,
}

impl BackgroundService {
    /// Create a new background service
    pub fn new(db: Arc<Database>) -> Self {
        Self {
            db,
            timer_states: RwLock::new(HashMap::new()),
            stopwatch_states: RwLock::new(HashMap::new()),
            running: RwLock::new(false),
        }
    }

    /// Start the background service
    pub async fn start(&self) -> anyhow::Result<()> {
        // Mark as running
        *self.running.write().await = true;

        // Load active timers and stopwatches from database
        self.load_active_timers().await?;
        self.load_active_stopwatches().await?;

        // Start the main loop
        let mut ticker = interval(Duration::from_millis(100)); // 10Hz update rate

        while *self.running.read().await {
            ticker.tick().await;
            
            // Process timers
            if let Err(e) = self.process_timers().await {
                eprintln!("Error processing timers: {}", e);
            }

            // Process stopwatches
            if let Err(e) = self.process_stopwatches().await {
                eprintln!("Error processing stopwatches: {}", e);
            }
        }

        Ok(())
    }

    /// Stop the background service
    pub async fn stop(&self) {
        *self.running.write().await = false;
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

    /// Process all active timers
    async fn process_timers(&self) -> anyhow::Result<()> {
        let timer_ids: Vec<String> = {
            let states = self.timer_states.read().await;
            states.keys().cloned().collect()
        };

        for timer_id in timer_ids {
            self.update_timer(&timer_id).await?;
        }

        Ok(())
    }

    /// Process all active stopwatches
    async fn process_stopwatches(&self) -> anyhow::Result<()> {
        let stopwatch_ids: Vec<String> = {
            let states = self.stopwatch_states.read().await;
            states.keys().cloned().collect()
        };

        for stopwatch_id in stopwatch_ids {
            self.update_stopwatch(&stopwatch_id).await?;
        }

        Ok(())
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
                            
                            // TODO: Trigger notification
                            println!("Timer completed: {}", timer.label);
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

/// Spawn the background service in a separate task
pub fn spawn_background_service(db: Arc<Database>) -> Arc<BackgroundService> {
    let service = Arc::new(BackgroundService::new(db));
    let service_clone = service.clone();

    tokio::spawn(async move {
        if let Err(e) = service_clone.start().await {
            eprintln!("Background service error: {}", e);
        }
    });

    service
}
