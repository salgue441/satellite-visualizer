package renderer

import "strings"

// Cell represents a single character cell in the frame with color.
type Cell struct {
	Char  rune
	Color string // ANSI escape sequence, empty for default
}

// Frame is a double-buffered text frame for flicker-free terminal rendering.
type Frame struct {
	Width, Height int
	front         [][]Cell
	back          [][]Cell
}

// NewFrame creates a frame with the given dimensions.
func NewFrame(width, height int) *Frame {
	f := &Frame{
		Width:  width,
		Height: height,
	}
	f.front = makeGrid(width, height)
	f.back = makeGrid(width, height)
	return f
}

func makeGrid(width, height int) [][]Cell {
	grid := make([][]Cell, height)
	for y := range grid {
		row := make([]Cell, width)
		for x := range row {
			row[x] = Cell{Char: ' '}
		}
		grid[y] = row
	}
	return grid
}

// Set writes a character and color to the back buffer at (x, y).
// Out-of-bounds writes are silently ignored.
func (f *Frame) Set(x, y int, ch rune, color string) {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return
	}
	f.back[y][x] = Cell{Char: ch, Color: color}
}

// Get returns the cell at (x, y) from the back buffer.
func (f *Frame) Get(x, y int) Cell {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return Cell{Char: ' '}
	}
	return f.back[y][x]
}

// Clear fills the back buffer with spaces and no color.
func (f *Frame) Clear() {
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			f.back[y][x] = Cell{Char: ' '}
		}
	}
}

// Swap swaps front and back buffers.
func (f *Frame) Swap() {
	f.front, f.back = f.back, f.front
}

// Render produces the ANSI string output from the front buffer.
// Uses ANSI reset between color changes. Adds newlines between rows.
func (f *Frame) Render() string {
	var sb strings.Builder
	currentColor := ""

	for y := 0; y < f.Height; y++ {
		if y > 0 {
			sb.WriteString("\n")
		}
		for x := 0; x < f.Width; x++ {
			cell := f.front[y][x]
			if cell.Color != currentColor {
				if currentColor != "" {
					sb.WriteString("\033[0m")
				}
				if cell.Color != "" {
					sb.WriteString(cell.Color)
				}
				currentColor = cell.Color
			}
			sb.WriteRune(cell.Char)
		}
	}
	sb.WriteString("\033[0m")
	return sb.String()
}
