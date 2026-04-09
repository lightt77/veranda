package service_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/service"
)

func setupTestServices(t *testing.T) (*service.TimerService, *service.StopwatchService, *service.JournalService, *service.SettingsService, func()) {
	tempDir := t.TempDir()
	cfg := &config.AppConfig{
		DataDir:      tempDir,
		DatabasePath: filepath.Join(tempDir, "test.db"),
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	timerRepo := repository.NewTimerRepository(database)
	stopwatchRepo := repository.NewStopwatchRepository(database)
	lapRepo := repository.NewStopwatchLapRepository(database)
	journalRepo := repository.NewJournalRepository(database)
	settingsRepo := repository.NewSettingsRepository(database)

	timerService := service.NewTimerService(timerRepo)
	stopwatchService := service.NewStopwatchService(stopwatchRepo, lapRepo)
	journalService := service.NewJournalService(journalRepo)
	settingsService := service.NewSettingsService(settingsRepo)

	cleanup := func() {
		database.Close()
		// tempDir is automatically cleaned up by Go's testing framework
	}

	return timerService, stopwatchService, journalService, settingsService, cleanup
}

func TestTimerService(t *testing.T) {
	timerService, _, _, _, cleanup := setupTestServices(t)
	defer cleanup()

	// Create timer
	timer, err := timerService.Create("Test Timer", 25*60*1000)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}
	if timer.Label != "Test Timer" {
		t.Errorf("Expected label 'Test Timer', got '%s'", timer.Label)
	}
	if timer.DurationMs != 25*60*1000 {
		t.Errorf("Expected duration 25min, got %d", timer.DurationMs)
	}

	// Start timer
	started, err := timerService.Start(timer.ID)
	if err != nil {
		t.Fatalf("Failed to start timer: %v", err)
	}
	if started.Status != models.TimerRunning {
		t.Errorf("Expected status running, got %s", started.Status)
	}
	if started.StartedAtMs == nil {
		t.Error("Expected started_at to be set")
	}

	// Pause timer
	time.Sleep(100 * time.Millisecond) // Let some time pass
	paused, err := timerService.Pause(timer.ID)
	if err != nil {
		t.Fatalf("Failed to pause timer: %v", err)
	}
	if paused.Status != models.TimerPaused {
		t.Errorf("Expected status paused, got %s", paused.Status)
	}
	if paused.RemainingMs >= 25*60*1000 {
		t.Error("Expected remaining time to have decreased")
	}

	// Resume timer
	resumed, err := timerService.Start(timer.ID)
	if err != nil {
		t.Fatalf("Failed to resume timer: %v", err)
	}
	if resumed.Status != models.TimerRunning {
		t.Errorf("Expected status running after resume, got %s", resumed.Status)
	}

	// Complete timer
	completed, err := timerService.Complete(timer.ID)
	if err != nil {
		t.Fatalf("Failed to complete timer: %v", err)
	}
	if completed.Status != models.TimerCompleted {
		t.Errorf("Expected status completed, got %s", completed.Status)
	}
	if completed.RemainingMs != 0 {
		t.Errorf("Expected remaining 0, got %d", completed.RemainingMs)
	}

	// Get running timers (should be empty)
	running, _ := timerService.GetRunning()
	if len(running) != 0 {
		t.Errorf("Expected 0 running timers, got %d", len(running))
	}
}

func TestTimerServiceQuickCreate(t *testing.T) {
	timerService, _, _, _, cleanup := setupTestServices(t)
	defer cleanup()

	// Create and start in one call
	timer, err := timerService.CreateQuick("Quick Timer", 5*60*1000)
	if err != nil {
		t.Fatalf("Failed to quick create timer: %v", err)
	}
	if timer.Status != models.TimerRunning {
		t.Errorf("Expected timer to be running, got %s", timer.Status)
	}

	// Verify AnyRunning returns true
	running, err := timerService.AnyRunning()
	if err != nil {
		t.Fatalf("Failed to check any running: %v", err)
	}
	if !running {
		t.Error("Expected AnyRunning to be true")
	}
}

