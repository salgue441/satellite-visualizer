package provider

import (
	"context"
	"errors"
	"satellite-visualizer/internal/domain"
	"sync"
	"testing"
	"time"
)

// callCountProvider tracks how many times FetchConstellation is called.
type callCountProvider struct {
	tles      []domain.TLE
	err       error
	available []string
	calls     int
}

func (m *callCountProvider) FetchConstellation(_ context.Context, _ string) ([]domain.TLE, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.tles, nil
}

func (m *callCountProvider) Available() []string {
	return m.available
}

func TestCache_FreshHit(t *testing.T) {
	inner := &callCountProvider{
		tles: []domain.TLE{{Name: "ISS", Line1: "l1", Line2: "l2"}},
	}

	cp := NewCached(inner, 5*time.Minute)

	// First call should fetch from inner.
	result, err := cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "ISS" {
		t.Fatalf("expected ISS, got %+v", result)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}

	// Second call should hit cache, not call inner again.
	result, err = cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "ISS" {
		t.Fatalf("expected ISS from cache, got %+v", result)
	}
	if inner.calls != 1 {
		t.Fatalf("expected still 1 inner call (cache hit), got %d", inner.calls)
	}
}

func TestCache_StaleTriggersRefetch(t *testing.T) {
	inner := &callCountProvider{
		tles: []domain.TLE{{Name: "ISS", Line1: "l1", Line2: "l2"}},
	}

	cp := NewCached(inner, 10*time.Millisecond)

	// First fetch populates cache.
	_, err := cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for cache to go stale.
	time.Sleep(20 * time.Millisecond)

	// Should refetch from inner.
	_, err = cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 2 {
		t.Fatalf("expected 2 inner calls after stale, got %d", inner.calls)
	}
}

func TestCache_FetchFailsWithStaleCache(t *testing.T) {
	inner := &callCountProvider{
		tles: []domain.TLE{{Name: "ISS", Line1: "l1", Line2: "l2"}},
	}

	cp := NewCached(inner, 10*time.Millisecond)

	// Populate cache.
	_, err := cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for cache to go stale, then make inner fail.
	time.Sleep(20 * time.Millisecond)
	inner.err = errors.New("network error")

	// Should return stale cached data.
	result, err := cp.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("expected stale cache fallback, got error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "ISS" {
		t.Fatalf("expected stale ISS data, got %+v", result)
	}
}

func TestCache_FetchFailsNoCache(t *testing.T) {
	inner := &callCountProvider{
		err: errors.New("network error"),
	}

	cp := NewCached(inner, 5*time.Minute)

	_, err := cp.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCache_ThreadSafety(t *testing.T) {
	inner := &callCountProvider{
		tles: []domain.TLE{{Name: "ISS", Line1: "l1", Line2: "l2"}},
	}

	cp := NewCached(inner, 1*time.Millisecond)

	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = cp.FetchConstellation(context.Background(), "stations")
		}()
	}

	wg.Wait()
	// If we reach here without panicking, the test passes.
}
