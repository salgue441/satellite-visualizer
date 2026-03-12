package renderer

import "fmt"

// OceanShade returns a styled cell for ocean areas.
// Shows subtle depth variation and optional grid lines.
func OceanShade(lat, normalZ float64, onGrid bool) (rune, string) {
	// Dark teal at edges, slightly lighter center
	r := int(3 + normalZ*8)
	g := int(10 + normalZ*18)
	b := int(20 + normalZ*30)

	if onGrid {
		// Faint grid line — slightly brighter teal foreground on ocean bg
		gr := int(8 + normalZ*15)
		gg := int(25 + normalZ*35)
		gb := int(40 + normalZ*50)
		return '·', fmt.Sprintf("\033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm", r, g, b, gr, gg, gb)
	}

	return ' ', fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// OceanShadeRGB returns an RGB color for ocean areas.
func OceanShadeRGB(normalZ float64, onGrid bool) RGB {
	r := uint8(10 + normalZ*20)
	g := uint8(30 + normalZ*50)
	b := uint8(80 + normalZ*90)
	if onGrid {
		r = uint8(min(255, int(r)+20))
		g = uint8(min(255, int(g)+20))
		b = uint8(min(255, int(b)+20))
	}
	return RGB{R: r, G: g, B: b}
}
