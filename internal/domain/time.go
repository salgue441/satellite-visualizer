package domain

import (
	"math"
	"time"
)

const (
	twoPi         = 2.0 * math.Pi
	secondsPerDay = 86400.0
)

// JulianDate converts a Go time.Time to Julian Date.
func JulianDate(t time.Time) float64 {
	y := float64(t.Year())
	m := float64(t.Month())
	d := float64(t.Day())
	hour := float64(t.Hour())
	min := float64(t.Minute())
	sec := float64(t.Second())

	jd := 367*y -
		math.Floor(7*(y+math.Floor((m+9)/12))/4) +
		math.Floor(275*m/9) +
		d +
		1721013.5 +
		(hour+min/60+sec/3600)/24

	return jd
}

// GMST returns Greenwich Mean Sidereal Time in radians for the given Julian Date.
func GMST(jd float64) float64 {
	// Julian centuries from J2000
	T := (jd - 2451545.0) / 36525.0

	// GMST in seconds of time
	gmstSec := 67310.54841 +
		(876600*3600+8640184.812866)*T +
		0.093104*T*T -
		6.2e-6*T*T*T

	// Convert from seconds to radians
	gmst := gmstSec * twoPi / secondsPerDay

	// Normalize to [0, 2π)
	gmst = math.Mod(gmst, twoPi)
	if gmst < 0 {
		gmst += twoPi
	}

	return gmst
}
