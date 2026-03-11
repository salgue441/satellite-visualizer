package panels

import (
	"fmt"
	"strings"
	"time"

	"satellite-visualizer/internal/ui/tui"
)

// StatusPanel shows system status information.
type StatusPanel struct {
	source    string // "CelesTrak" or "Space-Track"
	satCount  int
	fps       float64
	lastFetch time.Time
	staleData bool
	width     int
	styles    tui.Styles
}

// NewStatusPanel creates a status panel with the given width.
func NewStatusPanel(width int) *StatusPanel {
	return &StatusPanel{
		width:  width,
		styles: tui.DefaultStyles(),
	}
}

// SetSource sets the data source name.
func (p *StatusPanel) SetSource(source string) {
	p.source = source
}

// SetSatCount sets the number of tracked satellites.
func (p *StatusPanel) SetSatCount(count int) {
	p.satCount = count
}

// SetFPS sets the current frames per second.
func (p *StatusPanel) SetFPS(fps float64) {
	p.fps = fps
}

// SetLastFetch sets the time of the last data fetch.
func (p *StatusPanel) SetLastFetch(t time.Time) {
	p.lastFetch = t
}

// SetStale sets whether the data is stale.
func (p *StatusPanel) SetStale(stale bool) {
	p.staleData = stale
}

// Resize updates the panel width.
func (p *StatusPanel) Resize(width int) {
	p.width = width
}

// Render returns the status bar string.
func (p *StatusPanel) Render() string {
	var sb strings.Builder

	label := p.styles.StatusLabel
	value := p.styles.StatusValue

	sb.WriteString(label.Render("Source: "))
	sb.WriteString(value.Render(p.source))
	sb.WriteString("  ")
	sb.WriteString(label.Render("Sats: "))
	sb.WriteString(value.Render(fmt.Sprintf("%d", p.satCount)))
	sb.WriteString("  ")
	sb.WriteString(label.Render("FPS: "))
	sb.WriteString(value.Render(fmt.Sprintf("%.0f", p.fps)))
	sb.WriteString("  ")

	// Time since last fetch
	if !p.lastFetch.IsZero() {
		ago := time.Since(p.lastFetch)
		var agoStr string
		switch {
		case ago < time.Minute:
			agoStr = fmt.Sprintf("%ds ago", int(ago.Seconds()))
		case ago < time.Hour:
			agoStr = fmt.Sprintf("%dm ago", int(ago.Minutes()))
		default:
			agoStr = fmt.Sprintf("%dh ago", int(ago.Hours()))
		}
		sb.WriteString(value.Render(fmt.Sprintf("\u21bb %s", agoStr)))
	}

	if p.staleData {
		sb.WriteString("  ")
		sb.WriteString(p.styles.StatusWarn.Render("STALE DATA"))
	}

	return p.styles.StatusBar.Render(sb.String())
}
