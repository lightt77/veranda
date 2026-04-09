package starfield

import (
	"math"
	"time"

	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/moonposition"
	"github.com/soniakeys/meeus/v3/solar"
)

// Observer represents an observation location
type Observer struct {
	Latitude  float64 // -90 to 90 degrees
	Longitude float64 // -180 to 180 degrees (positive = East)
	Timezone  string
}

// HorizontalCoords represents position in the sky (altitude/azimuth)
type HorizontalCoords struct {
	Altitude float64 // -90 to 90 degrees (0 = horizon, 90 = zenith)
	Azimuth  float64 // 0 to 360 degrees (0 = North, 90 = East)
}

// AstronomicalCalculator provides celestial position calculations
// This interface allows for pluggable implementations
type AstronomicalCalculator interface {
	// CalculateStarPosition returns horizontal coordinates for a star at given time
	CalculateStarPosition(ra, dec float64, observer Observer, t time.Time) (HorizontalCoords, error)

	// CalculateBodyPosition returns horizontal coordinates for a solar system body
	CalculateBodyPosition(body int, observer Observer, t time.Time) (HorizontalCoords, error)

	// CalculateMoonPosition returns horizontal coordinates for Earth's moon
	CalculateMoonPosition(observer Observer, t time.Time) (HorizontalCoords, error)

	// CalculateMoonPhase returns moon phase (0-1) and illumination percentage
	CalculateMoonPhase(t time.Time) (phase, illumination float64)

	// GetBodyName returns the name of a solar system body
	GetBodyName(body int) string
}

// MeeusCalculator implements AstronomicalCalculator using meeus package
type MeeusCalculator struct{}

// NewMeeusCalculator creates a new astronomy calculator
func NewMeeusCalculator() *MeeusCalculator {
	return &MeeusCalculator{}
}

// toJulianDay converts time.Time to Julian Day
func toJulianDay(t time.Time) float64 {
	return julian.TimeToJD(t)
}

// calculateLocalSiderealTime calculates the local sidereal time in radians
func calculateLocalSiderealTime(jd, longitude float64) float64 {
	// Julian centuries since J2000.0
	T := (jd - 2451545.0) / 36525.0

	// Greenwich Mean Sidereal Time at 0h UT (in degrees)
	gmst := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*T*T - T*T*T/38710000.0

	// Normalize to 0-360 degrees
	gmst = math.Mod(gmst, 360.0)
	if gmst < 0 {
		gmst += 360.0
	}

	// Add longitude for local sidereal time
	lst := gmst + longitude
	lst = math.Mod(lst, 360.0)
	if lst < 0 {
		lst += 360.0
	}

	// Convert to radians
	return lst * math.Pi / 180.0
}

// CalculateStarPosition calculates horizontal coordinates for a star
func (c *MeeusCalculator) CalculateStarPosition(ra, dec float64, observer Observer, t time.Time) (HorizontalCoords, error) {
	// Convert to Julian Day
	jd := toJulianDay(t)

	// Convert observer latitude to radians
	latRad := observer.Latitude * math.Pi / 180.0

	// Calculate local sidereal time
	lstRad := calculateLocalSiderealTime(jd, observer.Longitude)

	// Convert RA from degrees to radians
	raRad := ra * math.Pi / 180.0
	decRad := dec * math.Pi / 180.0

	// Calculate hour angle in radians
	haRad := lstRad - raRad

	// Convert to horizontal coordinates using the formula
	// sin(alt) = sin(dec) * sin(lat) + cos(dec) * cos(lat) * cos(ha)
	sinAlt := math.Sin(decRad)*math.Sin(latRad) + math.Cos(decRad)*math.Cos(latRad)*math.Cos(haRad)
	altRad := math.Asin(sinAlt)

	// cos(az) = (sin(dec) - sin(alt) * sin(lat)) / (cos(alt) * cos(lat))
	cosAz := (math.Sin(decRad) - sinAlt*math.Sin(latRad)) / (math.Cos(altRad) * math.Cos(latRad))
	cosAz = math.Max(-1, math.Min(1, cosAz)) // Clamp to [-1, 1]
	azRad := math.Acos(cosAz)

	// Determine correct quadrant for azimuth
	// If sin(ha) > 0, azimuth = 360 - azimuth (azimuth measured westward from south)
	if math.Sin(haRad) > 0 {
		azRad = 2*math.Pi - azRad
	}

	// Convert to degrees
	altitude := altRad * 180.0 / math.Pi
	azimuth := azRad * 180.0 / math.Pi

	// Convert azimuth to our convention (0 = North, 90 = East)
	// Currently 0 = South, so add 180
	azimuth = math.Mod(azimuth+180.0, 360.0)

	return HorizontalCoords{
		Altitude: altitude,
		Azimuth:  azimuth,
	}, nil
}

