package application

import (
	"context"
	"errors"
	"satellite-visualizer/internal/domain"
	"testing"
	"time"
)

const (
	issLine1  = "1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990"
	issLine2  = "2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209246"
	noaaLine1 = "1 25338U 98030A   20045.44548192  .00000024  00000-0  24381-4 0  9997"
	noaaLine2 = "2 25338  98.7288  60.0249 0011274  89.0027 271.2352 14.25955633138981"
)

type mockProvider struct {
	tles map[string][]domain.TLE
	err  error
}

func (m *mockProvider) FetchConstellation(_ context.Context, name string) ([]domain.TLE, error) {
	if m.err != nil {
		return nil, m.err
	}
	tles, ok := m.tles[name]
	if !ok {
		return nil, domain.ErrConstellationNotFound
	}
	return tles, nil
}

func (m *mockProvider) Available() []string { return nil }

type mockPropagator struct {
	pos domain.Position
	vel domain.Velocity
	err error
}

func (m *mockPropagator) Propagate(_ domain.OrbitalElements, _ time.Time) (domain.Position, domain.Velocity, error) {
	return m.pos, m.vel, m.err
}

// callCountPropagator tracks calls and can fail on specific invocations.
type callCountPropagator struct {
	pos      domain.Position
	vel      domain.Velocity
	failOn   map[int]bool
	callNum  int
}

func (m *callCountPropagator) Propagate(_ domain.OrbitalElements, _ time.Time) (domain.Position, domain.Velocity, error) {
	m.callNum++
	if m.failOn[m.callNum] {
		return domain.Position{}, domain.Velocity{}, errors.New("propagation error")
	}
	return m.pos, m.vel, nil
}

func TestGetConstellations_Success(t *testing.T) {
	prov := &mockProvider{
		tles: map[string][]domain.TLE{
			"stations": {
				{Name: "ISS", Line1: issLine1, Line2: issLine2},
				{Name: "NOAA 15", Line1: noaaLine1, Line2: noaaLine2},
			},
		},
	}
	prop := &mockPropagator{
		pos: domain.Position{X: 1000, Y: 2000, Z: 3000},
		vel: domain.Velocity{X: 1, Y: 2, Z: 3},
	}

	tracker := NewTracker(prov, prop, []string{"stations"}, nil)
	result, err := tracker.GetConstellations(context.Background(), time.Date(2020, 2, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 constellation, got %d", len(result))
	}
	if result[0].Name != "stations" {
		t.Errorf("expected constellation name 'stations', got %q", result[0].Name)
	}
	if len(result[0].Satellites) != 2 {
		t.Errorf("expected 2 satellites, got %d", len(result[0].Satellites))
	}

	for _, s := range result[0].Satellites {
		if s.ConstellationName != "stations" {
			t.Errorf("expected ConstellationName 'stations', got %q", s.ConstellationName)
		}
		if s.Vel.X != 1 || s.Vel.Y != 2 || s.Vel.Z != 3 {
			t.Errorf("unexpected velocity: %+v", s.Vel)
		}
	}
}

func TestGetConstellations_MultipleConstellations(t *testing.T) {
	prov := &mockProvider{
		tles: map[string][]domain.TLE{
			"stations": {
				{Name: "ISS", Line1: issLine1, Line2: issLine2},
			},
			"weather": {
				{Name: "NOAA 15", Line1: noaaLine1, Line2: noaaLine2},
			},
		},
	}
	prop := &mockPropagator{
		pos: domain.Position{X: 1000, Y: 2000, Z: 3000},
		vel: domain.Velocity{X: 1, Y: 2, Z: 3},
	}

	tracker := NewTracker(prov, prop, []string{"stations", "weather"}, nil)
	result, err := tracker.GetConstellations(context.Background(), time.Date(2020, 2, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 constellations, got %d", len(result))
	}
	if result[0].Name != "stations" {
		t.Errorf("expected first constellation 'stations', got %q", result[0].Name)
	}
	if result[1].Name != "weather" {
		t.Errorf("expected second constellation 'weather', got %q", result[1].Name)
	}
}

func TestGetConstellations_PartialPropagationFailure(t *testing.T) {
	prov := &mockProvider{
		tles: map[string][]domain.TLE{
			"stations": {
				{Name: "ISS", Line1: issLine1, Line2: issLine2},
				{Name: "NOAA 15", Line1: noaaLine1, Line2: noaaLine2},
			},
		},
	}
	// Fail on the first call, succeed on the second.
	prop := &callCountPropagator{
		pos:    domain.Position{X: 1000, Y: 2000, Z: 3000},
		vel:    domain.Velocity{X: 1, Y: 2, Z: 3},
		failOn: map[int]bool{1: true},
	}

	tracker := NewTracker(prov, prop, []string{"stations"}, nil)
	result, err := tracker.GetConstellations(context.Background(), time.Date(2020, 2, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 constellation, got %d", len(result))
	}
	if len(result[0].Satellites) != 1 {
		t.Errorf("expected 1 satellite (partial failure), got %d", len(result[0].Satellites))
	}
	if result[0].Satellites[0].Name != "NOAA 15" {
		t.Errorf("expected surviving satellite 'NOAA 15', got %q", result[0].Satellites[0].Name)
	}
}

func TestGetConstellations_ProviderError(t *testing.T) {
	prov := &mockProvider{
		err: errors.New("network timeout"),
	}
	prop := &mockPropagator{}

	tracker := NewTracker(prov, prop, []string{"stations"}, nil)
	_, err := tracker.GetConstellations(context.Background(), time.Date(2020, 2, 14, 12, 0, 0, 0, time.UTC))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got: %v", err)
	}
}

func TestGetConstellations_AllPropagationFail(t *testing.T) {
	prov := &mockProvider{
		tles: map[string][]domain.TLE{
			"stations": {
				{Name: "ISS", Line1: issLine1, Line2: issLine2},
				{Name: "NOAA 15", Line1: noaaLine1, Line2: noaaLine2},
			},
		},
	}
	prop := &mockPropagator{
		err: errors.New("propagation always fails"),
	}

	tracker := NewTracker(prov, prop, []string{"stations"}, nil)
	_, err := tracker.GetConstellations(context.Background(), time.Date(2020, 2, 14, 12, 0, 0, 0, time.UTC))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrCalculationFailed) {
		t.Errorf("expected ErrCalculationFailed, got: %v", err)
	}
}
