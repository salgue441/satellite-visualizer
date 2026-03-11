package renderer

import (
	"math"
	"strings"
	"testing"
)

const tolerance = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

// --- RotateY tests ---

func TestRotateY_90Degrees(t *testing.T) {
	x, y, z := RotateY(1, 0, 0, math.Pi/2)
	if !almostEqual(x, 0) || !almostEqual(y, 0) || !almostEqual(z, -1) {
		t.Errorf("RotateY(1,0,0, π/2) = (%f,%f,%f), want (0,0,-1)", x, y, z)
	}
}

func TestRotateY_90Degrees_ZAxis(t *testing.T) {
	x, y, z := RotateY(0, 0, 1, math.Pi/2)
	if !almostEqual(x, 1) || !almostEqual(y, 0) || !almostEqual(z, 0) {
		t.Errorf("RotateY(0,0,1, π/2) = (%f,%f,%f), want (1,0,0)", x, y, z)
	}
}

func TestRotateY_ZeroDegrees(t *testing.T) {
	x, y, z := RotateY(3, 5, 7, 0)
	if !almostEqual(x, 3) || !almostEqual(y, 5) || !almostEqual(z, 7) {
		t.Errorf("RotateY(3,5,7, 0) = (%f,%f,%f), want (3,5,7)", x, y, z)
	}
}

// --- RotateX tests ---

func TestRotateX_90Degrees_YAxis(t *testing.T) {
	x, y, z := RotateX(0, 1, 0, math.Pi/2)
	if !almostEqual(x, 0) || !almostEqual(y, 0) || !almostEqual(z, 1) {
		t.Errorf("RotateX(0,1,0, π/2) = (%f,%f,%f), want (0,0,1)", x, y, z)
	}
}

func TestRotateX_90Degrees_ZAxis(t *testing.T) {
	x, y, z := RotateX(0, 0, 1, math.Pi/2)
	if !almostEqual(x, 0) || !almostEqual(y, -1) || !almostEqual(z, 0) {
		t.Errorf("RotateX(0,0,1, π/2) = (%f,%f,%f), want (0,-1,0)", x, y, z)
	}
}

// --- Project3DTo2D tests ---

func TestProject3DTo2D_Center(t *testing.T) {
	// sphereR=12 means sphere spans 12 rows from center. Screen 80x24.
	// Center: sx = 40 + 0 = 40, sy = 12 - 0 = 12
	sx, sy, visible := Project3DTo2D(0, 0, 1, 12.0, 80, 24)
	if !visible {
		t.Error("Center point should be visible")
	}
	if sx != 40 || sy != 12 {
		t.Errorf("Center point projected to (%d,%d), want (40,12)", sx, sy)
	}
}

func TestProject3DTo2D_RightEdge(t *testing.T) {
	// x=1 on unit sphere, sphereR=12, charAspect=2: sx = 40 + 1*12*2 = 64
	sx, _, visible := Project3DTo2D(1, 0, 1, 12.0, 80, 24)
	if !visible {
		t.Error("Right edge point should be visible")
	}
	if sx != 64 {
		t.Errorf("Right edge projected to sx=%d, want 64", sx)
	}
}

func TestProject3DTo2D_BehindSphere(t *testing.T) {
	_, _, visible := Project3DTo2D(0, 0, -1, 12.0, 80, 24)
	if visible {
		t.Error("Point behind sphere should not be visible")
	}
}

func TestProject3DTo2D_TopOfScreen(t *testing.T) {
	// y=1 on unit sphere, sphereR=12: sy = 12 - 1*12 = 0
	_, sy, visible := Project3DTo2D(0, 1, 1, 12.0, 80, 24)
	if !visible {
		t.Error("Top point should be visible")
	}
	if sy != 0 {
		t.Errorf("Top of screen sy = %d, want 0", sy)
	}
}

func TestProject3DTo2D_BottomOfScreen(t *testing.T) {
	// y=-0.9 on unit sphere, sphereR=12: sy = 12 - (-0.9)*12 = 22 (near bottom)
	_, sy, visible := Project3DTo2D(0, -0.9, 1, 12.0, 80, 24)
	if !visible {
		t.Error("Near-bottom point should be visible")
	}
	if sy != 22 {
		t.Errorf("Near-bottom sy = %d, want 22", sy)
	}
}

// --- Frame tests ---

func TestNewFrame_Dimensions(t *testing.T) {
	f := NewFrame(80, 24)
	if f.Width != 80 || f.Height != 24 {
		t.Errorf("Frame dimensions = (%d,%d), want (80,24)", f.Width, f.Height)
	}
}

func TestFrame_SetGet(t *testing.T) {
	f := NewFrame(80, 24)
	f.Set(5, 3, 'X', "\033[31m")
	c := f.Get(5, 3)
	if c.Char != 'X' {
		t.Errorf("Get Char = %c, want X", c.Char)
	}
	if c.Color != "\033[31m" {
		t.Errorf("Get Color = %q, want \\033[31m", c.Color)
	}
}

func TestFrame_Clear(t *testing.T) {
	f := NewFrame(10, 5)
	f.Set(3, 2, 'A', "\033[32m")
	f.Clear()
	c := f.Get(3, 2)
	if c.Char != ' ' {
		t.Errorf("After Clear, Char = %c, want space", c.Char)
	}
	if c.Color != "" {
		t.Errorf("After Clear, Color = %q, want empty", c.Color)
	}
}

func TestFrame_Render(t *testing.T) {
	f := NewFrame(3, 2)
	f.Set(0, 0, 'A', "")
	f.Set(1, 0, 'B', "")
	f.Set(2, 0, 'C', "")
	f.Set(0, 1, 'D', "")
	f.Set(1, 1, 'E', "")
	f.Set(2, 1, 'F', "")
	f.Swap()
	out := f.Render()
	if !strings.Contains(out, "ABC") {
		t.Errorf("Render should contain 'ABC', got %q", out)
	}
	if !strings.Contains(out, "\n") {
		t.Error("Render should contain newlines between rows")
	}
}

func TestFrame_RenderWithColors(t *testing.T) {
	f := NewFrame(2, 1)
	f.Set(0, 0, 'R', "\033[31m")
	f.Set(1, 0, 'G', "\033[32m")
	f.Swap()
	out := f.Render()
	if !strings.Contains(out, "\033[31m") {
		t.Error("Render should contain red ANSI code")
	}
	if !strings.Contains(out, "\033[32m") {
		t.Error("Render should contain green ANSI code")
	}
	if !strings.Contains(out, "\033[0m") {
		t.Error("Render should end with ANSI reset")
	}
}

func TestFrame_OutOfBoundsSet(t *testing.T) {
	f := NewFrame(10, 5)
	// Should not panic
	f.Set(-1, 0, 'X', "")
	f.Set(0, -1, 'X', "")
	f.Set(10, 0, 'X', "")
	f.Set(0, 5, 'X', "")
	f.Set(100, 100, 'X', "")
}
