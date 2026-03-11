package provider

import (
	"context"
	"errors"
	"satellite-visualizer/internal/domain"
	"testing"
)

type mockTLEProvider struct {
	tles      []domain.TLE
	err       error
	available []string
	called    bool
}

func (m *mockTLEProvider) FetchConstellation(_ context.Context, _ string) ([]domain.TLE, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.tles, nil
}

func (m *mockTLEProvider) Available() []string {
	return m.available
}

func TestFailover_PrimarySucceeds(t *testing.T) {
	primary := &mockTLEProvider{
		tles: []domain.TLE{{Name: "ISS", Line1: "l1", Line2: "l2"}},
	}
	secondary := &mockTLEProvider{
		tles: []domain.TLE{{Name: "NOAA", Line1: "l1", Line2: "l2"}},
	}

	fp := NewFailover(primary, secondary, nil)
	result, err := fp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "ISS" {
		t.Fatalf("expected ISS from primary, got %+v", result)
	}
	if secondary.called {
		t.Error("secondary should not have been called")
	}
}

func TestFailover_PrimaryFailsSecondarySucceeds(t *testing.T) {
	primary := &mockTLEProvider{
		err: errors.New("primary down"),
	}
	secondary := &mockTLEProvider{
		tles: []domain.TLE{{Name: "NOAA", Line1: "l1", Line2: "l2"}},
	}

	fp := NewFailover(primary, secondary, nil)
	result, err := fp.FetchConstellation(context.Background(), "weather")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "NOAA" {
		t.Fatalf("expected NOAA from secondary, got %+v", result)
	}
}

func TestFailover_BothFail(t *testing.T) {
	primary := &mockTLEProvider{err: errors.New("primary down")}
	secondary := &mockTLEProvider{err: errors.New("secondary down")}

	fp := NewFailover(primary, secondary, nil)
	_, err := fp.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got: %v", err)
	}
}

func TestFailover_AvailableMergesDeduplicated(t *testing.T) {
	primary := &mockTLEProvider{available: []string{"stations", "weather"}}
	secondary := &mockTLEProvider{available: []string{"weather", "gps"}}

	fp := NewFailover(primary, secondary, nil)
	avail := fp.Available()

	expected := map[string]bool{"stations": true, "weather": true, "gps": true}
	if len(avail) != len(expected) {
		t.Fatalf("expected %d items, got %d: %v", len(expected), len(avail), avail)
	}
	for _, name := range avail {
		if !expected[name] {
			t.Errorf("unexpected item in available: %q", name)
		}
	}
}
