package propagator

import (
	"math"
	"testing"
)

func TestSolveKepler(t *testing.T) {
	tests := []struct {
		name    string
		M       float64
		e       float64
		wantE   float64 // expected E, 0 means verify via Kepler's equation
		useTol  bool    // if true, compare E directly to wantE
		tol     float64
	}{
		{
			name:  "circular orbit e=0 M=1.0",
			M:     1.0,
			e:     0.0,
			wantE: 1.0,
			useTol: true,
			tol:   1e-10,
		},
		{
			name:  "low eccentricity e=0.001 M=1.0",
			M:     1.0,
			e:     0.001,
			wantE: 1.0,
			useTol: true,
			tol:   0.01,
		},
		{
			name: "moderate eccentricity e=0.5 M=pi/4",
			M:    math.Pi / 4,
			e:    0.5,
		},
		{
			name: "high eccentricity e=0.9 M=0.1",
			M:    0.1,
			e:    0.9,
		},
		{
			name:  "M=0 e=0.5",
			M:     0.0,
			e:     0.5,
			wantE: 0.0,
			useTol: true,
			tol:   1e-10,
		},
		{
			name:  "M=0 e=0.9",
			M:     0.0,
			e:     0.9,
			wantE: 0.0,
			useTol: true,
			tol:   1e-10,
		},
		{
			name: "M=pi e=0.5",
			M:    math.Pi,
			e:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			E := SolveKepler(tt.M, tt.e)

			if tt.useTol {
				if math.Abs(E-tt.wantE) > tt.tol {
					t.Errorf("SolveKepler(%v, %v) = %v, want %v (tol %v)", tt.M, tt.e, E, tt.wantE, tt.tol)
				}
			} else {
				// Verify Kepler's equation: E - e*sin(E) = M
				residual := E - tt.e*math.Sin(E) - tt.M
				if math.Abs(residual) > 1e-10 {
					t.Errorf("SolveKepler(%v, %v) = %v, Kepler residual = %v, want < 1e-10", tt.M, tt.e, E, residual)
				}
			}
		})
	}
}

func TestWrapTwoPi(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		want  float64
		tol   float64
	}{
		{"zero", 0, 0, 1e-10},
		{"2pi", TwoPi, 0, 1e-10},
		{"negative pi", -math.Pi, math.Pi, 1e-10},
		{"7pi", 7 * math.Pi, math.Pi, 1e-10},
		{"large negative", -10*math.Pi - math.Pi, math.Pi, 1e-10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapTwoPi(tt.angle)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("WrapTwoPi(%v) = %v, want %v", tt.angle, got, tt.want)
			}
		})
	}
}

func TestWrapNegPiToPi(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		want  float64
		tol   float64
	}{
		{"zero", 0, 0, 1e-10},
		{"2pi", TwoPi, 0, 1e-10},
		{"pi", math.Pi, -math.Pi, 1e-10},
		{"negative pi", -math.Pi, -math.Pi, 1e-10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapNegPiToPi(tt.angle)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("WrapNegPiToPi(%v) = %v, want %v", tt.angle, got, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name       string
		v, min, max float64
		want       float64
	}{
		{"within range", 5, 0, 10, 5},
		{"below min", -1, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("Clamp(%v, %v, %v) = %v, want %v", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
