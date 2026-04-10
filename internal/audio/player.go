// Package audio provides ambient sound playback functionality.
package audio

import (
	"fmt"
	"math"
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

// initSpeaker initializes the speaker once (thread-safe)
func initSpeaker() error {
	speakerOnce.Do(func() {
		speakerSampleRate = beep.SampleRate(44100)
		speakerInitErr = speaker.Init(speakerSampleRate, speakerSampleRate.N(time.Second/10))
	})
	return speakerInitErr
}

// AmbientStream represents a single ambient sound stream
type AmbientStream struct {
	filename   string
	volume     float64
	enabled    bool
	ctrl       *beep.Ctrl
	volumeCtrl *effects.Volume
	stopChan   chan struct{}
	stopOnce   sync.Once     // ensures stopChan is closed only once
	readyChan  chan struct{} // signals when volumeCtrl is initialized
	readyOnce  sync.Once     // ensures readyChan is closed only once
	wg         sync.WaitGroup
	isPlaying  bool
	isTest     bool // true if this is a test playback (non-looping)
}

// Player manages multiple parallel ambient sound streams
type Player struct {
	config      *config.AppConfig
	configDir   string
	soundsDir   string
	chimesDir   string
	ambienceDir string
	streams     map[string]*AmbientStream // key is filename
	globalVol   float64                   // global volume multiplier
	isFading    bool
	sampleRate  beep.SampleRate
	mutex       sync.RWMutex
	testStreams map[string]*AmbientStream // separate map for test streams
}

// NewPlayer creates a new audio player
func NewPlayer(cfg *config.AppConfig, configDir string) (*Player, error) {
	// Initialize speaker once (thread-safe)
	if err := initSpeaker(); err != nil {
		return nil, fmt.Errorf("failed to initialize speaker: %w", err)
	}

	player := &Player{
		config:      cfg,
		configDir:   configDir,
		soundsDir:   cfg.SoundsDir,
		chimesDir:   cfg.ChimesDir,
		ambienceDir: cfg.AmbienceDir,
		streams:     make(map[string]*AmbientStream),
		testStreams: make(map[string]*AmbientStream),
		globalVol:   1.0, // Default 100% global volume
		sampleRate:  speakerSampleRate,
	}

	return player, nil
}

// IsPlaying returns whether any ambient audio is currently playing
func (p *Player) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, stream := range p.streams {
		if stream.isPlaying {
			return true
		}
	}
	return false
}

// IsFading returns whether audio is currently fading in or out
func (p *Player) IsFading() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isFading
}

// PlayMultiple starts playing multiple ambient sounds in parallel
// Each sound plays at its configured volume
func (p *Player) PlayMultiple(sounds []config.AmbientSoundConfig) error {
	p.mutex.Lock()

	// Stop any existing streams first
	p.stopAllStreamsLocked()

	// Collect ready channels to wait for initialization
	var readyChans []chan struct{}

	// Start each enabled sound
	enabledCount := 0
	for _, sound := range sounds {
		if !sound.Enabled {
			continue
		}
		enabledCount++

		stream := &AmbientStream{
			filename:  sound.Filename,
			volume:    sound.Volume,
			enabled:   sound.Enabled,
			stopChan:  make(chan struct{}),
			readyChan: make(chan struct{}),
			isTest:    false,
		}

		p.streams[sound.Filename] = stream
		readyChans = append(readyChans, stream.readyChan)
		stream.wg.Add(1)
		go p.playbackLoop(stream)
	}

	// Release lock before waiting for ready channels to avoid deadlock
	// (playFile needs the lock to set volumeCtrl and close readyChan)
	p.mutex.Unlock()

	// Wait for all streams to initialize volume controls before returning
	// This ensures FadeIn will work correctly (with timeout)
	for _, readyChan := range readyChans {
		select {
		case <-readyChan:
			// Stream is ready
		case <-time.After(5 * time.Second):
			// Timeout - continue anyway
		}
	}

	return nil
}

