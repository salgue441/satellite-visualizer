package celestrak

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"satellite-visualizer/internal/domain"
)

const testTLEResponse = `ISS (ZARYA)
1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990
2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209246
NOAA 18
1 28654U 05018A   20045.21163223  .00000058  00000-0  52593-4 0  9993
2 28654  99.0373  72.5450 0013889 259.8738 100.0828 14.12587775762498
GOES 16
1 41866U 16071A   20045.12345678  .00000124  00000-0  00000+0 0  9995
2 41866   0.0558 269.6498 0000392 332.4346 167.9135  1.00270608121110`

func TestFetchConstellation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("GROUP")
		if group != "stations" {
			t.Errorf("expected GROUP=stations, got GROUP=%s", group)
		}
		format := r.URL.Query().Get("FORMAT")
		if format != "tle" {
			t.Errorf("expected FORMAT=tle, got FORMAT=%s", format)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTLEResponse))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, 5*time.Second)
	tles, err := client.FetchConstellation(context.Background(), "stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tles) != 3 {
		t.Fatalf("expected 3 TLEs, got %d", len(tles))
	}

	if tles[0].Name != "ISS (ZARYA)" {
		t.Errorf("expected name ISS (ZARYA), got %s", tles[0].Name)
	}
	if tles[0].Line1[0] != '1' {
		t.Errorf("expected line1 to start with '1', got %c", tles[0].Line1[0])
	}
	if tles[0].Line2[0] != '2' {
		t.Errorf("expected line2 to start with '2', got %c", tles[0].Line2[0])
	}

	if tles[1].Name != "NOAA 18" {
		t.Errorf("expected name NOAA 18, got %s", tles[1].Name)
	}
	if tles[2].Name != "GOES 16" {
		t.Errorf("expected name GOES 16, got %s", tles[2].Name)
	}
}

func TestFetchConstellation_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, 5*time.Second)
	_, err := client.FetchConstellation(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrConstellationNotFound) {
		t.Errorf("expected ErrConstellationNotFound, got %v", err)
	}
}

func TestFetchConstellation_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, 5*time.Second)
	_, err := client.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, domain.ErrConstellationNotFound) {
		t.Error("should not be ErrConstellationNotFound for 500")
	}
}

func TestFetchConstellation_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Only 2 lines — not divisible by 3
		w.Write([]byte("ISS (ZARYA)\n1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, 5*time.Second)
	_, err := client.FetchConstellation(context.Background(), "stations")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidTle) {
		t.Errorf("expected ErrInvalidTle, got %v", err)
	}
}

func TestAvailable(t *testing.T) {
	client := NewClient("http://example.com", 5*time.Second)
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

func TestFetchConstellation_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response — context should cancel before this completes
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTLEResponse))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.FetchConstellation(ctx, "stations")
	if err == nil {
		t.Fatal("expected error from canceled context, got nil")
	}
}
