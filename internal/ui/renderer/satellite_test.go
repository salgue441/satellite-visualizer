package renderer

import (
	"strings"
	"testing"

	"satellite-visualizer/internal/domain"
)

func TestRenderSatellites_VisibleSatellite(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "TEST-SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "starlink",
		},
	}

	RenderSatellites(f, sats, g)

	found := false
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '·' || cell.Char == '★' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected visible satellite to be rendered but it was not found")
	}
}

func TestRenderSatellites_HiddenSatellite(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "HIDDEN", Position: domain.Position{X: 0, Y: 0, Z: -(earthR + 400)}},
			ConstellationName: "starlink",
		},
	}

	RenderSatellites(f, sats, g)

	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '·' && strings.Contains(cell.Color, "255;255;255") {
				t.Errorf("expected hidden satellite not to be rendered, found at (%d, %d)", x, y)
				return
			}
		}
	}
}

func TestRenderSatellites_StationGetsStarChar(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "ISS", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "stations",
		},
	}

	RenderSatellites(f, sats, g)

	found := false
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '★' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected station satellite to be rendered with '★' but it was not found")
	}
}

func TestRenderSatellites_ConstellationColors(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()
	earthR := 6378.137

	for name, rgb := range ConstellationColors {
		t.Run(name, func(t *testing.T) {
			f.Clear()
			sats := []domain.SatelliteState{
				{
					Satellite:         domain.Satellite{Name: "SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
					ConstellationName: name,
				},
			}

			RenderSatellites(f, sats, g)

			found := false
			for y := 0; y < f.Height; y++ {
				for x := 0; x < f.Width; x++ {
					cell := f.Get(x, y)
					if cell.Char == '·' || cell.Char == '★' {
						// Verify the color contains the constellation's RGB values
						expectedFg := strings.Contains(cell.Color,
							strings.Join([]string{
								itoa(rgb[0]), itoa(rgb[1]), itoa(rgb[2]),
							}, ";"))
						if !expectedFg {
							t.Errorf("constellation %q: color %q doesn't contain expected RGB", name, cell.Color)
						}
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				t.Errorf("constellation %q: satellite character not found in frame", name)
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func TestRenderSatellites_UnknownConstellation(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()
	earthR := 6378.137

	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "MYSTERY", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "unknown-constellation",
		},
	}

	RenderSatellites(f, sats, g)

	found := false
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '·' && cell.Color != "" {
				// Should use default RGB (200,200,200)
				if !strings.Contains(cell.Color, "200;200;200") {
					t.Errorf("unknown constellation: got color %q, want default RGB", cell.Color)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected satellite with unknown constellation to be rendered but it was not found")
	}
}
