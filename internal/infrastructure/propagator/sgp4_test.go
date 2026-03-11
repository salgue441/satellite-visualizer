package propagator

import (
	"math"
	"testing"
	"time"

	"satellite-visualizer/internal/domain"
)

// ISS TLE for testing (epoch: 2020-02-14 04:27:39 UTC approx)
const (
	issLine1 = "1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990"
	issLine2 = "2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209246"
)

func parseISSElements(t *testing.T) domain.OrbitalElements {
	t.Helper()
	elems, err := domain.ParseTLE(issLine1, issLine2)
	if err != nil {
		t.Fatalf("failed to parse ISS TLE: %v", err)
	}
	return elems
}

func magnitude3(x, y, z float64) float64 {
	return math.Sqrt(x*x + y*y + z*z)
}

// epochTime returns the approximate time corresponding to the ISS TLE epoch.
// Epoch: 20045.18587073 -> 2020, day 45.18587073
// Day 45 of 2020 = February 14
func issEpochTime() time.Time {
	// 2020 day 45.18587073
	// Day 45 = Feb 14 (31 Jan + 14 Feb)
	// 0.18587073 * 24 = 4.4609 hours = 4h 27m 39s approx
	return time.Date(2020, 2, 14, 4, 27, 39, 0, time.UTC)
}

func TestSGP4Propagate_AtEpoch(t *testing.T) {
	elems := parseISSElements(t)
	prop := &SGP4Propagator{}

	epochT := issEpochTime()
	pos, vel, err := prop.Propagate(elems, epochT)
	if err != nil {
		t.Fatalf("Propagate at epoch failed: %v", err)
	}

	posMag := magnitude3(pos.X, pos.Y, pos.Z)
	velMag := magnitude3(vel.X, vel.Y, vel.Z)

	t.Logf("Position at epoch: (%.3f, %.3f, %.3f) km, magnitude: %.3f km", pos.X, pos.Y, pos.Z, posMag)
	t.Logf("Velocity at epoch: (%.6f, %.6f, %.6f) km/s, magnitude: %.6f km/s", vel.X, vel.Y, vel.Z, velMag)

	// ISS orbits at ~400 km altitude, Earth radius ~6378 km, so distance ~6778 km
	if posMag < 6500 || posMag > 7000 {
		t.Errorf("position magnitude %.3f km out of expected range [6500, 7000]", posMag)
	}

	// LEO velocity is ~7.5 km/s
	if velMag < 7.0 || velMag > 8.0 {
		t.Errorf("velocity magnitude %.6f km/s out of expected range [7.0, 8.0]", velMag)
	}
}

func TestSGP4Propagate_After60Minutes(t *testing.T) {
	elems := parseISSElements(t)
	prop := &SGP4Propagator{}

	epochT := issEpochTime()
	t0 := epochT
	t60 := epochT.Add(60 * time.Minute)

	pos0, _, err := prop.Propagate(elems, t0)
	if err != nil {
		t.Fatalf("Propagate at epoch failed: %v", err)
	}

	pos60, vel60, err := prop.Propagate(elems, t60)
	if err != nil {
		t.Fatalf("Propagate at t+60min failed: %v", err)
	}

	posMag := magnitude3(pos60.X, pos60.Y, pos60.Z)
	velMag := magnitude3(vel60.X, vel60.Y, vel60.Z)

	t.Logf("Position at t+60min: (%.3f, %.3f, %.3f) km, magnitude: %.3f km", pos60.X, pos60.Y, pos60.Z, posMag)
	t.Logf("Velocity at t+60min: (%.6f, %.6f, %.6f) km/s, magnitude: %.6f km/s", vel60.X, vel60.Y, vel60.Z, velMag)

	if posMag < 6500 || posMag > 7000 {
		t.Errorf("position magnitude %.3f km out of expected range [6500, 7000]", posMag)
	}

	if velMag < 7.0 || velMag > 8.0 {
		t.Errorf("velocity magnitude %.6f km/s out of expected range [7.0, 8.0]", velMag)
	}

	// Position must have changed
	dist := magnitude3(pos60.X-pos0.X, pos60.Y-pos0.Y, pos60.Z-pos0.Z)
	if dist < 1.0 {
		t.Errorf("position after 60 min should differ significantly from epoch, distance: %.3f km", dist)
	}
}

func TestSGP4Propagate_After1Day(t *testing.T) {
	elems := parseISSElements(t)
	prop := &SGP4Propagator{}

	t1day := issEpochTime().Add(24 * time.Hour)

	pos, vel, err := prop.Propagate(elems, t1day)
	if err != nil {
		t.Fatalf("Propagate at t+1day failed: %v", err)
	}

	posMag := magnitude3(pos.X, pos.Y, pos.Z)
	velMag := magnitude3(vel.X, vel.Y, vel.Z)

	t.Logf("Position at t+1day: (%.3f, %.3f, %.3f) km, magnitude: %.3f km", pos.X, pos.Y, pos.Z, posMag)
	t.Logf("Velocity at t+1day: (%.6f, %.6f, %.6f) km/s, magnitude: %.6f km/s", vel.X, vel.Y, vel.Z, velMag)

	if posMag < 6500 || posMag > 7000 {
		t.Errorf("position magnitude %.3f km out of expected range [6500, 7000]", posMag)
	}

	if velMag < 7.0 || velMag > 8.0 {
		t.Errorf("velocity magnitude %.6f km/s out of expected range [7.0, 8.0]", velMag)
	}
}

func TestSGP4Propagate_PositionChanges(t *testing.T) {
	elems := parseISSElements(t)
	prop := &SGP4Propagator{}

	epochT := issEpochTime()
	times := []time.Time{
		epochT,
		epochT.Add(10 * time.Minute),
		epochT.Add(30 * time.Minute),
		epochT.Add(90 * time.Minute),
	}

	positions := make([]domain.Position, len(times))
	for i, tt := range times {
		pos, _, err := prop.Propagate(elems, tt)
		if err != nil {
			t.Fatalf("Propagate at time[%d] failed: %v", i, err)
		}
		positions[i] = pos
	}

	// Every pair of positions should differ
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			dist := magnitude3(
				positions[i].X-positions[j].X,
				positions[i].Y-positions[j].Y,
				positions[i].Z-positions[j].Z,
			)
			if dist < 1.0 {
				t.Errorf("positions at time[%d] and time[%d] too close: %.3f km", i, j, dist)
			}
		}
	}
}
