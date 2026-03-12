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
		pb:     renderer.NewPixelBuffer(width, height*2),
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
