package renderer

import (
	"fmt"
	"testing"
)

func TestDebugPolygons(t *testing.T) {
	// Test each continent individually
	tests := []struct {
		name    string
		lat,lon float64
		polyIdx int
	}{
		// NYC should be in North America (index 0)
		{"NYC in NorthAm", 40.7, -74.0, 0},
		// Sydney in Australia (index 5)
		{"Sydney in Australia", -33.9, 151.2, 5},
		// Paris in Europe (index 3)
		{"Paris in Europe", 48.9, 2.3, 3},
		// Moscow in Asia (index 4) or Europe (index 3)
		{"Moscow in Europe", 55.8, 37.6, 3},
		{"Moscow in Asia", 55.8, 37.6, 4},
		// Mumbai in Asia (index 4)
		{"Mumbai in Asia", 19.1, 72.9, 4},
		// Cape Town in Africa (index 2)
		{"CapeTown in Africa", -34.0, 18.5, 2},
		// Berlin in Europe (index 3)
		{"Berlin in Europe", 52.5, 13.4, 3},
		// Rome in Europe (index 3)
		{"Rome in Europe", 41.9, 12.5, 3},
		// Jakarta in Java (index 10)
		{"Jakarta in Java", -6.2, 106.8, 10},
		// Havana in Cuba (index 21)
		{"Havana in Cuba", 23.1, -82.4, 21},
	}

	for _, tt := range tests {
		result := pointInPolygon(tt.lat, tt.lon, continents[tt.polyIdx])
		fmt.Printf("%s: %v (poly has %d points)\n", tt.name, result, len(continents[tt.polyIdx]))
	}
}
