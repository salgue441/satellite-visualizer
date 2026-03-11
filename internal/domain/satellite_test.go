package domain

import (
	"errors"
	"testing"
)

// --- Position ---

func TestPositionZeroValue(t *testing.T) {
	var p Position
	if p.X != 0 || p.Y != 0 || p.Z != 0 {
		t.Fatal("zero-value Position should have all fields as 0")
	}
}

func TestPositionFieldAssignment(t *testing.T) {
	p := Position{X: 1.1, Y: 2.2, Z: 3.3}
	if p.X != 1.1 || p.Y != 2.2 || p.Z != 3.3 {
		t.Fatal("Position fields not set correctly")
	}
}

// --- Velocity ---

func TestVelocityZeroValue(t *testing.T) {
	var v Velocity
	if v.X != 0 || v.Y != 0 || v.Z != 0 {
		t.Fatal("zero-value Velocity should have all fields as 0")
	}
}

func TestVelocityFieldAssignment(t *testing.T) {
	v := Velocity{X: 7.5, Y: -1.2, Z: 0.3}
	if v.X != 7.5 || v.Y != -1.2 || v.Z != 0.3 {
		t.Fatal("Velocity fields not set correctly")
	}
}

// --- GeoCoordinate ---

func TestGeoCoordinateZeroValue(t *testing.T) {
	var g GeoCoordinate
	if g.Latitude != 0 || g.Longitude != 0 || g.Altitude != 0 {
		t.Fatal("zero-value GeoCoordinate should have all fields as 0")
	}
}

func TestGeoCoordinateFieldAssignment(t *testing.T) {
	g := GeoCoordinate{Latitude: 40.7128, Longitude: -74.006, Altitude: 408.0}
	if g.Latitude != 40.7128 || g.Longitude != -74.006 || g.Altitude != 408.0 {
		t.Fatal("GeoCoordinate fields not set correctly")
	}
}

// --- TLE ---

func TestTLEFieldAssignment(t *testing.T) {
	tle := TLE{
		Name:  "ISS",
		Line1: "1 25544U 98067A   ...",
		Line2: "2 25544  51.6416 ...",
	}
	if tle.Name != "ISS" || tle.Line1 == "" || tle.Line2 == "" {
		t.Fatal("TLE fields not set correctly")
	}
}

// --- OrbitalElements ---

func TestOrbitalElementsZeroValue(t *testing.T) {
	var oe OrbitalElements
	if oe.Epoch != 0 || oe.Inclination != 0 || oe.NoradCatNo != 0 {
		t.Fatal("zero-value OrbitalElements should have all fields as 0")
	}
}

func TestOrbitalElementsFieldAssignment(t *testing.T) {
	oe := OrbitalElements{
		Epoch:         2459580.5,
		BStar:         0.00001,
		Inclination:   0.9006, // radians
		RAAN:          1.2345,
		Eccentricity:  0.0001,
		ArgPerigee:    0.5678,
		MeanAnomaly:   3.1415,
		MeanMotion:    15.49,
		ElementSetNo:  999,
		NoradCatNo:    25544,
		RevolutionNo:  30000,
	}
	if oe.Epoch != 2459580.5 {
		t.Fatal("OrbitalElements.Epoch not set correctly")
	}
	if oe.NoradCatNo != 25544 {
		t.Fatal("OrbitalElements.NoradCatNo not set correctly")
	}
	if oe.RevolutionNo != 30000 {
		t.Fatal("OrbitalElements.RevolutionNo not set correctly")
	}
}

// --- Satellite ---

func TestSatelliteFieldAssignment(t *testing.T) {
	s := Satellite{
		Name:   "ISS",
		RawTLE: TLE{Name: "ISS", Line1: "1 ...", Line2: "2 ..."},
		Position: Position{X: 100, Y: 200, Z: 300},
	}
	if s.Name != "ISS" {
		t.Fatal("Satellite.Name not set correctly")
	}
	if s.Position.X != 100 {
		t.Fatal("Satellite.Position not set correctly")
	}
}

// --- SatelliteState ---

func TestSatelliteStateEmbedsSatellite(t *testing.T) {
	state := SatelliteState{
		Satellite: Satellite{
			Name:     "ISS",
			RawTLE:   TLE{Name: "ISS", Line1: "1 ...", Line2: "2 ..."},
			Position: Position{X: 100, Y: 200, Z: 300},
		},
		Geo:               GeoCoordinate{Latitude: 40.0, Longitude: -74.0, Altitude: 408.0},
		Vel:               Velocity{X: 7.5, Y: -1.2, Z: 0.3},
		Visible:           true,
		ConstellationName: "ISS",
	}

	// Access embedded fields directly.
	if state.Name != "ISS" {
		t.Fatal("SatelliteState should promote Satellite.Name")
	}
	if state.Position.X != 100 {
		t.Fatal("SatelliteState should promote Satellite.Position")
	}
	if !state.Visible {
		t.Fatal("SatelliteState.Visible not set correctly")
	}
	if state.ConstellationName != "ISS" {
		t.Fatal("SatelliteState.ConstellationName not set correctly")
	}
	if state.Geo.Altitude != 408.0 {
		t.Fatal("SatelliteState.Geo not set correctly")
	}
	if state.Vel.X != 7.5 {
		t.Fatal("SatelliteState.Vel not set correctly")
	}
}

// --- Constellation ---

func TestConstellationFieldAssignment(t *testing.T) {
	c := Constellation{
		Name: "Starlink",
		Satellites: []SatelliteState{
			{
				Satellite: Satellite{Name: "SL-1"},
				Visible:   true,
			},
			{
				Satellite: Satellite{Name: "SL-2"},
				Visible:   false,
			},
		},
	}
	if c.Name != "Starlink" {
		t.Fatal("Constellation.Name not set correctly")
	}
	if len(c.Satellites) != 2 {
		t.Fatalf("expected 2 satellites, got %d", len(c.Satellites))
	}
	if c.Satellites[0].Name != "SL-1" {
		t.Fatal("Constellation satellite name mismatch")
	}
}

// --- Errors ---

func TestErrorsSentinelValues(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrInvalidTle", ErrInvalidTle},
		{"ErrCalculationFailed", ErrCalculationFailed},
		{"ErrConstellationNotFound", ErrConstellationNotFound},
		{"ErrStaleData", ErrStaleData},
		{"ErrAuthFailed", ErrAuthFailed},
		{"ErrProviderUnavailable", ErrProviderUnavailable},
	}

	for _, tc := range sentinels {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("%s should not be nil", tc.name)
			}
			if tc.err.Error() == "" {
				t.Fatalf("%s should have a non-empty message", tc.name)
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	all := []error{
		ErrInvalidTle,
		ErrCalculationFailed,
		ErrConstellationNotFound,
		ErrStaleData,
		ErrAuthFailed,
		ErrProviderUnavailable,
	}
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if errors.Is(all[i], all[j]) {
				t.Fatalf("error %d and %d should be distinct", i, j)
			}
		}
	}
}

// Verify IsVisible was removed from Satellite (compile-time check).
// If IsVisible still exists on *Satellite, this interface assertion would
// succeed — but we want it to NOT exist, so we don't add such an assertion.
// Instead, we just verify the SatelliteState.Visible field works.
func TestSatelliteHasNoIsVisibleMethod(t *testing.T) {
	type hasIsVisible interface {
		IsVisible() bool
	}
	var s interface{} = &Satellite{}
	if _, ok := s.(hasIsVisible); ok {
		t.Fatal("Satellite should no longer have an IsVisible() method")
	}
}
