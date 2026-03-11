package propagator

import "math"

const (
	// WGS84 Earth parameters
	EarthRadius     = 6378.137 // km (equatorial)
	EarthFlattening = 1.0 / 298.257223563
	Mu              = 398600.4418 // km³/s²

	// Gravitational zonal harmonics
	J2 = 0.00108262998905
	J3 = -0.00000253215306
	J4 = -0.00000161098761

	// Time constants
	MinutesPerDay = 1440.0
	SecondsPerDay = 86400.0

	// Mathematical constants
	TwoPi   = 2.0 * math.Pi
	Deg2Rad = math.Pi / 180.0
	Rad2Deg = 180.0 / math.Pi

	// SGP4 specific
	XKE    = 0.0743669161331734132       // 60.0 / sqrt(EarthRadius³ / Mu) with consistent units
	QOMS2T = 1.880279159015270643865e-09 // (120-78)^4 / EarthRadius^4 in er^4
	S      = 1.01222928                  // (1 + 78/EarthRadius) in earth radii
	AE     = 1.0                         // distance units (earth radii)
)
