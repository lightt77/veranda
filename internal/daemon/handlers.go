package daemon

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lightt77/veranda/internal/models"
)

// Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": s.config.AppVersion,
	})
}

// Timer handlers

func (s *Server) handleGetTimers(w http.ResponseWriter, r *http.Request) {
	timers, err := s.timerService.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s.enrichTimers(timers))
}

func (s *Server) handleGetRunningTimers(w http.ResponseWriter, r *http.Request) {
	timers, err := s.timerService.GetRunning()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s.enrichTimers(timers))
}

func (s *Server) handleGetTimer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	timer, err := s.timerService.Get(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichTimer(timer))
}

type createTimerRequest struct {
	Label      string `json:"label"`
	DurationMs int64  `json:"duration_ms"`
	Start      bool   `json:"start,omitempty"`
}

func (s *Server) handleCreateTimer(w http.ResponseWriter, r *http.Request) {
	var req createTimerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DurationMs <= 0 {
		respondError(w, http.StatusBadRequest, "duration must be positive")
		return
	}

	var timer interface{}
	var err error

	if req.Start {
		timer, err = s.timerService.CreateQuick(req.Label, req.DurationMs)
	} else {
		timer, err = s.timerService.Create(req.Label, req.DurationMs)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, timer)
}

func (s *Server) handleStartTimer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	timer, err := s.timerService.Start(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, timer)
}

func (s *Server) handlePauseTimer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	timer, err := s.timerService.Pause(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, timer)
}

func (s *Server) handleCompleteTimer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	timer, err := s.timerService.Complete(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, timer)
}

func (s *Server) handleDeleteTimer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := s.timerService.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Stopwatch handlers

func (s *Server) handleGetStopwatches(w http.ResponseWriter, r *http.Request) {
	stopwatches, err := s.stopwatchService.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s.enrichStopwatches(stopwatches))
}

func (s *Server) handleGetRunningStopwatches(w http.ResponseWriter, r *http.Request) {
	stopwatches, err := s.stopwatchService.GetRunning()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s.enrichStopwatches(stopwatches))
}

func (s *Server) handleGetStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	stopwatch, err := s.stopwatchService.Get(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichStopwatch(stopwatch))
}

type createStopwatchRequest struct {
	Label string `json:"label"`
	Start bool   `json:"start,omitempty"`
}

func (s *Server) handleCreateStopwatch(w http.ResponseWriter, r *http.Request) {
	var req createStopwatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var stopwatch *models.Stopwatch
	var err error

	if req.Start {
		stopwatch, err = s.stopwatchService.CreateQuick(req.Label)
	} else {
		stopwatch, err = s.stopwatchService.Create(req.Label)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, s.enrichStopwatch(stopwatch))
}

func (s *Server) handleStartStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	stopwatch, err := s.stopwatchService.Start(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichStopwatch(stopwatch))
}

func (s *Server) handleStopStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	stopwatch, err := s.stopwatchService.Stop(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichStopwatch(stopwatch))
}

func (s *Server) handleLapStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	stopwatch, lap, err := s.stopwatchService.Lap(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"stopwatch": s.enrichStopwatch(stopwatch),
		"lap":       lap,
	})
}

func (s *Server) handleResetStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	stopwatch, err := s.stopwatchService.Reset(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichStopwatch(stopwatch))
}

func (s *Server) handleGetLaps(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	laps, err := s.stopwatchService.GetLaps(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, laps)
}

func (s *Server) handleDeleteStopwatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := s.stopwatchService.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Journal handlers

func (s *Server) handleGetJournalEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := s.journalService.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

type createJournalRequest struct {
	Log string `json:"log"`
}

func (s *Server) handleCreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var req createJournalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := s.journalService.Create(req.Log)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

func (s *Server) handleGetRecentEntries(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	entries, err := s.journalService.GetRecent(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

// Settings handlers

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.settingsService.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (s *Server) handleGetSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, err := s.settingsService.Get(key)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"key": key, "value": value})
}

type setSettingRequest struct {
	Value string `json:"value"`
}