// CalculateBodyPosition calculates horizontal coordinates for a solar system body
// Uses simplified orbital elements for planets
func (c *MeeusCalculator) CalculateBodyPosition(body int, observer Observer, t time.Time) (HorizontalCoords, error) {
	jd := toJulianDay(t)

	var ra, dec float64

	switch body {
	case BodySun:
		// Calculate sun's apparent position
		sunLon := float64(solar.ApparentLongitude(jd).Deg())

		// Convert ecliptic longitude to equatorial (simplified, assuming ecliptic latitude is 0)
		epsilon := meanObliquity(jd)
		ra, dec = eclipticToEquatorial(sunLon, 0, epsilon)

	case BodyMoon:
		// Use moon position calculation
		return c.CalculateMoonPosition(observer, t)

	case BodyMercury, BodyVenus, BodyMars, BodyJupiter, BodySaturn, BodyUranus, BodyNeptune, BodyPluto:
		// Use simplified planet positions based on orbital elements
		ra, dec = c.simplifiedPlanetPosition(body, jd)

	default:
		return HorizontalCoords{}, nil
	}

	// Now convert equatorial to horizontal
	return c.CalculateStarPosition(ra, dec, observer, t)
}

// simplifiedPlanetPosition calculates approximate planet positions using orbital elements
// This is a simplified model accurate enough for visual representation
func (c *MeeusCalculator) simplifiedPlanetPosition(body int, jd float64) (ra, dec float64) {
	// Mean orbital elements for planets at J2000.0
	// Format: meanLongitude (deg), perihelion (deg), eccentricity, semiMajorAxis (AU), inclination (deg), node (deg)

	T := (jd - 2451545.0) / 36525.0 // Julian centuries since J2000.0

	var L0, omega, e, a, i, Omega float64

	switch body {
	case BodyMercury:
		L0, omega, e, a, i, Omega = 252.25084, 77.45645, 0.205630, 0.387098, 7.00487, 48.33167
	case BodyVenus:
		L0, omega, e, a, i, Omega = 181.97973, 131.53298, 0.006793, 0.723330, 3.39471, 76.68069
	case BodyMars:
		L0, omega, e, a, i, Omega = 355.43300, 336.060234, 0.093405, 1.523688, 1.84961, 49.57854
	case BodyJupiter:
		L0, omega, e, a, i, Omega = 34.351519, 14.75385, 0.048498, 5.20256, 1.30327, 100.55615
	case BodySaturn:
		L0, omega, e, a, i, Omega = 50.077444, 92.43194, 0.055548, 9.55475, 2.48888, 113.71504
	case BodyUranus:
		L0, omega, e, a, i, Omega = 314.055005, 170.96424, 0.046381, 19.18171, 0.773196, 74.22988
	case BodyNeptune:
		L0, omega, e, a, i, Omega = 304.348665, 44.97135, 0.009456, 30.05826, 1.76917, 131.72169
	case BodyPluto:
		// Pluto has a complex orbit, use approximate values
		L0, omega, e, a, i, Omega = 238.92881, 224.91678, 0.248852, 39.4817, 17.14175, 110.30347
	default:
		return 0, 0
	}

	// Calculate mean anomaly
	// For simplicity, assume mean motion based on orbital period
	period := math.Pow(a, 1.5)     // Kepler's 3rd law: T² = a³
	n := 360.0 / (period * 365.25) // Mean daily motion in degrees

	// Adjust mean longitude for time
	L := L0 + n*T*36525.0

	// Calculate mean anomaly
	M := L - omega
	M = math.Mod(M, 360.0)
	if M < 0 {
		M += 360.0
	}

	// Solve Kepler's equation (simplified - just use mean anomaly as approximation)
	// For visual purposes, this is sufficient
	E := M * math.Pi / 180.0 // Convert to radians

	// Calculate heliocentric coordinates
	x := a * (math.Cos(E) - e)
	y := a * math.Sqrt(1-e*e) * math.Sin(E)

	// Convert to ecliptic coordinates
	cosOmega := math.Cos(Omega * math.Pi / 180.0)
	sinOmega := math.Sin(Omega * math.Pi / 180.0)
	cosI := math.Cos(i * math.Pi / 180.0)
	sinI := math.Sin(i * math.Pi / 180.0)

	xEcl := (cosOmega*x - sinOmega*y*cosI)
	yEcl := (sinOmega*x + cosOmega*y*cosI)
	zEcl := y * sinI

	// Calculate longitude and latitude
	lambda := math.Atan2(yEcl, xEcl) * 180.0 / math.Pi
	beta := math.Atan2(zEcl, math.Sqrt(xEcl*xEcl+yEcl*yEcl)) * 180.0 / math.Pi

	// Convert to equatorial
	epsilon := meanObliquity(jd)
	ra, dec = eclipticToEquatorial(lambda, beta, epsilon)

	return ra, dec
}

// CalculateMoonPosition calculates horizontal coordinates for the Moon
func (c *MeeusCalculator) CalculateMoonPosition(observer Observer, t time.Time) (HorizontalCoords, error) {
	jd := toJulianDay(t)

	// Get moon's geocentric equatorial coordinates
	ra, dec, _ := moonposition.Position(jd)

	// Convert to degrees (ra and dec are in radians from moonposition.Position)
	raDeg := float64(ra) * 180.0 / math.Pi
	decDeg := float64(dec) * 180.0 / math.Pi

	// Now convert to horizontal
	return c.CalculateStarPosition(raDeg, decDeg, observer, t)
}

