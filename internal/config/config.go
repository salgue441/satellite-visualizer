package config

import (
	"flag"
	"os"
	"strings"
	"time"
)

// AppConfig holds all configurable parameters for the application.
type AppConfig struct {
	// CelesTrakURL is the API endpoint for fetching TLE data from CelesTrak.
	CelesTrakURL string

	// SpaceTrackURL is the API endpoint for the Space-Track.org service.
	SpaceTrackURL string

	// Constellations is the list of satellite groups to track (e.g., "starlink", "stations").
	Constellations []string

	// FetchTimeout is the maximum duration for external API requests.
	FetchTimeout time.Duration

	// FetchInterval is how often fresh TLE data is fetched from upstream sources.
	FetchInterval time.Duration

	// TargetFPS is the desired rendering frame rate.
	TargetFPS int

	// NoColor disables colored terminal output when true.
	NoColor bool

	// ObserverLat is the observer's latitude in decimal degrees.
	ObserverLat float64

	// ObserverLon is the observer's longitude in decimal degrees.
	ObserverLon float64

	// SpaceTrackUser is the Space-Track.org account username.
	SpaceTrackUser string

	// SpaceTrackPass is the Space-Track.org account password.
	SpaceTrackPass string
}

// DefaultConfig returns production defaults.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		CelesTrakURL:   "https://celestrak.org/NORAD/elements/gp.php",
		SpaceTrackURL:  "https://www.space-track.org",
		Constellations: []string{"stations", "starlink"},
		FetchTimeout:   10 * time.Second,
		FetchInterval:  15 * time.Minute,
		TargetFPS:      30,
	}
}

// Load creates config from defaults, env vars, and CLI flags.
// Priority order: defaults < environment variables < CLI flags.
func Load() *AppConfig {
	cfg := DefaultConfig()

	// Env vars
	if u := os.Getenv("SPACETRACK_USER"); u != "" {
		cfg.SpaceTrackUser = u
	}
	if p := os.Getenv("SPACETRACK_PASS"); p != "" {
		cfg.SpaceTrackPass = p
	}

	// CLI flags
	var constellations string
	flag.StringVar(&constellations, "constellations", "", "comma-separated constellation list")
	flag.IntVar(&cfg.TargetFPS, "fps", cfg.TargetFPS, "target frame rate")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "disable colors")
	flag.Float64Var(&cfg.ObserverLat, "observer-lat", 0, "observer latitude")
	flag.Float64Var(&cfg.ObserverLon, "observer-lon", 0, "observer longitude")
	flag.Parse()

	if constellations != "" {
		cfg.Constellations = strings.Split(constellations, ",")
	}

	return cfg
}

// HasSpaceTrack returns true if Space-Track credentials are configured.
func (c *AppConfig) HasSpaceTrack() bool {
	return c.SpaceTrackUser != "" && c.SpaceTrackPass != ""
}
