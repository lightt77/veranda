// Package audio provides ambient sound playback functionality.
package audio

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/lightt77/veranda/internal/config"
)

// Player manages ambient sound playback
type Player struct {
	config      *config.AppConfig
	soundsDir   string
	volume      float64
	isPlaying   bool
	isFading    bool
	currentFile string
	ctrl        *beep.Ctrl
	volumeCtrl  *effects.Volume
	sampleRate  beep.SampleRate
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mutex       sync.RWMutex
}

// NewPlayer creates a new audio player
func NewPlayer(cfg *config.AppConfig) (*Player, error) {
	soundsDir := cfg.SoundsDir

	// Initialize speaker with default sample rate
	sampleRate := beep.SampleRate(44100)
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize speaker: %w", err)
	}

	player := &Player{
		config:     cfg,
		soundsDir:  soundsDir,
		volume:     0.5, // Default 50% volume
		sampleRate: sampleRate,
		stopChan:   make(chan struct{}),
	}

	return player, nil
}

// IsPlaying returns whether audio is currently playing
func (p *Player) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isPlaying
}

// GetVolume returns the current volume (0.0 to 1.0)
func (p *Player) GetVolume() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume
}

// SetVolume sets the volume (0.0 to 1.0)
func (p *Player) SetVolume(vol float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.volume = math.Max(0.0, math.Min(1.0, vol))
	if p.volumeCtrl != nil {
		// Convert linear volume to dB (beep uses dB)
		// 0 volume = -60dB, 1 volume = 0dB
		db := 20 * math.Log10(p.volume)
		if p.volume == 0 {
			db = -60
		}
		p.volumeCtrl.Volume = db
		p.volumeCtrl.Silent = p.volume == 0
	}
}

// Play starts playing ambient sounds with optional shuffle
func (p *Player) Play(shuffle bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isPlaying {
		return nil // Already playing
	}

	files, err := p.getSoundFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no audio files found in %s", p.soundsDir)
	}

	if shuffle {
		rand.Shuffle(len(files), func(i, j int) {
			files[i], files[j] = files[j], files[i]
		})
	}

	p.isPlaying = true
	p.wg.Add(1)
	go p.playbackLoop(files)

	// Give playbackLoop a moment to initialize volumeCtrl
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop stops playback with optional fade out
func (p *Player) Stop(fadeOutMs int64) {
	p.mutex.Lock()
	if !p.isPlaying {
		p.mutex.Unlock()
		return
	}

	p.isFading = true
	p.mutex.Unlock()

	if fadeOutMs > 0 {
		p.fadeOut(fadeOutMs)
	}

	close(p.stopChan)
	p.wg.Wait()

	p.mutex.Lock()
	p.isPlaying = false
	p.isFading = false
	p.stopChan = make(chan struct{})
	p.mutex.Unlock()
}

// FadeIn gradually increases volume from 0 to target
func (p *Player) FadeIn(durationMs int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.volumeCtrl == nil || durationMs <= 0 {
		return
	}

	targetVolume := p.volume
	steps := int(durationMs / 50) // Update every 50ms
	if steps < 1 {
		steps = 1
	}

	go func() {
		for i := 0; i <= steps; i++ {
			select {
			case <-p.stopChan:
				return
			default:
				progress := float64(i) / float64(steps)
				currentVol := targetVolume * progress
				db := 20 * math.Log10(currentVol)
				if currentVol == 0 {
					db = -60
				}

				p.mutex.Lock()
				if p.volumeCtrl != nil {
					p.volumeCtrl.Volume = db
					p.volumeCtrl.Silent = currentVol == 0
				}
				p.mutex.Unlock()

				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

// playbackLoop continuously plays sound files
func (p *Player) playbackLoop(files []string) {
	defer p.wg.Done()

	for {
		for _, file := range files {
			select {
			case <-p.stopChan:
				return
			default:
				if err := p.playFile(file); err != nil {
					// Log error but continue with next file
					fmt.Fprintf(os.Stderr, "Error playing %s: %v\n", file, err)
				}
			}
		}
	}
}

// playFile plays a single audio file
func (p *Player) playFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	// Note: We don't defer f.Close() here because the file needs to stay open
	// until playback completes. The streamer will close it when done.

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to decode MP3: %w", err)
	}
	// Note: We don't defer streamer.Close() here because the streamer needs
	// to stay open until playback completes. The callback will close it.

	// Resample to speaker's sample rate if needed
	resampled := beep.Resample(4, format.SampleRate, p.sampleRate, streamer)

	// Calculate initial volume (use target volume directly)
	db := 20 * math.Log10(p.volume)
	if p.volume == 0 {
		db = -60
	}

	// Create volume control with target volume
	p.mutex.Lock()
	p.volumeCtrl = &effects.Volume{
		Streamer: resampled,
		Base:     2,
		Volume:   db,
		Silent:   p.volume == 0,
	}
	p.currentFile = filename
	p.mutex.Unlock()

	// Create controller for pause/stop
	p.ctrl = &beep.Ctrl{Streamer: p.volumeCtrl}

	// Play with callback when done
	done := make(chan struct{})
	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
		streamer.Close()
		f.Close()
		close(done)
	})))

	// Wait for playback to finish or stop signal
	select {
	case <-done:
		return nil
	case <-p.stopChan:
		speaker.Clear()
		streamer.Close()
		f.Close()
		return nil
	}
}

// fadeOut gradually decreases volume to 0
func (p *Player) fadeOut(durationMs int64) {
	p.mutex.Lock()
	if p.volumeCtrl == nil {
		p.mutex.Unlock()
		return
	}

	startVolume := p.volume
	p.mutex.Unlock()

	steps := int(durationMs / 50)
	if steps < 1 {
		steps = 1
	}

	for i := steps; i >= 0; i-- {
		progress := float64(i) / float64(steps)
		currentVol := startVolume * progress
		db := 20 * math.Log10(currentVol)
		if currentVol == 0 {
			db = -60
		}

		p.mutex.Lock()
		if p.volumeCtrl != nil {
			p.volumeCtrl.Volume = db
			p.volumeCtrl.Silent = currentVol == 0
		}
		p.mutex.Unlock()

		time.Sleep(50 * time.Millisecond)
	}
}

// getSoundFiles returns all MP3 files in the sounds directory
func (p *Player) getSoundFiles() ([]string, error) {
	entries, err := os.ReadDir(p.soundsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("sounds directory does not exist: %s", p.soundsDir)
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".mp3") {
			files = append(files, filepath.Join(p.soundsDir, name))
		}
	}

	return files, nil
}

// GetCurrentFile returns the currently playing file
func (p *Player) GetCurrentFile() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.currentFile
}
