package app

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/config"
	"satellite-visualizer/internal/domain"
	"satellite-visualizer/internal/ui/tui"
	"satellite-visualizer/internal/ui/tui/panels"
)

type focusedPanel int

const (
	focusGlobe focusedPanel = iota
	focusSidebar
)

// scheduledFetchMsg triggers a background data fetch after a delay.
type scheduledFetchMsg struct{}

// App is the root Bubbletea model composing all dashboard panels.
type App struct {
	// Panels
	globe   *panels.GlobePanel
	sidebar *panels.SidebarPanel
	details *panels.DetailsPanel
	status  *panels.StatusPanel
	help    *panels.HelpPanel

	// State
	keys           tui.KeyMap
	styles         tui.Styles
	focused        focusedPanel
	showHelp       bool
	paused         bool
	constellations []domain.Constellation
	allSatellites  []domain.SatelliteState

	// Dependencies
	tracker *application.Tracker
	cfg     *config.AppConfig

	// Dimensions
	width  int
	height int

	// Performance tracking
	lastRender time.Time
	frameCount int
	fps        float64
}

// NewApp creates the root application model.
func NewApp(tracker *application.Tracker, cfg *config.AppConfig) *App {
	keys := tui.DefaultKeyMap()
	styles := tui.DefaultStyles()

	return &App{
		globe:   panels.NewGlobePanel(60, 20),
		sidebar: panels.NewSidebarPanel(25, 20),
		details: panels.NewDetailsPanel(60, 5),
		status:  panels.NewStatusPanel(25),
		help:    panels.NewHelpPanel(keys, 60, 20),
		keys:    keys,
		styles:  styles,
		tracker: tracker,
		cfg:     cfg,
	}
}

// Init starts the tick loop and initial data fetch.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.tickCmd(),
		a.fetchCmd(),
	)
}

// tickCmd returns a command that sends TickMsg at the configured FPS.
func (a *App) tickCmd() tea.Cmd {
	fps := 30
	if a.cfg != nil && a.cfg.TargetFPS > 0 {
		fps = a.cfg.TargetFPS
	}
	interval := time.Duration(1000/fps) * time.Millisecond
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tui.TickMsg(t)
	})
}

// fetchCmd fetches constellation data in the background.
func (a *App) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		if a.tracker == nil {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		constellations, err := a.tracker.GetConstellations(ctx, time.Now())
		if err != nil {
			return tui.ErrMsg{Err: err}
		}
		return tui.FetchCompleteMsg{Constellations: constellations}
	}
}

// scheduleFetch returns a command that triggers a fetch after the configured interval.
func (a *App) scheduleFetch() tea.Cmd {
	interval := 15 * time.Minute
	if a.cfg != nil && a.cfg.FetchInterval > 0 {
		interval = a.cfg.FetchInterval
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return scheduledFetchMsg{}
	})
}

// Update handles all messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.resizePanels()
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case tui.TickMsg:
		if !a.paused {
			a.globe.Globe().RotationY += 0.02
		}
		a.updateFPS()
		return a, a.tickCmd()

	case tui.FetchCompleteMsg:
		a.constellations = msg.Constellations
		a.allSatellites = flattenSatellites(msg.Constellations)
		a.sidebar.Update(a.allSatellites)
		a.status.SetSatCount(len(a.allSatellites))
		a.status.SetSource("CelesTrak")
		a.status.SetLastFetch(time.Now())
		a.status.SetStale(false)
		return a, a.scheduleFetch()

	case scheduledFetchMsg:
		return a, a.fetchCmd()

	case tui.ErrMsg:
		a.status.SetStale(true)
		return a, a.scheduleFetch()
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit
	case key.Matches(msg, a.keys.Help):
		a.showHelp = !a.showHelp
	case key.Matches(msg, a.keys.Tab):
		if a.focused == focusGlobe {
			a.focused = focusSidebar
		} else {
			a.focused = focusGlobe
		}
	case key.Matches(msg, a.keys.Pause):
		a.paused = !a.paused
	case key.Matches(msg, a.keys.Left):
		a.globe.Globe().RotationY -= 0.1
	case key.Matches(msg, a.keys.Right):
		a.globe.Globe().RotationY += 0.1
	case key.Matches(msg, a.keys.Up):
		a.sidebar.MoveUp()
		if sel := a.sidebar.Selected(); sel != nil {
			a.details.SetSatellite(sel)
		}
	case key.Matches(msg, a.keys.Down):
		a.sidebar.MoveDown()
		if sel := a.sidebar.Selected(); sel != nil {
			a.details.SetSatellite(sel)
		}
	case key.Matches(msg, a.keys.ZoomIn):
		a.globe.Globe().Zoom *= 1.1
	case key.Matches(msg, a.keys.ZoomOut):
		g := a.globe.Globe()
		g.Zoom /= 1.1
		if g.Zoom < 0.5 {
			g.Zoom = 0.5
		}
	case key.Matches(msg, a.keys.Enter):
		if sel := a.sidebar.Selected(); sel != nil {
			a.details.SetSatellite(sel)
		}
	case key.Matches(msg, a.keys.Refresh):
		return a, a.fetchCmd()
	}
	return a, nil
}

