package renderer

import (
	"strings"
	"testing"
)

func TestGlobeRender_ProducesOutput(t *testing.T) {
	g := NewGlobe()
	f := NewFrame(40, 20)
	g.Render(f)
	f.Swap()
	output := f.Render()
	if len(strings.TrimSpace(output)) == 0 {
		t.Error("expected non-empty rendered output")
	}
}

func TestGlobeRender_ContainsLandAndOcean(t *testing.T) {
	g := NewGlobe()
	f := NewFrame(80, 40)
	g.Render(f)
	f.Swap() // Move rendered content to front buffer for inspection

	var hasLand, hasOcean bool
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.front[y][x]
			if cell.Color == "" {
				continue
			}
			// Ocean cells are spaces with background color only (48;2; but no 38;2;)
			// Land cells have both background AND foreground (48;2;...38;2;...)
			if cell.Char == ' ' && strings.Contains(cell.Color, "48;2;") && !strings.Contains(cell.Color, "38;2;") {
				hasOcean = true
			}
			if cell.Char != ' ' && strings.Contains(cell.Color, "38;2;") {
				hasLand = true
			}
		}
	}

	if !hasOcean {
		t.Error("expected ocean-colored cells in rendered frame")
	}
	if !hasLand {
		t.Error("expected land-colored cells in rendered frame")
	}
}

func TestIsLand_KnownPoints(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{"NYC", 40, -74, true},
		{"Gulf of Guinea", 0, 0, false},
		{"Sydney", -33, 151, true},
		{"North Pole", 90, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLand(tt.lat, tt.lon)
			if got != tt.want {
				t.Errorf("IsLand(%v, %v) = %v, want %v", tt.lat, tt.lon, got, tt.want)
			}
		})
	}
}

func TestIsLand_Ocean(t *testing.T) {
	oceanPoints := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Mid Pacific", 0, -160},
		{"Mid Atlantic", 30, -40},
		{"Southern Ocean", -60, 90},
		{"North Atlantic", 50, -30},
	}
	for _, tt := range oceanPoints {
		t.Run(tt.name, func(t *testing.T) {
			if IsLand(tt.lat, tt.lon) {
				t.Errorf("IsLand(%v, %v) = true, expected false (ocean)", tt.lat, tt.lon)
			}
		})
	}
}

func TestOceanShade_ReturnsTrueColor(t *testing.T) {
	for _, nz := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
		ch, color := OceanShade(20.0, nz, false)
		if ch != ' ' {
			t.Errorf("OceanShade(%v) returned char %q, want space", nz, ch)
		}
		if !strings.Contains(color, "48;2;") {
			t.Errorf("OceanShade(%v) should use true-color background, got %q", nz, color)
		}
	}
}

func TestAtmosphereGlow_ThinHalo(t *testing.T) {
	// Very close to edge: should produce visible glow
	ch, color := AtmosphereGlow(0.01)
	if ch != ' ' || color == "" {
		t.Errorf("AtmosphereGlow(0.01) = (%q, %q), want (' ', non-empty)", ch, color)
	}

	// Far from edge: should produce nothing
	ch, color = AtmosphereGlow(0.1)
	if color != "" {
		t.Errorf("AtmosphereGlow(0.1) should return empty color for far distance, got %q", color)
	}
}

func TestLandShade_VariesByLatitude(t *testing.T) {
	_, tropicalColor := LandShade(5, 0.8, false)
	_, polarColor := LandShade(75, 0.8, false)

	if tropicalColor == polarColor {
		t.Errorf("expected different colors for tropical (%v) and polar (%v) latitudes",
			tropicalColor, polarColor)
	}
}

func TestLandShade_ReturnsTrueColor(t *testing.T) {
	_, color := LandShade(20, 0.5, false)
	// Should have both background (48;2;) and foreground (38;2;) true-color
	if !strings.Contains(color, "48;2;") {
		t.Errorf("LandShade should use true-color background, got %q", color)
	}
	if !strings.Contains(color, "38;2;") {
		t.Errorf("LandShade should use true-color foreground, got %q", color)
	}
}
