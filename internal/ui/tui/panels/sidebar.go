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
		// Reserve lines for title + separator
		visibleLines := p.height - 2
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

// Render returns the styled satellite list string.
func (p *SidebarPanel) Render(focused bool) string {
	var sb strings.Builder

	title := p.styles.Title.Render("SATELLITES")
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(p.styles.Subtitle.Render(" " + strings.Repeat("\u2500", p.width-2)))
	sb.WriteString("\n")

	sats := p.filtered()
	visibleLines := p.height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	if len(sats) == 0 {
		sb.WriteString(p.styles.Subtitle.Render(" No satellites"))
		return sb.String()
	}

	end := p.offset + visibleLines
	if end > len(sats) {
		end = len(sats)
	}

	for i := p.offset; i < end; i++ {
		sat := sats[i]

		// Pick icon
		icon := "\u25cf" // ●
		if sat.ConstellationName == "stations" {
			icon = "\u2605" // ★
		}

		// Pick color from constellation colors
		color, ok := renderer.ConstellationColors[sat.ConstellationName]
		if !ok {
			color = renderer.DefaultSatColor
		}

		name := sat.Name
		if len(name) > p.width-4 {
			name = name[:p.width-4]
		}

		line := fmt.Sprintf(" %s %s", icon, name)

		if i == p.selected {
			// Apply color via raw ANSI since constellation colors are ANSI strings
			style := p.styles.Selected
			if color != "" {
				style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
			}
			sb.WriteString(style.Render(line))
		} else {
			// Use the constellation color via lipgloss
			style := p.styles.Unselected
			_ = color // constellation color used for the icon in the rendered string
			sb.WriteString(style.Render(line))
		}

		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	if p.searching {
		sb.WriteString("\n")
		sb.WriteString(p.styles.Subtitle.Render(fmt.Sprintf(" /%s", p.searchQuery)))
	}

	return sb.String()
}
