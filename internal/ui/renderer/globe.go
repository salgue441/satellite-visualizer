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

// Render draws the globe into the given pixel buffer using ray-sphere intersection.
// Each pixel is mapped to a ray; if it hits the unit sphere, we compute
// lat/lon after rotation and shade with RGB based on land/ocean and depth.
func (g *Globe) Render(pb *PixelBuffer) {
	pb.Clear()

	w, h := pb.Width, pb.Height
	if w == 0 || h == 0 {
		return
	}

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	termH := float64(h) / 2.0
	sphereR := termH * 0.45 * g.Zoom

	const pixelAspect = 1.0

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			nx := (float64(px) - cx) / (sphereR * pixelAspect)
			ny := -(float64(py) - cy) / sphereR

			r2 := nx*nx + ny*ny
			if r2 > 1.0 {
				dist := math.Sqrt(r2) - 1.0
				if dist < 0.07 {
					if c, ok := AtmosphereGlow(dist); ok {
						pb.Set(px, py, c)
						continue
					}
				}
				if c, ok := StarField(px, py); ok {
					pb.Set(px, py, c)
				}
				continue
			}

			nz := math.Sqrt(1.0 - r2)

			edgeAlpha := 1.0
			if r2 > 0.98 {
				edgeAlpha = (1.0 - math.Sqrt(r2)) / (1.0 - math.Sqrt(0.98))
				if edgeAlpha > 1.0 {
					edgeAlpha = 1.0
				}
				if edgeAlpha < 0.0 {
					edgeAlpha = 0.0
				}
			}

			wx, wy, wz := RotateY(nx, ny, nz, -g.RotationY)
			wx, wy, wz = RotateX(wx, wy, wz, -g.RotationX)

			lat := math.Asin(wy) * 180.0 / math.Pi
			lon := math.Atan2(wx, wz) * 180.0 / math.Pi

			onGrid := isGridLine(lat, lon)

			var c RGB
			if IsLand(lat, lon) {
				c = LandShade(lat, lon, nz, onGrid)
			} else {
				c = OceanShade(nz, onGrid)
			}

			if edgeAlpha < 1.0 {
				c.R = uint8(float64(c.R) * edgeAlpha)
				c.G = uint8(float64(c.G) * edgeAlpha)
				c.B = uint8(float64(c.B) * edgeAlpha)
			}

			pb.Set(px, py, c)
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
