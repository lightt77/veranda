-- Migration: Create todos table
CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    task_name TEXT NOT NULL,
    task_desc TEXT,
    project TEXT,
    labels TEXT, -- JSON array of labels
    status TEXT NOT NULL CHECK (status IN ('TODO', 'IN_PROGRESS', 'DONE', 'ARCHIVED')),
    last_status_update_ts TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_project ON todos(project);
CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
