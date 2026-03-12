package renderer

import (
	"strings"
	"testing"
)

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

func TestCompositeHalfBlocks_SameColor(t *testing.T) {
	pb := NewPixelBuffer(2, 4)
	red := RGB{R: 200, G: 0, B: 0}
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			pb.Set(x, y, red)
		}
	}
	output := pb.CompositeHalfBlocks()
	if !containsRune(output, '█') {
		t.Errorf("expected full block for same-color pairs, got: %q", output)
	}
}

func TestCompositeHalfBlocks_DifferentColors(t *testing.T) {
	pb := NewPixelBuffer(1, 2)
	pb.Set(0, 0, RGB{R: 255, G: 0, B: 0})
	pb.Set(0, 1, RGB{R: 0, G: 0, B: 255})
	output := pb.CompositeHalfBlocks()
	if !containsRune(output, '▀') {
		t.Errorf("expected upper half block for different colors, got: %q", output)
	}
	if !strings.Contains(output, "38;2;255;0;0") {
		t.Errorf("expected red foreground ANSI")
	}
	if !strings.Contains(output, "48;2;0;0;255") {
		t.Errorf("expected blue background ANSI")
	}
}

func TestCompositeHalfBlocks_Dimensions(t *testing.T) {
	pb := NewPixelBuffer(3, 6)
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
