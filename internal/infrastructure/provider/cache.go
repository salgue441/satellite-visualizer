package provider

import (
	"context"
	"log/slog"
	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/domain"
	"sync"
	"time"
)

// Compile-time interface check.
var _ application.TLEProvider = (*CachedProvider)(nil)

// CachedProvider wraps a TLEProvider with a thread-safe in-memory cache.
type CachedProvider struct {
	inner    application.TLEProvider
	cache    map[string]cacheEntry
	mu       sync.RWMutex
	staleTTL time.Duration
}

type cacheEntry struct {
	tles      []domain.TLE
	fetchedAt time.Time
}

// NewCached creates a CachedProvider that caches results from inner for staleTTL.
func NewCached(inner application.TLEProvider, staleTTL time.Duration) *CachedProvider {
	return &CachedProvider{
		inner:    inner,
		cache:    make(map[string]cacheEntry),
		staleTTL: staleTTL,
	}
}

// FetchConstellation returns cached data if fresh, otherwise fetches from inner.
// If inner fetch fails but cache exists (even stale), returns cached data.
func (c *CachedProvider) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error) {
	// Check for fresh cache entry.
	c.mu.RLock()
	entry, exists := c.cache[name]
	c.mu.RUnlock()

	if exists && time.Since(entry.fetchedAt) < c.staleTTL {
		return entry.tles, nil
	}

	// Cache miss or stale: fetch from inner.
	tles, err := c.inner.FetchConstellation(ctx, name)
	if err == nil {
		c.mu.Lock()
		c.cache[name] = cacheEntry{
			tles:      tles,
			fetchedAt: time.Now(),
		}
		c.mu.Unlock()
		return tles, nil
	}

	// Fetch failed: return stale cache if available.
	if exists {
		slog.Warn("fetch failed, returning stale cache",
			"constellation", name,
			"error", err,
			"cache_age", time.Since(entry.fetchedAt),
		)
		return entry.tles, nil
	}

	// No cache at all.
	return nil, err
}

// Available delegates to the inner provider.
func (c *CachedProvider) Available() []string {
	return c.inner.Available()
}
