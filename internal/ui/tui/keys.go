package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	ZoomIn     key.Binding
	ZoomOut    key.Binding
	Tab        key.Binding
	Enter      key.Binding
	Search     key.Binding
	Pause      key.Binding
	CycleConst key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Quit       key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("\u2191/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("\u2193/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("\u2190/h", "rotate left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("\u2192/l", "rotate right"),
		),
		ZoomIn: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "zoom in"),
		),
		ZoomOut: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "zoom out"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle focus"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "pause rotation"),
		),
		CycleConst: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cycle constellation"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh data"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
