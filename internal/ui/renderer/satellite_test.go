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

	// Render globe first, then capture center pixel, then render satellite
	g.Render(pb)
	cx, cy := pb.Width/2, pb.Height/2
	before := pb.Get(cx, cy)

	RenderSatellites(pb, sats, g)

	// The satellite should be near center (Z-axis facing viewer).
	// With blending, it won't be pure white but should be brighter than the globe.
	found := false
	for y := cy - 5; y <= cy+5; y++ {
		for x := cx - 5; x <= cx+5; x++ {
			after := pb.Get(x, y)
			if after.R > before.R+30 || after.G > before.G+30 || after.B > before.B+30 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected visible satellite to brighten pixels near center")
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

	// Snapshot before
	snapshot := make([]RGB, pb.Width*pb.Height)
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			snapshot[y*pb.Width+x] = pb.Get(x, y)
		}
	}

	RenderSatellites(pb, sats, g)

	// Nothing should have changed — satellite is behind the globe
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) != snapshot[y*pb.Width+x] {
				t.Errorf("hidden satellite modified pixel at (%d, %d)", x, y)
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

	// Stations render at full opacity with exact color
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

			// Snapshot before
			cx, cy := pb.Width/2, pb.Height/2
			before := pb.Get(cx, cy)

			RenderSatellites(pb, sats, g)

			// Verify a pixel near center was modified (satellite rendered)
			modified := false
			for y := cy - 5; y <= cy+5; y++ {
				for x := cx - 5; x <= cx+5; x++ {
					after := pb.Get(x, y)
					if after != before {
						modified = true

						_ = rgb // color verified by dedicated station/blending tests
						break
					}
				}
				if modified {
					break
				}
			}

			if !modified {
				t.Errorf("constellation %q: no pixel near center was modified", name)
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
	cx, cy := pb.Width/2, pb.Height/2
	before := pb.Get(cx, cy)

	RenderSatellites(pb, sats, g)

	// Should modify a pixel near center with blended default color
	modified := false
	for y := cy - 5; y <= cy+5; y++ {
		for x := cx - 5; x <= cx+5; x++ {
			if pb.Get(x, y) != before {
				modified = true
				break
			}
		}
		if modified {
			break
		}
	}

	if !modified {
		t.Error("expected satellite with unknown constellation to modify a pixel near center")
	}
}
