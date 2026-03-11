package propagator

import (
	"math"
	"testing"
	"time"
)

func TestJulianDate_J2000Epoch(t *testing.T) {
	// 2000-01-01 12:00 UTC = JD 2451545.0 (J2000 epoch)
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	got := JulianDate(j2000)
	want := 2451545.0
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("JulianDate(J2000) = %f, want %f", got, want)
	}
}

func TestJulianDate_UnixEpoch(t *testing.T) {
	// 1970-01-01 00:00 UTC = JD 2440587.5
	unix := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	got := JulianDate(unix)
	want := 2440587.5
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("JulianDate(Unix epoch) = %f, want %f", got, want)
	}
}

func TestGMST_AtJ2000(t *testing.T) {
	// At J2000 epoch (JD 2451545.0), GMST ≈ 4.8949612 radians
	got := GMST(2451545.0)
	want := 4.8949612
	if math.Abs(got-want) > 0.001 {
		t.Errorf("GMST(J2000) = %f, want %f", got, want)
	}
}

func TestGMST_InRange(t *testing.T) {
	// GMST should always be in [0, 2π)
	testJDs := []float64{2451545.0, 2451546.0, 2460000.0, 2440587.5}
	for _, jd := range testJDs {
		got := GMST(jd)
		if got < 0 || got >= TwoPi {
			t.Errorf("GMST(%f) = %f, not in [0, 2π)", jd, got)
		}
	}
}

func TestMinutesSinceEpoch_SameJD(t *testing.T) {
	jd := 2451545.0
	got := MinutesSinceEpoch(jd, jd)
	if got != 0 {
		t.Errorf("MinutesSinceEpoch(same JD) = %f, want 0", got)
	}
}

func TestMinutesSinceEpoch_OneDayApart(t *testing.T) {
	epoch := 2451545.0
	target := 2451546.0
	got := MinutesSinceEpoch(epoch, target)
	want := 1440.0
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("MinutesSinceEpoch(1 day apart) = %f, want %f", got, want)
	}
}
