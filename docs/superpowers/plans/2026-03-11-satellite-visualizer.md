# Satellite Orbit Visualizer Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a terminal-based satellite orbit visualizer with real-time 3D globe rendering, custom SGP4 propagation, and live TLE data from CelesTrak/Space-Track.

**Architecture:** Hexagonal architecture with domain entities, application ports, infrastructure adapters (CelesTrak, Space-Track, SGP4), and a Bubbletea TUI with dashboard layout. Background goroutines handle data fetching; the main loop handles propagation and rendering at ~30fps.

**Tech Stack:** Go 1.25, Bubbletea/Lipgloss/Bubbles, custom SGP4, no external math/orbital libraries.

**Spec:** `docs/superpowers/specs/2026-03-11-satellite-visualizer-design.md`

---

## Chunk 1: Domain Layer

### Task 1: Update Domain Entities

**Files:**
- Modify: `internal/domain/satellite.go`
- Modify: `internal/domain/errors.go`
- Test: `internal/domain/satellite_test.go`

- [ ] **Step 1: Write tests for domain entities**

Create `internal/domain/satellite_test.go`:

```go
package domain

import (
	"math"
	"testing"
)

func TestPositionZeroValue(t *testing.T) {
	var p Position
	if p.X != 0 || p.Y != 0 || p.Z != 0 {
		t.Error("zero-value Position should have all zeros")
	}
}

func TestVelocityMagnitude(t *testing.T) {
	v := Velocity{X: 3, Y: 4, Z: 0}
	mag := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if math.Abs(mag-5.0) > 1e-9 {
		t.Errorf("expected magnitude 5, got %f", mag)
	}
}

func TestGeoCoordinateRanges(t *testing.T) {
	g := GeoCoordinate{Latitude: 45.0, Longitude: -73.0, Altitude: 408.0}
	if g.Latitude < -90 || g.Latitude > 90 {
		t.Errorf("latitude out of range: %f", g.Latitude)
	}
	if g.Longitude < -180 || g.Longitude > 180 {
		t.Errorf("longitude out of range: %f", g.Longitude)
	}
}

func TestSatelliteStateEmbed(t *testing.T) {
	sat := Satellite{Name: "ISS", Position: Position{X: 1, Y: 2, Z: 3}}
	state := SatelliteState{
		Satellite:         sat,
		Geo:               GeoCoordinate{Latitude: 51.5, Longitude: -0.1, Altitude: 408},
		Vel:               Velocity{X: 7.0, Y: 0.5, Z: 0.1},
		Visible:           true,
		ConstellationName: "stations",
	}
	if state.Name != "ISS" {
		t.Errorf("embedded name should be ISS, got %s", state.Name)
	}
	if !state.Visible {
		t.Error("state should be visible")
	}
}

func TestConstellationSatelliteCount(t *testing.T) {
	c := Constellation{
		Name: "test",
		Satellites: []SatelliteState{
			{Satellite: Satellite{Name: "A"}},
			{Satellite: Satellite{Name: "B"}},
		},
	}
	if len(c.Satellites) != 2 {
		t.Errorf("expected 2 satellites, got %d", len(c.Satellites))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/carlossalguero/dev/apps/satellite-visualizer && go test ./internal/domain/ -v`
Expected: compilation errors for undefined types

- [ ] **Step 3: Update satellite.go with new entities**

Replace `internal/domain/satellite.go` with:

```go
// Package domain contains the core business entities and rules.
// It has zero dependencies on external frameworks or infrastructure.
package domain

// Position represents an object's location in 3D Cartesian coordinates.
// For orbital mechanics, this is typically measured in kilometers (km) relative
// to the Earth-Centered Inertial (ECI) frame.
type Position struct {
	X, Y, Z float64
}

// Velocity represents an object's velocity vector in km/s
// in the Earth-Centered Inertial (ECI) frame.
type Velocity struct {
	X, Y, Z float64
}

// GeoCoordinate represents a geodetic position on or above the Earth's surface.
type GeoCoordinate struct {
	Latitude  float64 // degrees, -90 to 90
	Longitude float64 // degrees, -180 to 180
	Altitude  float64 // km above WGS84 ellipsoid
}

// TLE (Two-Line Element) is the standard format used by NORAD to
// distribute orbital elements of Earth-orbiting objects.
type TLE struct {
	Name  string
	Line1 string
	Line2 string
}

// OrbitalElements contains the Keplerian elements parsed from a TLE.
// These are the inputs to the SGP4 propagation algorithm.
type OrbitalElements struct {
	Epoch           float64 // Julian date of element set epoch
	BStar           float64 // drag coefficient (1/earth-radii)
	Inclination     float64 // radians
	RAAN            float64 // Right Ascension of Ascending Node (radians)
	Eccentricity    float64 // dimensionless, 0 to ~1
	ArgPerigee      float64 // Argument of perigee (radians)
	MeanAnomaly     float64 // radians
	MeanMotion      float64 // revolutions per day
	ElementSetNo    int
	NoradCatNo      int
	RevolutionNo    int
}

// Satellite represents an active tracked object in orbit.
// It is an aggregate entity combining its raw identifier (TLE) and
// its calculated state in space.
type Satellite struct {
	Name     string
	RawTLE   TLE
	Position Position
}

// SatelliteState is the enriched satellite aggregate used by the UI layer.
// It embeds Satellite for ECI position (needed by the 3D renderer) and adds
// geodetic coordinates (needed by the details panel), velocity, and metadata.
type SatelliteState struct {
	Satellite
	Geo               GeoCoordinate
	Vel               Velocity
	Visible           bool
	ConstellationName string
}

// Constellation groups satellites by their orbital constellation.
type Constellation struct {
	Name       string
	Satellites []SatelliteState
}
```

