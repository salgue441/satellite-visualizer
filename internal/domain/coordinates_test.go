package domain

import (
	"math"
	"testing"
)

func TestECIToGeo_PointOnXAxis(t *testing.T) {
	// Point on X-axis at ~400km altitude, GMST=0
	// Should give lat≈0°, lon≈0°, alt≈400km
	pos := Position{X: 6778, Y: 0, Z: 0}
	geo := ECIToGeo(pos, 0)

	if math.Abs(geo.Latitude) > 1.0 {
		t.Errorf("expected latitude ≈ 0°, got %f°", geo.Latitude)
	}
	if math.Abs(geo.Longitude) > 1.0 {
		t.Errorf("expected longitude ≈ 0°, got %f°", geo.Longitude)
	}
	if math.Abs(geo.Altitude-400) > 10 {
		t.Errorf("expected altitude ≈ 400 km, got %f km", geo.Altitude)
	}
}

func TestECIToGeo_PointOnZAxis(t *testing.T) {
	// Point on Z-axis → north pole, lat≈90°
	pos := Position{X: 0, Y: 0, Z: 6778}
	geo := ECIToGeo(pos, 0)

	if math.Abs(geo.Latitude-90) > 1.0 {
		t.Errorf("expected latitude ≈ 90°, got %f°", geo.Latitude)
	}
}

func TestECIToGeo_ArbitraryPoint(t *testing.T) {
	// Verify reasonable values for an arbitrary ECI position
	pos := Position{X: 4000, Y: 4000, Z: 3000}
	geo := ECIToGeo(pos, 0)

	if geo.Latitude < -90 || geo.Latitude > 90 {
		t.Errorf("latitude out of range: %f°", geo.Latitude)
	}
	if geo.Longitude < -180 || geo.Longitude > 180 {
		t.Errorf("longitude out of range: %f°", geo.Longitude)
	}
	if geo.Altitude < 0 || geo.Altitude > 10000 {
		t.Errorf("altitude out of reasonable range: %f km", geo.Altitude)
	}
	// X=Y=4000 → lon≈45°
	if math.Abs(geo.Longitude-45) > 2.0 {
		t.Errorf("expected longitude ≈ 45°, got %f°", geo.Longitude)
	}
	// Positive Z → positive latitude
	if geo.Latitude <= 0 {
		t.Errorf("expected positive latitude for positive Z, got %f°", geo.Latitude)
	}
}

func TestECIToGeo_NegativeZ(t *testing.T) {
	// Negative Z → southern hemisphere (negative latitude)
	pos := Position{X: 6000, Y: 0, Z: -3000}
	geo := ECIToGeo(pos, 0)

	if geo.Latitude >= 0 {
		t.Errorf("expected negative latitude for negative Z, got %f°", geo.Latitude)
	}
}

func TestIsVisible_DirectlyOverhead(t *testing.T) {
	// Satellite directly overhead at 400km altitude → visible
	geo := GeoCoordinate{Latitude: 40.0, Longitude: -74.0, Altitude: 400}
	if !IsVisible(geo, 40.0, -74.0) {
		t.Error("expected satellite directly overhead to be visible")
	}
}

func TestIsVisible_OppositeSideOfEarth(t *testing.T) {
	// Satellite on opposite side of Earth → not visible
	geo := GeoCoordinate{Latitude: -40.0, Longitude: 106.0, Altitude: 400}
	if IsVisible(geo, 40.0, -74.0) {
		t.Error("expected satellite on opposite side of Earth to not be visible")
	}
}

func TestIsVisible_WithinHorizon(t *testing.T) {
	// Satellite at moderate distance but within horizon (ISS-like altitude)
	// At 400km, horizon distance is ~2300km ground distance (~20° central angle)
	geo := GeoCoordinate{Latitude: 42.0, Longitude: -74.0, Altitude: 400}
	if !IsVisible(geo, 40.0, -74.0) {
		t.Error("expected satellite within horizon to be visible")
	}
}

func TestIsVisible_BeyondHorizon(t *testing.T) {
	// Satellite at low altitude, far away → not visible
	// At 200km altitude, horizon angle is small (~7°)
	// Place satellite ~30° away - should be beyond horizon
	geo := GeoCoordinate{Latitude: 70.0, Longitude: -74.0, Altitude: 200}
	if IsVisible(geo, 40.0, -74.0) {
		t.Error("expected satellite beyond horizon to not be visible")
	}
}
