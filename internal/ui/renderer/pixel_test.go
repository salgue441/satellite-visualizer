package renderer

import (
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