// TestPlay starts playing a single ambient sound for testing (non-looping)
// Returns a stop function to stop the test playback
func (p *Player) TestPlay(filename string, volume float64) (func(), error) {
	// Resolve full path
	fullPath := filepath.Join(p.ambienceDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sound file not found: %s", filename)
	}

	p.mutex.Lock()

	// Stop any existing test for this file
	if existing, ok := p.testStreams[filename]; ok {
		p.stopStreamLocked(existing)
		delete(p.testStreams, filename)
	}

	stream := &AmbientStream{
		filename:  filename,
		volume:    volume,
		enabled:   true,
		stopChan:  make(chan struct{}),
		readyChan: make(chan struct{}), // Initialize ready channel
		isTest:    true,                // Non-looping
	}

	p.testStreams[filename] = stream
	stream.wg.Add(1)
	go p.playbackLoop(stream)
	p.mutex.Unlock()

	// Wait for playback to initialize volume control before fading (with timeout)
	select {
	case <-stream.readyChan:
		// Ready
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for audio initialization")
	}
	time.Sleep(50 * time.Millisecond)

	// Apply fade in
	p.fadeInStream(stream, config.AudioFadeInDuration)

	// Return stop function
	stopFunc := func() {
		p.mutex.Lock()
		stream, ok := p.testStreams[filename]
		if !ok {
			p.mutex.Unlock()
			return
		}
		// Get the stream but release lock before fadeOut (it acquires its own locks)
		p.mutex.Unlock()

		// Fade out without holding the main mutex
		p.fadeOutStream(stream, config.AudioFadeOutDuration)

		p.mutex.Lock()
		if s, stillThere := p.testStreams[filename]; stillThere && s == stream {
			p.stopStreamLocked(s)
			delete(p.testStreams, filename)
		}
		p.mutex.Unlock()
	}

	return stopFunc, nil
}

// StopTestPlay stops a specific test playback
func (p *Player) StopTestPlay(filename string) {
	p.mutex.Lock()
	stream, ok := p.testStreams[filename]
	if !ok {
		p.mutex.Unlock()
		return
	}
	p.mutex.Unlock()

	// Fade out without holding the main mutex
	p.fadeOutStream(stream, config.AudioFadeOutDuration)

	p.mutex.Lock()
	if s, stillThere := p.testStreams[filename]; stillThere && s == stream {
		p.stopStreamLocked(s)
		delete(p.testStreams, filename)
	}
	p.mutex.Unlock()
}

// IsTestPlaying returns whether a specific file is currently being tested
func (p *Player) IsTestPlaying(filename string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if stream, ok := p.testStreams[filename]; ok {
		return stream.isPlaying
	}
	return false
}

// SetTestVolume updates the volume of a playing test stream in real-time
func (p *Player) SetTestVolume(filename string, volume float64) {
	p.mutex.Lock()
	stream, ok := p.testStreams[filename]
	if !ok || stream.volumeCtrl == nil {
		p.mutex.Unlock()
		return
	}
	stream.volume = volume
	targetDB := volumeToDB(volume * p.globalVol)
	stream.volumeCtrl.Volume = targetDB
	stream.volumeCtrl.Silent = volume == 0
	p.mutex.Unlock()
}

// StopAllTestPlays stops all test playbacks
func (p *Player) StopAllTestPlays() {
	p.mutex.Lock()
	for filename, stream := range p.testStreams {
		p.stopStreamLocked(stream)
		delete(p.testStreams, filename)
	}
	p.mutex.Unlock()
}

// StopAllTestPlaysWithFade stops all test playbacks with fade out
func (p *Player) StopAllTestPlaysWithFade() {
	p.mutex.Lock()
	// Get all streams to fade out
	streams := make([]*AmbientStream, 0, len(p.testStreams))
	for _, stream := range p.testStreams {
		streams = append(streams, stream)
	}
	p.mutex.Unlock()

	// Fade out all streams first
	for _, stream := range streams {
		p.fadeOutStream(stream, config.AudioFadeOutDuration)
	}

	// Wait for fade out to complete
	time.Sleep(time.Duration(config.AudioFadeOutDuration) * time.Millisecond)

	// Now stop all streams
	p.mutex.Lock()
	for filename, stream := range p.testStreams {
		p.stopStreamLocked(stream)
		delete(p.testStreams, filename)
	}
	p.mutex.Unlock()
}

// Stop stops all playback with optional fade out
func (p *Player) Stop(fadeOutMs int64) {
	p.mutex.Lock()

	if fadeOutMs > 0 {
		p.isFading = true
		// Fade out all streams
		for _, stream := range p.streams {
			p.fadeOutStream(stream, fadeOutMs)
		}
		for _, stream := range p.testStreams {
			p.fadeOutStream(stream, fadeOutMs)
		}
	}

	p.stopAllStreamsLocked()
	p.isFading = false
	p.mutex.Unlock()
}

// stopAllStreamsLocked stops all streams (must hold lock)
func (p *Player) stopAllStreamsLocked() {
	// Stop all regular streams
	for filename, stream := range p.streams {
		p.stopStreamLocked(stream)
		delete(p.streams, filename)
	}
	// Stop all test streams
	for filename, stream := range p.testStreams {
		p.stopStreamLocked(stream)
		delete(p.testStreams, filename)
	}
}

// stopStreamLocked stops a single stream (must hold lock)
func (p *Player) stopStreamLocked(stream *AmbientStream) {
	if !stream.isPlaying {
		return
	}
	// Use stopOnce to ensure stopChan is only closed once
	stream.stopOnce.Do(func() {
		close(stream.stopChan)
	})
	stream.wg.Wait()
	stream.isPlaying = false
}

