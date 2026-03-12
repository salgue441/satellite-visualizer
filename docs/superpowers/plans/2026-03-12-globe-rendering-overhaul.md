# Globe Rendering Overhaul Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Overhaul the terminal globe renderer to use half-block characters for 2× vertical resolution, accurate continent polygons, richer colors, and smooth sphere edges — while maintaining 30 FPS.

**Architecture:** The pixel buffer (width × height×2 RGB values) replaces the current cell-based rendering. Globe/atmosphere/satellite code writes RGB pixels; a compositor pass merges pixel pairs into half-block terminal cells. Continent data is replaced wholesale with denser polygons.

**Tech Stack:** Go, Bubbletea, Lipgloss, ANSI true-color escape sequences, Unicode half-block characters (▀▄█)

**Spec deviations:**
- PixelBuffer is a standalone type (not a field on Frame) for cleaner separation of concerns
- Delta-rendering optimization is deferred to a follow-up task — the initial implementation renders the full buffer each frame, which should be sufficient at 30 FPS given the simple per-cell math
- The existing `Frame` type is kept but no longer used by globe rendering; it remains available for any non-globe UI elements that still reference it
- The ±0.5° coastal buffer in the existing `IsLand()` is removed; the new denser polygon data makes it unnecessary. All 23 test cities must pass without the buffer.

---

## Chunk 1: Pixel Buffer & Half-Block Compositor

### Task 1: RGB Type and Pixel Buffer

**Files:**
- Create: `internal/ui/renderer/pixel.go`
- Test: `internal/ui/renderer/pixel_test.go`

- [ ] **Step 1: Write the failing test for RGB and PixelBuffer**

```go
package renderer

import "testing"

func TestNewPixelBuffer(t *testing.T) {
	pb := NewPixelBuffer(10, 20)
	if pb.Width != 10 || pb.Height != 20 {
		t.Errorf("got %dx%d, want 10x20", pb.Width, pb.Height)
	}
}

func TestPixelBuffer_SetGet(t *testing.T) {
	pb := NewPixelBuffer(10, 20)
	c := RGB{R: 100, G: 150, B: 200}
	pb.Set(5, 10, c)
	got := pb.Get(5, 10)
	if got != c {
		t.Errorf("got %v, want %v", got, c)
	}
}

func TestPixelBuffer_OutOfBounds(t *testing.T) {
	pb := NewPixelBuffer(10, 20)
	// Should not panic
	pb.Set(-1, 0, RGB{})
	pb.Set(0, -1, RGB{})
	pb.Set(10, 0, RGB{})
	pb.Set(0, 20, RGB{})
	got := pb.Get(-1, 0)
	if got != (RGB{}) {
		t.Errorf("out of bounds Get should return zero RGB")
	}
}

func TestPixelBuffer_Clear(t *testing.T) {
	pb := NewPixelBuffer(5, 5)
	pb.Set(2, 2, RGB{R: 255, G: 0, B: 0})
	pb.Clear()
	got := pb.Get(2, 2)
	if got != (RGB{}) {
		t.Errorf("after Clear, got %v, want zero", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/renderer/ -run TestNewPixelBuffer -v`
Expected: FAIL — `NewPixelBuffer` undefined

- [ ] **Step 3: Write minimal implementation**

```go
package renderer

// RGB holds a single pixel color.
type RGB struct {
	R, G, B uint8
}

// PixelBuffer is a 2D grid of RGB pixels.
type PixelBuffer struct {
	Width, Height int
	pixels        []RGB // flat array, row-major
}

// NewPixelBuffer creates a pixel buffer initialized to black.
func NewPixelBuffer(width, height int) *PixelBuffer {
	return &PixelBuffer{
		Width:  width,
		Height: height,
		pixels: make([]RGB, width*height),
	}
}

// Set writes an RGB color at (x, y). Out-of-bounds writes are ignored.
func (pb *PixelBuffer) Set(x, y int, c RGB) {
	if x < 0 || y < 0 || x >= pb.Width || y >= pb.Height {
		return
	}
	pb.pixels[y*pb.Width+x] = c
}

// Get reads the RGB color at (x, y). Out-of-bounds returns zero RGB.
func (pb *PixelBuffer) Get(x, y int) RGB {
	if x < 0 || y < 0 || x >= pb.Width || y >= pb.Height {
		return RGB{}
	}
	return pb.pixels[y*pb.Width+x]
}

// Clear resets all pixels to black.
func (pb *PixelBuffer) Clear() {
	for i := range pb.pixels {
		pb.pixels[i] = RGB{}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/renderer/ -run "TestNewPixelBuffer|TestPixelBuffer" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/renderer/pixel.go internal/ui/renderer/pixel_test.go
git commit -m "feat(renderer): add RGB type and PixelBuffer"
```

