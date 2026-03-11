package renderer

import (
	"fmt"
	"testing"
)

func TestMajorCities(t *testing.T) {
	cities := []struct {
		name string
		lat, lon float64
		expectLand bool
	}{
		{"New York City", 40.7, -74.0, true},
		{"London", 51.5, -0.1, true},
		{"Tokyo", 35.7, 139.7, true},
		{"Sydney", -33.9, 151.2, true},
		{"Cairo", 30.0, 31.2, true},
		{"São Paulo", -23.5, -46.6, true},
		{"Paris", 48.9, 2.3, true},
		{"Moscow", 55.8, 37.6, true},
		{"Beijing", 39.9, 116.4, true},
		{"Mumbai", 19.1, 72.9, true},
		{"Buenos Aires", -34.6, -58.4, true},
		{"Lagos", 6.5, 3.4, true},
		{"Nairobi", -1.3, 36.8, true},
		{"Cape Town", -34.0, 18.5, true},
		{"Mexico City", 19.4, -99.1, true},
		{"Los Angeles", 34.1, -118.2, true},
		{"Chicago", 41.9, -87.6, true},
		{"Berlin", 52.5, 13.4, true},
		{"Rome", 41.9, 12.5, true},
		{"Delhi", 28.6, 77.2, true},
		{"Jakarta", -6.2, 106.8, true},
		{"Reykjavik", 64.1, -21.9, true},
		{"Havana", 23.1, -82.4, true},
		// Ocean points
		{"Mid Atlantic", 30.0, -40.0, false},
		{"Mid Pacific", 0.0, -160.0, false},
		{"Indian Ocean", -20.0, 70.0, false},
	}

	pass := 0
	fail := 0
	for _, c := range cities {
		result := IsLand(c.lat, c.lon)
		if result != c.expectLand {
			t.Errorf("FAIL: %s (lat=%.1f, lon=%.1f) expected_land=%v got=%v", c.name, c.lat, c.lon, c.expectLand, result)
			fail++
		} else {
			pass++
		}
	}
	fmt.Printf("\nResults: %d passed, %d failed out of %d\n", pass, fail, pass+fail)
}
