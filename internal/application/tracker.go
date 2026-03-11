package application

import (
	"context"
	"fmt"
	"log/slog"
	"satellite-visualizer/internal/domain"
	"time"
)

// Tracker coordinates the fetching of telemetry and the calculation of
// orbital physics for multiple constellations.
type Tracker struct {
	provider       TLEProvider
	propagator     OrbitPropagator
	constellations []string
	logger         *slog.Logger
}

// NewTracker creates a new Tracker instance using Dependency Injection.
// By injecting the interfaces, we decouple the business logic from the infrastructure.
func NewTracker(
	prov TLEProvider,
	prop OrbitPropagator,
	constellations []string,
	logger *slog.Logger,
) *Tracker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Tracker{
		provider:       prov,
		propagator:     prop,
		constellations: constellations,
		logger:         logger,
	}
}

// GetConstellations fetches TLE data for all configured constellations,
// propagates orbital elements, and returns the computed state for each satellite.
func (t *Tracker) GetConstellations(
	ctx context.Context,
	now time.Time,
) ([]domain.Constellation, error) {
	jd := domain.JulianDate(now)
	gmst := domain.GMST(jd)

	var result []domain.Constellation
	var lastErr error

	for _, name := range t.constellations {
		tles, err := t.provider.FetchConstellation(ctx, name)
		if err != nil {
			t.logger.Error("Failed to fetch constellation",
				slog.String("constellation", name),
				slog.Any("error", err),
			)
			lastErr = err
			continue
		}

		var states []domain.SatelliteState
		for _, tle := range tles {
			elements, err := domain.ParseTLE(tle.Line1, tle.Line2)
			if err != nil {
				t.logger.Warn("Failed to parse TLE",
					slog.String("name", tle.Name),
					slog.Any("error", err),
				)
				continue
			}

			pos, vel, err := t.propagator.Propagate(elements, now)
			if err != nil {
				t.logger.Warn("Propagation failed",
					slog.String("name", tle.Name),
					slog.Any("error", err),
				)
				continue
			}

			geo := domain.ECIToGeo(pos, gmst)

			states = append(states, domain.SatelliteState{
				Satellite: domain.Satellite{
					Name:     tle.Name,
					RawTLE:   tle,
					Position: pos,
				},
				Geo:               geo,
				Vel:               vel,
				ConstellationName: name,
			})
		}

		if len(states) > 0 {
			result = append(result, domain.Constellation{
				Name:       name,
				Satellites: states,
			})
		}
	}

	if len(result) == 0 && lastErr != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, lastErr)
	}
	if len(result) == 0 {
		return nil, domain.ErrCalculationFailed
	}

	return result, nil
}