### Task 2: Half-Block Compositor

**Files:**
- Modify: `internal/ui/renderer/pixel.go`
- Test: `internal/ui/renderer/pixel_test.go`

- [ ] **Step 1: Write the failing test for compositor**

Add `"strings"` to the imports in `pixel_test.go` (alongside `"testing"`), then add these tests:

```go
func TestCompositeHalfBlocks_SameColor(t *testing.T) {
	pb := NewPixelBuffer(2, 4) // 2 wide, 4 pixel-rows = 2 terminal rows
	red := RGB{R: 200, G: 0, B: 0}
	// Fill all pixels with red
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			pb.Set(x, y, red)
		}
	}
	output := pb.CompositeHalfBlocks()
	// Should contain full block character '█' since top and bottom are same
	if !containsRune(output, '█') {
		t.Errorf("expected full block for same-color pairs, got: %q", output)
	}
}

func TestCompositeHalfBlocks_DifferentColors(t *testing.T) {
	pb := NewPixelBuffer(1, 2) // 1 wide, 2 pixel-rows = 1 terminal row
	pb.Set(0, 0, RGB{R: 255, G: 0, B: 0})   // top = red
	pb.Set(0, 1, RGB{R: 0, G: 0, B: 255})   // bottom = blue
	output := pb.CompositeHalfBlocks()
	// Should contain upper half block '▀' with fg=red, bg=blue
	if !containsRune(output, '▀') {
		t.Errorf("expected upper half block for different colors, got: %q", output)
	}
	// Should contain ANSI for red foreground and blue background
	if !strings.Contains(output, "38;2;255;0;0") {
		t.Errorf("expected red foreground ANSI")
	}
	if !strings.Contains(output, "48;2;0;0;255") {
		t.Errorf("expected blue background ANSI")
	}
}

func TestCompositeHalfBlocks_Dimensions(t *testing.T) {
	pb := NewPixelBuffer(3, 6) // 3 wide, 6 pixel-rows = 3 terminal rows
	output := pb.CompositeHalfBlocks()
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 terminal rows, got %d", len(lines))
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/renderer/ -run "TestCompositeHalfBlocks" -v`
Expected: FAIL — `CompositeHalfBlocks` undefined

- [ ] **Step 3: Write minimal implementation**

Add to `pixel.go`:

```go
import (
	"fmt"
	"strings"
)

// CompositeHalfBlocks merges pixel pairs into half-block terminal output.
// Each pair of vertical pixels becomes one terminal cell using ▀ with
// fg = top pixel color, bg = bottom pixel color.
// Returns the full ANSI string ready for terminal display.
func (pb *PixelBuffer) CompositeHalfBlocks() string {
	termRows := pb.Height / 2
	var sb strings.Builder
	// Pre-allocate roughly: ~45 bytes per cell (ANSI) × width × rows
	sb.Grow(termRows * pb.Width * 45)

	prevFg := RGB{}
	prevBg := RGB{}
	firstCell := true

	for row := 0; row < termRows; row++ {
		if row > 0 {
			sb.WriteString("\033[0m\n")
			firstCell = true
		}
		for col := 0; col < pb.Width; col++ {
			top := pb.Get(col, row*2)
			bot := pb.Get(col, row*2+1)

			// Optimize: skip ANSI if colors unchanged from previous cell
			if firstCell || top != prevFg || bot != prevBg {
				if top == bot {
					sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm", top.R, top.G, top.B))
					sb.WriteRune('█')
				} else {
					sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
						top.R, top.G, top.B, bot.R, bot.G, bot.B))
					sb.WriteRune('▀')
				}
				prevFg = top
				prevBg = bot
				firstCell = false
			} else {
				if top == bot {
					sb.WriteRune('█')
				} else {
					sb.WriteRune('▀')
				}
			}
		}
	}
	sb.WriteString("\033[0m")
	return sb.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/renderer/ -run "TestCompositeHalfBlocks" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/renderer/pixel.go internal/ui/renderer/pixel_test.go
git commit -m "feat(renderer): add half-block compositor for pixel buffer"
```

