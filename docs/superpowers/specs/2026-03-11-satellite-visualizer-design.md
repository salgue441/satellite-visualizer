# Satellite Orbit Visualizer — Design Specification

## Overview

A terminal-based application (TUI) that connects to public aerospace APIs (CelesTrak + Space-Track.org), ingests live TLE telemetry, and renders a spinning ASCII-art 3D globe showing real-time satellite positions. Built in Go with Bubbletea, custom SGP4 orbital propagation, and a full dashboard layout.

## Architecture

Hexagonal architecture (ports & adapters):

```
cmd/satellite-visualizer/          → Entry point, DI wiring
internal/
  domain/                          → Entities, domain logic, zero external deps
  application/                     → Use cases (Tracker) + port interfaces
  infrastructure/
    celestrak/                     → TLEProvider adapter for CelesTrak API
    spacetrack/                    → TLEProvider adapter for Space-Track.org
    provider/                      → Failover wrapper (CelesTrak primary, Space-Track fallback)
    propagator/                    → Custom SGP4/SDP4 implementation
  ui/
    tui/                           → Bubbletea app, models, views
      panels/                      → Dashboard panel components
    renderer/                      → 3D globe engine, projection, continental data
```

## Domain Layer

### Existing Entities (kept)

- `Position` — X, Y, Z in km (ECI coordinates)
- `TLE` — Name, Line1, Line2
- `Satellite` — Name, TLE, Position
- Domain errors: `ErrInvalidTle`, `ErrCalculationFailed`, `ErrConstellationNotFound`

### New Entities

- `OrbitalElements` — Parsed from TLE: inclination, RAAN, eccentricity, argument of perigee, mean anomaly, mean motion, epoch, B* drag term, element set number
- `GeoCoordinate` — Latitude, longitude, altitude (km)
- `Velocity` — Vx, Vy, Vz in km/s
- `SatelliteState` — Satellite + GeoCoordinate + Velocity + visibility flag + constellation name
- `Constellation` — Name + slice of SatelliteState

### Domain Logic

- `ParseTLE(line1, line2 string) (OrbitalElements, error)` — Extract orbital elements with checksum validation
- `ECIToGeo(pos Position, gmst float64) GeoCoordinate` — ECI to lat/lon/alt conversion
- `IsVisible(geo GeoCoordinate, observerLat, observerLon float64) bool` — Proper horizon calculation

## Custom SGP4 Engine

Located in `internal/infrastructure/propagator/`:

```
propagator/
  sgp4.go           → Main propagator, implements OrbitPropagator interface
  constants.go      → WGS84 constants, gravitational params, time constants
  deep_space.go     → SDP4 extensions for deep-space satellites (period > 225 min)
  math_helpers.go   → Kepler equation solver, Newton-Raphson, trig helpers
  time.go           → Julian date conversions, GMST calculation
  tle_parser.go     → TLE line parsing into OrbitalElements
```

### Propagation pipeline per tick

1. Parse epoch from OrbitalElements, compute minutes-since-epoch
2. Secular perturbations — J2, J3, J4 zonal harmonics (Earth oblateness)
3. Periodic perturbations — Short-period gravity oscillations
4. Deep-space branch (SDP4) — Lunar/solar gravity if mean motion < 6.4 rev/day
5. Solve Kepler's equation (Newton-Raphson) for true anomaly
6. Convert orbital elements → ECI position + velocity
7. ECI → geodetic (lat/lon/alt) via GMST

### Accuracy target

~1 km position error near-Earth, ~10 km deep-space. Sufficient for visualization.

### Performance

SGP4 per satellite is ~microseconds. 5000+ Starlink sats at 30fps is achievable on a single goroutine.

## Infrastructure Adapters

### CelesTrak (`internal/infrastructure/celestrak/`)

- Implements `TLEProvider`
- HTTP GET to `celestrak.org/NORAD/elements/gp.php?GROUP=<name>&FORMAT=tle`
- Parses 3-line TLE format
- Built-in rate limiting

### Space-Track (`internal/infrastructure/spacetrack/`)

- Implements `TLEProvider`
- Auth via `www.space-track.org/ajaxauth/login` (username/password)
- REST API queries by constellation/NORAD ID
- Session cookie management
- Credentials from `SPACETRACK_USER` / `SPACETRACK_PASS` env vars

### Failover Provider (`internal/infrastructure/provider/`)

- Wraps both adapters behind `TLEProvider` interface
- CelesTrak first, Space-Track fallback
- Transparent to application layer

### Data Refresh

- Background goroutine per active constellation, fetches every ~15 min
- Thread-safe TLE cache (`sync.RWMutex`)
- Tracker reads from cache; fetcher writes — no contention on hot path
- Stale data detection with warning if cache age exceeds threshold

## Rendering Engine

Located in `internal/ui/renderer/`:

```
renderer/
  globe.go          → 3D sphere, rotation matrix, projection
  continents.go     → Simplified continental outline polygons (~2000 points)
  ocean.go          → Ocean fill with Unicode blocks + ANSI blue tones
  atmosphere.go     → Glow effect at globe edges using gradient characters
  projection.go     → 3D→2D orthographic projection, aspect ratio correction
  satellite.go      → Satellite dots with constellation-based color coding
  frame.go          → Double-buffered frame buffer
```

### Rendering pipeline per frame

