package domain

import "math"

const (
	// earthRadius is the WGS84 equatorial radius in kilometers.
	earthRadius = 6378.137

	// earthFlattening is the WGS84 flattening factor.
	earthFlattening = 1.0 / 298.257223563
)

// eSquared is the square of the first eccentricity of the WGS84 ellipsoid.
var eSquared = 2*earthFlattening - earthFlattening*earthFlattening

// ECIToGeo converts Earth-Centered Inertial coordinates to geodetic
// (latitude/longitude/altitude). The gmst parameter is the Greenwich Mean
// Sidereal Time in radians.
func ECIToGeo(pos Position, gmst float64) GeoCoordinate {
	r := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y)

	// Longitude: atan2(Y, X) - gmst, normalized to [-pi, pi]
	lon := math.Atan2(pos.Y, pos.X) - gmst
	// Normalize to [-pi, pi]
	lon = math.Mod(lon+math.Pi, 2*math.Pi)
	if lon < 0 {
		lon += 2 * math.Pi
	}
	lon -= math.Pi

	// Iterative latitude computation with WGS84 ellipsoid correction
	lat := math.Atan2(pos.Z, r)
	var N float64
	for i := 0; i < 5; i++ {
		sinLat := math.Sin(lat)
		N = earthRadius / math.Sqrt(1-eSquared*sinLat*sinLat)
		lat = math.Atan2(pos.Z+eSquared*N*sinLat, r)
	}

	// Altitude
	alt := r/math.Cos(lat) - N

	return GeoCoordinate{
		Latitude:  lat * 180.0 / math.Pi,
		Longitude: lon * 180.0 / math.Pi,
		Altitude:  alt,
	}
}

// IsVisible determines if a satellite at the given geographic coordinate is
// visible from an observer at (observerLat, observerLon) using great-circle
// distance and horizon angle. Lat/lon parameters are in degrees.
func IsVisible(geo GeoCoordinate, observerLat, observerLon float64) bool {
	satLat := geo.Latitude * math.Pi / 180.0
	satLon := geo.Longitude * math.Pi / 180.0
	obsLat := observerLat * math.Pi / 180.0
	obsLon := observerLon * math.Pi / 180.0

	dLat := satLat - obsLat
	dLon := satLon - obsLon

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(obsLat)*math.Cos(satLat)*math.Sin(dLon/2)*math.Sin(dLon/2)
	centralAngle := 2 * math.Asin(math.Sqrt(a))

	horizonAngle := math.Acos(earthRadius / (earthRadius + geo.Altitude))

	return centralAngle < horizonAngle
}