func (s *Server) handleSetSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req setSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.settingsService.Set(key, req.Value); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"key": key, "value": req.Value})
}

// Helper methods to enrich timer responses with computed fields

func (s *Server) enrichTimer(timer *models.Timer) map[string]interface{} {
	now := time.Now()
	currentRemaining := timer.GetCurrentRemaining(now)
	progress := float64(0)
	if timer.DurationMs > 0 {
		progress = float64(timer.DurationMs-currentRemaining) / float64(timer.DurationMs)
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
	}

	return map[string]interface{}{
		"id":                   timer.ID,
		"label":                timer.Label,
		"duration_ms":          timer.DurationMs,
		"remaining_ms":         timer.RemainingMs,
		"current_remaining_ms": currentRemaining,
		"status":               timer.Status,
		"started_at_ms":        timer.StartedAtMs,
		"paused_at_ms":         timer.PausedAtMs,
		"completed_at_ms":      timer.CompletedAtMs,
		"progress":             progress,
		"updated_at_ms":        timer.UpdatedAtMs,
	}
}

func (s *Server) enrichTimers(timers []*models.Timer) []map[string]interface{} {
	result := make([]map[string]interface{}, len(timers))
	for i, timer := range timers {
		result[i] = s.enrichTimer(timer)
	}
	return result
}

// enrichStopwatch enriches a stopwatch response with computed current elapsed time
func (s *Server) enrichStopwatch(stopwatch *models.Stopwatch) map[string]interface{} {
	now := time.Now()
	currentElapsed := stopwatch.GetCurrentElapsed(now)

	return map[string]interface{}{
		"id":                 stopwatch.ID,
		"label":              stopwatch.Label,
		"elapsed_ms":         stopwatch.ElapsedMs,
		"current_elapsed_ms": currentElapsed,
		"status":             stopwatch.Status,
		"started_at_ms":      stopwatch.StartedAtMs,
		"stopped_at_ms":      stopwatch.StoppedAtMs,
		"updated_at_ms":      stopwatch.UpdatedAtMs,
	}
}

func (s *Server) enrichStopwatches(stopwatches []*models.Stopwatch) []map[string]interface{} {
	result := make([]map[string]interface{}, len(stopwatches))
	for i, sw := range stopwatches {
		result[i] = s.enrichStopwatch(sw)
	}
	return result
}

// Daemon control handlers

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "shutting down"})
	// Shutdown in a goroutine so we can send the response first
	go s.StopAndExit()
}

// Ambience sound handlers

func (s *Server) handleAmbientStatus(w http.ResponseWriter, r *http.Request) {
	if s.ambienceService == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"message": "ambience sound not available",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"playing": s.ambienceService.IsPlaying(),
		"note":    "volume and file list now configured per-sound via TUI",
	})
}

func (s *Server) handleAmbientPlay(w http.ResponseWriter, r *http.Request) {
	if s.ambienceService == nil {
		respondError(w, http.StatusServiceUnavailable, "ambience sound not available")
		return
	}

	if err := s.ambienceService.Play(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "playing"})
}

func (s *Server) handleAmbientStop(w http.ResponseWriter, r *http.Request) {
	if s.ambienceService == nil {
		respondError(w, http.StatusServiceUnavailable, "ambience sound not available")
		return
	}

	s.ambienceService.StopPlayback()
	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

type setVolumeRequest struct {
	Volume float64 `json:"volume"`
}

func (s *Server) handleAmbientSetVolume(w http.ResponseWriter, r *http.Request) {
	if s.ambienceService == nil {
		respondError(w, http.StatusServiceUnavailable, "ambience sound not available")
		return
	}

	// Volume is now controlled per-sound via TUI/config, not via API
	respondError(w, http.StatusNotImplemented, "volume is now controlled per-sound via TUI. Use the Ambience tab tab.")
}

func (s *Server) handleAmbientRefresh(w http.ResponseWriter, r *http.Request) {
	if s.ambienceService == nil {
		respondError(w, http.StatusServiceUnavailable, "ambience sound not available")
		return
	}

	s.ambienceService.RefreshPlayback()
	respondJSON(w, http.StatusOK, map[string]string{"status": "refreshed"})
}
