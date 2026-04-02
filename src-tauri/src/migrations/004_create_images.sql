-- Migration: Create images table
CREATE TABLE IF NOT EXISTS images (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    local_path TEXT NOT NULL,
    downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    metadata TEXT -- JSON for additional metadata like dimensions, source info
);

CREATE INDEX IF NOT EXISTS idx_images_is_active ON images(is_active);
CREATE INDEX IF NOT EXISTS idx_images_downloaded_at ON images(downloaded_at);