### Task 3: Integrate Pixel Buffer into Globe Rendering

**Files:**
- Modify: `internal/ui/renderer/globe.go`
- Modify: `internal/ui/renderer/ocean.go`
- Modify: `internal/ui/renderer/atmosphere.go`
- Test: `internal/ui/renderer/globe_test.go`

This task rewires the globe renderer to write into a PixelBuffer instead of a Frame. The shading functions are updated to return RGB instead of (rune, string).

- [ ] **Step 1: Update OceanShade to return RGB**

Add new function to `ocean.go` (keep old one temporarily for test compat):

```go
// OceanShadeRGB returns the ocean pixel color for the given parameters.
func OceanShadeRGB(normalZ float64, onGrid bool) RGB {
	r := uint8(10 + normalZ*20)
	g := uint8(30 + normalZ*50)
	b := uint8(80 + normalZ*90)

	if onGrid {
		r = uint8(min(255, int(r)+20))
		g = uint8(min(255, int(g)+20))
		b = uint8(min(255, int(b)+20))
	}
	return RGB{R: r, G: g, B: b}
}
```

- [ ] **Step 2: Update LandShade to return RGB**

Add new function to `atmosphere.go`:

```go
// LandShadeRGB returns the land pixel color based on biome and lighting.
func LandShadeRGB(lat, lon, normalZ float64, onGrid bool) RGB {
	absLat := math.Abs(lat)

	var dark, lit RGB
	switch {
	case absLat > 75:
		dark = RGB{200, 220, 240}
		lit = RGB{230, 240, 250}
	case absLat > 55:
		dark = RGB{15, 70, 30}
		lit = RGB{30, 110, 45}
	case absLat > 35:
		dark = RGB{30, 100, 30}
		lit = RGB{50, 160, 50}
	case absLat > 15:
		if isDesertRegionV2(lat, lon) {
			dark = RGB{160, 130, 60}
			lit = RGB{200, 170, 90}
		} else {
			dark = RGB{40, 110, 30}
			lit = RGB{60, 170, 50}
		}
	default:
		dark = RGB{20, 120, 40}
		lit = RGB{40, 180, 60}
	}

	// Lerp between dark and lit based on normalZ
	r := uint8(float64(dark.R) + float64(int(lit.R)-int(dark.R))*normalZ)
	g := uint8(float64(dark.G) + float64(int(lit.G)-int(dark.G))*normalZ)
	b := uint8(float64(dark.B) + float64(int(lit.B)-int(dark.B))*normalZ)

	if onGrid {
		r = uint8(min(255, int(r)+30))
		g = uint8(min(255, int(g)+30))
		b = uint8(min(255, int(b)+30))
	}

	return RGB{R: r, G: g, B: b}
}

// isDesertRegionV2 checks latitude AND longitude for desert detection.
func isDesertRegionV2(lat, lon float64) bool {
	absLat := math.Abs(lat)
	if absLat < 18 || absLat > 35 {
		return false
	}
	// Northern hemisphere deserts
	if lat > 0 {
		// Sahara: lat 18-35, lon -17 to 40
		if lon >= -17 && lon <= 40 {
			return true
		}
		// Arabian: lat 18-35, lon 40 to 60
		if lon >= 40 && lon <= 60 {
			return true
		}
		// Thar: lat 23-30, lon 68 to 76
		if absLat >= 23 && absLat <= 30 && lon >= 68 && lon <= 76 {
			return true
		}
	}
	// Southern hemisphere: Australian interior
	if lat < 0 && lon >= 125 && lon <= 145 {
		return true
	}
	return false
}
```

- [ ] **Step 3: Update AtmosphereGlow to return RGB**

Add to `atmosphere.go`:

```go
// AtmosphereGlowRGB returns the atmosphere glow color for a point outside the sphere.
// Returns zero RGB if beyond glow range.
func AtmosphereGlowRGB(distFromEdge float64) (RGB, bool) {
	if distFromEdge >= 0.07 {
		return RGB{}, false
	}
	alpha := 1.0 - distFromEdge/0.07
	alpha *= alpha
	return RGB{
		R: uint8(8 * alpha),
		G: uint8(50 * alpha),
		B: uint8(90 * alpha),
	}, true
}

// StarFieldRGB returns a star color for the given position, or zero RGB + false if no star.
func StarFieldRGB(sx, sy int) (RGB, bool) {
	hash := (sx*7919 + sy*6271 + sx*sy*31) % 997
	if hash > 12 {
		return RGB{}, false
	}
	switch {
	case hash < 2:
		return RGB{180, 180, 200}, true
	case hash < 5:
		return RGB{100, 100, 130}, true
	case hash < 9:
		return RGB{50, 50, 70}, true
	default:
		return RGB{30, 30, 45}, true
	}
}
```

