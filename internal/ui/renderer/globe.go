package renderer

import "math"

// Globe represents a 3D globe that can be rendered to a Frame.
type Globe struct {
	Radius    float64
	RotationY float64 // longitude rotation (radians)
	RotationX float64 // latitude tilt (radians)
	Zoom      float64 // zoom factor (1.0 = default)
}

// NewGlobe creates a globe with default settings.
func NewGlobe() *Globe {
	return &Globe{Radius: 1.0, Zoom: 1.0}
}

// Render draws the globe into the given frame using ray-sphere intersection.
// Each terminal cell is mapped to a ray; if it hits the unit sphere, we compute
// lat/lon after rotation and shade with true-color based on land/ocean and depth.
func (g *Globe) Render(f *Frame) {
	f.Clear()

	w, h := f.Width, f.Height
	if w == 0 || h == 0 {
		return
	}

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0

	// Sphere radius in row-units. Fill ~90% of the vertical space.
	sphereR := float64(h) * 0.45 * g.Zoom

	// Terminal characters are roughly 2x taller than wide.
	const charAspect = 2.0

	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			// Normalize to unit-sphere coords [-1, 1].
			nx := (float64(sx) - cx) / (sphereR * charAspect)
			ny := -(float64(sy) - cy) / sphereR

			r2 := nx*nx + ny*ny
			if r2 > 1.0 {
				// Outside sphere — atmosphere halo or stars
				dist := math.Sqrt(r2) - 1.0
				if dist < 0.06 {
					ch, color := AtmosphereGlow(dist)
					if color != "" {
						f.Set(sx, sy, ch, color)
						continue
					}
				}
				// Sparse star field
				ch, color := StarField(sx, sy)
				if color != "" {
					f.Set(sx, sy, ch, color)
				}
				continue
			}

			// Point on sphere surface
			nz := math.Sqrt(1.0 - r2)

			// Apply rotation to get world coordinates
			wx, wy, wz := RotateY(nx, ny, nz, -g.RotationY)
			wx, wy, wz = RotateX(wx, wy, wz, -g.RotationX)

			// Convert to lat/lon
			lat := math.Asin(wy) * 180.0 / math.Pi
			lon := math.Atan2(wx, wz) * 180.0 / math.Pi

			// Check for grid lines (every 30° lat/lon)
			onGrid := isGridLine(lat, lon)

			// nz is the dot product with the view direction — natural lighting value.
			if IsLand(lat, lon) {
				ch, color := LandShade(lat, nz, onGrid)
				f.Set(sx, sy, ch, color)
			} else {
				ch, color := OceanShade(lat, nz, onGrid)
				f.Set(sx, sy, ch, color)
			}
		}
	}
}

// isGridLine returns true if the given lat/lon is near a grid line (every 30°).
func isGridLine(lat, lon float64) bool {
	// Grid every 30 degrees
	const spacing = 30.0
	const thickness = 1.2

	latMod := math.Mod(math.Abs(lat)+thickness/2, spacing)
	lonMod := math.Mod(math.Abs(lon)+thickness/2, spacing)

	return latMod < thickness || lonMod < thickness
}