// View composes the dashboard layout.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Help overlay takes over the entire screen
	if a.showHelp {
		helpView := a.help.Render()
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, helpView)
	}

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		PaddingLeft(1)
	pauseIndicator := ""
	if a.paused {
		pauseStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
		pauseIndicator = "  " + pauseStyle.Render("[PAUSED]")
	}
	titleBar := titleStyle.Render("SATELLITE TRACKER") + pauseIndicator
	titleBar = lipgloss.NewStyle().Width(a.width).Render(titleBar)

	// Layout: sidebar takes fixed width, globe gets the rest
	sidebarWidth := min(32, a.width/4)
	globeWidth := a.width - sidebarWidth - 4 // borders
	bottomHeight := 5
	globeHeight := a.height - bottomHeight - 5 // borders + bottom row + title

	if globeHeight < 5 {
		globeHeight = 5
	}

	a.globe.Resize(globeWidth, globeHeight)

	// Render panels with active/inactive borders based on focus
	globeBorder := a.styles.BorderInactive
	sidebarBorder := a.styles.BorderInactive
	if a.focused == focusGlobe {
		globeBorder = a.styles.BorderActive
	} else {
		sidebarBorder = a.styles.BorderActive
	}

	globeView := globeBorder.Width(globeWidth).Height(globeHeight).
		Render(a.globe.Render(a.allSatellites))
	sidebarView := sidebarBorder.Width(sidebarWidth).Height(globeHeight).
		Render(a.sidebar.Render(a.focused == focusSidebar))

	// Bottom row: details (left) + status (right)
	a.status.SetFPS(a.fps)
	detailsView := a.styles.BorderInactive.Width(globeWidth).Height(bottomHeight).
		Render(a.details.Render())
	statusView := a.styles.BorderInactive.Width(sidebarWidth).Height(bottomHeight).
		Render(a.status.Render())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, globeView, sidebarView)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, detailsView, statusView)

	return lipgloss.JoinVertical(lipgloss.Left, titleBar, topRow, bottomRow)
}

// flattenSatellites collects all satellites from all constellations into a single slice.
func flattenSatellites(constellations []domain.Constellation) []domain.SatelliteState {
	var all []domain.SatelliteState
	for _, c := range constellations {
		all = append(all, c.Satellites...)
	}
	return all
}

func (a *App) resizePanels() {
	sidebarWidth := min(32, a.width/4)
	bottomHeight := 5
	globeHeight := a.height - bottomHeight - 5 // title bar

	if globeHeight < 5 {
		globeHeight = 5
	}

	a.globe.Resize(a.width-sidebarWidth-4, globeHeight)
	a.sidebar = panels.NewSidebarPanel(sidebarWidth, globeHeight)
	a.sidebar.Update(a.allSatellites)
}

func (a *App) updateFPS() {
	a.frameCount++
	now := time.Now()
	elapsed := now.Sub(a.lastRender)
	if elapsed >= time.Second {
		a.fps = float64(a.frameCount) / elapsed.Seconds()
		a.frameCount = 0
		a.lastRender = now
	}
}