- [ ] **Step 4: Update errors.go with new error types**

Replace `internal/domain/errors.go` with:

```go
package domain

import "errors"

var (
	// ErrInvalidTle is returned when the provided TLE strings are malformed.
	ErrInvalidTle = errors.New("invalid or malformed TLE data")

	// ErrCalculationFailed is returned when the orbital propagation algorithm fails
	// to converge on a valid position for a given timestamp.
	ErrCalculationFailed = errors.New("orbital propagation algorithm failed to converge")

	// ErrConstellationNotFound is returned when the requested constellation is not found.
	ErrConstellationNotFound = errors.New("constellation not found")

	// ErrStaleData is returned when TLE cache age exceeds the acceptable threshold.
	ErrStaleData = errors.New("TLE data is stale")

	// ErrAuthFailed is returned when Space-Track authentication fails.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrProviderUnavailable is returned when all data sources are unreachable.
	ErrProviderUnavailable = errors.New("all data providers are unavailable")
)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/carlossalguero/dev/apps/satellite-visualizer && go test ./internal/domain/ -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/domain/satellite.go internal/domain/errors.go internal/domain/satellite_test.go
git commit -m "feat(domain): add new entities and error types for satellite state, velocity, geo coords"
```

---

### Task 2: TLE Parser

**Files:**
- Create: `internal/domain/tle_parser.go`
- Test: `internal/domain/tle_parser_test.go`

- [ ] **Step 1: Write TLE parser tests**

Create `internal/domain/tle_parser_test.go` with tests for:
- Valid TLE parsing (ISS reference TLE) extracting all orbital elements
- Checksum validation (valid line passes, corrupted line fails with ErrInvalidTle)
- Line length validation (lines must be 69 chars)
- Line number validation (line1 starts with '1', line2 starts with '2')
- Field extraction accuracy: inclination, eccentricity, mean motion, RAAN, etc.

Use this reference ISS TLE for tests:
```
Line1: "1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990"
Line2: "2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209242"
```

Expected parsed values:
- Inclination: 51.6443 degrees (convert to radians)
- Eccentricity: 0.0004615
- MeanMotion: 15.49163961 rev/day
- RAAN: 242.7420 degrees (convert to radians)
- NoradCatNo: 25544

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/ -run TestParseTLE -v`
Expected: compilation errors

- [ ] **Step 3: Implement TLE parser**

Create `internal/domain/tle_parser.go` implementing:
- `ParseTLE(line1, line2 string) (OrbitalElements, error)` — main entry point
- `validateChecksum(line string) bool` — TLE checksum: sum digits, '-' counts as 1, mod 10, compare to last char
- Field extraction using fixed-column positions per TLE spec:
  - Line 1: catalog number (cols 3-7), epoch year+day (cols 19-32), bstar (cols 54-61)
  - Line 2: inclination (cols 9-16), RAAN (cols 18-25), eccentricity (cols 27-33, implied decimal), arg perigee (cols 35-42), mean anomaly (cols 44-51), mean motion (cols 53-63)
- Convert angles from degrees to radians
- Parse the TLE epoch into Julian date

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain/ -run TestParseTLE -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/domain/tle_parser.go internal/domain/tle_parser_test.go
git commit -m "feat(domain): implement TLE parser with checksum validation"
```

---

### Task 3: Coordinate Transforms & Visibility

**Files:**
- Create: `internal/domain/coordinates.go`
- Test: `internal/domain/coordinates_test.go`

- [ ] **Step 1: Write coordinate transform tests**

Test `ECIToGeo` with known reference values:
- A point on the x-axis at GMST=0 should give lat=0, lon=0
- A point on the z-axis should give lat=90 (north pole)
- Known ISS position → expected lat/lon (use reference data)

