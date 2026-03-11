package config

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("Constellations", func(t *testing.T) {
		want := []string{"stations", "starlink"}
		if len(cfg.Constellations) != len(want) {
			t.Fatalf("got %d constellations, want %d", len(cfg.Constellations), len(want))
		}
		for i, v := range want {
			if cfg.Constellations[i] != v {
				t.Errorf("Constellations[%d] = %q, want %q", i, cfg.Constellations[i], v)
			}
		}
	})

	t.Run("TargetFPS", func(t *testing.T) {
		if cfg.TargetFPS != 30 {
			t.Errorf("TargetFPS = %d, want 30", cfg.TargetFPS)
		}
	})

	t.Run("CelesTrakURL", func(t *testing.T) {
		if !strings.HasPrefix(cfg.CelesTrakURL, "https://celestrak") {
			t.Errorf("CelesTrakURL = %q, want prefix %q", cfg.CelesTrakURL, "https://celestrak")
		}
	})

	t.Run("FetchTimeout", func(t *testing.T) {
		if cfg.FetchTimeout != 10*time.Second {
			t.Errorf("FetchTimeout = %v, want 10s", cfg.FetchTimeout)
		}
	})

	t.Run("FetchInterval", func(t *testing.T) {
		if cfg.FetchInterval != 15*time.Minute {
			t.Errorf("FetchInterval = %v, want 15m", cfg.FetchInterval)
		}
	})

	t.Run("NoColor", func(t *testing.T) {
		if cfg.NoColor {
			t.Error("NoColor = true, want false")
		}
	})

	t.Run("ObserverLat", func(t *testing.T) {
		if cfg.ObserverLat != 0 {
			t.Errorf("ObserverLat = %f, want 0", cfg.ObserverLat)
		}
	})

	t.Run("ObserverLon", func(t *testing.T) {
		if cfg.ObserverLon != 0 {
			t.Errorf("ObserverLon = %f, want 0", cfg.ObserverLon)
		}
	})
}

func TestHasSpaceTrack_WithCredentials(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SpaceTrackUser = "user@example.com"
	cfg.SpaceTrackPass = "secret"

	if !cfg.HasSpaceTrack() {
		t.Error("HasSpaceTrack() = false, want true when both credentials set")
	}
}

func TestHasSpaceTrack_MissingUser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SpaceTrackPass = "secret"

	if cfg.HasSpaceTrack() {
		t.Error("HasSpaceTrack() = true, want false when user is missing")
	}
}

func TestHasSpaceTrack_MissingPass(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SpaceTrackUser = "user@example.com"

	if cfg.HasSpaceTrack() {
		t.Error("HasSpaceTrack() = true, want false when password is missing")
	}
}

func TestHasSpaceTrack_NoCreds(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.HasSpaceTrack() {
		t.Error("HasSpaceTrack() = true, want false when no credentials set")
	}
}
