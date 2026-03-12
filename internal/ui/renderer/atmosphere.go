package renderer

import (
	"fmt"
	"math"
)

// AtmosphereGlow returns a subtle edge glow for cells just outside the sphere.
func AtmosphereGlow(distFromEdge float64) (rune, string) {
	if distFromEdge >= 0.05 {
		return ' ', ""
	}

	// Smooth fade: bright at edge, invisible by 0.05
	alpha := 1.0 - distFromEdge/0.05
	alpha *= alpha // quadratic falloff for smoothness

	r := int(5 * alpha)
	g := int(40 * alpha)
	b := int(70 * alpha)

	return ' ', fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// StarField returns a sparse star character for the space background.
func StarField(sx, sy int) (rune, string) {
	// Deterministic pseudo-random based on position
	hash := (sx*7919 + sy*6271 + sx*sy*31) % 997
	if hash > 12 {
		return ' ', ""
	}

	// Vary star brightness
	switch {
	case hash < 2:
		// Bright star
		return '*', "\033[38;2;180;180;200m"
	case hash < 5:
		// Medium star
		return '·', "\033[38;2;100;100;130m"
	case hash < 9:
		// Dim star
		return '.', "\033[38;2;50;50;70m"
	default:
		// Very dim
		return '.', "\033[38;2;30;30;45m"
	}
}

// LandShade returns a scattered ASCII character with colored foreground on dark background.
// Creates a "terminal/matrix" aesthetic with biome-aware coloring.
func LandShade(lat, normalZ float64, onGrid bool) (rune, string) {
	absLat := math.Abs(lat)

	// Pick character based on position hash for scatter effect
	hash := int((lat*17.3+normalZ*53.7)*100) % 100
	if hash < 0 {
		hash = -hash
	}

	var ch rune

	if onGrid {
		ch = '+'
	} else {
		switch {
		case normalZ > 0.6:
			// Bright face — denser characters
			switch {
			case hash < 20:
				ch = '▓'
			case hash < 40:
				ch = '▒'
			case hash < 55:
				ch = '░'
			case hash < 70:
				ch = '·'
			default:
				ch = '.'
			}
		case normalZ > 0.3:
			// Medium — sparser
			switch {
			case hash < 15:
				ch = '▒'
			case hash < 35:
				ch = '░'
			case hash < 50:
				ch = '·'
			default:
				ch = '.'
			}
		default:
			// Edge — very sparse
			switch {
			case hash < 15:
				ch = '░'
			case hash < 30:
				ch = '·'
			default:
				ch = ' '
			}
		}
	}

	// Biome-aware coloring
	var r, g, b int
	switch {
	case absLat > 75:
		// Polar ice — white/pale blue
		brightness := 0.5 + normalZ*0.5
		r = int(180 * brightness)
		g = int(200 * brightness)
		b = int(220 * brightness)
	case absLat > 55:
		// Boreal/taiga — dark green
		r = int(10 + normalZ*25)
		g = int(50 + normalZ*90)
		b = int(15 + normalZ*25)
	case absLat > 35:
		// Temperate — medium green
		r = int(15 + normalZ*30)
		g = int(70 + normalZ*130)
		b = int(15 + normalZ*30)
	case absLat > 15:
		// Subtropical — warm green / desert detection
		if isDesertRegion(lat, 0) {
			// Sandy/desert — amber/brown
			r = int(80 + normalZ*120)
			g = int(60 + normalZ*80)
			b = int(15 + normalZ*25)
		} else {
			// Lush tropical
			r = int(10 + normalZ*25)
			g = int(80 + normalZ*170)
			b = int(10 + normalZ*25)
		}
	default:
		// Equatorial — vivid green
		r = int(10 + normalZ*30)
		g = int(90 + normalZ*165)
		b = int(15 + normalZ*35)
	}

	if onGrid {
		// Grid lines on land — slightly brighter version of biome color
		r = min(255, r+40)
		g = min(255, g+40)
		b = min(255, b+40)
	}

	// Dark teal background matching ocean for seamless edges
	bgR := int(3 + normalZ*8)
	bgG := int(10 + normalZ*18)
	bgB := int(20 + normalZ*30)

	color := fmt.Sprintf("\033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm", bgR, bgG, bgB, r, g, b)
	return ch, color
}

// isDesertRegion returns true for approximate desert regions (Sahara, Arabia, etc.)
func isDesertRegion(lat, _ float64) bool {
	// Major desert belts are roughly 15-35° N and S
	absLat := math.Abs(lat)
	return absLat > 18 && absLat < 35
}

// LandShadeRGB returns an RGB color for land areas based on biome.
func LandShadeRGB(lat, lon, normalZ float64, onGrid bool) RGB {
	absLat := math.Abs(lat)
	var dark, lit RGB
	switch {
	case absLat > 75:
		dark = RGB{200, 220, 240}
		lit = RGB{230, 240, 250}
	case absLat > 55:
		dark = RGB{15, 70, 30}
		lit = RGB{30, 110, 45}
	case absLat > 35:
		dark = RGB{30, 100, 30}
		lit = RGB{50, 160, 50}
	case absLat > 15:
		if isDesertRegionV2(lat, lon) {
			dark = RGB{160, 130, 60}
			lit = RGB{200, 170, 90}
		} else {
			dark = RGB{40, 110, 30}
			lit = RGB{60, 170, 50}
		}
	default:
		dark = RGB{20, 120, 40}
		lit = RGB{40, 180, 60}
	}
	r := uint8(float64(dark.R) + float64(int(lit.R)-int(dark.R))*normalZ)
	g := uint8(float64(dark.G) + float64(int(lit.G)-int(dark.G))*normalZ)
	b := uint8(float64(dark.B) + float64(int(lit.B)-int(dark.B))*normalZ)
	if onGrid {
		r = uint8(min(255, int(r)+30))
		g = uint8(min(255, int(g)+30))
		b = uint8(min(255, int(b)+30))
	}
	return RGB{R: r, G: g, B: b}
}

func isDesertRegionV2(lat, lon float64) bool {
	absLat := math.Abs(lat)
	if absLat < 18 || absLat > 35 {
		return false
	}
	if lat > 0 {
		if lon >= -17 && lon <= 40 {
			return true
		} // Sahara
		if lon >= 40 && lon <= 60 {
			return true
		} // Arabian
		if absLat >= 23 && absLat <= 30 && lon >= 68 && lon <= 76 {
			return true
		} // Thar
	}
	if lat < 0 && lon >= 125 && lon <= 145 {
		return true
	} // Australian
	return false
}

// AtmosphereGlowRGB returns an RGB glow color for cells just outside the sphere.
func AtmosphereGlowRGB(distFromEdge float64) (RGB, bool) {
	if distFromEdge >= 0.07 {
		return RGB{}, false
	}
	alpha := 1.0 - distFromEdge/0.07
	alpha *= alpha
	return RGB{
		R: uint8(8 * alpha),
		G: uint8(50 * alpha),
		B: uint8(90 * alpha),
	}, true
}

// StarFieldRGB returns an RGB star color for the space background.
func StarFieldRGB(sx, sy int) (RGB, bool) {
	hash := (sx*7919 + sy*6271 + sx*sy*31) % 997
	if hash > 12 {
		return RGB{}, false
	}
	switch {
	case hash < 2:
		return RGB{180, 180, 200}, true
	case hash < 5:
		return RGB{100, 100, 130}, true
	case hash < 9:
		return RGB{50, 50, 70}, true
	default:
		return RGB{30, 30, 45}, true
	}
}
