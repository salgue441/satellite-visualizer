package panels

import (
	"fmt"
	"math"
	"strings"

	"satellite-visualizer/internal/domain"
	"satellite-visualizer/internal/ui/tui"

	"github.com/charmbracelet/lipgloss"
)

// DetailsPanel shows information about the selected satellite.
type DetailsPanel struct {
	satellite *domain.SatelliteState
	width     int
	height    int
	styles    tui.Styles
}

// NewDetailsPanel creates a details panel with the given dimensions.
func NewDetailsPanel(width, height int) *DetailsPanel {
	return &DetailsPanel{
		width:  width,
		height: height,
		styles: tui.DefaultStyles(),
	}
}

// SetSatellite sets the satellite to display details for.
func (p *DetailsPanel) SetSatellite(sat *domain.SatelliteState) {
	p.satellite = sat
}

// Resize updates the panel dimensions.
func (p *DetailsPanel) Resize(width, height int) {
	p.width = width
	p.height = height
}

// Render returns formatted satellite details.
func (p *DetailsPanel) Render() string {
	if p.satellite == nil {
		return p.styles.Subtitle.Render(" No satellite selected")
	}

	var sb strings.Builder
	sat := p.satellite

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	sb.WriteString(titleStyle.Render(fmt.Sprintf(" SELECTED: %s", sat.Name)))
	sb.WriteString("\n")

	// Compute velocity magnitude
	vel := math.Sqrt(sat.Vel.X*sat.Vel.X + sat.Vel.Y*sat.Vel.Y + sat.Vel.Z*sat.Vel.Z)

	// Format latitude
	latDir := "N"
	lat := sat.Geo.Latitude
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}

	// Format longitude
	lonDir := "E"
	lon := sat.Geo.Longitude
	if lon < 0 {
		lonDir = "W"
		lon = -lon
	}

	labelStyle := p.styles.StatusLabel
	valueStyle := p.styles.StatusValue

	sb.WriteString(fmt.Sprintf(" %s %s   %s %s\n",
		labelStyle.Render("Alt:"),
		valueStyle.Render(fmt.Sprintf("%.1f km", sat.Geo.Altitude)),
		labelStyle.Render("Lat:"),
		valueStyle.Render(fmt.Sprintf("%.1f\u00b0 %s", lat, latDir)),
	))

	sb.WriteString(fmt.Sprintf(" %s %s   %s %s\n",
		labelStyle.Render("Lon:"),
		valueStyle.Render(fmt.Sprintf("%.1f\u00b0 %s", lon, lonDir)),
		labelStyle.Render("Vel:"),
		valueStyle.Render(fmt.Sprintf("%.2f km/s", vel)),
	))

	sb.WriteString(fmt.Sprintf(" %s %s",
		labelStyle.Render("Constellation:"),
		valueStyle.Render(sat.ConstellationName),
	))

	return sb.String()
}