// CalculateMoonPhase returns the moon phase and illumination
func (c *MeeusCalculator) CalculateMoonPhase(t time.Time) (phase, illumination float64) {
	jd := toJulianDay(t)

	// Get sun's longitude
	sunLon := float64(solar.ApparentLongitude(jd).Deg())

	// Get moon's longitude
	ra, _, _ := moonposition.Position(jd)
	moonLon := float64(ra) * 180.0 / math.Pi

	// Calculate phase angle (elongation)
	phaseAngle := moonLon - sunLon
	for phaseAngle < 0 {
		phaseAngle += 360.0
	}
	for phaseAngle >= 360.0 {
		phaseAngle -= 360.0
	}

	// Convert to 0-1 phase (0=new moon, 0.5=full moon, 1=new moon again)
	phase = phaseAngle / 360.0

	// Calculate illumination percentage
	// 0° = 0% (new moon), 90° = 50%, 180° = 100% (full moon)
	illumination = (1.0 - math.Cos(phaseAngle*math.Pi/180.0)) / 2.0

	return phase, illumination
}

// GetBodyName returns the name of a solar system body
func (c *MeeusCalculator) GetBodyName(body int) string {
	switch body {
	case BodySun:
		return "Sun"
	case BodyMoon:
		return "Moon"
	case BodyMercury:
		return "Mercury"
	case BodyVenus:
		return "Venus"
	case BodyMars:
		return "Mars"
	case BodyJupiter:
		return "Jupiter"
	case BodySaturn:
		return "Saturn"
	case BodyUranus:
		return "Uranus"
	case BodyNeptune:
		return "Neptune"
	case BodyPluto:
		return "Pluto"
	default:
		return "Unknown"
	}
}

// meanObliquity returns the mean obliquity of the ecliptic
func meanObliquity(jd float64) float64 {
	// Meeus formula for mean obliquity
	T := (jd - 2451545.0) / 36525.0
	epsilon := 23.0 + 26.0/60.0 + 21.448/3600.0
	epsilon -= (46.8150*T + 0.00059*T*T - 0.001813*T*T*T) / 3600.0
	return epsilon
}

// eclipticToEquatorial converts ecliptic coordinates to equatorial
func eclipticToEquatorial(lambda, beta, epsilon float64) (ra, dec float64) {
	// Convert to radians
	lam := lambda * math.Pi / 180.0
	bet := beta * math.Pi / 180.0
	eps := epsilon * math.Pi / 180.0

	// Calculate right ascension
	// tan(ra) = (sin(lambda) * cos(epsilon) - tan(beta) * sin(epsilon)) / cos(lambda)
	y := math.Sin(lam)*math.Cos(eps) - math.Tan(bet)*math.Sin(eps)
	x := math.Cos(lam)
	ra = math.Atan2(y, x) * 180.0 / math.Pi
	if ra < 0 {
		ra += 360.0
	}

	// Calculate declination
	// sin(dec) = sin(beta) * cos(epsilon) + cos(beta) * sin(epsilon) * sin(lambda)
	dec = math.Asin(math.Sin(bet)*math.Cos(eps)+math.Cos(bet)*math.Sin(eps)*math.Sin(lam)) * 180.0 / math.Pi

	return ra, dec
}

// CalculateMoonPositionRelative calculates moon position relative to its parent planet
// This is a simplified model for visual representation
func CalculateMoonPositionRelative(parentAzimuth, parentAltitude float64, moonIndex int, t time.Time) (float64, float64) {
	// Create a simple orbital pattern based on time and moon index
	// This is for visual effect, not precise astronomical calculation

	// Convert time to a rotation angle
	seconds := float64(t.Unix())
	orbitPeriod := 1000.0 + float64(moonIndex)*200.0 // Different periods for different moons
	angle := seconds / orbitPeriod

	// Orbit radius (degrees in sky)
	radius := 2.0 + float64(moonIndex)*0.3

	// Calculate offset from parent
	deltaAz := math.Cos(angle) * radius
	deltaAlt := math.Sin(angle) * radius

	// Apply to parent position
	moonAz := math.Mod(parentAzimuth+deltaAz+360.0, 360.0)
	moonAlt := parentAltitude + deltaAlt

	// Clamp altitude
	if moonAlt > 90.0 {
		moonAlt = 90.0
	} else if moonAlt < -90.0 {
		moonAlt = -90.0
	}

	return moonAz, moonAlt
}

// IsVisible returns true if the object is above the horizon (with some margin)
func (h HorizontalCoords) IsVisible() bool {
	// Objects are visible if altitude > -5 degrees (just below horizon for atmospheric refraction)
	return h.Altitude > -5.0
}

// GetBrightness returns brightness factor (0-1) based on altitude
// Objects near horizon are dimmed
func (h HorizontalCoords) GetBrightness() float64 {
	if h.Altitude < 0 {
		return 0
	}
	if h.Altitude > 10.0 {
		return 1.0
	}
	// Linear fade from horizon to 10 degrees
	return h.Altitude / 10.0
}
