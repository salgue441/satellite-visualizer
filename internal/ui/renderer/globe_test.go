package renderer

import "testing"

func TestGlobeRender_ProducesOutput(t *testing.T) {
	g := NewGlobe()
	pb := NewPixelBuffer(40, 40)
	g.Render(pb)
	output := pb.CompositeHalfBlocks()
	if len(output) == 0 {
		t.Error("expected non-empty rendered output")
	}
}

func TestGlobeRender_ContainsLandAndOcean(t *testing.T) {
	g := NewGlobe()
	pb := NewPixelBuffer(80, 80)
	g.Render(pb)
	hasLand := false
	hasOcean := false
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			c := pb.Get(x, y)
			if c == (RGB{}) {
				continue
			}
			if c.B > c.R && c.B > c.G && c.B > 50 {
				hasOcean = true
			}
			if c.G > c.R && c.G > c.B && c.G > 50 {
				hasLand = true
			}
		}
	}
	if !hasOcean {
		t.Error("expected ocean-colored pixels")
	}
	if !hasLand {
		t.Error("expected land-colored pixels")
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
			if got := IsLand(tt.lat, tt.lon); got != tt.want {
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
				t.Errorf("IsLand(%v, %v) = true, expected false", tt.lat, tt.lon)
			}
		})
	}
}

func TestOceanShadeRGB_Range(t *testing.T) {
	for _, nz := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
		c := OceanShadeRGB(nz, false)
		if c.B <= c.R || c.B <= c.G {
			t.Errorf("OceanShadeRGB(%v): blue should dominate, got %v", nz, c)
		}
	}
}

func TestAtmosphereGlowRGB_ThinHalo(t *testing.T) {
	c, ok := AtmosphereGlowRGB(0.01)
	if !ok || c == (RGB{}) {
		t.Errorf("AtmosphereGlowRGB(0.01) should return visible glow")
	}
	_, ok = AtmosphereGlowRGB(0.1)
	if ok {
		t.Errorf("AtmosphereGlowRGB(0.1) should return false")
	}
}

func TestLandShadeRGB_VariesByLatitude(t *testing.T) {
	tropical := LandShadeRGB(5, 0, 0.8, false)
	polar := LandShadeRGB(80, 0, 0.8, false)
	if tropical == polar {
		t.Errorf("expected different colors for tropical %v and polar %v", tropical, polar)
	}
}