- [ ] **Step 4: Rewrite Globe.Render to use PixelBuffer**

Replace the `Render` method in `globe.go`:

```go
// Render draws the globe into the given pixel buffer using ray-sphere intersection.
// The pixel buffer should have height = 2× terminal rows for half-block rendering.
func (g *Globe) Render(pb *PixelBuffer) {
	pb.Clear()

	w, h := pb.Width, pb.Height
	if w == 0 || h == 0 {
		return
	}

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0

	// Sphere radius in pixel-row units. Fill ~90% of vertical space.
	// Since pixel height = 2× terminal height, divide by 2 to match terminal proportions.
	termH := float64(h) / 2.0
	sphereR := termH * 0.45 * g.Zoom

	// Terminal characters are ~2x taller than wide. In pixel space, each pixel-row
	// is half a terminal row, so the effective aspect ratio is 1.0 (2.0 / 2).
	const pixelAspect = 1.0

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			nx := (float64(px) - cx) / (sphereR * pixelAspect)
			ny := -(float64(py) - cy) / sphereR

			r2 := nx*nx + ny*ny
			if r2 > 1.0 {
				dist := math.Sqrt(r2) - 1.0
				if dist < 0.07 {
					if c, ok := AtmosphereGlowRGB(dist); ok {
						pb.Set(px, py, c)
						continue
					}
				}
				if c, ok := StarFieldRGB(px, py); ok {
					pb.Set(px, py, c)
				}
				continue
			}

			nz := math.Sqrt(1.0 - r2)

			// Edge anti-aliasing: blend with background near sphere boundary
			edgeAlpha := 1.0
			if r2 > 0.98 {
				edgeAlpha = (1.0 - math.Sqrt(r2)) / (1.0 - math.Sqrt(0.98))
				if edgeAlpha > 1.0 {
					edgeAlpha = 1.0
				}
				if edgeAlpha < 0.0 {
					edgeAlpha = 0.0
				}
			}

			wx, wy, wz := RotateY(nx, ny, nz, -g.RotationY)
			wx, wy, wz = RotateX(wx, wy, wz, -g.RotationX)

			lat := math.Asin(wy) * 180.0 / math.Pi
			lon := math.Atan2(wx, wz) * 180.0 / math.Pi

			onGrid := isGridLine(lat, lon)

			var c RGB
			if IsLand(lat, lon) {
				c = LandShadeRGB(lat, lon, nz, onGrid)
			} else {
				c = OceanShadeRGB(nz, onGrid)
			}

			// Apply edge anti-aliasing
			if edgeAlpha < 1.0 {
				c.R = uint8(float64(c.R) * edgeAlpha)
				c.G = uint8(float64(c.G) * edgeAlpha)
				c.B = uint8(float64(c.B) * edgeAlpha)
			}

			pb.Set(px, py, c)
		}
	}
}
```

- [ ] **Step 5: Update tests for new Globe.Render signature**

Update `globe_test.go` — the tests need to use PixelBuffer instead of Frame:

