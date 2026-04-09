package repository_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
	"github.com/lightt77/veranda/internal/repository"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	tempDir := t.TempDir()
	cfg := &config.AppConfig{
		DataDir:      tempDir,
		DatabasePath: filepath.Join(tempDir, "test.db"),
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

func TestTimerRepository(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewTimerRepository(database)
	now := time.Now()

	// Create timer
	timer := models.NewTimer(now.Unix(), "Test Timer", 25*60*1000, now)
	err := repo.Create(timer)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(timer.ID)
	if err != nil {
		t.Fatalf("Failed to get timer: %v", err)
	}
	if retrieved.Label != "Test Timer" {
		t.Errorf("Expected label 'Test Timer', got '%s'", retrieved.Label)
	}
	if retrieved.DurationMs != 25*60*1000 {
		t.Errorf("Expected duration 25min, got %d", retrieved.DurationMs)
	}

	// Update timer
	timer.Label = "Updated Timer"
	timer.Start(now)
	err = repo.Update(timer)
	if err != nil {
		t.Fatalf("Failed to update timer: %v", err)
	}

	// Verify update
	updated, _ := repo.GetByID(timer.ID)
	if updated.Label != "Updated Timer" {
		t.Errorf("Expected updated label, got '%s'", updated.Label)
	}
	if updated.Status != models.TimerRunning {
		t.Errorf("Expected status running, got %s", updated.Status)
	}

	// Get running timers
	running, err := repo.GetRunning()
	if err != nil {
		t.Fatalf("Failed to get running timers: %v", err)
	}
	if len(running) != 1 {
		t.Errorf("Expected 1 running timer, got %d", len(running))
	}

	// Soft delete
	err = repo.SoftDelete(timer.ID)
	if err != nil {
		t.Fatalf("Failed to soft delete timer: %v", err)
	}

	// Verify soft delete
	_, err = repo.GetByID(timer.ID)
	if err == nil {
		t.Error("Expected error for deleted timer, got nil")
	}

	// Get all should return empty
	all, _ := repo.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 timers after delete, got %d", len(all))
	}
}

func TestStopwatchRepository(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewStopwatchRepository(database)
	lapRepo := repository.NewStopwatchLapRepository(database)
	now := time.Now()

	// Create stopwatch
	stopwatch := models.NewStopwatch(now.Unix(), "Test Stopwatch", now)
	err := repo.Create(stopwatch)
	if err != nil {
		t.Fatalf("Failed to create stopwatch: %v", err)
	}

	// Start and update
	stopwatch.Start(now)
	err = repo.Update(stopwatch)
	if err != nil {
		t.Fatalf("Failed to update stopwatch: %v", err)
	}

	// Create lap
	lapDuration := int64(5000) // 5 seconds
	totalElapsed := stopwatch.GetCurrentElapsed(now) + lapDuration
	lap := models.NewStopwatchLap(stopwatch.ID, 1, lapDuration, totalElapsed, now)
	err = lapRepo.Create(lap)
	if err != nil {
		t.Fatalf("Failed to create lap: %v", err)
	}
	if lap.ID == 0 {
		t.Error("Expected lap ID to be set after insert")
	}

	// Get laps
	laps, err := lapRepo.GetByStopwatchID(stopwatch.ID)
	if err != nil {
		t.Fatalf("Failed to get laps: %v", err)
	}
	if len(laps) != 1 {
		t.Errorf("Expected 1 lap, got %d", len(laps))
	}
}

func TestJournalRepository(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewJournalRepository(database)
	now := time.Now()

	// Create entry
	entry := models.NewJournalEntry(now.Unix(), "Test journal entry", now)
	err := repo.Create(entry)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Get all
	entries, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Get recent
	recent, err := repo.GetRecent(10)
	if err != nil {
		t.Fatalf("Failed to get recent: %v", err)
	}
	if len(recent) != 1 {
		t.Errorf("Expected 1 recent entry, got %d", len(recent))
	}
}

func TestSettingsRepository(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewSettingsRepository(database)

	// Set value
	err := repo.Set("test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to set setting: %v", err)
	}

	// Get value
	value, err := repo.Get("test_key")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Get non-existent
	value, err = repo.Get("non_existent")
	if err != nil {
		t.Fatalf("Failed to get non-existent: %v", err)
	}
	if value != "" {
		t.Errorf("Expected empty string for non-existent, got '%s'", value)
	}

	// Get with default
	value, err = repo.GetOrDefault("non_existent", "default_value")
	if err != nil {
		t.Fatalf("Failed to get with default: %v", err)
	}
	if value != "default_value" {
		t.Errorf("Expected default value, got '%s'", value)
	}

	// Update value
	err = repo.Set("test_key", "updated_value")
	if err != nil {
		t.Fatalf("Failed to update setting: %v", err)
	}

	value, _ = repo.Get("test_key")
	if value != "updated_value" {
		t.Errorf("Expected 'updated_value', got '%s'", value)
	}
}
