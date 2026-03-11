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
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
		return hint.Render(" Press tab to focus sidebar, ↑↓ to browse, enter to select")
	}

	var sb strings.Builder
	sat := p.satellite

	// Title line
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	sb.WriteString(titleStyle.Render(fmt.Sprintf(" %s", sat.Name)))
	constStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	sb.WriteString(constStyle.Render(fmt.Sprintf("  [%s]", sat.ConstellationName)))
	sb.WriteString("\n")

	// Velocity magnitude
	vel := math.Sqrt(sat.Vel.X*sat.Vel.X + sat.Vel.Y*sat.Vel.Y + sat.Vel.Z*sat.Vel.Z)

	// Format coordinates
	latDir := "N"
	lat := sat.Geo.Latitude
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}
	lonDir := "E"
	lon := sat.Geo.Longitude
	if lon < 0 {
		lonDir = "W"
		lon = -lon
	}

	label := p.styles.StatusLabel
	value := p.styles.StatusValue

	sb.WriteString(fmt.Sprintf(" %s %-12s  %s %-14s  %s %s\n",
		label.Render("Alt:"),
		value.Render(fmt.Sprintf("%.0f km", sat.Geo.Altitude)),
		label.Render("Lat:"),
		value.Render(fmt.Sprintf("%.2f° %s", lat, latDir)),
		label.Render("Vel:"),
		value.Render(fmt.Sprintf("%.2f km/s", vel)),
	))

	sb.WriteString(fmt.Sprintf(" %s %-12s  %s %-14s",
		label.Render("     "),
		value.Render(""),
		label.Render("Lon:"),
		value.Render(fmt.Sprintf("%.2f° %s", lon, lonDir)),
	))

	return sb.String()
}