```go
package renderer

import (
	"testing"
)

func TestGlobeRender_ProducesOutput(t *testing.T) {
	g := NewGlobe()
	pb := NewPixelBuffer(40, 40) // 40 wide, 40 pixel-rows = 20 terminal rows
	g.Render(pb)
	output := pb.CompositeHalfBlocks()
	if len(output) == 0 {
		t.Error("expected non-empty rendered output")
	}
}

func TestGlobeRender_ContainsLandAndOcean(t *testing.T) {
	g := NewGlobe()
	pb := NewPixelBuffer(80, 80) // 80 wide, 80 pixel-rows = 40 terminal rows
	g.Render(pb)

	hasLand := false
	hasOcean := false
	// Ocean center color at nz=1: RGB(30, 80, 170)
	// Land equatorial lit: RGB(40, 180, 60)
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			c := pb.Get(x, y)
			if c == (RGB{}) {
				continue
			}
			// Ocean: blue-dominant (b > r and b > g)
			if c.B > c.R && c.B > c.G && c.B > 50 {
				hasOcean = true
			}
			// Land: green-dominant (g > r and g > b)
			if c.G > c.R && c.G > c.B && c.G > 50 {
				hasLand = true
			}
		}
	}

	if !hasOcean {
		t.Error("expected ocean-colored pixels in rendered buffer")
	}
	if !hasLand {
		t.Error("expected land-colored pixels in rendered buffer")
	}
}

func TestIsLand_KnownPoints(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{"NYC", 40, -74, true},
		{"Gulf of Guinea", 0, 0, false},
		{"Sydney", -33, 151, true},
		{"North Pole", 90, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLand(tt.lat, tt.lon)
			if got != tt.want {
				t.Errorf("IsLand(%v, %v) = %v, want %v", tt.lat, tt.lon, got, tt.want)
			}
		})
	}
}

func TestIsLand_Ocean(t *testing.T) {
	oceanPoints := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Mid Pacific", 0, -160},
		{"Mid Atlantic", 30, -40},
		{"Southern Ocean", -60, 90},
		{"North Atlantic", 50, -30},
	}
	for _, tt := range oceanPoints {
		t.Run(tt.name, func(t *testing.T) {
			if IsLand(tt.lat, tt.lon) {
				t.Errorf("IsLand(%v, %v) = true, expected false (ocean)", tt.lat, tt.lon)
			}
		})
	}
}

func TestOceanShadeRGB_Range(t *testing.T) {
	for _, nz := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
		c := OceanShadeRGB(nz, false)
		// Blue should dominate for ocean
		if c.B <= c.R || c.B <= c.G {
			t.Errorf("OceanShadeRGB(%v): blue should dominate, got %v", nz, c)
		}
	}
}

func TestAtmosphereGlowRGB_ThinHalo(t *testing.T) {
	c, ok := AtmosphereGlowRGB(0.01)
	if !ok || c == (RGB{}) {
		t.Errorf("AtmosphereGlowRGB(0.01) should return visible glow")
	}
	_, ok = AtmosphereGlowRGB(0.1)
	if ok {
		t.Errorf("AtmosphereGlowRGB(0.1) should return false for far distance")
	}
}

func TestLandShadeRGB_VariesByLatitude(t *testing.T) {
	tropical := LandShadeRGB(5, 0, 0.8, false)
	polar := LandShadeRGB(80, 0, 0.8, false)
	if tropical == polar {
		t.Errorf("expected different colors for tropical %v and polar %v", tropical, polar)
	}
}
```

- [ ] **Step 6: Run all tests**

Run: `go test ./internal/ui/renderer/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/renderer/globe.go internal/ui/renderer/ocean.go internal/ui/renderer/atmosphere.go internal/ui/renderer/globe_test.go
git commit -m "feat(renderer): rewrite globe rendering to use pixel buffer with RGB shading"
```

### Task 4: Update GlobePanel and Satellite Rendering

**Files:**
- Modify: `internal/ui/tui/panels/globe.go`
- Modify: `internal/ui/renderer/satellite.go`
- Test: `internal/ui/renderer/satellite_test.go`

- [ ] **Step 1: Rewrite RenderSatellites to work with PixelBuffer**

Replace `satellite.go`'s `RenderSatellites`:

```go
// RenderSatellites draws satellite positions into the pixel buffer.
// Must be called after Globe.Render() and before CompositeHalfBlocks().
func RenderSatellites(pb *PixelBuffer, satellites []domain.SatelliteState, g *Globe) {
	earthRadius := 6378.137

	// Sphere radius must match globe.go Render()
	termH := float64(pb.Height) / 2.0
	sphereR := termH * 0.45 * g.Zoom

	const pixelAspect = 1.0

	for _, sat := range satellites {
		dist := math.Sqrt(sat.Position.X*sat.Position.X +
			sat.Position.Y*sat.Position.Y +
			sat.Position.Z*sat.Position.Z)
		if dist == 0 {
			continue
		}

		scale := 1.0 + (dist/earthRadius-1.0)*0.02
		if scale < 1.01 {
			scale = 1.01
		}
		nx := sat.Position.X / earthRadius * scale
		ny := sat.Position.Y / earthRadius * scale
		nz := sat.Position.Z / earthRadius * scale

		rx, ry, rz := RotateY(nx, ny, nz, -g.RotationY)
		rx, ry, rz = RotateX(rx, ry, rz, -g.RotationX)

		if rz <= 0 {
			continue
		}

		// Project to pixel coordinates (not terminal cell coordinates)
		cx := float64(pb.Width) / 2.0
		cy := float64(pb.Height) / 2.0
		px := int(cx + rx*sphereR*pixelAspect)
		py := int(cy - ry*sphereR)

		if px < 0 || px >= pb.Width || py < 0 || py >= pb.Height {
			continue
		}

		rgb, ok := ConstellationColors[sat.ConstellationName]
		if !ok {
			rgb = DefaultSatRGB
		}
		satColor := RGB{R: uint8(rgb[0]), G: uint8(rgb[1]), B: uint8(rgb[2])}

		// Write satellite pixel. For stations (ISS), write a 2×2 block for visibility.
		pb.Set(px, py, satColor)
		if sat.ConstellationName == "stations" {
			pb.Set(px+1, py, satColor)
			pb.Set(px, py+1, satColor)
			pb.Set(px+1, py+1, satColor)
		}
	}
}
```