Test `IsVisible` with:
- Satellite directly overhead (alt 400km, same lat/lon as observer) → visible
- Satellite on opposite side of Earth → not visible
- Satellite at horizon edge → boundary case

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/ -run "TestECIToGeo|TestIsVisible" -v`

- [ ] **Step 3: Implement coordinate transforms**

Create `internal/domain/coordinates.go`:
- `ECIToGeo(pos Position, gmst float64) GeoCoordinate` — convert ECI to geodetic using:
  - longitude = atan2(Y, X) - gmst, normalized to [-180, 180]
  - latitude = atan2(Z, sqrt(X² + Y²)) with iterative WGS84 correction
  - altitude = distance from center - Earth radius at that latitude
- `IsVisible(geo GeoCoordinate, observerLat, observerLon float64) bool` — compute great-circle angular distance, compare against horizon angle based on altitude

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain/ -run "TestECIToGeo|TestIsVisible" -v`

- [ ] **Step 5: Commit**

```bash
git add internal/domain/coordinates.go internal/domain/coordinates_test.go
git commit -m "feat(domain): implement ECI-to-geodetic coordinate transform and visibility check"
```

---

## Chunk 2: SGP4 Engine

### Task 4: Constants & Time Utilities

**Files:**
- Create: `internal/infrastructure/propagator/constants.go`
- Create: `internal/infrastructure/propagator/time.go`
- Test: `internal/infrastructure/propagator/time_test.go`

- [ ] **Step 1: Write time utility tests**

Test `JulianDate` conversion:
- 2000-01-01 12:00 UTC = JD 2451545.0 (J2000 epoch)
- 1970-01-01 00:00 UTC = JD 2440587.5

Test `GMST` (Greenwich Mean Sidereal Time):
- At J2000 epoch, GMST ≈ 4.8949612 radians
- Verify GMST increases ~360.98° per solar day

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infrastructure/propagator/ -run "TestJulian|TestGMST" -v`

- [ ] **Step 3: Create constants.go**

WGS84 constants:
- EarthRadius = 6378.137 km
- EarthFlattening = 1/298.257223563
- Mu (gravitational param) = 398600.4418 km³/s²
- J2 = 0.00108262998905
- J3 = -0.00000253215306
- J4 = -0.00000161098761
- MinutesPerDay = 1440.0
- TwoPi = 2π
- XKE = 60.0 / sqrt(EarthRadius³ / Mu)

- [ ] **Step 4: Implement time.go**

- `JulianDate(t time.Time) float64` — convert Go time.Time to Julian Date
- `JulianDateOfEpoch(epochYear, epochDay float64) float64` — convert TLE epoch to JD
- `GMST(jd float64) float64` — Greenwich Mean Sidereal Time in radians
- `MinutesSinceEpoch(epochJD, targetJD float64) float64`

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/propagator/ -run "TestJulian|TestGMST" -v`

- [ ] **Step 6: Commit**

```bash
git add internal/infrastructure/propagator/constants.go internal/infrastructure/propagator/time.go internal/infrastructure/propagator/time_test.go
git commit -m "feat(propagator): add WGS84 constants and Julian date/GMST time utilities"
```

---

### Task 5: Math Helpers & Kepler Solver

**Files:**
- Create: `internal/infrastructure/propagator/math_helpers.go`
- Test: `internal/infrastructure/propagator/math_helpers_test.go`

- [ ] **Step 1: Write Kepler solver tests**

Table-driven tests for `SolveKepler(meanAnomaly, eccentricity float64) float64`:
- Circular orbit (e=0): E should equal M
- Low eccentricity (e=0.001, M=1.0): E ≈ M (nearly equal)
- Moderate eccentricity (e=0.5, M=π/4): verify against known solution
- High eccentricity (e=0.9, M=0.1): verify convergence
- Edge: M=0 → E=0 for any e
- Edge: M=π → verify solution

Test `WrapTwoPi(angle float64) float64`:
- 0 → 0, 2π → 0, -π → π, 7π → π

