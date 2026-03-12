package renderer

import (
	"testing"

	"satellite-visualizer/internal/domain"
)

func TestRenderSatellites_VisibleSatellite(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "TEST-SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "starlink",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	// Starlink is white (255,255,255) — look for that pixel
	white := RGB{255, 255, 255}
	found := false
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == white {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected visible satellite to be rendered but no white pixel was found")
	}
}

func TestRenderSatellites_HiddenSatellite(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "HIDDEN", Position: domain.Position{X: 0, Y: 0, Z: -(earthR + 400)}},
			ConstellationName: "starlink",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	white := RGB{255, 255, 255}
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == white {
				t.Errorf("expected hidden satellite not to be rendered, found white pixel at (%d, %d)", x, y)
				return
			}
		}
	}
}

func TestRenderSatellites_StationGets2x2Block(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "ISS", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "stations",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	stationColor := RGB{255, 80, 80}
	count := 0
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == stationColor {
				count++
			}
		}
	}

	if count < 4 {
		t.Errorf("expected station satellite to have at least 4 red pixels (2x2 block), got %d", count)
	}
}

func TestRenderSatellites_ConstellationColors(t *testing.T) {
	earthR := 6378.137

	for name, rgb := range ConstellationColors {
		t.Run(name, func(t *testing.T) {
			pb := NewPixelBuffer(80, 80)
			g := NewGlobe()

			sats := []domain.SatelliteState{
				{
					Satellite:         domain.Satellite{Name: "SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
					ConstellationName: name,
				},
			}

			g.Render(pb)
			RenderSatellites(pb, sats, g)

			expected := RGB{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2])}
			found := false
			for y := 0; y < pb.Height; y++ {
				for x := 0; x < pb.Width; x++ {
					if pb.Get(x, y) == expected {
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				t.Errorf("constellation %q: expected pixel with RGB(%d,%d,%d) not found", name, rgb[0], rgb[1], rgb[2])
			}
		})
	}
}

func TestRenderSatellites_UnknownConstellation(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()
	earthR := 6378.137

	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "MYSTERY", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "unknown-constellation",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	defaultColor := RGB{200, 200, 200}
	found := false
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == defaultColor {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected satellite with unknown constellation to use default RGB(200,200,200) but pixel was not found")
	}
}
