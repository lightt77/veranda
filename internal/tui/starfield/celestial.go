// Package starfield provides celestial object definitions and starfield rendering
package starfield

import "github.com/charmbracelet/lipgloss"

// ObjectType represents the type of celestial object
type ObjectType int

const (
	ObjectTypeStar ObjectType = iota
	ObjectTypePlanet
	ObjectTypeMoon
	ObjectTypeSun
	ObjectTypeGalaxy
)

// CelestialObject represents a celestial body in the starfield
type CelestialObject struct {
	Name        string
	Type        ObjectType
	Symbol      string
	Color       lipgloss.Color
	ApparentMag float64 // Apparent magnitude (brightness)
	// For stars: RA/Dec in degrees (J2000 epoch)
	RightAscension float64 // 0-360 degrees
	Declination    float64 // -90 to +90 degrees
	// For solar system bodies: Astronomy Body value
	Body int // For planets, Sun, Moon (Body values from astronomy package)
	// For moons: parent planet
	ParentPlanet int // Body value of parent planet
	// For galaxies
	IsExtended bool // true for galaxies (fuzzy appearance)
}

// Star catalog - Top 20 brightest stars + named stars
// RA and Dec in degrees (J2000 epoch)
// Apparent magnitude from -1.46 (Sirius) to ~2.0
var StarCatalog = []CelestialObject{
	// Top 20 brightest stars
	{Name: "Sirius", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: -1.46, RightAscension: 101.287, Declination: -16.716},
	{Name: "Canopus", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: -0.74, RightAscension: 95.988, Declination: -52.696},
	{Name: "Arcturus", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#a6e3a1"), ApparentMag: -0.05, RightAscension: 213.915, Declination: 19.183},
	{Name: "Alpha Centauri", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: -0.27, RightAscension: 219.902, Declination: -60.833},
	{Name: "Vega", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 0.03, RightAscension: 279.234, Declination: 38.783},
	{Name: "Capella", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: 0.08, RightAscension: 79.172, Declination: 45.998},
	{Name: "Rigel", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 0.13, RightAscension: 78.634, Declination: -8.202},
	{Name: "Procyon", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: 0.38, RightAscension: 114.825, Declination: 5.225},
	{Name: "Betelgeuse", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f38ba8"), ApparentMag: 0.50, RightAscension: 88.793, Declination: 7.407},
	{Name: "Achernar", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 0.46, RightAscension: 24.429, Declination: -57.237},
	{Name: "Hadar", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: 0.61, RightAscension: 210.956, Declination: -60.373},
	{Name: "Altair", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#cdd6f4"), ApparentMag: 0.77, RightAscension: 297.696, Declination: 8.868},
	{Name: "Acrux", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.40, RightAscension: 186.650, Declination: -63.056},
	{Name: "Aldebaran", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f38ba8"), ApparentMag: 0.87, RightAscension: 68.980, Declination: 16.509},
	{Name: "Antares", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f38ba8"), ApparentMag: 1.06, RightAscension: 247.352, Declination: -26.432},
	{Name: "Spica", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.04, RightAscension: 201.298, Declination: -11.161},
	{Name: "Pollux", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#f9e2af"), ApparentMag: 1.16, RightAscension: 116.329, Declination: 28.026},
	{Name: "Fomalhaut", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.16, RightAscension: 344.413, Declination: -29.622},
	{Name: "Deneb", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.25, RightAscension: 310.358, Declination: 45.280},
	{Name: "Mimosa", Type: ObjectTypeStar, Symbol: "★", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.25, RightAscension: 191.930, Declination: -59.689},

	// Additional notable stars
	{Name: "Regulus", Type: ObjectTypeStar, Symbol: "*", Color: lipgloss.Color("#f9e2af"), ApparentMag: 1.35, RightAscension: 152.093, Declination: 11.967},
	{Name: "Adhara", Type: ObjectTypeStar, Symbol: "*", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.50, RightAscension: 104.656, Declination: -28.972},
	{Name: "Castor", Type: ObjectTypeStar, Symbol: "*", Color: lipgloss.Color("#cdd6f4"), ApparentMag: 1.58, RightAscension: 113.650, Declination: 31.889},
	{Name: "Gacrux", Type: ObjectTypeStar, Symbol: "*", Color: lipgloss.Color("#f38ba8"), ApparentMag: 1.64, RightAscension: 187.791, Declination: -57.113},
	{Name: "Bellatrix", Type: ObjectTypeStar, Symbol: "*", Color: lipgloss.Color("#89b4fa"), ApparentMag: 1.64, RightAscension: 81.283, Declination: 6.350},
}

// Solar system bodies (Body values for astronomy package)
// These are constants used by the astronomy library
const (
	BodySun     = 0
	BodyMoon    = 1
	BodyMercury = 2
	BodyVenus   = 3
	BodyEarth   = 4
	BodyMars    = 5
	BodyJupiter = 6
	BodySaturn  = 7
	BodyUranus  = 8
	BodyNeptune = 9
	BodyPluto   = 10
)

// SolarSystemCatalog - Sun, Moon, and 9 planets
var SolarSystemCatalog = []CelestialObject{
	{Name: "Sun", Type: ObjectTypeSun, Symbol: "☉", Color: lipgloss.Color("#f9e2af"), ApparentMag: -26.74, Body: BodySun},
	{Name: "Moon", Type: ObjectTypeMoon, Symbol: "○", Color: lipgloss.Color("#cdd6f4"), ApparentMag: -12.74, Body: BodyMoon},
	{Name: "Mercury", Type: ObjectTypePlanet, Symbol: "•", Color: lipgloss.Color("#9399b2"), ApparentMag: -2.48, Body: BodyMercury},
	{Name: "Venus", Type: ObjectTypePlanet, Symbol: "✦", Color: lipgloss.Color("#f5e0dc"), ApparentMag: -4.92, Body: BodyVenus},
	{Name: "Mars", Type: ObjectTypePlanet, Symbol: "⬢", Color: lipgloss.Color("#f38ba8"), ApparentMag: -2.94, Body: BodyMars},
	{Name: "Jupiter", Type: ObjectTypePlanet, Symbol: "⬡", Color: lipgloss.Color("#fab387"), ApparentMag: -2.94, Body: BodyJupiter},
	{Name: "Saturn", Type: ObjectTypePlanet, Symbol: "🪐", Color: lipgloss.Color("#f9e2af"), ApparentMag: -0.55, Body: BodySaturn},
	{Name: "Uranus", Type: ObjectTypePlanet, Symbol: "◆", Color: lipgloss.Color("#89dceb"), ApparentMag: 5.38, Body: BodyUranus},
	{Name: "Neptune", Type: ObjectTypePlanet, Symbol: "◆", Color: lipgloss.Color("#89b4fa"), ApparentMag: 7.67, Body: BodyNeptune},
	{Name: "Pluto", Type: ObjectTypePlanet, Symbol: "·", Color: lipgloss.Color("#6c7086"), ApparentMag: 15.1, Body: BodyPluto},
}

// MoonCatalog - Major moons of the solar system
// Note: Positions calculated relative to their parent planets
var MoonCatalog = []CelestialObject{
	// Jupiter's Galilean moons
	{Name: "Io", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#f9e2af"), ApparentMag: 5.0, ParentPlanet: BodyJupiter},
	{Name: "Europa", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#cdd6f4"), ApparentMag: 5.3, ParentPlanet: BodyJupiter},
	{Name: "Ganymede", Type: ObjectTypeMoon, Symbol: "•", Color: lipgloss.Color("#b4befe"), ApparentMag: 4.6, ParentPlanet: BodyJupiter},
	{Name: "Callisto", Type: ObjectTypeMoon, Symbol: "•", Color: lipgloss.Color("#6c7086"), ApparentMag: 5.7, ParentPlanet: BodyJupiter},

	// Saturn's major moons
	{Name: "Titan", Type: ObjectTypeMoon, Symbol: "•", Color: lipgloss.Color("#f9e2af"), ApparentMag: 8.3, ParentPlanet: BodySaturn},
	{Name: "Rhea", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#cdd6f4"), ApparentMag: 9.7, ParentPlanet: BodySaturn},
	{Name: "Iapetus", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#6c7086"), ApparentMag: 11.1, ParentPlanet: BodySaturn},
	{Name: "Dione", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#cdd6f4"), ApparentMag: 10.4, ParentPlanet: BodySaturn},

	// Mars' moons
	{Name: "Phobos", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#6c7086"), ApparentMag: 11.8, ParentPlanet: BodyMars},
	{Name: "Deimos", Type: ObjectTypeMoon, Symbol: "·", Color: lipgloss.Color("#6c7086"), ApparentMag: 12.9, ParentPlanet: BodyMars},
}

// GalaxyCatalog - Bright galaxies visible in amateur telescopes
// RA/Dec in degrees
var GalaxyCatalog = []CelestialObject{
	// Naked eye visible
	{Name: "Andromeda Galaxy", Type: ObjectTypeGalaxy, Symbol: "∴", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 3.44, RightAscension: 10.685, Declination: 41.269, IsExtended: true},
	{Name: "Large Magellanic Cloud", Type: ObjectTypeGalaxy, Symbol: "∴", Color: lipgloss.Color("#f5c0e7"), ApparentMag: 0.9, RightAscension: 80.894, Declination: -69.756, IsExtended: true},
	{Name: "Small Magellanic Cloud", Type: ObjectTypeGalaxy, Symbol: "∴", Color: lipgloss.Color("#f5c0e7"), ApparentMag: 2.7, RightAscension: 16.260, Declination: -72.285, IsExtended: true},

	// Bright telescopic galaxies
	{Name: "Triangulum Galaxy", Type: ObjectTypeGalaxy, Symbol: "∴", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 5.72, RightAscension: 23.462, Declination: 30.660, IsExtended: true},
	{Name: "Centaurus A", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 6.84, RightAscension: 201.365, Declination: -43.019, IsExtended: true},
	{Name: "Bode's Galaxy", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 6.94, RightAscension: 148.888, Declination: 69.065, IsExtended: true},    // M81
	{Name: "Whirlpool Galaxy", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 8.36, RightAscension: 202.469, Declination: 47.195, IsExtended: true}, // M51
	{Name: "Cigar Galaxy", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 8.41, RightAscension: 148.969, Declination: 69.680, IsExtended: true},     // M82
	{Name: "Sombrero Galaxy", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 8.98, RightAscension: 189.997, Declination: -11.623, IsExtended: true}, // M104
	{Name: "Pinwheel Galaxy", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 7.86, RightAscension: 210.802, Declination: 54.349, IsExtended: true},  // M101
}

// Leo Triplet is a group, adding individual galaxies
var LeoTripletCatalog = []CelestialObject{
	{Name: "M65", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 9.3, RightAscension: 169.733, Declination: 13.093, IsExtended: true},
	{Name: "M66", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 8.9, RightAscension: 170.063, Declination: 12.991, IsExtended: true},
	{Name: "NGC 3628", Type: ObjectTypeGalaxy, Symbol: "·", Color: lipgloss.Color("#f5c2e7"), ApparentMag: 9.5, RightAscension: 170.485, Declination: 13.589, IsExtended: true},
}

// GetAllCelestialObjects returns the complete catalog
func GetAllCelestialObjects() []CelestialObject {
	var all []CelestialObject
	all = append(all, StarCatalog...)
	all = append(all, SolarSystemCatalog...)
	all = append(all, MoonCatalog...)
	all = append(all, GalaxyCatalog...)
	all = append(all, LeoTripletCatalog...)
	return all
}
