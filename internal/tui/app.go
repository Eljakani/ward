package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/eventbus"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/tui/components"
	"github.com/eljakani/ward/internal/tui/theme"
	"github.com/eljakani/ward/internal/tui/views"
)

const version = "0.2.0"

// App is the root Bubble Tea model for Laravel Ward.
type App struct {
	bus        *eventbus.EventBus
	targetPath string

	// Layout
	width  int
	height int
	ready  bool

	// View state
	activeView ViewID

	// Theme and help
	theme   *theme.Theme
	keys    KeyMap
	help    help.Model
	spinner spinner.Model

	// Shared state (updated from EventBus messages)
	projectName    string
	laravelVersion string
	phpVersion     string
	packageCount   int
	currentStage   models.PipelineStage
	scanners       []models.ScannerInfo
	findings       []models.Finding
	severityCounts map[models.Severity]int
	eventLog       []eventbus.Event
	report         *models.ScanReport
	scanRunning    bool
	scanComplete   bool
	scanError      error

	// Sub-views
	scanView    *views.ScanView
	resultsView *views.ResultsView
}

// NewApp creates the root TUI model.
func NewApp(bus *eventbus.EventBus, targetPath string) *App {
	t := theme.DefaultTheme()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(t.Colors.StageActive)

	h := help.New()
	h.ShowAll = false

	return &App{
		bus:            bus,
		targetPath:     targetPath,
		theme:          t,
		keys:           DefaultKeyMap(),
		help:           h,
		spinner:        s,
		activeView:     ViewScan,
		severityCounts: make(map[models.Severity]int),
		scanners:       make([]models.ScannerInfo, 0),
		findings:       make([]models.Finding, 0),
		eventLog:       make([]eventbus.Event, 0, 100),
		scanView:       views.NewScanView(t),
	}
}

// Init returns the initial commands.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		tickCmd(),
	)
}

// Update handles all incoming messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.propagateSize()
		return a, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
			return a, nil
		case key.Matches(msg, a.keys.Tab):
			if a.scanComplete && a.activeView == ViewScan {
				a.activeView = ViewResults
				return a, nil
			} else if a.activeView == ViewResults {
				// Delegate tab to results view for panel switching
				cmd := a.resultsView.HandleKey(msg)
				return a, cmd
			}
			return a, nil
		case key.Matches(msg, a.keys.Escape):
			if a.activeView == ViewResults {
				a.activeView = ViewScan
				return a, nil
			}
			return a, nil
		}

		// Delegate to active view
		cmd := a.delegateKeyToView(msg)
		cmds = append(cmds, cmd)

	case eventbus.BusEventMsg:
		cmd := a.handleBusEvent(msg.Event)
		cmds = append(cmds, cmd)

	case tickMsg:
		// Forward tick for spinner animation in scan view
		if a.activeView == ViewScan {
			cmd := a.scanView.Tick(msg)
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, tickCmd())

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case switchViewMsg:
		a.activeView = msg.view
	}

	return a, tea.Batch(cmds...)
}

