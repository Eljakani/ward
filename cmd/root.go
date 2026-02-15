package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/spf13/cobra"
)

var (
	verbose   bool
	noColor   bool
	outputFmt string
)

var rootCmd = &cobra.Command{
	Use:   "ward",
	Short: "Laravel security scanner",
	Long:  banner.Render(Version),
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
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "tui", "output format: tui, json, text")
}
