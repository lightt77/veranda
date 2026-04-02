-- Migration: Create stopwatches and stopwatch_laps tables
CREATE TABLE IF NOT EXISTS stopwatches (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    elapsed_seconds INTEGER DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('CREATED', 'RUNNING', 'PAUSED', 'STOPPED')),
    started_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stopwatch_laps (
    id TEXT PRIMARY KEY,
    stopwatch_id TEXT NOT NULL,
    lap_number INTEGER NOT NULL,
    lap_duration_ms INTEGER NOT NULL,
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stopwatch_id) REFERENCES stopwatches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stopwatches_status ON stopwatches(status);
CREATE INDEX IF NOT EXISTS idx_stopwatch_laps_stopwatch_id ON stopwatch_laps(stopwatch_id);
