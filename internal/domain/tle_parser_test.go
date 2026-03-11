package domain

import (
	"math"
	"testing"
)

const (
	issLine1 = "1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990"
	issLine2 = "2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209246"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestParseTLE_ValidISS(t *testing.T) {
	elems, err := ParseTLE(issLine1, issLine2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elems.NoradCatNo != 25544 {
		t.Errorf("NoradCatNo: got %d, want 25544", elems.NoradCatNo)
	}
	if elems.ElementSetNo != 999 {
		t.Errorf("ElementSetNo: got %d, want 999", elems.ElementSetNo)
	}
	if elems.RevolutionNo != 20924 {
		t.Errorf("RevolutionNo: got %d, want 20924", elems.RevolutionNo)
	}

	// Eccentricity
	if !almostEqual(elems.Eccentricity, 0.0004615, 1e-7) {
		t.Errorf("Eccentricity: got %v, want 0.0004615", elems.Eccentricity)
	}

	// MeanMotion (rev/day)
	if !almostEqual(elems.MeanMotion, 15.49163961, 1e-6) {
		t.Errorf("MeanMotion: got %v, want 15.49163961", elems.MeanMotion)
	}

	// Angles in radians
	deg2rad := math.Pi / 180.0

	if !almostEqual(elems.Inclination, 51.6443*deg2rad, 1e-4) {
		t.Errorf("Inclination: got %v, want ~%v", elems.Inclination, 51.6443*deg2rad)
	}
	if !almostEqual(elems.RAAN, 242.7420*deg2rad, 1e-4) {
		t.Errorf("RAAN: got %v, want ~%v", elems.RAAN, 242.7420*deg2rad)
	}
	if !almostEqual(elems.ArgPerigee, 225.0295*deg2rad, 1e-4) {
		t.Errorf("ArgPerigee: got %v, want ~%v", elems.ArgPerigee, 225.0295*deg2rad)
	}
	if !almostEqual(elems.MeanAnomaly, 296.6842*deg2rad, 1e-4) {
		t.Errorf("MeanAnomaly: got %v, want ~%v", elems.MeanAnomaly, 296.6842*deg2rad)
	}

	// BStar
	if !almostEqual(elems.BStar, 0.25302e-4, 1e-9) {
		t.Errorf("BStar: got %v, want 0.25302e-4", elems.BStar)
	}

	// Epoch (Julian Date for 2020, day 45.18587073)
	// JD of Jan 1 2020 = 2458849.5, epoch = 2458849.5 + 45.18587073 - 1 = 2458893.68587073
	expectedEpoch := 2458849.5 + 45.18587073 - 1.0
	if !almostEqual(elems.Epoch, expectedEpoch, 1e-5) {
		t.Errorf("Epoch: got %v, want ~%v", elems.Epoch, expectedEpoch)
	}
}

func TestParseTLE_ChecksumValid(t *testing.T) {
	// ISS lines should pass checksum
	_, err := ParseTLE(issLine1, issLine2)
	if err != nil {
		t.Errorf("valid checksum rejected: %v", err)
	}
}

func TestParseTLE_ChecksumInvalidLine1(t *testing.T) {
	// Corrupt last digit of line1
	bad := issLine1[:68] + "1"
	_, err := ParseTLE(bad, issLine2)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for bad line1 checksum, got %v", err)
	}
}

func TestParseTLE_ChecksumInvalidLine2(t *testing.T) {
	// Corrupt last digit of line2
	bad := issLine2[:68] + "0"
	_, err := ParseTLE(issLine1, bad)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for bad line2 checksum, got %v", err)
	}
}

func TestParseTLE_LineTooShort(t *testing.T) {
	_, err := ParseTLE(issLine1[:60], issLine2)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for short line1, got %v", err)
	}

	_, err = ParseTLE(issLine1, issLine2[:60])
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for short line2, got %v", err)
	}
}

func TestParseTLE_LineTooLong(t *testing.T) {
	_, err := ParseTLE(issLine1+"X", issLine2)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for long line1, got %v", err)
	}

	_, err = ParseTLE(issLine1, issLine2+"X")
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for long line2, got %v", err)
	}
}

func TestParseTLE_WrongLineNumber(t *testing.T) {
	// Swap line numbers
	bad1 := "2" + issLine1[1:]
	_, err := ParseTLE(bad1, issLine2)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for line1 starting with '2', got %v", err)
	}

	bad2 := "1" + issLine2[1:]
	_, err = ParseTLE(issLine1, bad2)
	if err != ErrInvalidTle {
		t.Errorf("expected ErrInvalidTle for line2 starting with '1', got %v", err)
	}
}

func TestParseTLE_AnotherSatellite(t *testing.T) {
	// NOAA 15 TLE
	line1 := "1 25338U 98030A   20045.44548192  .00000024  00000-0  24381-4 0  9997"
	line2 := "2 25338  98.7288  60.0249 0011274  89.0027 271.2352 14.25955633138981"

	elems, err := ParseTLE(line1, line2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elems.NoradCatNo != 25338 {
		t.Errorf("NoradCatNo: got %d, want 25338", elems.NoradCatNo)
	}

	if !almostEqual(elems.Eccentricity, 0.0011274, 1e-7) {
		t.Errorf("Eccentricity: got %v, want 0.0011274", elems.Eccentricity)
	}

	if !almostEqual(elems.MeanMotion, 14.25955633, 1e-6) {
		t.Errorf("MeanMotion: got %v, want 14.25955633", elems.MeanMotion)
	}

	deg2rad := math.Pi / 180.0
	if !almostEqual(elems.Inclination, 98.7288*deg2rad, 1e-4) {
		t.Errorf("Inclination: got %v, want ~%v", elems.Inclination, 98.7288*deg2rad)
	}

	if elems.RevolutionNo != 13898 {
		t.Errorf("RevolutionNo: got %d, want 13898", elems.RevolutionNo)
	}
}
