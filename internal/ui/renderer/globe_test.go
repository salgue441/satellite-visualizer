package renderer

import (
	"strings"
	"testing"
)

func TestGlobeRender_ProducesOutput(t *testing.T) {
	g := NewGlobe()
	f := NewFrame(40, 20)
	g.Render(f)
	output := f.Render()
	if len(strings.TrimSpace(output)) == 0 {
		t.Error("expected non-empty rendered output")
	}
}

func TestGlobeRender_ContainsLandAndOcean(t *testing.T) {
	g := NewGlobe()
	f := NewFrame(80, 40)
	g.Render(f)

	var hasLand, hasOcean bool
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			cell := f.front[y][x]
			if cell.Color == "" {
				continue
			}
			// Ocean colors contain ;5;17m through ;5;27m (blue range)
			if strings.Contains(cell.Color, ";5;17m") ||
				strings.Contains(cell.Color, ";5;19m") ||
				strings.Contains(cell.Color, ";5;21m") ||
				strings.Contains(cell.Color, ";5;24m") ||
				strings.Contains(cell.Color, ";5;27m") {
				hasOcean = true
			}
			// Land colors contain green/brown codes
			if strings.Contains(cell.Color, ";5;34m") ||
				strings.Contains(cell.Color, ";5;28m") ||
				strings.Contains(cell.Color, ";5;130m") ||
				strings.Contains(cell.Color, ";5;136m") ||
				strings.Contains(cell.Color, ";5;231m") {
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

func TestOceanShade_ReturnsValidChars(t *testing.T) {
	validChars := map[rune]bool{'░': true, '▒': true, '▓': true, '█': true}
	for _, nz := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
		ch, color := OceanShade(nz)
		if !validChars[ch] {
			t.Errorf("OceanShade(%v) returned invalid char %q", nz, ch)
		}
		if color == "" {
			t.Errorf("OceanShade(%v) returned empty color", nz)
		}
	}
}

func TestAtmosphereGlow_ReturnsValidChars(t *testing.T) {
	validChars := map[rune]bool{'·': true, '°': true, ' ': true}
	for _, d := range []float64{0.0, 0.05, 0.1, 0.14} {
		ch, color := AtmosphereGlow(d)
		if !validChars[ch] {
			t.Errorf("AtmosphereGlow(%v) returned invalid char %q", d, ch)
		}
		if color == "" {
			t.Errorf("AtmosphereGlow(%v) returned empty color", d)
		}
	}
}

func TestLandShade_VariesByLatitude(t *testing.T) {
	_, tropicalColor := LandShade(5, 0.8)
	_, polarColor := LandShade(75, 0.8)

	if tropicalColor == polarColor {
		t.Errorf("expected different colors for tropical (%v) and polar (%v) latitudes",
			tropicalColor, polarColor)
	}
}
