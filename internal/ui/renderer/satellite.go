package renderer

import (
	"fmt"
	"math"
	"strings"

	"satellite-visualizer/internal/domain"
)

// ConstellationColors maps constellation names to RGB foreground colors.
var ConstellationColors = map[string][3]int{
	"stations":     {255, 80, 80},  // red
	"starlink":     {255, 255, 255}, // white
	"gps-ops":      {255, 220, 50},  // yellow
	"oneweb":       {255, 160, 50},  // orange
	"iridium-NEXT": {80, 220, 255},  // cyan
	"galileo":      {180, 100, 255}, // purple
	"glo-ops":      {80, 255, 80},   // green
}

// DefaultSatRGB is used when a constellation has no specific color.
var DefaultSatRGB = [3]int{200, 200, 200}

// RenderSatellites draws satellite positions onto the frame.
// Satellites appear as bright colored symbols on the dark globe background.
func RenderSatellites(f *Frame, satellites []domain.SatelliteState, g *Globe) {
	earthRadius := 6378.137 // km

	// Sphere radius in row-units, must match globe.go Render()
	sphereR := float64(f.Height) * 0.45 * g.Zoom

	for _, sat := range satellites {
		dist := math.Sqrt(sat.Position.X*sat.Position.X +
			sat.Position.Y*sat.Position.Y +
			sat.Position.Z*sat.Position.Z)
		if dist == 0 {
			continue
		}

		// Normalize to slightly above unit sphere for visibility
		scale := 1.0 + (dist/earthRadius-1.0)*0.02 // subtle offset
		if scale < 1.01 {
			scale = 1.01
		}
		nx := sat.Position.X / earthRadius * scale
		ny := sat.Position.Y / earthRadius * scale
		nz := sat.Position.Z / earthRadius * scale

		// Apply same rotation as globe
		rx, ry, rz := RotateY(nx, ny, nz, -g.RotationY)
		rx, ry, rz = RotateX(rx, ry, rz, -g.RotationX)

		// Only render if on visible hemisphere
		if rz <= 0 {
			continue
		}

		sx, sy, visible := Project3DTo2D(rx, ry, rz, sphereR, f.Width, f.Height)
		if !visible {
			continue
		}

		// Pick character
		ch := '·'
		if sat.ConstellationName == "stations" {
			ch = '★'
		}

		// Get constellation color
		rgb, ok := ConstellationColors[sat.ConstellationName]
		if !ok {
			rgb = DefaultSatRGB
		}

		// Use the existing background from whatever is already rendered at this cell
		existing := f.Get(sx, sy)
		bgColor := ""
		if existing.Color != "" {
			// Extract background color (48;2;R;G;B) if present
			if idx := strings.Index(existing.Color, "48;2;"); idx >= 0 {
				end := strings.IndexByte(existing.Color[idx+5:], 'm')
				if end >= 0 {
					bgColor = existing.Color[idx : idx+5+end+1]
				}
			}
		}
		if bgColor == "" {
			bgColor = "48;2;5;15;30m"
		}
		color := fmt.Sprintf("\033[%s\033[38;2;%d;%d;%dm", bgColor, rgb[0], rgb[1], rgb[2])
		f.Set(sx, sy, ch, color)
	}
}
