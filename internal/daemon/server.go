// Package daemon provides the HTTP server for cross-process communication.
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/service"
)

// Server represents the HTTP daemon server
type Server struct {
	router           *chi.Mux
	httpServer       *http.Server
	config           *config.AppConfig
	timerService     *service.TimerService
	stopwatchService *service.StopwatchService
	journalService   *service.JournalService
	settingsService  *service.SettingsService
	ambienceService  *audio.AmbienceService

	// Activity tracking for auto-shutdown
	lastActivity  time.Time
	activityMutex sync.RWMutex
	idleTimeout   time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewServer creates a new daemon server
func NewServer(
	cfg *config.AppConfig,
	timerService *service.TimerService,
	stopwatchService *service.StopwatchService,
	journalService *service.JournalService,
	settingsService *service.SettingsService,
	ambienceService *audio.AmbienceService,
) *Server {
	r := chi.NewRouter()

	s := &Server{
		router:           r,
		config:           cfg,
		timerService:     timerService,
		stopwatchService: stopwatchService,
		journalService:   journalService,
		settingsService:  settingsService,
		ambienceService:  ambienceService,
		lastActivity:     time.Now(),
		idleTimeout:      30 * time.Second, // Shutdown after 30s of inactivity
		stopChan:         make(chan struct{}),
	}

	s.setupRoutes()
	return s
}

// RecordActivity marks that there was recent activity
func (s *Server) RecordActivity() {
	s.activityMutex.Lock()
	s.lastActivity = time.Now()
	s.activityMutex.Unlock()
}

// GetLastActivity returns the time of last activity
func (s *Server) GetLastActivity() time.Time {
	s.activityMutex.RLock()
	defer s.activityMutex.RUnlock()
	return s.lastActivity
}

// HasWork returns true if there's active work to do
func (s *Server) HasWork() bool {
	// Check for running timers
	timersRunning, _ := s.timerService.AnyRunning()
	if timersRunning {
		return true
	}

	// Check for running stopwatches
	stopwatchesRunning, _ := s.stopwatchService.AnyRunning()
	if stopwatchesRunning {
		return true
	}

	// Check if ambience sound is playing
	if s.ambienceService != nil && s.ambienceService.IsPlaying() {
		return true
	}

	return false
}

// IsIdle returns true if daemon has been idle for longer than timeout
func (s *Server) IsIdle() bool {
	s.activityMutex.RLock()
	idleTime := time.Since(s.lastActivity)
	s.activityMutex.RUnlock()

	return idleTime > s.idleTimeout && !s.HasWork()
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(jsonContentType)
	s.router.Use(s.activityMiddleware)

	// Health check
	s.router.Get("/health", s.handleHealth)

	// Daemon control
	s.router.Post("/daemon/shutdown", s.handleShutdown)

	// Timer routes
	s.router.Route("/timers", func(r chi.Router) {
		r.Get("/", s.handleGetTimers)
		r.Post("/", s.handleCreateTimer)
		r.Get("/running", s.handleGetRunningTimers)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetTimer)
			r.Post("/start", s.handleStartTimer)
			r.Post("/pause", s.handlePauseTimer)
			r.Post("/stop", s.handlePauseTimer) // Alias
			r.Post("/complete", s.handleCompleteTimer)
			r.Delete("/", s.handleDeleteTimer)
		})
	})

	// Stopwatch routes
	s.router.Route("/stopwatches", func(r chi.Router) {
		r.Get("/", s.handleGetStopwatches)
		r.Post("/", s.handleCreateStopwatch)
		r.Get("/running", s.handleGetRunningStopwatches)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetStopwatch)
			r.Get("/laps", s.handleGetLaps)
			r.Post("/start", s.handleStartStopwatch)
			r.Post("/stop", s.handleStopStopwatch)
			r.Post("/lap", s.handleLapStopwatch)
			r.Post("/reset", s.handleResetStopwatch)
			r.Delete("/", s.handleDeleteStopwatch)
		})
	})

	// Journal routes
	s.router.Route("/journal", func(r chi.Router) {
		r.Get("/", s.handleGetJournalEntries)
		r.Post("/", s.handleCreateJournalEntry)
		r.Get("/recent", s.handleGetRecentEntries)
	})

	// Settings routes
	s.router.Route("/settings", func(r chi.Router) {
		r.Get("/", s.handleGetSettings)
		r.Get("/{key}", s.handleGetSetting)
		r.Put("/{key}", s.handleSetSetting)
	})

	// Ambient sound routes
	s.router.Route("/ambient", func(r chi.Router) {
		r.Get("/status", s.handleAmbientStatus)
		r.Post("/play", s.handleAmbientPlay)
		r.Post("/stop", s.handleAmbientStop)
		r.Put("/volume", s.handleAmbientSetVolume)
	})
}

// activityMiddleware records API activity
func (s *Server) activityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.RecordActivity()
		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server and idle monitoring
func (s *Server) Start() error {
	addr := fmt.Sprintf("localhost:%d", s.config.DaemonPort)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	fmt.Printf("Daemon server starting on %s\n", addr)

	// Start HTTP server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Daemon server error: %v\n", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Start idle monitoring
	s.wg.Add(1)
	go s.monitorIdle()

	return nil
}

// monitorIdle checks for idle state and shuts down when appropriate
func (s *Server) monitorIdle() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.IsIdle() {
				fmt.Println("Daemon idle - no active timers or ambient sound. Shutting down...")
				// Graceful shutdown and exit
				go s.StopAndExit()
				return
			}
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	close(s.stopChan)

	if s.httpServer == nil {
		s.wg.Wait()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	s.wg.Wait()

	return err
}

// StopAndExit gracefully shuts down and exits the process
func (s *Server) StopAndExit() error {
	err := s.Stop()
	os.Exit(0)
	return err
}

// Middleware
func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

// GetRouter returns the chi router for testing
func (s *Server) GetRouter() http.Handler {
	return s.router
}
