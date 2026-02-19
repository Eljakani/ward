package cmd

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/eljakani/ward/internal/updater"
	"github.com/spf13/cobra"
)

var (
	verbose   bool
	noColor   bool
	outputFmt string
)

// updateNotice is populated asynchronously by PersistentPreRun.
var (
	updateNotice string
	updateOnce   sync.Once
	updateDone   = make(chan struct{})
)

var rootCmd = &cobra.Command{
	Use:   "ward",
	Short: "Laravel security scanner",
	Long:  banner.Render(Version),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Launch async update check — never blocks
		go func() {
			defer close(updateDone)
			wardDir, err := config.Dir()
			if err != nil {
				return
			}
			updateNotice = updater.CheckForUpdate(Version, wardDir)
		}()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Wait for the check to finish (≤2s due to HTTP timeout)
		<-updateDone
		if updateNotice != "" {
			updateStyle := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#E65100", Dark: "#FFB74D"}).
				Bold(true)
			borderStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#E65100", Dark: "#FFB74D"}).
				Padding(0, 1)
			fmt.Println()
			fmt.Println(borderStyle.Render(updateStyle.Render("⬆ Update Available") + "\n" + updateNotice))
			fmt.Println()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.RenderWithBox(Version))
		fmt.Println()

		dim := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"})
		accent := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#B388FF"}).
			Bold(true)

		fmt.Println(dim.Render("  Usage:"))
		fmt.Println(accent.Render("    ward init") + dim.Render("           Initialize ~/.ward config"))
		fmt.Println(accent.Render("    ward scan <path>") + dim.Render("    Scan a Laravel project"))
		fmt.Println(accent.Render("    ward version") + dim.Render("        Print version info"))
		fmt.Println(accent.Render("    ward --help") + dim.Render("         Show all commands"))
		fmt.Println()
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "tui", "output mode: tui (interactive), or comma-separated formats (json,sarif,html,markdown)")
}
