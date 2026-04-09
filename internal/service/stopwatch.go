package service

import (
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/models"
	"github.com/lightt77/veranda/internal/repository"
)

// LapNotifier interface for lap notifications
type LapNotifier interface {
	StopLapNotification(stopwatchName string, lapNumber int, lapDurationMs int64)
}

// StopwatchService handles stopwatch business logic
type StopwatchService struct {
	stopwatchRepo    *repository.StopwatchRepository
	stopwatchLapRepo *repository.StopwatchLapRepository
	lapNotifier      LapNotifier
}

// NewStopwatchService creates a new stopwatch service
func NewStopwatchService(
	stopwatchRepo *repository.StopwatchRepository,
	stopwatchLapRepo *repository.StopwatchLapRepository,
) *StopwatchService {
	return &StopwatchService{
		stopwatchRepo:    stopwatchRepo,
		stopwatchLapRepo: stopwatchLapRepo,
	}
}

// SetLapNotifier sets the lap notifier
func (s *StopwatchService) SetLapNotifier(notifier LapNotifier) {
	s.lapNotifier = notifier
}

// Create creates a new stopwatch
func (s *StopwatchService) Create(label string) (*models.Stopwatch, error) {
	now := time.Now()
	stopwatch := models.NewStopwatch(GenerateID(), label, now)

	if err := s.stopwatchRepo.Create(stopwatch); err != nil {
		return nil, fmt.Errorf("failed to create stopwatch: %w", err)
	}

	return stopwatch, nil
}

// CreateQuick creates and starts a stopwatch immediately
func (s *StopwatchService) CreateQuick(label string) (*models.Stopwatch, error) {
	stopwatch, err := s.Create(label)
	if err != nil {
		return nil, err
	}

	return s.Start(stopwatch.ID)
}

// Start starts a stopwatch
func (s *StopwatchService) Start(id int64) (*models.Stopwatch, error) {
	stopwatch, err := s.stopwatchRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("stopwatch not found: %w", err)
	}

	if stopwatch.Status == models.StopwatchRunning {
		return nil, fmt.Errorf("stopwatch is already running")
	}

	now := time.Now()
	stopwatch.Start(now)

	if err := s.stopwatchRepo.Update(stopwatch); err != nil {
		return nil, fmt.Errorf("failed to start stopwatch: %w", err)
	}

	return stopwatch, nil
}

// Stop stops a stopwatch
func (s *StopwatchService) Stop(id int64) (*models.Stopwatch, error) {
	stopwatch, err := s.stopwatchRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("stopwatch not found: %w", err)
	}

	if stopwatch.Status != models.StopwatchRunning {
		return nil, fmt.Errorf("stopwatch is not running")
	}

	now := time.Now()
	stopwatch.Stop(now)

	if err := s.stopwatchRepo.Update(stopwatch); err != nil {
		return nil, fmt.Errorf("failed to stop stopwatch: %w", err)
	}

	return stopwatch, nil
}

// Lap records a lap on a running stopwatch
func (s *StopwatchService) Lap(id int64) (*models.Stopwatch, *models.StopwatchLap, error) {
	stopwatch, err := s.stopwatchRepo.GetByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("stopwatch not found: %w", err)
	}

	if stopwatch.Status != models.StopwatchRunning {
		return nil, nil, fmt.Errorf("stopwatch is not running")
	}

	now := time.Now()

	// Get next lap number
	lastLapNum, err := s.stopwatchLapRepo.GetLastLapNumber(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get lap number: %w", err)
	}
	lapNumber := lastLapNum + 1

	// Calculate lap duration
	lapDuration := stopwatch.GetCurrentElapsed(now) - stopwatch.ElapsedMs

	// Record the lap on the stopwatch model (updates elapsed_ms)
	stopwatch.Lap(now)

	// Create lap record
	totalElapsed := stopwatch.ElapsedMs
	lap := models.NewStopwatchLap(id, lapNumber, lapDuration, totalElapsed, now)

	// Save both
	if err := s.stopwatchRepo.Update(stopwatch); err != nil {
		return nil, nil, fmt.Errorf("failed to update stopwatch: %w", err)
	}

	if err := s.stopwatchLapRepo.Create(lap); err != nil {
		return nil, nil, fmt.Errorf("failed to create lap: %w", err)
	}

	// Notify about the lap
	if s.lapNotifier != nil {
		s.lapNotifier.StopLapNotification(stopwatch.Label, lapNumber, lapDuration)
	}

	return stopwatch, lap, nil
}

// Reset resets a stopwatch to zero
func (s *StopwatchService) Reset(id int64) (*models.Stopwatch, error) {
	stopwatch, err := s.stopwatchRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("stopwatch not found: %w", err)
	}

	now := time.Now()
	stopwatch.Reset(now)

	// Delete all laps for this stopwatch
	if err := s.stopwatchLapRepo.DeleteByStopwatchID(id); err != nil {
		return nil, fmt.Errorf("failed to clear laps: %w", err)
	}

	if err := s.stopwatchRepo.Update(stopwatch); err != nil {
		return nil, fmt.Errorf("failed to reset stopwatch: %w", err)
	}

	return stopwatch, nil
}

// Get retrieves a stopwatch by ID
func (s *StopwatchService) Get(id int64) (*models.Stopwatch, error) {
	return s.stopwatchRepo.GetByID(id)
}

// GetAll retrieves all non-deleted stopwatches
func (s *StopwatchService) GetAll() ([]*models.Stopwatch, error) {
	return s.stopwatchRepo.GetAll()
}

// GetRunning retrieves all running stopwatches
func (s *StopwatchService) GetRunning() ([]*models.Stopwatch, error) {
	return s.stopwatchRepo.GetRunning()
}

// GetLaps retrieves all laps for a stopwatch
func (s *StopwatchService) GetLaps(stopwatchID int64) ([]*models.StopwatchLap, error) {
	return s.stopwatchLapRepo.GetByStopwatchID(stopwatchID)
}

// Delete soft-deletes a stopwatch
func (s *StopwatchService) Delete(id int64) error {
	return s.stopwatchRepo.SoftDelete(id)
}

// AnyRunning returns true if any stopwatch is currently running
func (s *StopwatchService) AnyRunning() (bool, error) {
	running, err := s.stopwatchRepo.GetRunning()
	if err != nil {
		return false, err
	}
	return len(running) > 0, nil
}
