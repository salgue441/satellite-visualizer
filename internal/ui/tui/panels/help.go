package panels

import (
	"fmt"
	"strings"

	"satellite-visualizer/internal/ui/tui"
)

// HelpPanel renders a keyboard shortcut reference overlay.
type HelpPanel struct {
	keys   tui.KeyMap
	width  int
	height int
	styles tui.Styles
}

// NewHelpPanel creates a help panel with the given key map and dimensions.
func NewHelpPanel(keys tui.KeyMap, width, height int) *HelpPanel {
	return &HelpPanel{
		keys:   keys,
		width:  width,
		height: height,
		styles: tui.DefaultStyles(),
	}
}

// Resize updates the panel dimensions.
func (p *HelpPanel) Resize(width, height int) {
	p.width = width
	p.height = height
}

// Render returns the help overlay string.
func (p *HelpPanel) Render() string {
	var sb strings.Builder

	title := p.styles.Title.Render("KEYBOARD SHORTCUTS")
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(p.styles.Subtitle.Render(" " + strings.Repeat("\u2500", 36)))
	sb.WriteString("\n")

	// Two-column layout: left column and right column bindings
	type binding struct {
		key  string
		desc string
	}

	left := []binding{
		{p.keys.Up.Help().Key, p.keys.Up.Help().Desc},
		{p.keys.Down.Help().Key, p.keys.Down.Help().Desc},
		{p.keys.Left.Help().Key, p.keys.Left.Help().Desc},
		{p.keys.Right.Help().Key, p.keys.Right.Help().Desc},
		{p.keys.Enter.Help().Key, p.keys.Enter.Help().Desc},
		{p.keys.Search.Help().Key, p.keys.Search.Help().Desc},
		{p.keys.Help.Help().Key, p.keys.Help.Help().Desc},
	}

	right := []binding{
		{p.keys.ZoomIn.Help().Key, p.keys.ZoomIn.Help().Desc},
		{p.keys.ZoomOut.Help().Key, p.keys.ZoomOut.Help().Desc},
		{p.keys.Tab.Help().Key, p.keys.Tab.Help().Desc},
		{p.keys.Pause.Help().Key, p.keys.Pause.Help().Desc},
		{p.keys.CycleConst.Help().Key, p.keys.CycleConst.Help().Desc},
		{p.keys.Refresh.Help().Key, p.keys.Refresh.Help().Desc},
		{p.keys.Quit.Help().Key, p.keys.Quit.Help().Desc},
	}

	keyStyle := p.styles.HelpKey
	descStyle := p.styles.HelpDesc

	rows := len(left)
	if len(right) > rows {
		rows = len(right)
	}

	for i := 0; i < rows; i++ {
		var leftStr, rightStr string

		if i < len(left) {
			leftStr = fmt.Sprintf(" %s  %s",
				keyStyle.Render(fmt.Sprintf("%-7s", left[i].key)),
				descStyle.Render(fmt.Sprintf("%-14s", left[i].desc)),
			)
		} else {
			leftStr = strings.Repeat(" ", 24)
		}

		if i < len(right) {
			rightStr = fmt.Sprintf(" %s  %s",
				keyStyle.Render(fmt.Sprintf("%-7s", right[i].key)),
				descStyle.Render(right[i].desc),
			)
		}

		sb.WriteString(leftStr)
		sb.WriteString(rightStr)

		if i < rows-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
