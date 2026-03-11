package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/config"
	"satellite-visualizer/internal/infrastructure/celestrak"
	"satellite-visualizer/internal/infrastructure/propagator"
	"satellite-visualizer/internal/infrastructure/provider"
	"satellite-visualizer/internal/infrastructure/spacetrack"
	tuiapp "satellite-visualizer/internal/ui/tui/app"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	// Build TLE provider chain
	celestrakClient := celestrak.NewClient(cfg.CelesTrakURL, cfg.FetchTimeout)
	var tleProv application.TLEProvider = celestrakClient

	if cfg.HasSpaceTrack() {
		spacetrackClient := spacetrack.NewClient(
			cfg.SpaceTrackURL, cfg.SpaceTrackUser, cfg.SpaceTrackPass, cfg.FetchTimeout,
		)
		tleProv = provider.NewFailover(celestrakClient, spacetrackClient, logger)
	}

	cachedProv := provider.NewCached(tleProv, cfg.FetchInterval)

	// Build propagator
	prop := &propagator.SGP4Propagator{}

	// Build tracker
	tracker := application.NewTracker(cachedProv, prop, cfg.Constellations, logger)

	// Build and run TUI
	a := tuiapp.NewApp(tracker, cfg)
	p := tea.NewProgram(a, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
