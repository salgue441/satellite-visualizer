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
