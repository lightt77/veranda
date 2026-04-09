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

var (
	speakerOnce       sync.Once
	speakerInitErr    error
	speakerSampleRate beep.SampleRate
)

// Player manages ambient sound playback
type Player struct {
	config      *config.AppConfig
	soundsDir   string
	chimesDir   string
	ambienceDir string
	volume      float64
	isPlaying   bool
	isFading    bool
	isFadeStop  bool // true when stopping after a fade-out
	currentFile string
	ctrl        *beep.Ctrl
	volumeCtrl  *effects.Volume
	sampleRate  beep.SampleRate
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mutex       sync.RWMutex
}

// initSpeaker initializes the speaker once (thread-safe)
func initSpeaker() error {
	speakerOnce.Do(func() {
		speakerSampleRate = beep.SampleRate(44100)
		speakerInitErr = speaker.Init(speakerSampleRate, speakerSampleRate.N(time.Second/10))
	})
	return speakerInitErr
}

// NewPlayer creates a new audio player
func NewPlayer(cfg *config.AppConfig) (*Player, error) {
	// Initialize speaker once (thread-safe)
	if err := initSpeaker(); err != nil {
		return nil, fmt.Errorf("failed to initialize speaker: %w", err)
	}

	player := &Player{
		config:      cfg,
		soundsDir:   cfg.SoundsDir,
		chimesDir:   cfg.ChimesDir,
		ambienceDir: cfg.AmbienceDir,
		volume:      0.5, // Default 50% volume
		sampleRate:  speakerSampleRate,
		stopChan:    make(chan struct{}),
	}

	return player, nil
}

// IsPlaying returns whether audio is currently playing
func (p *Player) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isPlaying
}

// IsFading returns whether audio is currently fading in or out
func (p *Player) IsFading() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isFading
}

// StartFadeOut starts fading out the volume and stops playback after fade completes
func (p *Player) StartFadeOut(durationMs int64) {
	p.mutex.Lock()
	if !p.isPlaying || p.isFading {
		p.mutex.Unlock()
		return
	}
	p.isFading = true
	p.isFadeStop = true
	p.mutex.Unlock()

	go func() {
		p.fadeOut(durationMs)
		// After fade out completes, signal stop (volume is already 0)
		p.stopAfterFade()
		p.mutex.Lock()
		p.isFading = false
		p.isFadeStop = false
		p.mutex.Unlock()
	}()
}

// stopAfterFade signals playback to stop after fade-out (volume already at 0)
func (p *Player) stopAfterFade() {
	p.mutex.Lock()
	if !p.isPlaying {
		p.mutex.Unlock()
		return
	}
	p.mutex.Unlock()

	close(p.stopChan)
	p.wg.Wait()

	p.mutex.Lock()
	p.isPlaying = false
	p.stopChan = make(chan struct{})
	p.mutex.Unlock()
}

// stopImmediately stops playback without fade out
func (p *Player) stopImmediately() {
	p.mutex.Lock()
	if !p.isPlaying {
		p.mutex.Unlock()
		return
	}
	p.isFadeStop = false
	p.mutex.Unlock()

	close(p.stopChan)
	p.wg.Wait()

	p.mutex.Lock()
	p.isPlaying = false
	p.stopChan = make(chan struct{})
	p.mutex.Unlock()
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
	p.isFadeStop = fadeOutMs > 0
	p.mutex.Unlock()

	if fadeOutMs > 0 {
		p.fadeOut(fadeOutMs)
	}

	close(p.stopChan)
	p.wg.Wait()

	p.mutex.Lock()
	p.isPlaying = false
	p.isFading = false
	p.isFadeStop = false
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

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to decode MP3: %w", err)
	}

	// Resample to speaker's sample rate if needed
	resampled := beep.Resample(4, format.SampleRate, p.sampleRate, streamer)

	// Create volume control wrapper
	p.mutex.Lock()
	volCtrl := &effects.Volume{
		Streamer: resampled,
		Base:     2,
		Volume:   0, // Start silent for fade-in
		Silent:   true,
	}
	p.volumeCtrl = volCtrl
	targetVol := p.volume
	p.currentFile = filename
	p.mutex.Unlock()

	// Set initial volume
	if targetVol > 0 {
		db := 20 * math.Log10(targetVol)
		volCtrl.Volume = db
		volCtrl.Silent = false
	}

	// Play with volume control
	done := make(chan struct{})
	speaker.Play(beep.Seq(volCtrl, beep.Callback(func() {
		streamer.Close()
		f.Close()
		close(done)
	})))

	// Wait for playback to finish or stop signal
	select {
	case <-done:
		return nil
	case <-p.stopChan:
		p.mutex.RLock()
		isFadeStop := p.isFadeStop
		p.mutex.RUnlock()

		if !isFadeStop {
			// Hard stop - clear speaker immediately
			speaker.Clear()
		} else {
			// Fade stop - volume is already 0, just signal and wait for natural end
			// Give a small delay for the last buffered audio to play at 0 volume
			time.Sleep(100 * time.Millisecond)
		}
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

// getSoundFiles returns all MP3 files in the ambience directory
func (p *Player) getSoundFiles() ([]string, error) {
	p.mutex.RLock()
	ambienceDir := p.ambienceDir
	p.mutex.RUnlock()

	entries, err := os.ReadDir(ambienceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ambience directory does not exist: %s", ambienceDir)
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
			files = append(files, filepath.Join(ambienceDir, name))
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

// PlayOneShot plays a single MP3 file once asynchronously (non-blocking)
// The filename should be just the filename (not full path), which will be resolved
// relative to the chimes directory. Returns error if file doesn't exist.
func (p *Player) PlayOneShot(filename string) error {
	p.mutex.RLock()
	chimesDir := p.chimesDir
	p.mutex.RUnlock()

	// Resolve full path in chimes directory
	fullPath := filepath.Join(chimesDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("sound file not found in chimes directory: %s", filename)
	}

	// Open the file
	f, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open sound file: %w", err)
	}

	// Decode MP3
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to decode MP3: %w", err)
	}

	// Resample to speaker's sample rate if needed
	resampled := beep.Resample(4, format.SampleRate, p.sampleRate, streamer)

	// Play asynchronously - speaker.Play is non-blocking
	// Use a callback to close resources when done
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		streamer.Close()
		f.Close()
	})))

	return nil
}
