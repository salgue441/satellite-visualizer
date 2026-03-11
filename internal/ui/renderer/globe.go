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

// Render draws the globe into the given frame.
func (g *Globe) Render(f *Frame) {
	f.Clear()

	w, h := f.Width, f.Height
	radius := g.Radius * g.Zoom

	// For each terminal cell, trace a ray to the sphere
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			// Map screen coords to normalized [-1, 1] space
			// Account for terminal aspect ratio (chars are ~2:1 height:width)
			nx := (2.0*float64(sx)/float64(w) - 1.0) / (radius * 2.0) // * 2 for aspect ratio
			ny := -(2.0*float64(sy)/float64(h) - 1.0) / radius

			// Check if this pixel hits the sphere
			r2 := nx*nx + ny*ny
			if r2 > 1.0 {
				// Outside sphere - check for atmosphere
				dist := math.Sqrt(r2) - 1.0
				if dist < 0.15 { // atmosphere ring
					ch, color := AtmosphereGlow(dist)
					f.Set(sx, sy, ch, color)
				}
				continue
			}

			// Calculate z on sphere surface
			nz := math.Sqrt(1.0 - r2)

			// Apply rotation to get world coordinates
			wx, wy, wz := RotateY(nx, ny, nz, -g.RotationY)
			wx, wy, wz = RotateX(wx, wy, wz, -g.RotationX)

			// Convert to lat/lon
			lat := math.Asin(wy) * 180.0 / math.Pi
			lon := math.Atan2(wx, wz) * 180.0 / math.Pi

			// Determine if land or ocean
			if IsLand(lat, lon) {
				// Land - green/brown based on latitude
				ch, color := LandShade(lat, nz)
				f.Set(sx, sy, ch, color)
			} else {
				// Ocean - blue shading
				ch, color := OceanShade(nz)
				f.Set(sx, sy, ch, color)
			}
		}
	}

	f.Swap()
}
