# satellite-visualizer

A real-time satellite tracker that renders an interactive 3D globe in your terminal. It fetches live TLE (Two-Line Element) data from CelesTrak and Space-Track.org, propagates orbits using SGP4, and displays satellites on a rotating Earth with accurate coastlines, ocean shading, and atmosphere glow.

https://github.com/user-attachments/assets/demo.mp4

<video src="assets/demo.mp4" width="100%" autoplay loop muted></video>

## Features

- **3D globe rendering** — ray-traced sphere with land/ocean shading, coastlines, grid lines, atmosphere glow, and star field
- **Real-time SGP4 propagation** — computes satellite positions from TLE orbital elements, including deep-space corrections
- **Live TLE data** — fetches from CelesTrak with optional Space-Track.org failover
- **Constellation support** — track Starlink, ISS, and other NORAD satellite groups
- **Interactive TUI** — built with Bubble Tea; rotate, zoom, search, and select satellites
- **Caching** — configurable refresh interval to avoid hammering upstream APIs

## Requirements

- Go 1.25+
- A terminal with true-color support (256-color minimum)

## Installation

```sh
git clone git@github.com:salgue441/satellite-visualizer.git
cd satellite-visualizer
make build
```

The binary is placed in `bin/satellite-visualizer`.

## Usage

```sh
# Run with defaults (ISS + Starlink)
./bin/satellite-visualizer

# Track specific constellations
./bin/satellite-visualizer -constellations starlink,gps-ops,stations

# Set observer location
./bin/satellite-visualizer -observer-lat 25.6866 -observer-lon -100.3161

# Adjust frame rate
./bin/satellite-visualizer -fps 60

# Disable colors
./bin/satellite-visualizer -no-color
```

### Space-Track.org (optional)

For failover TLE data, set credentials as environment variables:

```sh
export SPACETRACK_USER="your@email.com"
export SPACETRACK_PASS="your_password"
./bin/satellite-visualizer
```

## Keybindings

| Key | Action |
|---|---|
| `arrow keys` / `hjkl` | Navigate / rotate globe |
| `+` / `-` | Zoom in / out |
| `space` | Pause rotation |
| `tab` | Cycle panel focus |
| `enter` | Select satellite |
| `/` | Search |
| `c` | Cycle constellation |
| `r` | Refresh TLE data |
| `?` | Toggle help |
| `q` / `Ctrl+C` | Quit |

## Project Structure

```
cmd/satellite-visualizer/    Entry point
internal/
  application/               Tracker service and port interfaces
  config/                    Configuration (flags, env vars, defaults)
  domain/                    TLE parsing, orbital elements, coordinates
  infrastructure/
    celestrak/               CelesTrak API client
    spacetrack/              Space-Track.org API client
    propagator/              SGP4/SDP4 orbit propagator
    provider/                Caching and failover wrappers
  ui/
    renderer/                Globe, projection, pixel buffer, shading
    tui/                     Bubble Tea app, panels, keybindings, styles
```

## Development

```sh
make test     # Run all tests
make lint     # Run go vet
make clean    # Remove build artifacts
make all      # Clean + build + test
```

## License

MIT
