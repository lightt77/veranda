package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/db"
)

func TestOpen(t *testing.T) {
	// Use temp directory for test
	tempDir := t.TempDir()
	cfg := &config.AppConfig{
		DataDir:      tempDir,
		DatabasePath: filepath.Join(tempDir, "test.db"),
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Verify tables were created by checking one exists
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='timers'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected timers table to exist, got count: %d", count)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
