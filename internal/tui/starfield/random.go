package starfield

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Star represents a twinkling star in the background
type Star struct {
	X          int
	Y          int
	Char       string
	Color      lipgloss.Color
	Brightness int // 0=dark, 1=medium, 2=bright
}

// Star characters by brightness level (single-width for proper alignment)
var starChars = [][]string{
	{"·", "."}, // Dim
	{"*", "+"}, // Medium
	{"★", "✦"}, // Bright
}

// Star colors (Catppuccin Mocha tints)
var starColors = []lipgloss.Color{
	lipgloss.Color("#cba6f7"), // Mauve
	lipgloss.Color("#b4befe"), // Lavender
	lipgloss.Color("#f5c2e7"), // Pink
	lipgloss.Color("#cdd6f4"), // Text (white-blue)
}

// RandomStarfield manages random twinkling stars
type RandomStarfield struct {
	stars          []Star
	width          int
	height         int
	starAreaHeight int
}

// NewRandomStarfield creates a new random starfield
func NewRandomStarfield() *RandomStarfield {
	return &RandomStarfield{
		stars: []Star{},
	}
}

// UpdateChar sets the star's character based on brightness
func (s *Star) UpdateChar() {
	chars := starChars[s.Brightness]
	s.Char = chars[rand.Intn(len(chars))]
}

// Resize updates the starfield dimensions
func (rs *RandomStarfield) Resize(width, height, starAreaHeight int) {
	rs.width = width
	rs.height = height
	rs.starAreaHeight = starAreaHeight

	// Regenerate stars if empty or size changed significantly
	if len(rs.stars) == 0 {
		rs.generateStars()
	}
}

// generateStars creates random stars in the upper area with random count (5-15)
func (rs *RandomStarfield) generateStars() {
	if rs.width == 0 || rs.starAreaHeight <= 0 {
		return
	}

	// Random star count between 5 and 15
	starCount := 5 + rand.Intn(11)
	rs.stars = make([]Star, 0, starCount)

	for len(rs.stars) < starCount {
		// Place stars only in the star area (upper portion)
		x := rand.Intn(rs.width)
		y := rand.Intn(rs.starAreaHeight)

		color := starColors[rand.Intn(len(starColors))]
		brightness := rand.Intn(3)

		star := Star{
			X:          x,
			Y:          y,
			Color:      color,
			Brightness: brightness,
		}
		star.UpdateChar()
		rs.stars = append(rs.stars, star)
	}
}

// Twinkle randomly changes star brightness and handles star lifecycle
func (rs *RandomStarfield) Twinkle() {
	if len(rs.stars) == 0 {
		rs.generateStars()
		return
	}

	// Twinkle existing stars (change brightness)
	twinkleCount := 1 + rand.Intn(2)
	for i := 0; i < twinkleCount; i++ {
		idx := rand.Intn(len(rs.stars))
		change := rand.Intn(3) - 1
		rs.stars[idx].Brightness += change
		if rs.stars[idx].Brightness < 0 {
			rs.stars[idx].Brightness = 0
		}
		if rs.stars[idx].Brightness > 2 {
			rs.stars[idx].Brightness = 2
		}
		rs.stars[idx].UpdateChar()
	}

	// 1% chance: a star disappears and a new one appears elsewhere (swap)
	if rand.Float32() < 0.01 && len(rs.stars) > 0 {
		// Remove a random star
		removeIdx := rand.Intn(len(rs.stars))
		rs.stars = append(rs.stars[:removeIdx], rs.stars[removeIdx+1:]...)

		// Add a new star at a random position
		rs.spawnStar()
	}

	// 1% chance: spawn an extra star (up to 20 max)
	if rand.Float32() < 0.01 && len(rs.stars) < 20 {
		rs.spawnStar()
	}

	// 1% chance: remove a star without replacement (down to 3 min)
	if rand.Float32() < 0.01 && len(rs.stars) > 3 {
		removeIdx := rand.Intn(len(rs.stars))
		rs.stars = append(rs.stars[:removeIdx], rs.stars[removeIdx+1:]...)
	}
}

// spawnStar adds a new star at a random position
func (rs *RandomStarfield) spawnStar() {
	if rs.width == 0 || rs.starAreaHeight <= 0 {
		return
	}

	x := rand.Intn(rs.width)
	y := rand.Intn(rs.starAreaHeight)
	color := starColors[rand.Intn(len(starColors))]
	// New stars start dim and brighten over time
	brightness := rand.Intn(2)

	newStar := Star{
		X:          x,
		Y:          y,
		Color:      color,
		Brightness: brightness,
	}
	newStar.UpdateChar()
	rs.stars = append(rs.stars, newStar)
}

// GetStars returns all stars for rendering
func (rs *RandomStarfield) GetStars() []Star {
	return rs.stars
}

// TwinkleCmd creates a command that triggers star twinkling
func TwinkleCmd() <-chan time.Time {
	return time.Tick(800 * time.Millisecond)
}