// FadeIn gradually increases volume for all streams from 0 to target
func (p *Player) FadeIn(durationMs int64) {
	p.mutex.RLock()
	streams := make([]*AmbientStream, 0, len(p.streams))
	for _, s := range p.streams {
		streams = append(streams, s)
	}
	p.mutex.RUnlock()

	for _, stream := range streams {
		p.fadeInStream(stream, durationMs)
	}
}

// fadeInStream gradually increases volume for a single stream
func (p *Player) fadeInStream(stream *AmbientStream, durationMs int64) {
	if stream.volumeCtrl == nil || durationMs <= 0 {
		return
	}

	targetVolume := stream.volume * p.globalVol

	steps := int(durationMs / 50)
	if steps < 1 {
		steps = 1
	}

	go func() {
		for i := 0; i <= steps; i++ {
			select {
			case <-stream.stopChan:
				return
			default:
				progress := float64(i) / float64(steps)
				currentVol := targetVolume * progress
				db := volumeToDB(currentVol)

				p.mutex.Lock()
				if stream.volumeCtrl != nil {
					stream.volumeCtrl.Volume = db
					stream.volumeCtrl.Silent = currentVol == 0
				}
				p.mutex.Unlock()

				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

// fadeOutStream gradually decreases volume to 0 for a single stream
// This runs asynchronously in a goroutine to avoid blocking the caller
func (p *Player) fadeOutStream(stream *AmbientStream, durationMs int64) {
	if stream.volumeCtrl == nil || durationMs <= 0 {
		return
	}

	startVolume := stream.volume * p.globalVol

	steps := int(durationMs / 50)
	if steps < 1 {
		steps = 1
	}

	go func() {
		for i := steps; i >= 0; i-- {
			select {
			case <-stream.stopChan:
				return
			default:
				progress := float64(i) / float64(steps)
				currentVol := startVolume * progress
				db := volumeToDB(currentVol)

				p.mutex.Lock()
				if stream.volumeCtrl != nil {
					stream.volumeCtrl.Volume = db
					stream.volumeCtrl.Silent = currentVol == 0
				}
				p.mutex.Unlock()

				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

// playbackLoop continuously plays a sound file
// Note: Test playback also loops for continuous testing
func (p *Player) playbackLoop(stream *AmbientStream) {
	defer stream.wg.Done()

	fullPath := filepath.Join(p.ambienceDir, stream.filename)

	for {
		select {
		case <-stream.stopChan:
			return
		default:
			if err := p.playFile(stream, fullPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error playing %s: %v\n", stream.filename, err)
				// Signal ready channel even on error to prevent hanging
				stream.readyOnce.Do(func() {
					close(stream.readyChan)
				})
				// Don't loop on error, exit
				return
			}
			// Loop continuously (both regular and test playback)
		}
	}
}

// playFile plays a single audio file into a stream
func (p *Player) playFile(stream *AmbientStream, filename string) error {
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
		Volume:   -60, // Start silent
		Silent:   true,
	}
	stream.volumeCtrl = volCtrl
	stream.isPlaying = true
	// Signal that volume control is ready for fade operations (only once)
	stream.readyOnce.Do(func() {
		close(stream.readyChan)
	})
	p.mutex.Unlock()

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
	case <-stream.stopChan:
		// Just close the streamer and file - don't call speaker.Clear()
		// as that would stop ALL streams, not just this one
		streamer.Close()
		f.Close()
		return nil
	}
}

// PlayOneShot plays a single MP3 file once asynchronously (non-blocking)
// The filename should be just the filename (not full path), which will be resolved
// relative to the chimes directory. Returns error if file doesn't exist.
func (p *Player) PlayOneShot(filename string, volume float64) error {
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

	// Apply volume
	db := volumeToDB(volume)
	volCtrl := &effects.Volume{
		Streamer: resampled,
		Base:     2,
		Volume:   db,
		Silent:   volume == 0,
	}

	// Play asynchronously - speaker.Play is non-blocking
	speaker.Play(beep.Seq(volCtrl, beep.Callback(func() {
		streamer.Close()
		f.Close()
	})))

	return nil
}

// GetSoundFiles returns all MP3 files in the ambience directory
func (p *Player) GetSoundFiles() ([]string, error) {
	p.mutex.RLock()
	ambienceDir := p.ambienceDir
	p.mutex.RUnlock()

	return p.getSoundFilesFromDir(ambienceDir)
}

// getSoundFilesFromDir returns all MP3 files in the given directory
func (p *Player) getSoundFilesFromDir(ambienceDir string) ([]string, error) {
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
			files = append(files, name)
		}
	}

	return files, nil
}

// volumeToDB converts linear volume (0.0-1.0) to decibels
func volumeToDB(volume float64) float64 {
	if volume <= 0 {
		return -60
	}
	db := 20 * math.Log10(volume)
	if db < -60 {
		return -60
	}
	return db
}