// View renders the entire screen.
func (a *App) View() string {
	if !a.ready {
		return "\n  Initializing Laravel Ward..."
	}

	header := a.renderHeader()
	footer := a.renderFooter()

	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
	contentH := a.height - headerH - footerH
	if contentH < 4 {
		contentH = 4
	}

	var content string
	switch a.activeView {
	case ViewScan:
		content = a.scanView.View(a.width, contentH)
	case ViewResults:
		if a.resultsView != nil {
			content = a.resultsView.View(a.width, contentH)
		} else {
			content = a.theme.Muted.Render("\n  No results available yet.")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (a *App) handleBusEvent(event eventbus.Event) tea.Cmd {
	// Append to event log (capped at 200)
	a.eventLog = append(a.eventLog, event)
	if len(a.eventLog) > 200 {
		a.eventLog = a.eventLog[len(a.eventLog)-200:]
	}

	switch event.Type {
	case eventbus.EventScanStarted:
		data := event.Data.(eventbus.ScanStartedData)
		a.scanRunning = true
		a.projectName = data.ProjectName

	case eventbus.EventContextResolved:
		data := event.Data.(eventbus.ContextResolvedData)
		if data.ProjectName != "" {
			a.projectName = data.ProjectName
		}
		a.laravelVersion = data.LaravelVersion
		a.phpVersion = data.PHPVersion
		a.packageCount = data.PackageCount
		a.scanView.UpdateProjectInfo(a.projectName, a.laravelVersion, a.phpVersion, a.packageCount)
		a.propagateSize() // recalculate layout with info bar

	case eventbus.EventStageStarted:
		data := event.Data.(eventbus.StageStartedData)
		a.currentStage = data.Stage
		a.scanView.UpdateStage(a.currentStage)

	case eventbus.EventScannerRegistered:
		data := event.Data.(eventbus.ScannerRegisteredData)
		a.scanners = append(a.scanners, models.ScannerInfo{
			Name:        data.Name,
			Description: data.Description,
			Status:      models.ScannerPending,
		})
		a.scanView.UpdateScanners(a.scanners)

	case eventbus.EventScannerStarted:
		data := event.Data.(eventbus.ScannerStartedData)
		a.updateScannerStatus(data.Name, models.ScannerRunning, 0)

	case eventbus.EventScannerCompleted:
		data := event.Data.(eventbus.ScannerCompletedData)
		a.updateScannerStatus(data.Name, models.ScannerDone, data.FindingCount)

	case eventbus.EventScannerFailed:
		data := event.Data.(eventbus.ScannerFailedData)
		a.updateScannerStatus(data.Name, models.ScannerError, 0)
		for i := range a.scanners {
			if a.scanners[i].Name == data.Name {
				a.scanners[i].Error = data.Error
				break
			}
		}

	case eventbus.EventFindingDiscovered:
		data := event.Data.(eventbus.FindingDiscoveredData)
		a.findings = append(a.findings, data.Finding)
		a.severityCounts[data.Finding.Severity]++
		a.scanView.UpdateStats(a.severityCounts)

	case eventbus.EventScanCompleted:
		data := event.Data.(eventbus.ScanCompletedData)
		a.report = data.Report
		a.scanRunning = false
		a.scanComplete = true
		a.scanView.SetScanComplete(true)
		a.resultsView = views.NewResultsView(a.theme, a.report)
		a.propagateSize()
		return switchViewCmd(ViewResults)

	case eventbus.EventScanFailed:
		data := event.Data.(eventbus.ScanFailedData)
		a.scanRunning = false
		a.scanError = data.Error
	}

	a.scanView.UpdateEventLog(a.eventLog)

	return nil
}

func (a *App) updateScannerStatus(name string, status models.ScannerStatus, findingCount int) {
	for i := range a.scanners {
		if a.scanners[i].Name == name {
			a.scanners[i].Status = status
			if findingCount > 0 {
				a.scanners[i].FindingCount = findingCount
			}
			break
		}
	}
	a.scanView.UpdateScanners(a.scanners)
}

func (a *App) propagateSize() {
	// Use actual rendered heights for accurate layout
	header := a.renderHeader()
	footer := a.renderFooter()
	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
	contentH := a.height - headerH - footerH
	if contentH < 4 {
		contentH = 4
	}

	a.scanView.SetSize(a.width, contentH)
	if a.resultsView != nil {
		a.resultsView.SetSize(a.width, contentH)
	}
	a.help.Width = a.width
}

func (a *App) delegateKeyToView(msg tea.KeyMsg) tea.Cmd {
	switch a.activeView {
	case ViewScan:
		a.scanView.HandleKey(msg)
		return nil
	case ViewResults:
		if a.resultsView != nil {
			return a.resultsView.HandleKey(msg)
		}
	}
	return nil
}

func (a *App) renderHeader() string {
	data := components.HeaderData{
		ProjectName:    a.projectName,
		LaravelVersion: a.laravelVersion,
		PHPVersion:     a.phpVersion,
		PackageCount:   a.packageCount,
		ToolVersion:    version,
		ScanRunning:    a.scanRunning,
		ScanComplete:   a.scanComplete,
		ScanError:      a.scanError != nil,
	}
	if a.projectName == "" {
		data.ProjectName = a.targetPath
	}
	return components.RenderHeader(data, a.theme, a.width)
}

func (a *App) renderFooter() string {
	return components.RenderFooter(a.help, a.keys, a.theme, a.width)
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func switchViewCmd(v ViewID) tea.Cmd {
	return func() tea.Msg { return switchViewMsg{view: v} }
}
