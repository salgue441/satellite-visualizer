package renderer

import "math"

// RotateY rotates a 3D point around the Y axis by angle radians.
func RotateY(x, y, z, angle float64) (float64, float64, float64) {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	return x*cosA + z*sinA, y, -x*sinA + z*cosA
}

// RotateX rotates a 3D point around the X axis by angle radians.
func RotateX(x, y, z, angle float64) (float64, float64, float64) {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	return x, y*cosA - z*sinA, y*sinA + z*cosA
}

// Project3DTo2D performs orthographic projection from 3D to terminal coordinates.
// Takes a 3D point (already rotated), sphere radius in row-units, and screen dimensions.
// Returns screen x, y coordinates and whether the point is visible (z > 0, front-facing).
// Accounts for terminal character aspect ratio (~2:1 height:width).
func Project3DTo2D(x, y, z, sphereR float64, screenW, screenH int) (sx, sy int, visible bool) {
	if z < 0 {
		return 0, 0, false
	}

	const charAspect = 2.0

	cx := float64(screenW) / 2.0
	cy := float64(screenH) / 2.0

	// x is in unit-sphere coords; scale to screen cells with aspect correction
	sx = int(cx + x*sphereR*charAspect)
	sy = int(cy - y*sphereR)

	if sx < 0 || sx >= screenW || sy < 0 || sy >= screenH {
		return 0, 0, false
	}

	return sx, sy, true
}