- [ ] **Step 2: Update GlobePanel to use PixelBuffer**

Replace `panels/globe.go`:

```go
package panels

import (
	"satellite-visualizer/internal/domain"
	"satellite-visualizer/internal/ui/renderer"
)

// GlobePanel wraps the 3D globe renderer.
type GlobePanel struct {
	globe  *renderer.Globe
	pb     *renderer.PixelBuffer
	width  int
	height int
}

// NewGlobePanel creates a globe panel with given dimensions.
func NewGlobePanel(width, height int) *GlobePanel {
	return &GlobePanel{
		globe:  renderer.NewGlobe(),
		pb:     renderer.NewPixelBuffer(width, height*2), // 2× vertical for half-blocks
		width:  width,
		height: height,
	}
}

// Resize updates the panel dimensions.
func (p *GlobePanel) Resize(width, height int) {
	p.width = width
	p.height = height
	p.pb = renderer.NewPixelBuffer(width, height*2)
}

// Globe returns the underlying globe for rotation/zoom control.
func (p *GlobePanel) Globe() *renderer.Globe {
	return p.globe
}

// Render draws the globe and satellites, returns the rendered string.
func (p *GlobePanel) Render(satellites []domain.SatelliteState) string {
	p.globe.Render(p.pb)
	renderer.RenderSatellites(p.pb, satellites, p.globe)
	return p.pb.CompositeHalfBlocks()
}
```

- [ ] **Step 3: Update satellite tests**

Rewrite `satellite_test.go` to use PixelBuffer:

```go
package renderer

import (
	"testing"

	"satellite-visualizer/internal/domain"
)

func TestRenderSatellites_VisibleSatellite(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "TEST-SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "starlink",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	// The satellite should create a white pixel somewhere near center
	found := false
	white := RGB{255, 255, 255}
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == white {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("expected visible satellite pixel but it was not found")
	}
}

func TestRenderSatellites_HiddenSatellite(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "HIDDEN", Position: domain.Position{X: 0, Y: 0, Z: -(earthR + 400)}},
			ConstellationName: "starlink",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	white := RGB{255, 255, 255}
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == white {
				t.Errorf("expected hidden satellite not to be rendered, found white pixel at (%d, %d)", x, y)
				return
			}
		}
	}
}

func TestRenderSatellites_StationGetsLargerMarker(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()

	earthR := 6378.137
	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "ISS", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "stations",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	stationColor := RGB{255, 80, 80}
	count := 0
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == stationColor {
				count++
			}
		}
	}

	// Station should have a 2×2 block = 4 pixels
	if count < 4 {
		t.Errorf("expected station to have at least 4 pixels, got %d", count)
	}
}

func TestRenderSatellites_ConstellationColors(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()
	earthR := 6378.137

	for name, rgb := range ConstellationColors {
		t.Run(name, func(t *testing.T) {
			pb.Clear()
			g.Render(pb)

			sats := []domain.SatelliteState{
				{
					Satellite:         domain.Satellite{Name: "SAT", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
					ConstellationName: name,
				},
			}
			RenderSatellites(pb, sats, g)

			expected := RGB{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2])}
			found := false
			for y := 0; y < pb.Height; y++ {
				for x := 0; x < pb.Width; x++ {
					if pb.Get(x, y) == expected {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				t.Errorf("constellation %q: expected pixel with color %v not found", name, expected)
			}
		})
	}
}

func TestRenderSatellites_UnknownConstellation(t *testing.T) {
	pb := NewPixelBuffer(80, 80)
	g := NewGlobe()
	earthR := 6378.137

	sats := []domain.SatelliteState{
		{
			Satellite:         domain.Satellite{Name: "MYSTERY", Position: domain.Position{X: 0, Y: 0, Z: earthR + 400}},
			ConstellationName: "unknown-constellation",
		},
	}

	g.Render(pb)
	RenderSatellites(pb, sats, g)

	defaultColor := RGB{200, 200, 200}
	found := false
	for y := 0; y < pb.Height; y++ {
		for x := 0; x < pb.Width; x++ {
			if pb.Get(x, y) == defaultColor {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected satellite with default color not found")
	}
}
```

