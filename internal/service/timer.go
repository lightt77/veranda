package service

import (
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/models"
	"github.com/lightt77/veranda/internal/repository"
)

// TimerService handles timer business logic
type TimerService struct {
	timerRepo *repository.TimerRepository
}

// NewTimerService creates a new timer service
func NewTimerService(timerRepo *repository.TimerRepository) *TimerService {
	return &TimerService{
		timerRepo: timerRepo,
	}
}

// Create creates a new timer with the given duration and label
func (s *TimerService) Create(label string, durationMs int64) (*models.Timer, error) {
	if durationMs <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}

	now := time.Now()
	timer := models.NewTimer(GenerateID(), label, durationMs, now)

	if err := s.timerRepo.Create(timer); err != nil {
		return nil, fmt.Errorf("failed to create timer: %w", err)
	}

	return timer, nil
}

// CreateQuick creates and starts a timer immediately (quick syntax)
func (s *TimerService) CreateQuick(label string, durationMs int64) (*models.Timer, error) {
	timer, err := s.Create(label, durationMs)
	if err != nil {
		return nil, err
	}

	return s.Start(timer.ID)
}

// Start starts a timer
func (s *TimerService) Start(id int64) (*models.Timer, error) {
	timer, err := s.timerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("timer not found: %w", err)
	}

	if timer.Status == models.TimerRunning {
		return nil, fmt.Errorf("timer is already running")
	}

	if timer.Status == models.TimerCompleted {
		return nil, fmt.Errorf("timer is already completed")
	}

	now := time.Now()
	if timer.Status == models.TimerPaused {
		timer.Resume(now)
	} else {
		timer.Start(now)
	}

	if err := s.timerRepo.Update(timer); err != nil {
		return nil, fmt.Errorf("failed to start timer: %w", err)
	}

	return timer, nil
}

// Pause pauses a running timer
func (s *TimerService) Pause(id int64) (*models.Timer, error) {
	timer, err := s.timerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("timer not found: %w", err)
	}

	if timer.Status != models.TimerRunning {
		return nil, fmt.Errorf("timer is not running")
	}

	now := time.Now()
	timer.Pause(now)

	if err := s.timerRepo.Update(timer); err != nil {
		return nil, fmt.Errorf("failed to pause timer: %w", err)
	}

	return timer, nil
}

// Stop stops a timer (alias for pause)
func (s *TimerService) Stop(id int64) (*models.Timer, error) {
	return s.Pause(id)
}

// Complete marks a timer as completed
func (s *TimerService) Complete(id int64) (*models.Timer, error) {
	timer, err := s.timerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("timer not found: %w", err)
	}

	now := time.Now()
	timer.Complete(now)

	if err := s.timerRepo.Update(timer); err != nil {
		return nil, fmt.Errorf("failed to complete timer: %w", err)
	}

	return timer, nil
}

// CheckExpired checks if a running timer has expired and marks it complete
func (s *TimerService) CheckExpired(id int64) (*models.Timer, bool, error) {
	timer, err := s.timerRepo.GetByID(id)
	if err != nil {
		return nil, false, fmt.Errorf("timer not found: %w", err)
	}

	if timer.Status != models.TimerRunning {
		return timer, false, nil
	}

	now := time.Now()
	if timer.IsExpired(now) {
		timer.Complete(now)
		if err := s.timerRepo.Update(timer); err != nil {
			return nil, false, fmt.Errorf("failed to complete expired timer: %w", err)
		}
		return timer, true, nil
	}

	return timer, false, nil
}

// Get retrieves a timer by ID
func (s *TimerService) Get(id int64) (*models.Timer, error) {
	return s.timerRepo.GetByID(id)
}

// GetAll retrieves all non-deleted timers
func (s *TimerService) GetAll() ([]*models.Timer, error) {
	return s.timerRepo.GetAll()
}

// GetRunning retrieves all running timers
func (s *TimerService) GetRunning() ([]*models.Timer, error) {
	return s.timerRepo.GetRunning()
}

// Delete soft-deletes a timer
func (s *TimerService) Delete(id int64) error {
	return s.timerRepo.SoftDelete(id)
}

// AnyRunning returns true if any timer is currently running
func (s *TimerService) AnyRunning() (bool, error) {
	running, err := s.timerRepo.GetRunning()
	if err != nil {
		return false, err
	}
	return len(running) > 0, nil
}
