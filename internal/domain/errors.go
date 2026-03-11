package domain

import "errors"

var (
	// ErrInvalidTle is returned when the provided TLE strings are malformed.
	ErrInvalidTle = errors.New("invalid or malformed TLE data")

	// ErrCalculationFailed is returned when the orbital propagation algorithm fails
	// to converge on a valid position for a given timestamp.
	ErrCalculationFailed = errors.New("orbital propagation algorithm failed to converge")

	// ErrConstellationNotFound is returned when the requested constellation is not found.
	ErrConstellationNotFound = errors.New("constellation not found")

	// ErrStaleData is returned when the cached data is too old to be considered reliable.
	ErrStaleData = errors.New("data is stale and needs to be refreshed")

	// ErrAuthFailed is returned when authentication with an external provider fails.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrProviderUnavailable is returned when an external data provider cannot be reached.
	ErrProviderUnavailable = errors.New("data provider is unavailable")
)
