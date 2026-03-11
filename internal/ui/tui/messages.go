package tui

import (
	"satellite-visualizer/internal/domain"
	"time"
)

// TickMsg triggers a propagation + render cycle.
type TickMsg time.Time

// DataUpdateMsg delivers fresh TLE data from background fetchers.
type DataUpdateMsg struct {
	Constellation string
	TLEs          []domain.TLE
}

// SelectMsg indicates the user selected a satellite.
type SelectMsg struct {
	Satellite domain.SatelliteState
}

// ErrMsg wraps errors from background operations.
type ErrMsg struct {
	Err error
}

// FetchCompleteMsg indicates a background fetch completed.
type FetchCompleteMsg struct {
	Constellations []domain.Constellation
}
