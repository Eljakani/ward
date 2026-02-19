package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/eventbus"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/orchestrator"
	"github.com/eljakani/ward/internal/tui"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a Laravel project for security issues",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// If --output specifies formats (not "tui"), override config and run headless
		if outputFmt != "tui" {
			cfg.Output.Formats = parseOutputFormats(outputFmt)
			return runHeadless(cfg, targetPath)
		}

		// If no TTY available, fall back to headless
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return runHeadless(cfg, targetPath)
		}

		return runWithTUI(cfg, targetPath)
	},
}

func runWithTUI(cfg *config.WardConfig, targetPath string) error {
	bus := eventbus.New()
	model := tui.NewApp(bus, targetPath, Version)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	bridge := eventbus.NewBridge(bus, p)
	bridge.Start()
	defer bridge.Stop()

	go func() {
		orch := orchestrator.New(bus, cfg, targetPath, Version)
		if err := orch.Run(context.Background()); err != nil {
			bus.Publish(eventbus.NewEvent(eventbus.EventScanFailed, eventbus.ScanFailedData{
				Error: err,
			}))
		}
	}()

	_, err := p.Run()
	return err
}

func runHeadless(cfg *config.WardConfig, targetPath string) error {
	fmt.Println(banner.Render(Version))

	bus := eventbus.New()

	dim := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"})
	accent := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#B388FF"}).Bold(true)
	sevStyles := map[models.Severity]*lipgloss.Style{
		models.SeverityCritical: ptr(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5252")).Bold(true)),
		models.SeverityHigh:     ptr(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB74D")).Bold(true)),
		models.SeverityMedium:   ptr(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD54F"))),
		models.SeverityLow:      ptr(lipgloss.NewStyle().Foreground(lipgloss.Color("#81C784"))),
		models.SeverityInfo:     ptr(lipgloss.NewStyle().Foreground(lipgloss.Color("#64B5F6"))),
	}

	// Print events as they happen
	bus.Subscribe(eventbus.EventStageStarted, func(e eventbus.Event) {
		data := e.Data.(eventbus.StageStartedData)
		fmt.Printf("  %s %s\n", accent.Render("●"), data.Stage)
	})

	bus.Subscribe(eventbus.EventFindingDiscovered, func(e eventbus.Event) {
		data := e.Data.(eventbus.FindingDiscoveredData)
		f := data.Finding
		style := sevStyles[f.Severity]
		fmt.Printf("    %s %s\n", style.Render(fmt.Sprintf("[%s]", f.Severity)), f.Title)
	})

	bus.Subscribe(eventbus.EventScannerCompleted, func(e eventbus.Event) {
		data := e.Data.(eventbus.ScannerCompletedData)
		fmt.Printf("  %s %s — %d findings\n", dim.Render("✓"), data.Name, data.FindingCount)
	})

	bus.Subscribe(eventbus.EventLogMessage, func(e eventbus.Event) {
		data := e.Data.(eventbus.LogMessageData)
		fmt.Printf("  %s %s\n", dim.Render("["+data.Level+"]"), data.Message)
	})

	bus.Subscribe(eventbus.EventScanCompleted, func(e eventbus.Event) {
		data := e.Data.(eventbus.ScanCompletedData)
		r := data.Report
		counts := r.CountBySeverity()
		fmt.Println()
		fmt.Printf("  %s %d findings in %s\n", accent.Render("Done."), len(r.Findings), r.Duration.Round(1e6))
		for _, sev := range []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityInfo} {
			if c := counts[sev]; c > 0 {
				style := sevStyles[sev]
				fmt.Printf("    %s %d\n", style.Render(fmt.Sprintf("%-10s", sev)), c)
			}
		}
		fmt.Println()
	})

	orch := orchestrator.New(bus, cfg, targetPath, Version)
	return orch.Run(context.Background())
}

func ptr(s lipgloss.Style) *lipgloss.Style { return &s }

// parseOutputFormats splits a comma-separated format string into a list.
// e.g. "json" → ["json"], "json,sarif" → ["json", "sarif"]
func parseOutputFormats(s string) []string {
	var formats []string
	for _, f := range strings.Split(s, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			formats = append(formats, f)
		}
	}
	return formats
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