Test `Clamp(v, min, max float64) float64`

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infrastructure/propagator/ -run "TestSolveKepler|TestWrap|TestClamp" -v`

- [ ] **Step 3: Implement math_helpers.go**

- `SolveKepler(M, e float64) float64` — Newton-Raphson iteration: E_{n+1} = E_n - (E_n - e*sin(E_n) - M) / (1 - e*cos(E_n)), max 10 iterations, tolerance 1e-12
- `WrapTwoPi(angle float64) float64` — normalize angle to [0, 2π)
- `WrapNegPiToPi(angle float64) float64` — normalize to [-π, π)
- `Clamp(v, min, max float64) float64`
- `Acosh(x float64) float64` — for hyperbolic orbits

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/propagator/ -run "TestSolveKepler|TestWrap|TestClamp" -v`

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/propagator/math_helpers.go internal/infrastructure/propagator/math_helpers_test.go
git commit -m "feat(propagator): implement Kepler equation solver and math utilities"
```

---

### Task 6: SGP4 Propagator (Near-Earth)

**Files:**
- Create: `internal/infrastructure/propagator/sgp4.go`
- Test: `internal/infrastructure/propagator/sgp4_test.go`

- [ ] **Step 1: Write SGP4 tests using NORAD reference data**

Use the ISS TLE and known position/velocity at specific timestamps. Table-driven tests:

```go
// Reference: ISS TLE
// 1 25544U 98067A   20045.18587073  .00000950  00000-0  25302-4 0  9990
// 2 25544  51.6443 242.7420 0004615 225.0295 296.6842 15.49163961209242
//
// At t=0 min from epoch:
// Position (km): X≈-2695.77, Y≈-4288.99, Z≈4505.09 (approximate)
// Velocity (km/s): Vx≈5.31, Vy≈-4.25, Vz≈-2.53 (approximate)
```

Test that `Propagate` returns position within 10km of reference and velocity within 0.1 km/s. These tolerances are loose enough for our custom implementation.

Also test:
- Propagation at t=0 (epoch)
- Propagation at t=60 min
- Propagation at t=1440 min (1 day)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infrastructure/propagator/ -run TestSGP4 -v`

- [ ] **Step 3: Implement SGP4 propagator**

Create `internal/infrastructure/propagator/sgp4.go`:

The `SGP4Propagator` struct implements `application.OrbitPropagator`:

```go
type SGP4Propagator struct{}

func (p *SGP4Propagator) Propagate(
    elements domain.OrbitalElements,
    t time.Time,
) (domain.Position, domain.Velocity, error)
```

Implementation follows the standard SGP4 algorithm:

1. **Initialization from elements:** Convert mean motion to rad/min, compute semi-major axis (a₀), recover original mean motion (n₀), compute perigee, determine near-earth vs deep-space
2. **Secular effects:** Update RAAN, argument of perigee, and mean anomaly for J2/J3/J4 drag and gravitational perturbations over time since epoch
3. **Solve Kepler's equation** for eccentric anomaly
4. **Short-period periodics:** Apply oscillatory corrections from J2
5. **Convert to position/velocity:** From updated orbital elements to ECI coordinates using standard orientation math

Key internal functions:
- `initSGP4(elements) *sgp4State` — precompute constants from initial elements
- `propagate(state *sgp4State, tSinceMin float64) (Position, Velocity, error)` — run propagation for given minutes since epoch
- Near-earth branch (mean motion >= 6.4 rev/day, most satellites)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/propagator/ -run TestSGP4 -v`
Expected: all PASS (within tolerance)

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/propagator/sgp4.go internal/infrastructure/propagator/sgp4_test.go
git commit -m "feat(propagator): implement SGP4 near-earth orbital propagation"
```

---

### Task 7: SDP4 Deep-Space Extension

**Files:**
- Create: `internal/infrastructure/propagator/deep_space.go`
- Test: `internal/infrastructure/propagator/deep_space_test.go`

- [ ] **Step 1: Write deep-space tests**

Use a known deep-space satellite TLE (e.g., Vanguard 1, NORAD 5, or a GEO satellite with mean motion < 6.4 rev/day). Test propagation at t=0 and t=1440 min. Tolerances: 50km position (deep space is harder).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infrastructure/propagator/ -run TestDeepSpace -v`

- [ ] **Step 3: Implement deep_space.go**

Add deep-space (SDP4) extensions:
- `isDeepSpace(meanMotion float64) bool` — true if meanMotion < 6.4 rev/day
- Lunar-solar gravitational perturbations
- Resonance terms for 12-hour and 24-hour orbits
- Integrate into `SGP4Propagator.Propagate` — check if deep-space, apply additional corrections

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/propagator/ -run TestDeepSpace -v`

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/propagator/deep_space.go internal/infrastructure/propagator/deep_space_test.go
git commit -m "feat(propagator): add SDP4 deep-space orbital corrections"
```

---

## Chunk 3: Configuration & Application Layer

### Task 8: Config Rewrite

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write config tests**

Test `DefaultConfig()` returns expected defaults:
- Constellations: ["stations", "starlink"]
- TargetFPS: 30
- CelesTrakURL: "https://celestrak.org/NORAD/elements/gp.php"
- FetchTimeout: 10s
- NoColor: false
- ObserverLat/Lon: 0

Test `ParseFlags` with CLI args:
- `--constellations=gps,stations` → Constellations: ["gps", "stations"]
- `--fps=60` → TargetFPS: 60
- `--no-color` → NoColor: true
- `--observer-lat=40.7 --observer-lon=-74.0` → correct values

Test env var resolution for Space-Track creds.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -v`