- [ ] **Step 4: Run all tests**

Run: `go test ./internal/ui/renderer/ -v && go test ./internal/ui/tui/... -v`
Expected: PASS (compilation may require removing old Frame references)

- [ ] **Step 5: Clean up old Frame-based code**

Remove the old `Frame` type from `frame.go` if no other code references it. If the debug_test.go or projection_test.go still use Frame, keep it but mark it deprecated. Remove the old `OceanShade`, `LandShade`, `AtmosphereGlow`, `StarField` functions that returned `(rune, string)`.

- [ ] **Step 6: Run full test suite**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/renderer/satellite.go internal/ui/renderer/satellite_test.go internal/ui/tui/panels/globe.go internal/ui/renderer/frame.go internal/ui/renderer/atmosphere.go internal/ui/renderer/ocean.go internal/ui/renderer/globe_test.go internal/ui/renderer/debug_test.go
git commit -m "feat(renderer): integrate pixel buffer into globe panel and satellite rendering"
```

## Chunk 2: Continent Data Overhaul

### Task 5: Replace Continent Polygon Data

**Files:**
- Modify: `internal/ui/renderer/continents.go`
- Test: `internal/ui/renderer/continents_test.go`

This is the largest single change. The existing ~537 points are replaced with ~2500+ points for accurate coastlines.

- [ ] **Step 1: Add bounding box optimization to IsLand**

Before replacing data, add the bbox optimization that will help with the denser data:

```go
// ContinentBBox holds precomputed bounding boxes for each continent polygon.
type ContinentBBox struct {
	MinLat, MaxLat, MinLon, MaxLon float64
}

var continentBBoxes []ContinentBBox

func init() {
	continentBBoxes = make([]ContinentBBox, len(continents))
	for i, poly := range continents {
		if len(poly) == 0 {
			continue
		}
		bb := ContinentBBox{
			MinLat: poly[0].Lat, MaxLat: poly[0].Lat,
			MinLon: poly[0].Lon, MaxLon: poly[0].Lon,
		}
		for _, p := range poly[1:] {
			if p.Lat < bb.MinLat {
				bb.MinLat = p.Lat
			}
			if p.Lat > bb.MaxLat {
				bb.MaxLat = p.Lat
			}
			if p.Lon < bb.MinLon {
				bb.MinLon = p.Lon
			}
			if p.Lon > bb.MaxLon {
				bb.MaxLon = p.Lon
			}
		}
		// Add small margin for edge cases
		bb.MinLat -= 0.5
		bb.MaxLat += 0.5
		bb.MinLon -= 0.5
		bb.MaxLon += 0.5
		continentBBoxes[i] = bb
	}
}

