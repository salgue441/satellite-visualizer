package renderer

import (
	"fmt"
	"strings"
)

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

// CompositeHalfBlocks merges pixel pairs into half-block terminal output.
func (pb *PixelBuffer) CompositeHalfBlocks() string {
	termRows := pb.Height / 2
	var sb strings.Builder
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

			if firstCell || top != prevFg || bot != prevBg {
				if top == bot {
					fmt.Fprintf(&sb, "\033[38;2;%d;%d;%dm", top.R, top.G, top.B)
					sb.WriteRune('█')
				} else {
					fmt.Fprintf(&sb, "\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm",
						top.R, top.G, top.B, bot.R, bot.G, bot.B)
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
