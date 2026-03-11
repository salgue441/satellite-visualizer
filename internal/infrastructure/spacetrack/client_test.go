package spacetrack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/domain"
)

// Compile-time interface check.
var _ application.TLEProvider = (*Client)(nil)

const testTLEResponse = `ISS (ZARYA)
1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990
2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209246
NOAA 18
1 28654U 05018A   20045.21163223  .00000058  00000-0  52593-4 0  9993
2 28654  99.0373  72.5450 0013889 259.8738 100.0828 14.12587775762498`

func TestFetchConstellation_AuthAndFetch(t *testing.T) {
	var loginCalled atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ajaxauth/login":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST for login, got %s", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.FormValue("identity") != "testuser" {
				t.Errorf("expected identity=testuser, got %s", r.FormValue("identity"))
			}
			if r.FormValue("password") != "testpass" {
				t.Errorf("expected password=testpass, got %s", r.FormValue("password"))
			}
			loginCalled.Add(1)
			http.SetCookie(w, &http.Cookie{
				Name:  "chocolatechip",
				Value: "session-token-123",
			})
			w.WriteHeader(http.StatusOK)

		default:
			// Verify auth cookie is present
			cookie, err := r.Cookie("chocolatechip")
			if err != nil || cookie.Value != "session-token-123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testTLEResponse))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "testuser", "testpass", 5*time.Second)
	tles, err := client.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loginCalled.Load() != 1 {
		t.Errorf("expected login to be called once, got %d", loginCalled.Load())
	}

	if len(tles) != 2 {
		t.Fatalf("expected 2 TLEs, got %d", len(tles))
	}

	if tles[0].Name != "ISS (ZARYA)" {
		t.Errorf("expected name ISS (ZARYA), got %s", tles[0].Name)
	}
	if tles[0].Line1[0] != '1' {
		t.Errorf("expected line1 to start with '1', got %c", tles[0].Line1[0])
	}
	if tles[1].Name != "NOAA 18" {
		t.Errorf("expected name NOAA 18, got %s", tles[1].Name)
	}

	// Second call should reuse the session (no extra login)
	_, err = client.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if loginCalled.Load() != 1 {
		t.Errorf("expected login to still be 1, got %d (session reuse failed)", loginCalled.Load())
	}
}

func TestFetchConstellation_AuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "baduser", "badpass", 5*time.Second)
	_, err := client.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrAuthFailed) {
		t.Errorf("expected ErrAuthFailed, got %v", err)
	}
}

func TestFetchConstellation_EmptyCredentials(t *testing.T) {
	client := NewClient("http://example.com", "", "", 5*time.Second)
	_, err := client.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrAuthFailed) {
		t.Errorf("expected ErrAuthFailed, got %v", err)
	}
}

func TestFetchConstellation_RetryOnExpiredSession(t *testing.T) {
	var fetchCount atomic.Int32
	var loginCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ajaxauth/login":
			loginCount.Add(1)
			http.SetCookie(w, &http.Cookie{
				Name:  "chocolatechip",
				Value: "new-session",
			})
			w.WriteHeader(http.StatusOK)
		default:
			count := fetchCount.Add(1)
			if count == 1 {
				// First fetch returns 401 (expired session)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testTLEResponse))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "user", "pass", 5*time.Second)
	// Pre-set a stale cookie to trigger the retry path
	client.authCookie = &http.Cookie{Name: "chocolatechip", Value: "stale"}

	tles, err := client.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tles) != 2 {
		t.Fatalf("expected 2 TLEs, got %d", len(tles))
	}
	if loginCount.Load() != 1 {
		t.Errorf("expected 1 re-auth, got %d", loginCount.Load())
	}
}

func TestAvailable(t *testing.T) {
	client := NewClient("http://example.com", "u", "p", 5*time.Second)
	available := client.Available()

	if len(available) == 0 {
		t.Fatal("expected non-empty list")
	}

	has := func(name string) bool {
		for _, n := range available {
			if n == name {
				return true
			}
		}
		return false
	}

	if !has("stations") {
		t.Error("expected 'stations' in available list")
	}
	if !has("starlink") {
		t.Error("expected 'starlink' in available list")
	}
}
