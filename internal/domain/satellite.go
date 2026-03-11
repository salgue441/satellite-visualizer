// Package domain contains the core business entities and rules.
// It has zero dependencies on external frameworks or infrastructure.
package domain

// Position represents an object's location in 3D Cartesian coordinates.
// For orbital mechanics, this is typically measured in kilometers (km) relative
// to the Earth-Centered Inertial (ECI) frame.
type Position struct {
	X, Y, Z float64
}

// Velocity represents an object's velocity vector in 3D Cartesian coordinates.
// Measured in kilometers per second (km/s) in the Earth-Centered Inertial (ECI) frame.
type Velocity struct {
	X, Y, Z float64
}

// GeoCoordinate represents a geographic position on or above the Earth's surface.
// Latitude and Longitude are in degrees, Altitude is in kilometers above sea level.
type GeoCoordinate struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// TLE (Two-Line Element) is the standard format used by NORAD to
// distribute orbital elements of Earth-orbiting objects.
type TLE struct {
	// Name of the object.
	Name string

	// Line1 contains the first line of the TLE.
	Line1 string

	// Line2 contains the second line of the TLE.
	Line2 string
}

// OrbitalElements contains the Keplerian orbital elements extracted from a TLE.
// All angular values (Inclination, RAAN, ArgPerigee, MeanAnomaly) are in radians.
type OrbitalElements struct {
	// Epoch is the Julian date at which the elements are valid.
	Epoch float64

	// BStar is the drag term (inverse Earth radii).
	BStar float64

	// Inclination is the angle between the orbital plane and the equatorial plane (radians).
	Inclination float64

	// RAAN is the Right Ascension of the Ascending Node (radians).
	RAAN float64

	// Eccentricity describes the shape of the orbit (dimensionless, 0 = circular).
	Eccentricity float64

	// ArgPerigee is the Argument of Perigee (radians).
	ArgPerigee float64

	// MeanAnomaly is the mean anomaly at epoch (radians).
	MeanAnomaly float64

	// MeanMotion is the number of revolutions per day.
	MeanMotion float64

	// ElementSetNo is the element set number assigned by NORAD.
	ElementSetNo int

	// NoradCatNo is the NORAD catalog number for the object.
	NoradCatNo int

	// RevolutionNo is the revolution number at epoch.
	RevolutionNo int
}

// Satellite represents an active tracked object in orbit.
// It is an aggregate entity combining its raw identifier (TLE) and
// its calculated state in space.
type Satellite struct {
	Name     string
	RawTLE   TLE
	Position Position
}

// SatelliteState extends Satellite with computed geographic coordinates,
// velocity, visibility, and constellation membership.
type SatelliteState struct {
	Satellite

	// Geo is the satellite's geographic position (lat/lon/alt).
	Geo GeoCoordinate

	// Vel is the satellite's velocity vector in the ECI frame.
	Vel Velocity

	// Visible indicates whether the satellite is visible from the observer's position.
	Visible bool

	// ConstellationName is the name of the constellation this satellite belongs to.
	ConstellationName string
}

// Constellation represents a named group of satellites that share a common purpose
// or operator (e.g., Starlink, GPS, Iridium).
type Constellation struct {
	// Name is the constellation identifier.
	Name string

	// Satellites contains the current state of each satellite in the constellation.
	Satellites []SatelliteState
}
