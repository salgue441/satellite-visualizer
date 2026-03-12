package renderer

import (
	"math"

	"satellite-visualizer/internal/domain"
)

// ConstellationColors maps constellation names to RGB foreground colors.
var ConstellationColors = map[string][3]int{
	"stations":     {255, 80, 80},   // red
	"starlink":     {255, 255, 255}, // white
	"gps-ops":      {255, 220, 50},  // yellow
	"oneweb":       {255, 160, 50},  // orange
	"iridium-NEXT": {80, 220, 255},  // cyan
	"galileo":      {180, 100, 255}, // purple
	"glo-ops":      {80, 255, 80},   // green
}

// DefaultSatRGB is used when a constellation has no specific color.
var DefaultSatRGB = [3]int{200, 200, 200}

// RenderSatellites draws satellite positions onto the pixel buffer.
// Satellites are blended with the underlying globe pixel using screen blending
// so the globe remains visible underneath dense constellations.
func RenderSatellites(pb *PixelBuffer, satellites []domain.SatelliteState, g *Globe) {
	earthRadius := 6378.137
	termH := float64(pb.Height) / 2.0
	sphereR := termH * 0.80 * g.Zoom
	const pixelAspect = 1.0

	// Track which pixels already have a satellite so we don't stack blends.
	// Multiple satellites hitting the same pixel would accumulate toward white.
	occupied := make([]bool, pb.Width*pb.Height)

	for _, sat := range satellites {
		dist := math.Sqrt(sat.Position.X*sat.Position.X +
			sat.Position.Y*sat.Position.Y +
			sat.Position.Z*sat.Position.Z)
		if dist == 0 {
			continue
		}

		// Project satellite position slightly above the globe surface.
		altRatio := dist/earthRadius - 1.0 // 0.0 at surface, ~0.086 for Starlink
		scale := 1.0 + altRatio*0.15       // spread higher orbits out more
		if scale < 1.02 {
			scale = 1.02
		}

		nx := sat.Position.X / earthRadius * scale
		ny := sat.Position.Y / earthRadius * scale
		nz := sat.Position.Z / earthRadius * scale

		rx, ry, rz := RotateY(nx, ny, nz, -g.RotationY)
		rx, ry, rz = RotateX(rx, ry, rz, -g.RotationX)
		if rz <= 0 {
			continue
		}

		cx := float64(pb.Width) / 2.0
		cy := float64(pb.Height) / 2.0
		px := int(cx + rx*sphereR*pixelAspect)
		py := int(cy - ry*sphereR)
		if px < 0 || px >= pb.Width || py < 0 || py >= pb.Height {
			continue
		}

		rgb, ok := ConstellationColors[sat.ConstellationName]
		if !ok {
			rgb = DefaultSatRGB
		}

		// Stations get full opacity and a 2x2 block (they're few and important).
		if sat.ConstellationName == "stations" {
			blended := RGB{R: uint8(rgb[0]), G: uint8(rgb[1]), B: uint8(rgb[2])}
			pb.Set(px, py, blended)
			pb.Set(px+1, py, blended)
			pb.Set(px, py+1, blended)
			pb.Set(px+1, py+1, blended)
			continue
		}

		// Skip if this pixel already has a satellite — prevents blend stacking.
		idx := py*pb.Width + px
		if occupied[idx] {
			continue
		}
		occupied[idx] = true

		// Blend satellite color with the underlying globe pixel.
		bg := pb.Get(px, py)
		const alpha = 0.4
		blended := RGB{
			R: uint8(float64(bg.R)*(1-alpha) + float64(rgb[0])*alpha),
			G: uint8(float64(bg.G)*(1-alpha) + float64(rgb[1])*alpha),
			B: uint8(float64(bg.B)*(1-alpha) + float64(rgb[2])*alpha),
		}
		pb.Set(px, py, blended)
	}
}