- [ ] **Step 3: Rewrite config.go**

```go
package config

import (
	"flag"
	"os"
	"strings"
	"time"
)

// AppConfig holds all configurable parameters for the application.
type AppConfig struct {
	CelesTrakURL   string
	SpaceTrackURL  string
	Constellations []string
	FetchTimeout   time.Duration
	FetchInterval  time.Duration
	TargetFPS      int
	NoColor        bool
	ObserverLat    float64
	ObserverLon    float64
	SpaceTrackUser string
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
func Load() *AppConfig {
	cfg := DefaultConfig()

	// Env vars (Space-Track credentials)
	if u := os.Getenv("SPACETRACK_USER"); u != "" {
		cfg.SpaceTrackUser = u
	}
	if p := os.Getenv("SPACETRACK_PASS"); p != "" {
		cfg.SpaceTrackPass = p
	}

	// CLI flags override defaults
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): rewrite config with multi-constellation, FPS, observer position support"
```

---

### Task 9: Update Application Ports & Tracker

**Files:**
- Modify: `internal/application/ports.go`
- Modify: `internal/application/tracker.go`
- Test: `internal/application/tracker_test.go`

- [ ] **Step 1: Write tracker tests**

Test with mock TLEProvider and OrbitPropagator:

```go
// mockProvider returns canned TLE data
// mockPropagator returns canned position/velocity

// TestGetConstellations_Success: fetch 2 constellations, verify all propagated
// TestGetConstellations_PartialFailure: one satellite fails propagation, rest succeed
// TestGetConstellations_ProviderError: provider fails, returns ErrProviderUnavailable
// TestGetConstellations_AllPropagationFail: returns ErrCalculationFailed
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/application/ -v`

- [ ] **Step 3: Update ports.go**

```go
package application

import (
	"context"
	"satellite-visualizer/internal/domain"
	"time"
)

// TLEProvider defines the contract for fetching satellite telemetry.
type TLEProvider interface {
	FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)
	Available() []string
}

// OrbitPropagator defines the contract for the physics engine.
type OrbitPropagator interface {
	Propagate(elements domain.OrbitalElements, t time.Time) (domain.Position, domain.Velocity, error)
}
```

- [ ] **Step 4: Update tracker.go**

Rewrite `Tracker` to:
- Accept `constellations []string` in constructor
- `GetConstellations(ctx, now) ([]Constellation, error)` iterates constellations, calls `ParseTLE` then `Propagate` for each satellite, builds `SatelliteState` with ECI + geo coords + velocity
- Wrap fetch errors with `ErrProviderUnavailable` (not `ErrConstellationNotFound`)
- Compute GMST once per call, convert ECI→Geo for each satellite

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/application/ -v`

- [ ] **Step 6: Commit**

```bash
git add internal/application/ports.go internal/application/tracker.go internal/application/tracker_test.go
git commit -m "feat(application): update ports for velocity/multi-constellation, rewrite tracker"
```

---

## Chunk 4: Infrastructure Adapters

### Task 10: CelesTrak Adapter

**Files:**
- Create: `internal/infrastructure/celestrak/client.go`
- Test: `internal/infrastructure/celestrak/client_test.go`

- [ ] **Step 1: Write CelesTrak tests**

Use `httptest.NewServer` to return canned 3-line TLE responses. Test:
- Successful fetch of 3 satellites → returns 3 TLEs with correct names/lines
- HTTP 404 → returns error
- Malformed response (incomplete TLE) → returns error
- `Available()` returns curated list

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infrastructure/celestrak/ -v`

- [ ] **Step 3: Implement CelesTrak client**

```go
package celestrak

// Client implements application.TLEProvider for CelesTrak.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client
func (c *Client) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)
func (c *Client) Available() []string
```

- Parse 3-line TLE format: line 0 = name (trimmed), line 1 = "1 ...", line 2 = "2 ..."
- Build URL: `baseURL + "?GROUP=" + name + "&FORMAT=tle"`
- Return `domain.ErrConstellationNotFound` on 404

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/celestrak/ -v`

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/celestrak/
git commit -m "feat(celestrak): implement TLEProvider adapter for CelesTrak API"
```

---

### Task 11: Space-Track Adapter

**Files:**
- Create: `internal/infrastructure/spacetrack/client.go`
- Test: `internal/infrastructure/spacetrack/client_test.go`

- [ ] **Step 1: Write Space-Track tests**

Use `httptest.NewServer`. Test:
- Auth flow: login POST with credentials, then fetch with session cookie
- Successful TLE fetch after auth
- Auth failure → `ErrAuthFailed`
- Missing credentials → `ErrAuthFailed`
- `Available()` returns curated list

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement Space-Track client**

