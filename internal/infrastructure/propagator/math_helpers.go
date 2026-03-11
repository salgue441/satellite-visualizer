package propagator

import "math"

// SolveKepler solves Kepler's equation M = E - e*sin(E) for eccentric anomaly E
// using Newton-Raphson iteration with max 10 iterations and tolerance 1e-12.
func SolveKepler(M, e float64) float64 {
	E := M // initial guess
	for i := 0; i < 10; i++ {
		f := E - e*math.Sin(E) - M
		fp := 1.0 - e*math.Cos(E)
		delta := f / fp
		E -= delta
		if math.Abs(delta) < 1e-12 {
			break
		}
	}
	return E
}

// WrapTwoPi normalizes an angle to [0, 2*pi).
func WrapTwoPi(angle float64) float64 {
	a := math.Mod(angle, TwoPi)
	if a < 0 {
		a += TwoPi
	}
	return a
}

// WrapNegPiToPi normalizes an angle to [-pi, pi).
func WrapNegPiToPi(angle float64) float64 {
	a := WrapTwoPi(angle)
	if a >= math.Pi {
		a -= TwoPi
	}
	return a
}

// Clamp restricts v to the range [min, max].
func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
