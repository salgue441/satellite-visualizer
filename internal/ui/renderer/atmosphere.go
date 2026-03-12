package renderer

import "math"

// LandShade returns an RGB color for land areas based on biome.
func LandShade(lat, lon, normalZ float64, onGrid bool) RGB {
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
		if isDesertRegion(lat, lon) {
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

func isDesertRegion(lat, lon float64) bool {
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

// AtmosphereGlow returns an RGB glow color for cells just outside the sphere.
func AtmosphereGlow(distFromEdge float64) (RGB, bool) {
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

// StarField returns an RGB star color for the space background.
func StarField(sx, sy int) (RGB, bool) {
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
