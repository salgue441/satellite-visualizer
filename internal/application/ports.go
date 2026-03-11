package application

import (
	"context"
	"satellite-visualizer/internal/domain"
	"time"
)

// TLEProvider defines the contract for fetching satellite telemetry.
// Whether this data comes from CelesTrak, Space-Track, a local file,
// or a mock during testing, the application layer does not care.
type TLEProvider interface {
	// FetchConstellation retrieves TLEs for a named satellite group.
	FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)

	// Available returns the list of constellation group names this provider supports.
	Available() []string
}

// OrbitPropagator defines the contract for the physics engine.
// It abstracts the mathematical calculations required to predict
// an object's position and velocity in 3D space.
type OrbitPropagator interface {
	// Propagate takes parsed orbital elements and a timestamp, returning
	// the predicted ECI position and velocity.
	Propagate(elements domain.OrbitalElements, t time.Time) (domain.Position, domain.Velocity, error)
}
