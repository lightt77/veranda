package starfield

import (
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// VisibleObject represents a celestial object visible in the sky
type VisibleObject struct {
	Object     CelestialObject
	Azimuth    float64 // 0-360 degrees
	Altitude   float64 // -90 to 90 degrees
	ScreenX    int
	ScreenY    int
	Brightness float64 // 0-1 based on altitude and apparent magnitude
	IsVisible  bool
}

// RealisticStarfield renders the actual night sky based on observer location and time
type RealisticStarfield struct {
	calculator     AstronomicalCalculator
	observer       Observer
	visibleObjects []VisibleObject
	lastUpdate     time.Time
	width          int
	height         int
	starAreaHeight int
}

// NewRealisticStarfield creates a new realistic starfield
func NewRealisticStarfield(observer Observer) *RealisticStarfield {
	return &RealisticStarfield{
		calculator:     NewMeeusCalculator(),
		observer:       observer,
		visibleObjects: []VisibleObject{},
		lastUpdate:     time.Time{},
	}
}

// SetObserver updates the observer location
func (rs *RealisticStarfield) SetObserver(observer Observer) {
	rs.observer = observer
	rs.lastUpdate = time.Time{} // Force recalculation
}

// Resize updates the starfield dimensions
func (rs *RealisticStarfield) Resize(width, height, starAreaHeight int) {
	rs.width = width
	rs.height = height
	rs.starAreaHeight = starAreaHeight
	rs.lastUpdate = time.Time{} // Force recalculation
}

// Update recalculates all object positions if needed
// Returns true if positions were updated
func (rs *RealisticStarfield) Update(t time.Time) bool {
	// Update every 120 seconds (2 minutes) as specified
	if !rs.lastUpdate.IsZero() && t.Sub(rs.lastUpdate) < 120*time.Second {
		return false
	}

	rs.lastUpdate = t
	rs.calculatePositions(t)
	return true
}

// calculatePositions computes the positions of all celestial objects
func (rs *RealisticStarfield) calculatePositions(t time.Time) {
	rs.visibleObjects = []VisibleObject{}

	if rs.width == 0 || rs.starAreaHeight <= 0 {
		return
	}

	// Calculate positions for all catalog objects
	catalog := GetAllCelestialObjects()

	for _, obj := range catalog {
		visible := rs.calculateObjectPosition(obj, t)
		if visible.IsVisible {
			rs.visibleObjects = append(rs.visibleObjects, visible)
		}
	}

	// Sort by brightness (brighter objects last, so they render on top)
	sort.Slice(rs.visibleObjects, func(i, j int) bool {
		return rs.visibleObjects[i].Brightness < rs.visibleObjects[j].Brightness
	})
}

// calculateObjectPosition calculates the position of a single celestial object
func (rs *RealisticStarfield) calculateObjectPosition(obj CelestialObject, t time.Time) VisibleObject {
	visible := VisibleObject{
		Object:    obj,
		IsVisible: false,
	}

	var coords HorizontalCoords
	var err error

	switch obj.Type {
	case ObjectTypeStar, ObjectTypeGalaxy:
		// Stars and galaxies have fixed RA/Dec
		coords, err = rs.calculator.CalculateStarPosition(obj.RightAscension, obj.Declination, rs.observer, t)

	case ObjectTypeSun:
		// Sun is special - only visible during day
		coords, err = rs.calculator.CalculateBodyPosition(BodySun, rs.observer, t)

	case ObjectTypeMoon:
		if obj.Name == "Moon" {
			// Earth's Moon
			coords, err = rs.calculator.CalculateMoonPosition(rs.observer, t)
		} else {
			// Moons of planets - calculate relative to parent
			// This is a simplified visual representation
			parentCoords, parentErr := rs.calculator.CalculateBodyPosition(obj.ParentPlanet, rs.observer, t)
			if parentErr != nil || !parentCoords.IsVisible() {
				return visible
			}

			// Calculate moon position relative to parent
			moonAz, moonAlt := CalculateMoonPositionRelative(
				parentCoords.Azimuth,
				parentCoords.Altitude,
				getMoonIndex(obj),
				t,
			)

			coords = HorizontalCoords{
				Azimuth:  moonAz,
				Altitude: moonAlt,
			}
		}

	case ObjectTypePlanet:
		// Planets
		coords, err = rs.calculator.CalculateBodyPosition(obj.Body, rs.observer, t)
	}

	if err != nil {
		return visible
	}

	// Check if object is visible (above horizon with margin)
	if !coords.IsVisible() {
		return visible
	}

	// Calculate brightness based on altitude and apparent magnitude
	altitudeBrightness := coords.GetBrightness()
	magnitudeBrightness := calculateMagnitudeBrightness(obj.ApparentMag)

	visible.Azimuth = coords.Azimuth
	visible.Altitude = coords.Altitude
	visible.Brightness = altitudeBrightness * magnitudeBrightness
	visible.IsVisible = true

	// Map to screen coordinates
	visible.ScreenX, visible.ScreenY = rs.mapToScreen(coords.Azimuth, coords.Altitude)

	return visible
}

// mapToScreen converts azimuth/altitude to screen coordinates
func (rs *RealisticStarfield) mapToScreen(azimuth, altitude float64) (x, y int) {
	// Map azimuth (0-360) to X (0 to width)
	// Azimuth: 0 = North, 90 = East, 180 = South, 270 = West
	x = int((azimuth / 360.0) * float64(rs.width))
	if x >= rs.width {
		x = rs.width - 1
	}
	if x < 0 {
		x = 0
	}

	// Map altitude (0-90) to Y (bottom to top of star area)
	// Altitude: 0 = horizon, 90 = zenith (directly overhead)
	// We use 90% of the star area height to leave some space
	altitudeFactor := altitude / 90.0
	y = rs.starAreaHeight - int(altitudeFactor*float64(rs.starAreaHeight)*0.9) - 1

	if y < 0 {
		y = 0
	}
	if y >= rs.starAreaHeight {
		y = rs.starAreaHeight - 1
	}

	return x, y
}

// calculateMagnitudeBrightness converts apparent magnitude to brightness factor (0-1)
// Lower magnitude = brighter object
func calculateMagnitudeBrightness(magnitude float64) float64 {
	// Brightness scale based on magnitude
	// mag < -10: extremely bright (Sun)
	// mag < -5: very bright (Moon)
	// mag < 0: bright (Sirius, Canopus, planets)
	// mag < 3: visible (bright stars)
	// mag < 6: visible to naked eye
	// mag > 6: telescope only

	switch {
	case magnitude < -10:
		return 1.0
	case magnitude < -5:
		return 0.95
	case magnitude < 0:
		return 0.9
	case magnitude < 2:
		return 0.8
	case magnitude < 4:
		return 0.6
	case magnitude < 6:
		return 0.4
	case magnitude < 8:
		return 0.25
	case magnitude < 10:
		return 0.15
	case magnitude < 12:
		return 0.1
	default:
		return 0.05
	}
}

// GetVisibleObjects returns all visible objects for rendering
func (rs *RealisticStarfield) GetVisibleObjects() []VisibleObject {
	return rs.visibleObjects
}

// GetObserver returns the current observer
func (rs *RealisticStarfield) GetObserver() Observer {
	return rs.observer
}

// GetLastUpdate returns when positions were last calculated
func (rs *RealisticStarfield) GetLastUpdate() time.Time {
	return rs.lastUpdate
}

// GetObjectSymbol returns the appropriate symbol based on object type and brightness
func GetObjectSymbol(obj CelestialObject, brightness float64) string {
	// For stars, use different symbols based on brightness
	if obj.Type == ObjectTypeStar {
		if brightness > 0.8 {
			return "★"
		} else if brightness > 0.5 {
			return "✦"
		} else if brightness > 0.3 {
			return "*"
		} else {
			return "·"
		}
	}

	// For other objects, use the configured symbol
	return obj.Symbol
}

// GetObjectColor returns the color with brightness adjustment
func GetObjectColor(obj CelestialObject, brightness float64) lipgloss.Color {
	// For now, return the configured color
	// In the future, could adjust brightness of the color
	return obj.Color
}

// getMoonIndex returns an index for the moon to create different orbital periods
func getMoonIndex(obj CelestialObject) int {
	// Map moon names to indices
	moonIndices := map[string]int{
		"Io":       0,
		"Europa":   1,
		"Ganymede": 2,
		"Callisto": 3,
		"Titan":    4,
		"Rhea":     5,
		"Iapetus":  6,
		"Dione":    7,
		"Phobos":   8,
		"Deimos":   9,
	}

	if idx, ok := moonIndices[obj.Name]; ok {
		return idx
	}
	return 0
}

// MumbaiObserver returns an observer at Mumbai (default location)
func MumbaiObserver() Observer {
	return Observer{
		Latitude:  19.0760,
		Longitude: 72.8777,
		Timezone:  "Asia/Kolkata",
	}
}

// ObserverFromCity creates an observer from a city
func ObserverFromCity(city struct {
	Latitude  float64
	Longitude float64
	Timezone  string
}) Observer {
	return Observer{
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
		Timezone:  city.Timezone,
	}
}