```go
package spacetrack

type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	cookie     *http.Cookie // session cookie after auth
}

func NewClient(baseURL, username, password string, timeout time.Duration) *Client
func (c *Client) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)
func (c *Client) Available() []string
func (c *Client) authenticate(ctx context.Context) error
```

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/spacetrack/
git commit -m "feat(spacetrack): implement TLEProvider adapter for Space-Track API"
```

---

### Task 12: Failover Provider & TLE Cache

**Files:**
- Create: `internal/infrastructure/provider/failover.go`
- Create: `internal/infrastructure/provider/cache.go`
- Test: `internal/infrastructure/provider/failover_test.go`
- Test: `internal/infrastructure/provider/cache_test.go`

- [ ] **Step 1: Write failover & cache tests**

Failover tests:
- Primary succeeds → returns primary data
- Primary fails, secondary succeeds → returns secondary data
- Both fail → returns `ErrProviderUnavailable`
- `Available()` merges both providers' lists (deduplicated)

Cache tests:
- `Set` + `Get` within TTL → returns data
- `Get` after TTL → returns data + `ErrStaleData`
- `Get` for missing key → returns nil, false
- Thread-safety: concurrent reads/writes don't panic

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement failover.go**

```go
package provider

type FailoverProvider struct {
	primary   application.TLEProvider
	secondary application.TLEProvider
	logger    *slog.Logger
}

func NewFailover(primary, secondary application.TLEProvider, logger *slog.Logger) *FailoverProvider
func (f *FailoverProvider) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)
func (f *FailoverProvider) Available() []string
```

- [ ] **Step 4: Implement cache.go**

```go
package provider

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

func NewCached(inner application.TLEProvider, staleTTL time.Duration) *CachedProvider
func (c *CachedProvider) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error)
func (c *CachedProvider) Available() []string
```

- [ ] **Step 5: Run tests to verify they pass**

- [ ] **Step 6: Commit**

```bash
git add internal/infrastructure/provider/
git commit -m "feat(provider): implement failover provider and thread-safe TLE cache"
```

---

## Chunk 5: Rendering Engine

### Task 13: Frame Buffer & Projection

**Files:**
- Create: `internal/ui/renderer/frame.go`
- Create: `internal/ui/renderer/projection.go`
- Test: `internal/ui/renderer/projection_test.go`

- [ ] **Step 1: Write projection tests**

Test `Project3DTo2D`:
- Point at center of sphere, facing camera → center of screen
- Point on right edge → right side of screen
- Point behind sphere (z < 0) → marked as occluded
- Aspect ratio correction: terminal chars are ~2:1

Test `RotatePoint`:
- Rotate (1,0,0) by 90° around Z → (0,1,0)
- Rotate (0,0,1) by 90° around X → (0,-1,0)

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement frame.go**

```go
package renderer

// Frame is a double-buffered text frame for flicker-free rendering.
type Frame struct {
	Width, Height int
	cells         [][]Cell
	back          [][]Cell
}

type Cell struct {
	Char  rune
	Color string // ANSI escape sequence
}

func NewFrame(width, height int) *Frame
func (f *Frame) Set(x, y int, ch rune, color string)
func (f *Frame) Clear()
func (f *Frame) Swap()
func (f *Frame) Render() string // produces ANSI string output
```

- [ ] **Step 4: Implement projection.go**

```go
package renderer

// RotatePoint applies rotation around Y (longitude) and X (latitude) axes.
func RotatePoint(x, y, z, rotY, rotX float64) (float64, float64, float64)

// Project3DTo2D performs orthographic projection from 3D to terminal coordinates.
// Returns screen x, y and whether the point is visible (not occluded by sphere).
func Project3DTo2D(x, y, z, radius float64, screenW, screenH int) (sx, sy int, visible bool)
```

- [ ] **Step 5: Run tests, commit**

```bash
git add internal/ui/renderer/frame.go internal/ui/renderer/projection.go internal/ui/renderer/projection_test.go
git commit -m "feat(renderer): implement double-buffered frame and 3D-to-2D projection"
```

---

### Task 14: Globe Surface Rendering

**Files:**
- Create: `internal/ui/renderer/globe.go`
- Create: `internal/ui/renderer/continents.go`
- Create: `internal/ui/renderer/ocean.go`
- Create: `internal/ui/renderer/atmosphere.go`
- Test: `internal/ui/renderer/globe_test.go`

- [ ] **Step 1: Write globe rendering tests**

Test `RenderGlobe` produces a frame with:
- Non-empty output of expected dimensions
- Contains land characters (green ANSI codes present)
- Contains ocean characters (blue ANSI codes present)
- Atmosphere ring present at edges

Test continent data:
- `ContinentPoints` is non-empty
- All lat/lon values within valid ranges

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement continents.go**

Store simplified continental outline polygons as `[]GeoPoint{{Lat, Lon}}` slices. Major landmasses: North America, South America, Europe, Africa, Asia, Australia, Antarctica. ~2000 total points. Each continent as a named slice.

Key function: `IsLand(lat, lon float64) bool` — ray-casting point-in-polygon test against all continental outlines.

- [ ] **Step 4: Implement ocean.go**

```go
// OceanShade returns the appropriate Unicode block character and ANSI color
// for an ocean cell based on its depth/position on the sphere.
func OceanShade(normalZ float64) (rune, string)
```

Uses `░▒▓` blocks with blues from `38;5;17` (dark) to `38;5;27` (lighter) based on surface normal dot product with light direction.

- [ ] **Step 5: Implement atmosphere.go**

```go
// AtmosphereGlow returns the character and color for atmosphere cells
// at the edge of the globe silhouette.
func AtmosphereGlow(distFromEdge float64) (rune, string)
```

Gradient from bright cyan at edge to transparent. Uses `·` and `°` characters.

- [ ] **Step 6: Implement globe.go**

```go
package renderer

