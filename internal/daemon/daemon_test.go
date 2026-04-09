package daemon_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/daemon"
	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/service"
)

func setupTestDaemon(t *testing.T) (*daemon.Server, *config.AppConfig, func()) {
	cfg := &config.AppConfig{
		AppName:      "veranda-test",
		AppVersion:   "0.1.0-test",
		DataDir:      t.TempDir(),
		DatabasePath: t.TempDir() + "/test.db",
		DaemonPort:   0, // Let OS assign port
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

	// Ambient service may fail without audio hardware, that's ok
	ambientService, _ := audio.NewAmbientService(cfg, timerService, stopwatchService)

	server := daemon.NewServer(
		cfg,
		timerService,
		stopwatchService,
		journalService,
		settingsService,
		ambientService,
	)

	cleanup := func() {
		server.Stop()
		database.Close()
	}

	return server, cfg, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	server, _, cleanup := setupTestDaemon(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	server.GetRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestTimerEndpoints(t *testing.T) {
	server, _, cleanup := setupTestDaemon(t)
	defer cleanup()

	// Create timer
	createReq := map[string]interface{}{
		"label":       "Test Timer",
		"duration_ms": 25 * 60 * 1000,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/timers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var timer map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &timer); err != nil {
		t.Fatalf("Failed to decode timer: %v", err)
	}

	if timer["label"] != "Test Timer" {
		t.Errorf("Expected label 'Test Timer', got '%v'", timer["label"])
	}

	// Get all timers
	req = httptest.NewRequest("GET", "/timers", nil)
	rr = httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var timers []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &timers); err != nil {
		t.Fatalf("Failed to decode timers: %v", err)
	}

	if len(timers) != 1 {
		t.Errorf("Expected 1 timer, got %d", len(timers))
	}
}

func TestClient(t *testing.T) {
	// Create a test server
	cfg := &config.AppConfig{
		DaemonPort: 9999, // Use a test port
	}

	database, err := db.Open(&config.AppConfig{
		DatabasePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	timerRepo := repository.NewTimerRepository(database)
	stopwatchRepo := repository.NewStopwatchRepository(database)
	lapRepo := repository.NewStopwatchLapRepository(database)
	journalRepo := repository.NewJournalRepository(database)
	settingsRepo := repository.NewSettingsRepository(database)

	timerService := service.NewTimerService(timerRepo)
	stopwatchService := service.NewStopwatchService(stopwatchRepo, lapRepo)
	journalService := service.NewJournalService(journalRepo)
	settingsService := service.NewSettingsService(settingsRepo)

	server := daemon.NewServer(
		cfg,
		timerService,
		stopwatchService,
		journalService,
		settingsService,
		nil,
	)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test client
	client := daemon.NewClient(9999)

	// Test health check
	if !client.IsRunning() {
		t.Error("Expected daemon to be running")
	}

	health, err := client.GetHealth()
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", health["status"])
	}
}