// IsLand returns true if the given lat/lon falls on a landmass.
// Uses bounding-box pre-check and ray-casting point-in-polygon test.
func IsLand(lat, lon float64) bool {
	for i, poly := range continents {
		bb := continentBBoxes[i]
		if lat < bb.MinLat || lat > bb.MaxLat || lon < bb.MinLon || lon > bb.MaxLon {
			continue
		}
		if pointInPolygon(lat, lon, poly) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run existing continent tests to verify bbox optimization works**

Run: `go test ./internal/ui/renderer/ -run "TestMajorCities|TestIsLand" -v`
Expected: PASS — bbox optimization is transparent

- [ ] **Step 3: Commit bbox optimization**

```bash
git add internal/ui/renderer/continents.go
git commit -m "perf(renderer): add bounding-box optimization to IsLand"
```

- [ ] **Step 4: Replace continent polygon data**

Replace the entire `continents` variable in `continents.go` with accurate, denser coastline polygons (~2500+ points). This is a large data replacement. The new data should include:

- North America: detailed coastline including Florida peninsula, Gulf Coast, Great Lakes outline, Alaska, Pacific coast
- South America: detailed Brazilian bulge, Patagonia, Chilean coast
- Africa: Horn of Africa, West African coast detail, Cape of Good Hope
- Europe: Scandinavia detail, Mediterranean peninsulas (Italy boot, Iberian, Greece), British Isles separation
- Asia: Indian subcontinent, Korean peninsula, Southeast Asian detail, Russian Arctic coast
- Australia: detailed coastline including bays
- Antarctica: simplified but complete outline
- Islands: Greenland, Japan (all main islands), UK, Ireland, Indonesia (Sumatra, Java, Borneo, Sulawesi), Philippines, New Zealand, Madagascar, Sri Lanka, Taiwan, Iceland, Cuba

**Key accuracy requirements:**
- Florida peninsula must be clearly visible
- Italy boot shape recognizable
- Indian subcontinent clearly defined
- Korean peninsula distinct
- UK/Ireland separate from continental Europe
- Japan archipelago recognizable

**Implementation note:** Generate this data from a simplified world coastline dataset. Each polygon should trace the coastline clockwise with ~1-2 degree spacing for major landmasses and ~0.5-1 degree for important smaller features.

- [ ] **Step 5: Run continent tests against new data**

Run: `go test ./internal/ui/renderer/ -run "TestMajorCities" -v`
Expected: ALL 23 cities PASS, ALL 3 ocean points PASS

If any cities fail, adjust polygon data until all pass.

- [ ] **Step 6: Run full test suite**

Run: `go test ./internal/ui/renderer/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/renderer/continents.go
git commit -m "feat(renderer): replace continent data with accurate high-density coastlines"
```

## Chunk 3: Visual Polish & Integration Testing

### Task 6: Manual Visual Verification

**Files:** None (testing only)

- [ ] **Step 1: Build and run the application**

Run: `go build -o satvis ./cmd/satellite-visualizer && ./satvis`

- [ ] **Step 2: Visual checks**

Verify in the running terminal:
1. Globe is round with smooth edges (no staircasing)
2. Continents are recognizable — can you identify Africa, Americas, Europe, Asia, Australia?
3. Ocean is a rich blue gradient (not near-black)
4. Land shows green/brown biomes with depth shading
5. Grid lines visible as subtle brighter lines
6. Stars visible in background space
7. Atmosphere glow visible as thin blue halo
8. Rotation is smooth (no flickering or tearing)
9. FPS counter shows ≥25 FPS

- [ ] **Step 3: Fix any visual issues found**

Address problems and re-test. Common issues:
- Aspect ratio wrong: adjust `pixelAspect` constant
- Continents too small/large: adjust polygon data
- Colors too bright/dark: tweak RGB ranges in shade functions
- Edge anti-aliasing too aggressive: adjust the `0.98` threshold

- [ ] **Step 4: Run full test suite one final time**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Final commit if any fixes were made**

```bash
git add -A
git commit -m "fix(renderer): visual polish adjustments from manual testing"
```

### Task 7: Remove Deprecated Code

**Files:**
- Modify: `internal/ui/renderer/frame.go`
- Modify: `internal/ui/renderer/atmosphere.go`
- Modify: `internal/ui/renderer/ocean.go`

- [ ] **Step 1: Keep Frame type — it's still tested in projection_test.go**

The `Frame` type is tested by 7 tests in `projection_test.go` (lines 110-185): `TestNewFrame_Dimensions`, `TestFrame_SetGet`, `TestFrame_Clear`, `TestFrame_Render`, `TestFrame_RenderWithColors`, `TestFrame_OutOfBoundsSet`. Keep `frame.go` as-is. The `Frame` type is no longer used by globe rendering but remains as a general-purpose terminal cell buffer.

- [ ] **Step 2: Remove old (rune, string) shading functions**

Remove `OceanShade`, `LandShade`, `AtmosphereGlow`, `StarField`, and `isDesertRegion` — keeping only the RGB versions.

- [ ] **Step 3: Rename RGB functions**

Rename `OceanShadeRGB` → `OceanShade`, `LandShadeRGB` → `LandShade`, etc. now that the old ones are gone. Update all callers.

- [ ] **Step 4: Run full test suite**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/renderer/
git commit -m "refactor(renderer): remove deprecated Frame type and old shading functions"
```
