package renderer

import (
	"math"

	"satellite-visualizer/internal/domain"
)

// ConstellationColors maps constellation names to ANSI color codes.
var ConstellationColors = map[string]string{
	"stations":     "\033[38;5;196m", // red (ISS etc)
	"starlink":     "\033[38;5;255m", // white
	"gps-ops":      "\033[38;5;226m", // yellow
	"oneweb":       "\033[38;5;208m", // orange
	"iridium-NEXT": "\033[38;5;51m",  // cyan
	"galileo":      "\033[38;5;135m", // purple
	"glo-ops":      "\033[38;5;46m",  // green
}

// DefaultSatColor is used when a constellation has no specific color.
var DefaultSatColor = "\033[38;5;250m" // light gray

// RenderSatellites draws satellite positions onto the frame.
// Each satellite's ECI position is projected onto the globe using the same
// rotation parameters as the globe rendering.
func RenderSatellites(f *Frame, satellites []domain.SatelliteState, g *Globe) {
	earthRadius := 6378.137 // km

	for _, sat := range satellites {
		// Normalize position to unit sphere (ECI coords are in km)
		dist := math.Sqrt(sat.Position.X*sat.Position.X +
			sat.Position.Y*sat.Position.Y +
			sat.Position.Z*sat.Position.Z)
		if dist == 0 {
			continue
		}

		// Normalize to slightly above unit sphere (satellites are above surface)
		scale := 1.02 // slightly above globe surface for visibility
		nx := sat.Position.X / earthRadius * scale
		ny := sat.Position.Y / earthRadius * scale
		nz := sat.Position.Z / earthRadius * scale

		// Apply same rotation as globe
		rx, ry, rz := RotateY(nx, ny, nz, -g.RotationY)
		rx, ry, rz = RotateX(rx, ry, rz, -g.RotationX)

		// Only render if on visible hemisphere (z > 0)
		if rz <= 0 {
			continue
		}

		// Project to screen coordinates
		sx, sy, visible := Project3DTo2D(rx, ry, rz, g.Radius*g.Zoom, f.Width, f.Height)
		if !visible {
			continue
		}

		// Pick character and color
		ch := '●'
		if sat.ConstellationName == "stations" {
			ch = '★' // special char for ISS/stations
		}

		color, ok := ConstellationColors[sat.ConstellationName]
		if !ok {
			color = DefaultSatColor
		}

		f.Set(sx, sy, ch, color)
	}
}
