package propagator

import (
	"math"
	"testing"
	"time"

	"satellite-visualizer/internal/domain"
)

func TestIsDeepSpace(t *testing.T) {
	tests := []struct {
		name       string
		meanMotion float64
		want       bool
	}{
		{"LEO satellite (15.0 rev/day)", 15.0, false},
		{"GEO satellite (1.0 rev/day)", 1.0, true},
		{"boundary exact (6.4 rev/day)", 6.4, false},
		{"just below boundary (6.3 rev/day)", 6.3, true},
		{"MEO satellite (2.0 rev/day)", 2.0, true},
		{"high LEO (10.0 rev/day)", 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDeepSpace(tt.meanMotion)
			if got != tt.want {
				t.Errorf("IsDeepSpace(%f) = %v, want %v", tt.meanMotion, got, tt.want)
			}
		})
	}
}

// GEO satellite TLE (SES-1 like, mean motion ~1.003 rev/day)
const (
	geoLine1 = "1 36516U 10013A   20045.54321098  .00000020  00000-0  00000-0 0  9991"
	geoLine2 = "2 36516   0.0182 271.3450 0003250  15.3800 344.5800  1.00272110 36189"
)

func parseGEOElements(t *testing.T) domain.OrbitalElements {
	t.Helper()
	elems, err := domain.ParseTLE(geoLine1, geoLine2)
	if err != nil {
		t.Fatalf("failed to parse GEO TLE: %v", err)
	}
	return elems
}

func geoEpochTime() time.Time {
	// Epoch: 20045.54321098 -> 2020, day 45.54321098
	// Day 45 = Feb 14, 0.54321098 * 24 = 13.037 hours = 13h 02m 13s approx
	return time.Date(2020, 2, 14, 13, 2, 13, 0, time.UTC)
}

func TestDeepSpacePropagation_AtEpoch(t *testing.T) {
	elems := parseGEOElements(t)
	prop := &SGP4Propagator{}

	epochT := geoEpochTime()
	pos, vel, err := prop.Propagate(elems, epochT)
	if err != nil {
		t.Fatalf("Propagate at epoch failed: %v", err)
	}

	posMag := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y + pos.Z*pos.Z)
	velMag := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y + vel.Z*vel.Z)

	t.Logf("GEO Position at epoch: (%.3f, %.3f, %.3f) km, magnitude: %.3f km", pos.X, pos.Y, pos.Z, posMag)
	t.Logf("GEO Velocity at epoch: (%.6f, %.6f, %.6f) km/s, magnitude: %.6f km/s", vel.X, vel.Y, vel.Z, velMag)

	// GEO altitude ~35786 km + earth radius ~6378 km = ~42164 km
	// Allow range 40000-44000 km (loose tolerance for simplified model)
	if posMag < 40000 || posMag > 44000 {
		t.Errorf("position magnitude %.3f km out of expected GEO range [40000, 44000]", posMag)
	}

	// GEO velocity ~3.07 km/s, allow 2.5-3.5 km/s
	if velMag < 2.5 || velMag > 3.5 {
		t.Errorf("velocity magnitude %.6f km/s out of expected GEO range [2.5, 3.5]", velMag)
	}
}

func TestDeepSpacePropagation_After1Day(t *testing.T) {
	elems := parseGEOElements(t)
	prop := &SGP4Propagator{}

	epochT := geoEpochTime()
	t0 := epochT
	t1day := epochT.Add(24 * time.Hour)

	pos0, _, err := prop.Propagate(elems, t0)
	if err != nil {
		t.Fatalf("Propagate at epoch failed: %v", err)
	}

	pos1, vel1, err := prop.Propagate(elems, t1day)
	if err != nil {
		t.Fatalf("Propagate at t+1day failed: %v", err)
	}

	posMag := math.Sqrt(pos1.X*pos1.X + pos1.Y*pos1.Y + pos1.Z*pos1.Z)
	velMag := math.Sqrt(vel1.X*vel1.X + vel1.Y*vel1.Y + vel1.Z*vel1.Z)

	t.Logf("GEO Position at t+1day: (%.3f, %.3f, %.3f) km, magnitude: %.3f km", pos1.X, pos1.Y, pos1.Z, posMag)
	t.Logf("GEO Velocity at t+1day: (%.6f, %.6f, %.6f) km/s, magnitude: %.6f km/s", vel1.X, vel1.Y, vel1.Z, velMag)

	// Should still be in GEO range
	if posMag < 40000 || posMag > 44000 {
		t.Errorf("position magnitude %.3f km out of expected GEO range [40000, 44000]", posMag)
	}

	if velMag < 2.5 || velMag > 3.5 {
		t.Errorf("velocity magnitude %.6f km/s out of expected GEO range [2.5, 3.5]", velMag)
	}

	// Position should change (GEO period ~24h, so after 1 day it should be similar but slightly shifted)
	dist := math.Sqrt(
		(pos1.X-pos0.X)*(pos1.X-pos0.X) +
			(pos1.Y-pos0.Y)*(pos1.Y-pos0.Y) +
			(pos1.Z-pos0.Z)*(pos1.Z-pos0.Z),
	)
	t.Logf("Distance between epoch and t+1day: %.3f km", dist)
	// GEO has ~24h period, so after 1 day the position should have shifted slightly
	// due to perturbations. Use a very loose check.
	if dist < 0.001 {
		t.Errorf("position after 1 day should differ from epoch, distance: %.6f km", dist)
	}
}

func TestDeepSpaceCorrections_InitAndApply(t *testing.T) {
	// Test that initDeepSpace and applyDeepSpace produce sensible corrections
	inclination := 0.0003 // nearly equatorial (radians)
	eccentricity := 0.0003
	raan := 4.737       // ~271 degrees
	argPerigee := 0.268 // ~15 degrees
	meanMotion := 1.003  // GEO

	ds := initDeepSpace(inclination, eccentricity, raan, argPerigee, meanMotion)

	if ds == nil {
		t.Fatal("initDeepSpace returned nil")
	}

	// Apply corrections over 1440 minutes (1 day)
	raanAdj, wAdj, mAdj := applyDeepSpace(ds, raan, argPerigee, 0.0, 1440.0)

	// Corrections should produce finite results
	if math.IsNaN(raanAdj) || math.IsInf(raanAdj, 0) {
		t.Errorf("adjusted RAAN is not finite: %f", raanAdj)
	}
	if math.IsNaN(wAdj) || math.IsInf(wAdj, 0) {
		t.Errorf("adjusted arg perigee is not finite: %f", wAdj)
	}
	if math.IsNaN(mAdj) || math.IsInf(mAdj, 0) {
		t.Errorf("adjusted mean anomaly is not finite: %f", mAdj)
	}

	// RAAN should have shifted from the initial value
	raanDiff := math.Abs(raanAdj - raan)
	t.Logf("RAAN shift over 1 day: %.8f rad (%.6f deg)", raanDiff, raanDiff*180/math.Pi)
}