type Globe struct {
	Radius    float64
	RotationY float64 // longitude rotation (radians)
	RotationX float64 // latitude tilt (radians)
	Zoom      float64
}

func NewGlobe() *Globe

// Render draws the globe into the given frame.
// For each terminal cell in the frame:
// 1. Map to sphere surface via inverse projection
// 2. Apply rotation to get lat/lon
// 3. Check IsLand → land color, else ocean shade
// 4. Draw atmosphere at edges
func (g *Globe) Render(f *Frame)
```

The rendering loop iterates screen pixels, traces a ray to the sphere, finds the lat/lon at the intersection, and colors accordingly.

- [ ] **Step 7: Run tests, commit**

```bash
git add internal/ui/renderer/globe.go internal/ui/renderer/continents.go internal/ui/renderer/ocean.go internal/ui/renderer/atmosphere.go internal/ui/renderer/globe_test.go
git commit -m "feat(renderer): implement 3D globe with continental outlines, ocean shading, atmosphere glow"
```

---

### Task 15: Satellite Rendering

**Files:**
- Create: `internal/ui/renderer/satellite.go`
- Test: `internal/ui/renderer/satellite_test.go`

- [ ] **Step 1: Write satellite rendering tests**

Test `RenderSatellites`:
- Satellite on visible hemisphere → dot appears in frame
- Satellite on hidden hemisphere → not rendered
- ISS gets star character `★`
- Different constellations get different colors

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement satellite.go**

```go
package renderer

// ConstellationColors maps constellation names to ANSI color codes.
var ConstellationColors = map[string]string{
	"stations": "\033[38;5;196m", // red
	"starlink": "\033[38;5;255m", // white
	"gps-ops":  "\033[38;5;226m", // yellow
	"oneweb":   "\033[38;5;208m", // orange
	"iridium":  "\033[38;5;51m",  // cyan
	"galileo":  "\033[38;5;135m", // purple
	"glo-ops":  "\033[38;5;46m",  // green
}

// RenderSatellites draws satellite positions onto the frame.
func RenderSatellites(f *Frame, satellites []domain.SatelliteState, g *Globe)
```

- [ ] **Step 4: Run tests, commit**

```bash
git add internal/ui/renderer/satellite.go internal/ui/renderer/satellite_test.go
git commit -m "feat(renderer): implement satellite position rendering with constellation colors"
```

---

## Chunk 6: TUI Dashboard

### Task 16: Bubbletea Messages, Keys & Styles

**Files:**
- Create: `internal/ui/tui/messages.go`
- Create: `internal/ui/tui/keys.go`
- Create: `internal/ui/tui/styles.go`

- [ ] **Step 1: Create messages.go**

```go
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
```

- [ ] **Step 2: Create keys.go**

Define key bindings using bubbles/key:
- Rotation: arrows + hjkl
- Zoom: +/-
- Tab: cycle focus
- Enter: select
- /: search
- Space: pause
- c: cycle constellation
- r: refresh
- ?: help
- q: quit

- [ ] **Step 3: Create styles.go**

Lipgloss styles for:
- Border styles (rounded)
- Title styles (bold, colored)
- Selected item highlight
- Status bar style
- Panel dimensions (responsive to terminal size)

- [ ] **Step 4: Commit**

```bash
git add internal/ui/tui/messages.go internal/ui/tui/keys.go internal/ui/tui/styles.go
git commit -m "feat(tui): add message types, key bindings, and lipgloss styles"
```

---

### Task 17: Dashboard Panels

**Files:**
- Create: `internal/ui/tui/panels/globe.go`
- Create: `internal/ui/tui/panels/sidebar.go`
- Create: `internal/ui/tui/panels/details.go`
- Create: `internal/ui/tui/panels/status.go`
- Create: `internal/ui/tui/panels/help.go`

- [ ] **Step 1: Implement globe.go panel**

Wraps the renderer.Globe and renderer.Frame. On each tick:
- Updates globe rotation
- Calls renderer to produce frame
- Returns rendered string for Bubbletea View()

- [ ] **Step 2: Implement sidebar.go panel**

Scrollable satellite list with:
- Filterable by constellation
- Search by name (/ key)
- Highlighted selection
- Shows satellite icon + name + constellation

- [ ] **Step 3: Implement details.go panel**

Shows selected satellite info:
- Name, NORAD catalog number
- Altitude, Latitude, Longitude
- Velocity magnitude
- Constellation name
- Visibility status

- [ ] **Step 4: Implement status.go panel**

Bottom status bar showing:
- Active data source (CelesTrak/Space-Track)
- Total satellite count
- Current FPS
- Time since last data refresh
- "STALE DATA" warning when applicable

- [ ] **Step 5: Implement help.go panel**

Overlay panel (toggle with ?) showing all key bindings in a formatted table.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/tui/panels/
git commit -m "feat(tui): implement dashboard panels (globe, sidebar, details, status, help)"
```

