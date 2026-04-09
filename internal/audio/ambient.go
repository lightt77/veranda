package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/service"
)

// AmbientService manages ambient sound based on timer/stopwatch state
type AmbientService struct {
	player           *Player
	timerService     *service.TimerService
	stopwatchService *service.StopwatchService
	config           *config.AppConfig
	configDir        string
	isAutoPlaying    bool
	mutex            sync.RWMutex
	stopMonitorChan  chan struct{}
}

// NewAmbientService creates a new ambient sound service
func NewAmbientService(
	cfg *config.AppConfig,
	configDir string,
	timerService *service.TimerService,
	stopwatchService *service.StopwatchService,
) (*AmbientService, error) {
	player, err := NewPlayer(cfg, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio player: %w", err)
	}

	return &AmbientService{
		player:           player,
		timerService:     timerService,
		stopwatchService: stopwatchService,
		config:           cfg,
		configDir:        configDir,
		stopMonitorChan:  make(chan struct{}),
	}, nil
}

// Start starts the ambient sound service with auto-play monitoring
func (s *AmbientService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Start monitoring for active timers/stopwatches
	go s.monitorActivity()

	return nil
}

// Stop stops the ambient sound service
func (s *AmbientService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	close(s.stopMonitorChan)
	s.player.Stop(config.AudioFadeOutDuration)
}

// Play starts playing ambient sounds immediately based on config
func (s *AmbientService) Play() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load user config to get ambient sound settings
	userConfig := config.LoadUserConfig(s.configDir)

	// Play all enabled ambient sounds
	if err := s.player.PlayMultiple(userConfig.AmbientSounds); err != nil {
		return err
	}

	// Mark as NOT auto-played (manual control)
	s.isAutoPlaying = false

	// Fade in
	s.player.FadeIn(config.AudioFadeInDuration)
	fmt.Println("Ambient sound started (manual)")
	return nil
}

// StopPlayback stops ambient sound playback
func (s *AmbientService) StopPlayback() {
	s.player.Stop(config.AudioFadeOutDuration)
}

// IsPlaying returns whether ambient sound is playing
func (s *AmbientService) IsPlaying() bool {
	return s.player.IsPlaying()
}

// monitorActivity monitors timers and stopwatches to auto-start/stop ambient sound
func (s *AmbientService) monitorActivity() {
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
func (s *AmbientService) checkAndUpdatePlayback() {
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

	// If a timer is about to end and we're auto-playing, stop playback
	if len(timersAboutToEnd) > 0 && s.player.IsPlaying() && s.isAutoPlaying {
		s.player.Stop(config.AudioFadeOutDuration)
		s.isAutoPlaying = false
		fmt.Println("Ambient sound stopping (timer about to end)")
		return
	}

	if anyActive && !s.player.IsPlaying() {
		// Load user config to get current ambient sound settings
		userConfig := config.LoadUserConfig(s.configDir)

		// Start ambient sound with all enabled sounds
		if err := s.player.PlayMultiple(userConfig.AmbientSounds); err != nil {
			fmt.Printf("Error starting ambient sound: %v\n", err)
			return
		}
		s.player.FadeIn(config.AudioFadeInDuration)
		s.isAutoPlaying = true
		fmt.Println("Ambient sound started (timer/stopwatch active)")
	} else if !anyActive && s.player.IsPlaying() && s.isAutoPlaying {
		// Stop if no longer active (and we were auto-playing)
		s.player.Stop(config.AudioFadeOutDuration)
		s.isAutoPlaying = false
		fmt.Println("Ambient sound stopped (no active timers)")
	}
}

// GetPlayer returns the underlying audio player for use by other services
func (s *AmbientService) GetPlayer() *Player {
	return s.player
}
