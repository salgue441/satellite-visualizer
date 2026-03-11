package provider

import (
	"context"
	"fmt"
	"log/slog"
	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/domain"
)

// Compile-time interface check.
var _ application.TLEProvider = (*FailoverProvider)(nil)

// FailoverProvider wraps two TLEProviders, trying primary first, falling back to secondary.
type FailoverProvider struct {
	primary   application.TLEProvider
	secondary application.TLEProvider
	logger    *slog.Logger
}

// NewFailover creates a FailoverProvider that tries primary first, then secondary.
func NewFailover(primary, secondary application.TLEProvider, logger *slog.Logger) *FailoverProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &FailoverProvider{
		primary:   primary,
		secondary: secondary,
		logger:    logger,
	}
}

// FetchConstellation tries primary first, falls back to secondary.
// Returns ErrProviderUnavailable only if both fail.
func (f *FailoverProvider) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error) {
	tles, err := f.primary.FetchConstellation(ctx, name)
	if err == nil {
		return tles, nil
	}

	f.logger.Warn("primary provider failed, trying secondary",
		"constellation", name,
		"error", err,
	)

	tles, secErr := f.secondary.FetchConstellation(ctx, name)
	if secErr == nil {
		return tles, nil
	}

	f.logger.Error("both providers failed",
		"constellation", name,
		"primary_error", err,
		"secondary_error", secErr,
	)

	return nil, fmt.Errorf("both providers failed: primary: %v, secondary: %v: %w",
		err, secErr, domain.ErrProviderUnavailable)
}

// Available returns merged deduplicated list from both providers.
func (f *FailoverProvider) Available() []string {
	seen := make(map[string]struct{})
	var result []string

	for _, name := range f.primary.Available() {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			result = append(result, name)
		}
	}
	for _, name := range f.secondary.Available() {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			result = append(result, name)
		}
	}

	return result
}
