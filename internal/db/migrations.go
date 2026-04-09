package db

// timersTable stores timer data with millisecond precision
const timersTable = `
CREATE TABLE IF NOT EXISTS timers (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds (natural sortable ID)
	label TEXT NOT NULL,                       -- Timer name/description
	duration_ms INTEGER NOT NULL,              -- Total duration in milliseconds
	remaining_ms INTEGER NOT NULL,             -- Remaining time in milliseconds
	status TEXT NOT NULL DEFAULT 'pending',    -- pending, running, paused, completed
	started_at_ms INTEGER,                     -- When timer started (epoch ms)
	paused_at_ms INTEGER,                      -- When timer was paused (epoch ms)
	completed_at_ms INTEGER,                   -- When timer completed (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_timers_status ON timers(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_timers_deleted_at ON timers(deleted_at);
`

// stopwatchesTable stores stopwatch data
const stopwatchesTable = `
CREATE TABLE IF NOT EXISTS stopwatches (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	label TEXT NOT NULL,                       -- Stopwatch name/description
	elapsed_ms INTEGER NOT NULL DEFAULT 0,     -- Total elapsed time in milliseconds
	status TEXT NOT NULL DEFAULT 'stopped',    -- stopped, running
	started_at_ms INTEGER,                     -- When stopwatch started (epoch ms)
	stopped_at_ms INTEGER,                     -- When stopwatch stopped (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_stopwatches_status ON stopwatches(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stopwatches_deleted_at ON stopwatches(deleted_at);
`

// stopwatchLapsTable stores individual lap records for stopwatches
const stopwatchLapsTable = `
CREATE TABLE IF NOT EXISTS stopwatch_laps (
	id INTEGER PRIMARY KEY AUTOINCREMENT,      -- Auto-increment ID
	stopwatch_id INTEGER NOT NULL,             -- Reference to stopwatches.id (epoch seconds)
	lap_number INTEGER NOT NULL,               -- Sequential lap number (1, 2, 3...)
	lap_duration_ms INTEGER NOT NULL,          -- Duration of this specific lap
	total_elapsed_ms INTEGER NOT NULL,         -- Total elapsed time at lap completion
	created_at_ms INTEGER NOT NULL,            -- When lap was created (epoch ms)
	
	FOREIGN KEY (stopwatch_id) REFERENCES stopwatches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_laps_stopwatch_id ON stopwatch_laps(stopwatch_id);
CREATE INDEX IF NOT EXISTS idx_laps_lap_number ON stopwatch_laps(stopwatch_id, lap_number);
`

// journalEntriesTable stores journal log entries
const journalEntriesTable = `
CREATE TABLE IF NOT EXISTS journal_entries (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	log TEXT NOT NULL,                         -- Journal entry content
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	created_at_ms INTEGER NOT NULL             -- Creation timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_journal_deleted_at ON journal_entries(deleted_at);
CREATE INDEX IF NOT EXISTS idx_journal_created_at ON journal_entries(created_at_ms);
`

// scheduledTasksTable stores recurring/scheduled tasks (future feature)
const scheduledTasksTable = `
CREATE TABLE IF NOT EXISTS scheduled_tasks (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	task_name TEXT NOT NULL,                   -- Task identifier/name
	schedule_type TEXT NOT NULL,               -- cron, interval, once
	schedule_data TEXT NOT NULL,               -- JSON: cron expression or interval config
	next_due_at_ms INTEGER,                    -- Next scheduled run time (epoch ms)
	last_completed_at_ms INTEGER,              -- Last completion time (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	created_at_ms INTEGER NOT NULL,            -- Creation timestamp (epoch ms)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_next_due ON scheduled_tasks(next_due_at_ms) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_deleted_at ON scheduled_tasks(deleted_at);
`

// settingsTable stores application settings as key-value pairs
const settingsTable = `
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,                      -- Setting identifier
	value TEXT NOT NULL,                       -- Setting value (JSON or string)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);
`
