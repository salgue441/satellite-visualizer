package renderer

// OceanShade returns an RGB color for ocean areas.
func OceanShade(normalZ float64, onGrid bool) RGB {
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