func TestStopwatchService(t *testing.T) {
	_, stopwatchService, _, _, cleanup := setupTestServices(t)
	defer cleanup()

	// Create stopwatch
	sw, err := stopwatchService.Create("Test Stopwatch")
	if err != nil {
		t.Fatalf("Failed to create stopwatch: %v", err)
	}
	if sw.Label != "Test Stopwatch" {
		t.Errorf("Expected label 'Test Stopwatch', got '%s'", sw.Label)
	}

	// Start stopwatch
	started, err := stopwatchService.Start(sw.ID)
	if err != nil {
		t.Fatalf("Failed to start stopwatch: %v", err)
	}
	if started.Status != models.StopwatchRunning {
		t.Errorf("Expected status running, got %s", started.Status)
	}

	// Wait and record lap
	time.Sleep(50 * time.Millisecond)
	lapped, lap, err := stopwatchService.Lap(sw.ID)
	if err != nil {
		t.Fatalf("Failed to record lap: %v", err)
	}
	if lap.LapNumber != 1 {
		t.Errorf("Expected lap number 1, got %d", lap.LapNumber)
	}
	if lap.LapDurationMs <= 0 {
		t.Error("Expected positive lap duration")
	}
	if lapped.ElapsedMs <= 0 {
		t.Error("Expected positive elapsed time")
	}

	// Record another lap
	time.Sleep(50 * time.Millisecond)
	_, lap2, err := stopwatchService.Lap(sw.ID)
	if err != nil {
		t.Fatalf("Failed to record second lap: %v", err)
	}
	if lap2.LapNumber != 2 {
		t.Errorf("Expected lap number 2, got %d", lap2.LapNumber)
	}

	// Get laps
	laps, err := stopwatchService.GetLaps(sw.ID)
	if err != nil {
		t.Fatalf("Failed to get laps: %v", err)
	}
	if len(laps) != 2 {
		t.Errorf("Expected 2 laps, got %d", len(laps))
	}

	// Stop stopwatch
	stopped, err := stopwatchService.Stop(sw.ID)
	if err != nil {
		t.Fatalf("Failed to stop stopwatch: %v", err)
	}
	if stopped.Status != models.StopwatchStopped {
		t.Errorf("Expected status stopped, got %s", stopped.Status)
	}

	// Reset stopwatch
	reset, err := stopwatchService.Reset(sw.ID)
	if err != nil {
		t.Fatalf("Failed to reset stopwatch: %v", err)
	}
	if reset.ElapsedMs != 0 {
		t.Errorf("Expected elapsed 0 after reset, got %d", reset.ElapsedMs)
	}

	// Laps should be deleted
	laps, _ = stopwatchService.GetLaps(sw.ID)
	if len(laps) != 0 {
		t.Errorf("Expected 0 laps after reset, got %d", len(laps))
	}
}

func TestStopwatchServiceQuickCreate(t *testing.T) {
	_, stopwatchService, _, _, cleanup := setupTestServices(t)
	defer cleanup()

	// Create and start in one call
	sw, err := stopwatchService.CreateQuick("Quick Stopwatch")
	if err != nil {
		t.Fatalf("Failed to quick create stopwatch: %v", err)
	}
	if sw.Status != models.StopwatchRunning {
		t.Errorf("Expected stopwatch to be running, got %s", sw.Status)
	}
}

func TestJournalService(t *testing.T) {
	_, _, journalService, _, cleanup := setupTestServices(t)
	defer cleanup()

	// Create entry
	entry, err := journalService.Create("Test journal entry content")
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}
	if entry.Log != "Test journal entry content" {
		t.Errorf("Expected log content, got '%s'", entry.Log)
	}

	// Get all
	entries, err := journalService.GetAll()
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Get recent
	recent, err := journalService.GetRecent(10)
	if err != nil {
		t.Fatalf("Failed to get recent: %v", err)
	}
	if len(recent) != 1 {
		t.Errorf("Expected 1 recent entry, got %d", len(recent))
	}
}

func TestSettingsService(t *testing.T) {
	_, _, _, settingsService, cleanup := setupTestServices(t)
	defer cleanup()

	// Set value
	err := settingsService.Set("test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to set setting: %v", err)
	}

	// Get value
	value, err := settingsService.Get("test_key")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Get with default (should return actual value)
	value, err = settingsService.GetOrDefault("test_key", "default")
	if err != nil {
		t.Fatalf("Failed to get with default: %v", err)
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Get with default for non-existent (should return default)
	value, err = settingsService.GetOrDefault("non_existent", "default")
	if err != nil {
		t.Fatalf("Failed to get non-existent: %v", err)
	}
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}
}

func TestCommonUtilities(t *testing.T) {
	// Test ID generation (Unix epoch seconds)
	id1 := service.GenerateID()
	id2 := service.GenerateID()
	// IDs should be non-decreasing (may be equal if generated in same second)
	if id1 > id2 {
		t.Error("IDs should be non-decreasing")
	}
	// Verify ID is reasonable (within last minute)
	now := time.Now().Unix()
	if id1 < now-60 || id1 > now+60 {
		t.Error("ID should be a reasonable Unix timestamp")
	}

	// Test duration parsing
	ms, err := service.ParseDurationMs("25m")
	if err != nil {
		t.Fatalf("Failed to parse duration: %v", err)
	}
	if ms != 25*60*1000 {
		t.Errorf("Expected 25min in ms, got %d", ms)
	}

	// Test duration parsing with hours
	ms, err = service.ParseDurationMs("1h30m")
	if err != nil {
		t.Fatalf("Failed to parse duration: %v", err)
	}
	if ms != 90*60*1000 {
		t.Errorf("Expected 90min in ms, got %d", ms)
	}

	// Test duration formatting
	formatted := service.FormatDurationMs(25*60*1000 + 30*1000)
	if formatted != "25m 30s" {
		t.Errorf("Expected '25m 30s', got '%s'", formatted)
	}

	formatted = service.FormatDurationMs(90 * 60 * 1000)
	if formatted != "1h 30m" {
		t.Errorf("Expected '1h 30m', got '%s'", formatted)
	}
}
