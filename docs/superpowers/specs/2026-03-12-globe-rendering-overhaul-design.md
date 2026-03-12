# Globe Rendering Overhaul Design

## Goal

Improve the satellite visualizer's globe rendering to produce a more realistic, visually appealing Earth with accurate continent shapes, richer colors, smoother edges, and proper lighting — while maintaining 30 FPS.

## Current Problems

- Continents render as block-drawing characters (`▓`, `▒`, `░`, `·`, `.`) that don't form recognizable shapes due to coarse/inaccurate polygon data
- Ocean is near-black teal (peak `RGB(11, 28, 50)` at center), too dark and monotone
- Globe edges are jagged/blocky (one cell = one pixel)
- Lighting makes edges too dark, flattening the 3D effect
- Continent polygon data (~537 points) is too coarse or has coordinate errors
- `isDesertRegion()` only checks latitude, incorrectly desertifying non-desert regions (southern US, southern China, Japan)

## Design

### 1. Half-Block Rendering Engine

Replace the current one-pixel-per-cell renderer with a half-block system that treats each terminal cell as two vertical pixels.

**Characters used:**
- `▀` (upper half block) — foreground color on top, background color on bottom
- `▄` (lower half block) — background color on top, foreground color on bottom
- `█` (full block) — both pixels same color
- ` ` (space with background) — both pixels are background color

**Implementation:**
- Build a pixel buffer at `width × (height × 2)` resolution, separate from the existing `Frame` cell grid
- Each pixel stores an `RGB` color value (no character — characters are determined during compositing)
- Run the ray-sphere intersection loop over this pixel grid
- Compositor pass merges each vertical pixel pair into one terminal cell: pick the half-block character, set fg = top pixel's color, bg = bottom pixel's color (using `▀`)
- The sphere appears noticeably rounder with smoother edges

**Frame type changes:**
- `Frame` gains a new `PixelBuffer` field: `[][]RGB` at `width × (height×2)`
- New methods: `SetPixel(x, y, color)`, `GetPixel(x, y) RGB`
- `Render()` updated to composite pixel pairs into half-block ANSI output
- Existing `Set()`/`Get()` cell methods remain for non-globe UI elements (satellites overlay on top)

**ANSI output considerations:**
- Every cell requires both fg and bg colors (~40 bytes ANSI per cell)
- On 120×40 terminal: ~192KB/frame, ~5.7 MB/s at 30 FPS
- Optimize with delta-rendering: only emit cells that changed from previous frame
- Compare previous pixel buffer to current; skip unchanged cell pairs

**Terminal compatibility:**
- Half-block characters (`U+2580`, `U+2584`, `U+2588`) work correctly in all modern terminal emulators (iTerm2, Alacritty, Kitty, GNOME Terminal, Windows Terminal)
- No fallback path needed — these are universally supported in terminals that support true color (which is already required)

**Performance:** Ray-sphere math is trivial per pixel. 100% more pixel samples (width stays same, height doubles). Combined with delta-rendering optimization, well within 30 FPS budget.

### 2. Continent Data Overhaul

Replace `continents.go` polygon data entirely with accurate, denser coastline polygons.

**Requirements:**
- ~2000-3000 polygon points total (up from ~537)
- Recognizable coastlines for all continents
- Major islands: UK, Japan, Indonesia, New Zealand, Philippines, etc.
- Important peninsulas: Italy, Korea, Florida, India, Scandinavia, Iberian
- Pre-simplified for terminal resolution — no need for sub-degree precision

**Optimization:**
- Add bounding-box pre-check per continent polygon in `IsLand()`
- Skip full ray-cast if lat/lon is outside the continent's bounding rectangle
- Pre-compute bounding boxes at init time
- Note: the main performance cost is the 2× pixel loop, not `IsLand()` — bbox is a preventive measure

**Coastal buffer removal:**
- Remove the ±0.5° coastal buffer hack
- **Acceptance criterion:** All existing test cities in `continents_test.go` (23 cities including Mumbai, Cape Town, Lagos) must pass against the new polygon data WITHOUT the buffer
- Run tests before and after to validate

### 3. Color Palette & Lighting

#### Ocean
- Color range interpolated from dark to lit based on `normalZ`:
  - `r = 10 + normalZ × 20` (10–30)
  - `g = 30 + normalZ × 50` (30–80)
  - `b = 80 + normalZ × 90` (80–170)
- Grid lines: slightly brighter blue (+20 to each channel), rendered as pixel color variation (no separate character in half-block system)
- Note: brighter ocean will affect satellite marker compositing — satellite renderer must read pixel colors from the buffer, not parse ANSI strings

