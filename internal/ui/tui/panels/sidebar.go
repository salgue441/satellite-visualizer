package panels

import (
	"fmt"
	"strings"

	"satellite-visualizer/internal/domain"
	"satellite-visualizer/internal/ui/renderer"
	"satellite-visualizer/internal/ui/tui"

	"github.com/charmbracelet/lipgloss"
)

// SidebarPanel shows a scrollable list of satellites.
type SidebarPanel struct {
	satellites  []domain.SatelliteState
	selected    int
	offset      int // scroll offset
	height      int
	width       int
	filter      string // constellation filter
	searchQuery string
	searching   bool
	styles      tui.Styles
}

// NewSidebarPanel creates a sidebar panel with the given dimensions.
func NewSidebarPanel(width, height int) *SidebarPanel {
	return &SidebarPanel{
		width:  width,
		height: height,
		styles: tui.DefaultStyles(),
	}
}

// Update sets the current satellite list.
func (p *SidebarPanel) Update(satellites []domain.SatelliteState) {
	p.satellites = satellites
	if p.selected >= len(p.satellites) {
		p.selected = max(0, len(p.satellites)-1)
	}
}

// MoveUp navigates up in the list.
func (p *SidebarPanel) MoveUp() {
	if p.selected > 0 {
		p.selected--
		if p.selected < p.offset {
			p.offset = p.selected
		}
	}
}

// MoveDown navigates down in the list.
func (p *SidebarPanel) MoveDown() {
	if p.selected < len(p.satellites)-1 {
		p.selected++
		visibleLines := p.height - 3
		if visibleLines < 1 {
			visibleLines = 1
		}
		if p.selected >= p.offset+visibleLines {
			p.offset = p.selected - visibleLines + 1
		}
	}
}

// Selected returns the currently selected satellite, if any.
func (p *SidebarPanel) Selected() *domain.SatelliteState {
	if len(p.satellites) == 0 || p.selected >= len(p.satellites) {
		return nil
	}
	return &p.satellites[p.selected]
}

// SetFilter sets the constellation filter.
func (p *SidebarPanel) SetFilter(filter string) {
	p.filter = filter
}

// SetSearch sets the search query and mode.
func (p *SidebarPanel) SetSearch(query string, active bool) {
	p.searchQuery = query
	p.searching = active
}

// filtered returns the satellites matching the current filter and search query.
func (p *SidebarPanel) filtered() []domain.SatelliteState {
	if p.filter == "" && p.searchQuery == "" {
		return p.satellites
	}
	var result []domain.SatelliteState
	for _, s := range p.satellites {
		if p.filter != "" && s.ConstellationName != p.filter {
			continue
		}
		if p.searchQuery != "" && !strings.Contains(
			strings.ToLower(s.Name), strings.ToLower(p.searchQuery)) {
			continue
		}
		result = append(result, s)
	}
	return result
}

// constellationLipglossColor converts a constellation name to a lipgloss color string.
func constellationLipglossColor(name string) string {
	rgb, ok := renderer.ConstellationColors[name]
	if !ok {
		rgb = renderer.DefaultSatRGB
	}
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

// Render returns the styled satellite list string.
func (p *SidebarPanel) Render(focused bool) string {
	var sb strings.Builder

	title := p.styles.Title.Render("SATELLITES")
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	sb.WriteString(fmt.Sprintf("%s %s", title, countStyle.Render(fmt.Sprintf("(%d)", len(p.satellites)))))
	sb.WriteString("\n")
	sb.WriteString(p.styles.Subtitle.Render(strings.Repeat("─", p.width)))
	sb.WriteString("\n")

	sats := p.filtered()
	visibleLines := p.height - 3
	if visibleLines < 1 {
		visibleLines = 1
	}

	if len(sats) == 0 {
		sb.WriteString(p.styles.Subtitle.Render(" No satellites"))
		return sb.String()
	}

	end := min(p.offset+visibleLines, len(sats))

	if p.offset > 0 {
		sb.WriteString(p.styles.Subtitle.Render(fmt.Sprintf(" ▲ %d more", p.offset)))
		sb.WriteString("\n")
		end = min(p.offset+visibleLines-1, len(sats))
	}

	for i := p.offset; i < end; i++ {
		sat := sats[i]

		icon := "●"
		if sat.ConstellationName == "stations" {
			icon = "★"
		}

		name := sat.Name
		maxName := p.width - 4
		if maxName < 4 {
			maxName = 4
		}
		if len(name) > maxName {
			name = name[:maxName-1] + "…"
		}

		line := fmt.Sprintf(" %s %s", icon, name)

		if i == p.selected {
			padded := line
			for len(padded) < p.width {
				padded += " "
			}
			selectedStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("16")).
				Background(lipgloss.Color("39"))
			sb.WriteString(selectedStyle.Render(padded))
		} else {
			// Color the icon with constellation color
			iconColor := constellationLipglossColor(sat.ConstellationName)
			iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(iconColor))
			nameStyle := p.styles.Unselected

			sb.WriteString(fmt.Sprintf(" %s %s", iconStyle.Render(icon), nameStyle.Render(name)))
		}

		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	remaining := len(sats) - end
	if remaining > 0 {
		sb.WriteString("\n")
		sb.WriteString(p.styles.Subtitle.Render(fmt.Sprintf(" ▼ %d more", remaining)))
	}

	if p.searching {
		sb.WriteString("\n")
		sb.WriteString(p.styles.Subtitle.Render(fmt.Sprintf(" /%s", p.searchQuery)))
	}

	return sb.String()
}
