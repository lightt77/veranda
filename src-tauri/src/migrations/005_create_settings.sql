-- Migration: Create settings table
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT OR IGNORE INTO settings (key, value) VALUES 
    ('notification_sound_enabled', 'true'),
    ('notification_sound_name', 'default'),
    ('background_opacity', '0.3'),
    ('theme', 'dark');
