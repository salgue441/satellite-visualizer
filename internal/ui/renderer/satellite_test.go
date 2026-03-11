package renderer

import (
	"testing"

	"satellite-visualizer/internal/domain"
)

func TestRenderSatellites_VisibleSatellite(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()

	// Place satellite at positive Z in ECI (front-facing, no rotation).
	// Use position at earth's surface along +Z axis so it's visible.
	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "TEST-SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "starlink",
		},
	}

	RenderSatellites(f, sats, g)

	// The satellite should be rendered somewhere near the center of the frame.
	// Search the back buffer for the bullet character.
	found := false
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '●' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected visible satellite to be rendered with '●' but it was not found")
	}
}

func TestRenderSatellites_HiddenSatellite(t *testing.T) {
	f := NewFrame(80, 40)
	g := NewGlobe()

	// Place satellite behind the globe (negative Z in ECI, no rotation).
	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "HIDDEN", Position: domain.Position{X: 0, Y: 0, Z: -(earthR + 400)}},
			ConstellationName: "starlink",
		},
	}

	RenderSatellites(f, sats, g)

	// Should NOT find any satellite character.
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.Get(x, y)
			if cell.Char == '●' || cell.Char == '★' {
				t.Errorf("expected hidden satellite not to be rendered, but found '%c' at (%d, %d)", cell.Char, x, y)
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

	tests := []struct {
		constellation string
		wantColor     string
	}{
		{"stations", "\033[38;5;196m"},
		{"starlink", "\033[38;5;255m"},
		{"gps-ops", "\033[38;5;226m"},
		{"oneweb", "\033[38;5;208m"},
		{"iridium-NEXT", "\033[38;5;51m"},
		{"galileo", "\033[38;5;135m"},
		{"glo-ops", "\033[38;5;46m"},
	}

	for _, tt := range tests {
		t.Run(tt.constellation, func(t *testing.T) {
			f.Clear()
			sats := []domain.SatelliteState{
				{
					Satellite:         domain.Satellite{Name: "SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
					ConstellationName: tt.constellation,
				},
			}

			RenderSatellites(f, sats, g)

			found := false
			for y := 0; y < f.Height; y++ {
				for x := 0; x < f.Width; x++ {
					cell := f.Get(x, y)
					if cell.Char == '●' || cell.Char == '★' {
						if cell.Color != tt.wantColor {
							t.Errorf("constellation %q: got color %q, want %q", tt.constellation, cell.Color, tt.wantColor)
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
				t.Errorf("constellation %q: satellite character not found in frame", tt.constellation)
			}
		})
	}
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
			if cell.Char == '●' {
				if cell.Color != DefaultSatColor {
					t.Errorf("unknown constellation: got color %q, want %q", cell.Color, DefaultSatColor)
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
		t.Error("expected satellite with unknown constellation to be rendered with '●' but it was not found")
	}
}