#### Land (latitude-based biomes, interpolated by normalZ)
Each biome defines a dark (edge) and lit (center) color. Final color = `lerp(dark, lit, normalZ)`.

| Latitude | Biome | Dark (edge) | Lit (center) |
|----------|-------|-------------|--------------|
| <15° | Equatorial | RGB(20, 120, 40) | RGB(40, 180, 60) |
| 15-35° desert | Desert | RGB(160, 130, 60) | RGB(200, 170, 90) |
| 15-35° non-desert | Subtropical | RGB(40, 110, 30) | RGB(60, 170, 50) |
| 35-55° | Temperate | RGB(30, 100, 30) | RGB(50, 160, 50) |
| 55-75° | Boreal | RGB(15, 70, 30) | RGB(30, 110, 45) |
| >75° | Polar | RGB(200, 220, 240) | RGB(230, 240, 250) |

**Lighting model:**
- Simple lerp between dark and lit colors: `color = dark + (lit - dark) × normalZ`
- This replaces the current per-channel formulas and the inconsistent `base × (0.3 + 0.7 × normalZ)` approach
- At `normalZ=0` (edge): dark color. At `normalZ=1` (center): lit color.

**Desert detection improvement:**
- Add longitude boundaries to `isDesertRegion()`:
  - Sahara: lat 18-35°, lon -17° to 40°
  - Arabian: lat 18-35°, lon 40° to 60°
  - Thar: lat 23-30°, lon 68° to 76°
  - Australian interior: lat -18° to -30°, lon 125° to 145°
- All other land in the 15-35° band uses Subtropical colors

#### Atmosphere
- Keep existing starfield (works well)
- Increase atmosphere glow threshold from 0.05 to 0.07 for a slightly wider halo
- Atmosphere color: `RGB(8 × alpha, 50 × alpha, 90 × alpha)` (slightly brighter cyan-blue)

### 4. Sphere Edge Anti-Aliasing

For pixels near the sphere boundary, blend sphere color with background color based on distance from sphere edge. Creates a 1-pixel soft transition instead of a hard staircase. Implemented as a lerp during the existing ray-sphere test — negligible cost.

### 5. Satellite Rendering in Half-Block System

The current satellite renderer (`satellite.go`) parses ANSI escape strings to extract background colors for compositing. This will not work with the half-block pixel buffer.

**New approach:**
- Satellites are rendered AFTER the globe pixel buffer is complete but BEFORE the half-block compositing pass
- For each satellite screen position `(sx, sy)`:
  1. Convert to pixel coordinates: `px = sx`, `py = sy × 2` (or `sy × 2 + 1` depending on which half)
  2. Determine which pixel row the satellite maps to based on its precise projected Y position
  3. Write the satellite color directly into the pixel buffer at `(px, py)`
  4. If two satellites map to the same cell but different halves, both are preserved naturally (one in fg, one in bg after compositing)
- Satellite markers use a distinct bright color that stands out against both ocean and land
- The `★` character for ISS is a special case: render it as a full cell override (both halves get the station color) since it needs to be visually prominent

**No ANSI string parsing needed** — satellite compositing works directly on the pixel buffer.

## Files Changed

| File | Change |
|------|--------|
| `renderer/frame.go` | Add `PixelBuffer` type, `SetPixel`/`GetPixel` methods, half-block compositor in `Render()`, delta-rendering optimization |
| `renderer/globe.go` | Iterate over pixel grid instead of cell grid, edge anti-aliasing, updated sphere intersection loop |
| `renderer/continents.go` | Complete replacement with accurate polygon data (~2000-3000 points), bounding-box pre-check optimization |
| `renderer/ocean.go` | Richer blue palette with per-channel `normalZ` interpolation, returns RGB values instead of ANSI |
| `renderer/atmosphere.go` | Updated `LandShade()` to use lerp between dark/lit biome colors, improved `isDesertRegion()` with longitude boundaries, stronger atmosphere glow (threshold 0.07), returns RGB values |
| `renderer/satellite.go` | Rewrite compositing to work on pixel buffer instead of ANSI string parsing, handle half-block pixel placement |

## Files Unchanged

- `renderer/projection.go` — math is correct as-is
- All TUI panel code (`panels/`, `app/`, `styles.go`, etc.)
- All domain/infrastructure code
- Starfield generation logic
- Grid line logic (30° spacing)

## Performance Target

30 FPS maintained. Half-block doubling adds 100% more pixel samples, offset by:
- Delta-rendering (only emit changed cells)
- Bounding-box optimization on continent lookups
- Direct pixel buffer operations instead of ANSI string parsing for satellite compositing