---

### Task 18: Main TUI App Model

**Files:**
- Create: `internal/ui/tui/app.go`
- Test: `internal/ui/tui/app_test.go`

- [ ] **Step 1: Write app model tests**

Test:
- Initial model has correct default state
- Key 'q' returns quit command
- Tab cycles focused panel
- TickMsg triggers propagation update
- DataUpdateMsg updates satellite list
- Window resize adjusts layout

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement app.go**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type focusedPanel int

const (
	focusGlobe focusedPanel = iota
	focusSidebar
)

// App is the root Bubbletea model composing all dashboard panels.
type App struct {
	globe      panels.GlobePanel
	sidebar    panels.SidebarPanel
	details    panels.DetailsPanel
	status     panels.StatusPanel
	help       panels.HelpPanel
	showHelp   bool
	focused    focusedPanel
	tracker    *application.Tracker
	config     *config.AppConfig
	width      int
	height     int
	quitting   bool
}

func NewApp(tracker *application.Tracker, cfg *config.AppConfig) *App

// Init starts the tick loop and background data fetchers.
func (a *App) Init() tea.Cmd

// Update handles all messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd)

// View composes the dashboard layout.
func (a *App) View() string
```

Layout composition in View():
- Use lipgloss.JoinHorizontal/JoinVertical
- Globe panel takes ~70% width, sidebar ~30%
- Details panel below globe, status panel below sidebar
- Help overlay centered on top when visible

- [ ] **Step 4: Run tests, commit**

```bash
git add internal/ui/tui/app.go internal/ui/tui/app_test.go
git commit -m "feat(tui): implement main Bubbletea app model with dashboard layout"
```

---

## Chunk 7: Entry Point & Integration

### Task 19: Install Dependencies

- [ ] **Step 1: Install Bubbletea ecosystem**

```bash
cd /home/carlossalguero/dev/apps/satellite-visualizer
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

- [ ] **Step 2: Commit go.mod and go.sum**

```bash
git add go.mod go.sum
git commit -m "chore: add bubbletea, lipgloss, bubbles dependencies"
```

---

### Task 20: Main Entry Point

**Files:**
- Create: `cmd/satellite-visualizer/main.go`

- [ ] **Step 1: Implement main.go**

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"satellite-visualizer/internal/config"
	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/infrastructure/celestrak"
	"satellite-visualizer/internal/infrastructure/spacetrack"
	"satellite-visualizer/internal/infrastructure/provider"
	"satellite-visualizer/internal/infrastructure/propagator"
	"satellite-visualizer/internal/ui/tui"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background data fetcher
	go tracker.StartFetching(ctx, cfg.FetchInterval)

	// Build and run TUI
	app := tui.NewApp(tracker, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Handle OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify build**

Run: `make build`
Expected: binary at `bin/satellite-visualizer`

- [ ] **Step 3: Commit**

```bash
git add cmd/satellite-visualizer/main.go
git commit -m "feat: implement main entry point with DI wiring and graceful shutdown"
```

---

### Task 21: Integration Test & Polish

- [ ] **Step 1: Run full test suite**

Run: `make test`
Fix any failures.

- [ ] **Step 2: Run linter**

Run: `make lint`
Fix any issues.

- [ ] **Step 3: Manual smoke test**

Run: `make build && make run`
Verify:
- Globe renders with continental outlines
- Satellites appear (requires network for live TLE data)
- Key bindings work (rotate, zoom, select)
- Graceful quit with 'q'

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "fix: address integration issues from smoke testing"
```