1. Apply rotation matrix (auto-spin + user input)
2. Z-sort back-to-front for occlusion
3. Orthographic projection 3D→2D, accounting for terminal char aspect ratio (~2:1)
4. Draw globe surface: land (green/brown tones), ocean (blue tones, `░▒▓█`)
5. Draw continental outlines (brighter lines for definition)
6. Draw atmosphere glow (1-2 cell ring, dim cyan/blue gradient)
7. Draw satellites: colored dots (`●`, `◉`, `★` for ISS), color per constellation
8. Write double-buffered frame via ANSI escape sequences

### Color scheme

- Land: greens (`38;5;34`) to browns (`38;5;130`)
- Ocean: dark to mid blues (`38;5;17` to `38;5;27`)
- Atmosphere: cyan glow (`38;5;44`)
- Satellites: per-constellation (Starlink=white, GPS=yellow, ISS=red, etc.)

## TUI Dashboard

Built with Bubbletea in `internal/ui/tui/`:

```
tui/
  app.go            → Root model, composes all panels
  keys.go           → Key bindings
  styles.go         → Lipgloss styles
  messages.go       → Custom messages (tick, data update, selection)
  panels/
    globe.go        → Globe viewport (wraps renderer)
    sidebar.go      → Satellite list with scroll + search
    details.go      → Selected satellite info
    status.go       → Bottom bar: source, last fetch, count, FPS
    help.go         → Keyboard shortcut overlay (toggle '?')
```

### Layout

```
┌─────────────────────────────────┬──────────────────┐
│                                 │  SATELLITES       │
│                                 │  ─────────────── │
│         3D GLOBE                │  ★ ISS            │
│      (main viewport)            │  ● Starlink-1234  │
│                                 │  ● Starlink-1235  │
│                                 │  ● GPS-IIR-10     │
│                                 │  ...              │
├─────────────────────────────────┼──────────────────┤
│  SELECTED: ISS (ZARYA)          │  STATUS           │
│  Alt: 408km  Lat: 42.3 N       │  Source: CelesTrak│
│  Lon: 71.1 W  Vel: 7.66 km/s   │  Sats: 147       │
│  Constellation: stations        │  FPS: 30  ~ 14m  │
└─────────────────────────────────┴──────────────────┘
```

### Key bindings

- `arrows` / `hjkl` — Rotate globe
- `+/-` — Zoom in/out
- `Tab` — Cycle panel focus
- `Enter` — Select satellite
- `/` — Search satellites
- `Space` — Pause/resume auto-rotation
- `c` — Cycle constellation filter
- `r` — Force refresh TLE data
- `?` — Toggle help overlay
- `q` / `Ctrl+C` — Quit

### Bubbletea messages

- `TickMsg` — ~30fps, triggers propagation + render
- `DataUpdateMsg` — Background fetcher delivers new TLEs
- `SelectMsg` — User selected satellite, update details panel
- `ResizeMsg` — Terminal resize, recalculate layout

## Concurrency Model

### Goroutine architecture

```
main goroutine
  └─ Bubbletea event loop (UI thread)
       ├─ input events
       ├─ TickMsg → propagate + render
       └─ DataUpdateMsg → swap TLE cache

background goroutines:
  ├─ TLE Fetcher (1 per active constellation)
  │    └─ sleeps ~15min, fetches, sends DataUpdateMsg
  ├─ Failover Monitor (1)
  │    └─ health-checks primary source, switches if down
  └─ FPS Ticker (1)
       └─ sends TickMsg at ~30fps via Bubbletea Cmd
```

### Synchronization

- TLE cache: `sync.RWMutex` (fetchers write-lock, propagator read-locks)
- No channels on hot render path
- `Cmd` system for goroutine→UI communication
- `context.Context` with cancellation for clean shutdown

### Graceful shutdown

1. `q` / `Ctrl+C` pressed
2. Cancel root context → all goroutines drain
3. Bubbletea restores terminal (alt screen, cursor)
4. Clean exit

### Error resilience

- Failing fetcher: log + retry with exponential backoff, never crashes app
- All sources fail: show last-known positions + "STALE DATA" indicator
- SGP4 errors for individual satellites: skip + log, rest continue

## Configuration & CLI

```
satellite-visualizer [flags]

Flags:
  --constellations    Comma-separated list (default: "stations,starlink")
  --fps               Target frame rate (default: 30)
  --no-color          Disable colors for basic terminals
  --observer-lat      Observer latitude (default: 0)
  --observer-lon      Observer longitude (default: 0)

Environment:
  SPACETRACK_USER     Space-Track.org username
  SPACETRACK_PASS     Space-Track.org password
```

### Config resolution order

1. CLI flags (highest)
2. Environment variables
3. `~/.config/satellite-visualizer/config.yaml` (optional)
4. Built-in defaults (lowest)

## Constellations

### Curated defaults

ISS (stations), Starlink, GPS, OneWeb, Iridium, Galileo, GLONASS

### Custom

Users can add any CelesTrak group name via `--constellations` flag.

## Testing Strategy

### Unit tests

- Domain: TLE parsing, coordinate transforms, orbital element extraction
- SGP4: NORAD reference test cases (table-driven)
- Kepler solver: edge cases (circular, near-parabolic, convergence failure)
- TLE parser: valid/invalid lines, checksums

### Integration tests

- CelesTrak/Space-Track adapters: `httptest.Server` with canned responses
- Failover provider: primary fails → fallback, both fail → stale data
- Tracker: full TLE fetch → propagation → SatelliteState flow

### Rendering tests

- Frame buffer snapshots for known orientations
- Projection math: known 3D→2D mappings
- Aspect ratio correction

### Conventions

- Standard `testing` package only, no mock frameworks
- Test doubles via interfaces
- `*_test.go` co-located with source
