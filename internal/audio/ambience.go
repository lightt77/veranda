package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/service"
)

// AmbienceService manages ambience sound based on timer/stopwatch state
type AmbienceService struct {
	player           *Player
	timerService     *service.TimerService
	stopwatchService *service.StopwatchService
	config           *config.AppConfig
	configDir        string
	isAutoPlaying    bool
	mutex            sync.RWMutex
	stopMonitorChan  chan struct{}
}

// NewAmbienceService creates a new ambience sound service
func NewAmbienceService(
	cfg *config.AppConfig,
	configDir string,
	timerService *service.TimerService,
	stopwatchService *service.StopwatchService,
) (*AmbienceService, error) {
	player, err := NewPlayer(cfg, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio player: %w", err)
	}

	return &AmbienceService{
		player:           player,
		timerService:     timerService,
		stopwatchService: stopwatchService,
		config:           cfg,
		configDir:        configDir,
		stopMonitorChan:  make(chan struct{}),
	}, nil
}

// Start starts the ambience sound service with auto-play monitoring
func (s *AmbienceService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Start monitoring for active timers/stopwatches
	go s.monitorActivity()

	return nil
}

// Stop stops the ambience sound service
func (s *AmbienceService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	close(s.stopMonitorChan)
	s.player.Stop(config.AudioFadeOutDuration)
}

// Play starts playing ambience sounds immediately based on config
func (s *AmbienceService) Play() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load user config to get ambience sound settings
	userConfig := config.LoadUserConfig(s.configDir)

	// Play all enabled ambience sounds
	if err := s.player.PlayMultiple(userConfig.AmbienceSounds); err != nil {
		return err
	}

	// Mark as NOT auto-played (manual control)
	s.isAutoPlaying = false

	// Fade in
	s.player.FadeIn(config.AudioFadeInDuration)
	return nil
}

// StopPlayback stops ambience sound playback
func (s *AmbienceService) StopPlayback() {
	s.player.Stop(config.AudioFadeOutDuration)
}

// IsPlaying returns whether ambience sound is playing
func (s *AmbienceService) IsPlaying() bool {
	return s.player.IsPlaying()
}

// monitorActivity monitors timers and stopwatches to auto-start/stop ambience sound
func (s *AmbienceService) monitorActivity() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopMonitorChan:
			return
		case <-ticker.C:
			s.checkAndUpdatePlayback()
		}
	}
}

// checkAndUpdatePlayback checks if any timer/stopwatch is running and updates playback
func (s *AmbienceService) checkAndUpdatePlayback() {
	// Load user config first to check if ambience is enabled globally
	userConfig := config.LoadUserConfig(s.configDir)

	// Check if any timer is about to end (within fade out duration)
	timersAboutToEnd, err := s.timerService.GetRunningTimersAboutToEnd(config.AudioFadeOutDuration)
	if err != nil {
		fmt.Printf("Error checking timers about to end: %v\n", err)
		return
	}

	// Check if any timer is running
	timersRunning, err := s.timerService.AnyRunning()
	if err != nil {
		fmt.Printf("Error checking timers: %v\n", err)
		return
	}

	// Check if any stopwatch is running
	stopwatchesRunning, err := s.stopwatchService.AnyRunning()
	if err != nil {
		fmt.Printf("Error checking stopwatches: %v\n", err)
		return
	}

	anyActive := timersRunning || stopwatchesRunning

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If ambience is disabled globally, stop any auto-playing audio
	if !userConfig.AmbienceEnabled && s.player.IsPlaying() && s.isAutoPlaying {
		s.player.Stop(config.AudioFadeOutDuration)
		s.isAutoPlaying = false
		return
	}

	// If a timer is about to end and we're auto-playing, stop playback
	if len(timersAboutToEnd) > 0 && s.player.IsPlaying() && s.isAutoPlaying {
		s.player.Stop(config.AudioFadeOutDuration)
		s.isAutoPlaying = false
		return
	}

	// Only auto-start if ambience is enabled globally
	if userConfig.AmbienceEnabled && anyActive && !s.player.IsPlaying() {
		// Start ambience sound with all enabled sounds
		if err := s.player.PlayMultiple(userConfig.AmbienceSounds); err != nil {
			return
		}
		s.player.FadeIn(config.AudioFadeInDuration)
		s.isAutoPlaying = true
	} else if !anyActive && s.player.IsPlaying() && s.isAutoPlaying {
		// Stop if no longer active (and we were auto-playing)
		s.player.Stop(config.AudioFadeOutDuration)
		s.isAutoPlaying = false
	}
}

// GetPlayer returns the underlying audio player for use by other services
func (s *AmbienceService) GetPlayer() *Player {
	return s.player
}
