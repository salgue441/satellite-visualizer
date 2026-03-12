# Globe Rendering Overhaul Design

## Goal

Improve the satellite visualizer's globe rendering to produce a more realistic, visually appealing Earth with accurate continent shapes, richer colors, smoother edges, and proper lighting — while maintaining 30 FPS.

## Current Problems

- Continents render as scattered ASCII characters (`.+x-`) that don't form recognizable shapes
- Ocean is near-black teal, too dark and monotone
- Globe edges are jagged/blocky (one cell = one pixel)
- Lighting makes edges too dark, flattening the 3D effect
- Continent polygon data is too coarse or has coordinate errors

## Design

### 1. Half-Block Rendering Engine

Replace the current one-pixel-per-cell renderer with a half-block system that treats each terminal cell as two vertical pixels.

**Characters used:**
- `▀` (upper half block) — foreground color on top, background color on bottom
- `▄` (lower half block) — background color on top, foreground color on bottom
- `█` (full block) — both pixels same color
- ` ` (space with background) — both pixels are background color

**Implementation:**
- Build a pixel buffer at `width × (height × 2)` resolution
- Run the ray-sphere intersection loop over this pixel grid
- Compositor pass merges each vertical pixel pair into one cell: pick the half-block character, set fg = one pixel's color, bg = the other's
- The sphere appears noticeably rounder with smoother edges

**Performance:** Ray-sphere math is trivial per pixel. ~50% more pixel samples (width stays same, height doubles), well within 30 FPS budget.

### 2. Continent Data Overhaul

Replace `continents.go` polygon data entirely with accurate, denser coastline polygons.

**Requirements:**
- ~2000-3000 polygon points total (up from ~500-600)
- Recognizable coastlines for all continents
- Major islands: UK, Japan, Indonesia, New Zealand, Philippines, etc.
- Important peninsulas: Italy, Korea, Florida, India, Scandinavia, Iberian
- Pre-simplified for terminal resolution — no need for sub-degree precision

**Optimization:**
- Add bounding-box pre-check per continent polygon in `IsLand()`
- Skip full ray-cast if lat/lon is outside the continent's bounding rectangle
- Keeps performance flat even with 3-5× more polygon points

**Remove** the coastal buffer hack (±0.5°) — accurate data makes it unnecessary.

### 3. Color Palette & Lighting

#### Ocean
- Deep ocean: `RGB(10, 30, 80)` → lit: `RGB(30, 80, 170)`
- Smooth gradient based on `normalZ` depth value
- Grid lines: slightly brighter blue, no separate character needed

#### Land (latitude-based biomes, richer colors)
| Latitude | Biome | Dark (edge) | Lit (center) |
|----------|-------|-------------|--------------|
| <15° | Equatorial | RGB(20, 120, 40) | RGB(40, 180, 60) |
| 15-35° desert | Desert | RGB(160, 130, 60) | RGB(200, 170, 90) |
| 15-35° non-desert | Subtropical | RGB(40, 110, 30) | RGB(60, 170, 50) |
| 35-55° | Temperate | RGB(30, 100, 30) | RGB(50, 160, 50) |
| 55-75° | Boreal | RGB(15, 70, 30) | RGB(30, 110, 45) |
| >75° | Polar | RGB(200, 220, 240) | RGB(230, 240, 250) |

#### Lighting model
- Diffuse: `color = base × (0.3 + 0.7 × normalZ)`
- 30% ambient ensures edges stay visible
- Replaces current depth-based shading that makes edges too dark

#### Atmosphere
- Keep existing starfield (works well)
- Slightly stronger cyan-blue halo around sphere edge

### 4. Sphere Edge Anti-Aliasing

For pixels near the sphere boundary, blend sphere color with background color based on distance from sphere edge. Creates a 1-pixel soft transition instead of a hard staircase. Implemented as a lerp during the existing ray-sphere test — negligible cost.

## Files Changed

| File | Change |
|------|--------|
| `renderer/frame.go` | New pixel buffer at 2× vertical resolution, half-block compositor |
| `renderer/globe.go` | Pixel grid iteration, new lighting, edge anti-aliasing |
| `renderer/continents.go` | Complete replacement with accurate polygon data + bbox optimization |
| `renderer/ocean.go` | Richer blue palette, smoother depth gradient |
| `renderer/atmosphere.go` | Updated biome colors in LandShade(), stronger glow |
| `renderer/satellite.go` | Adapt to new pixel buffer |

## Files Unchanged

- `renderer/projection.go` — math is correct as-is
- All TUI panel code (`panels/`, `app/`, `styles.go`, etc.)
- All domain/infrastructure code
- Starfield generation logic
- Grid line logic (30° spacing)

## Performance Target

30 FPS maintained. Half-block doubling adds ~50% more pixel samples, offset by bounding-box optimization on continent lookups.
