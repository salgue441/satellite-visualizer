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
// Satellites appear as bright colored pixels on the dark globe background.
func RenderSatellites(pb *PixelBuffer, satellites []domain.SatelliteState, g *Globe) {
	earthRadius := 6378.137
	termH := float64(pb.Height) / 2.0
	sphereR := termH * 0.45 * g.Zoom
	const pixelAspect = 1.0

	for _, sat := range satellites {
		dist := math.Sqrt(sat.Position.X*sat.Position.X +
			sat.Position.Y*sat.Position.Y +
			sat.Position.Z*sat.Position.Z)
		if dist == 0 {
			continue
		}
		scale := 1.0 + (dist/earthRadius-1.0)*0.02
		if scale < 1.01 {
			scale = 1.01
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
		satColor := RGB{R: uint8(rgb[0]), G: uint8(rgb[1]), B: uint8(rgb[2])}

		pb.Set(px, py, satColor)
		if sat.ConstellationName == "stations" {
			pb.Set(px+1, py, satColor)
			pb.Set(px, py+1, satColor)
			pb.Set(px+1, py+1, satColor)
		}
	}
}
