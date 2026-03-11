package renderer

import (
	"fmt"
	"math"
)

// AtmosphereGlow returns the character and ANSI color for atmosphere cells
// at the edge of the globe. distFromEdge is how far outside the sphere (0 = edge).
func AtmosphereGlow(distFromEdge float64) (rune, string) {
	var ch rune
	var colorCode int

	switch {
	case distFromEdge < 0.04:
		ch = '°'
		colorCode = 44 // bright cyan
	case distFromEdge < 0.08:
		ch = '·'
		colorCode = 44
	default:
		ch = '·'
		colorCode = 240 // dim gray
	}

	color := fmt.Sprintf("\033[38;5;%dm", colorCode)
	return ch, color
}

// LandShade returns the character and ANSI color for a land cell.
// Varies by latitude (tropics = green, poles = brown/white) and normalZ for brightness.
func LandShade(lat, normalZ float64) (rune, string) {
	absLat := math.Abs(lat)

	var ch rune
	var colorCode int

	// Character based on brightness (normalZ)
	if normalZ > 0.5 {
		ch = '█'
	} else {
		ch = '▓'
	}

	// Color based on latitude bands
	switch {
	case absLat > 70:
		// Polar - white/bright
		colorCode = 231
	case absLat > 50:
		// Temperate - brown
		colorCode = 136
	case absLat > 30:
		// Temperate - darker green/brown
		colorCode = 130
	default:
		// Tropical - green
		if normalZ > 0.5 {
			colorCode = 34 // bright green
		} else {
			colorCode = 28 // dark green
		}
	}

	color := fmt.Sprintf("\033[38;5;%dm", colorCode)
	return ch, color
}
