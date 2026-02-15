package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the ~/.ward config directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(banner.Render(Version))

		dir, err := config.Init(initForce)
		if err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}

		success := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#69F0AE"}).
			Bold(true)
		dim := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"})
		accent := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#B388FF"})

		fmt.Println(success.Render("  Initialized Ward configuration."))
		fmt.Println()
		fmt.Println(dim.Render("  Created:"))
		fmt.Println(accent.Render(fmt.Sprintf("    %s/config.yaml", dir)) + dim.Render("       main config"))
		fmt.Println(accent.Render(fmt.Sprintf("    %s/rules/", dir)) + dim.Render("            custom rules"))
		fmt.Println(accent.Render(fmt.Sprintf("    %s/reports/", dir)) + dim.Render("          scan reports"))
		fmt.Println(accent.Render(fmt.Sprintf("    %s/store/", dir)) + dim.Render("            result store"))
		fmt.Println()
		fmt.Println(dim.Render("  Edit config.yaml to customize scan behaviour."))
		fmt.Println(dim.Render("  Drop .yaml rule files into rules/ to add custom rules."))
		fmt.Println()

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config files")
	rootCmd.AddCommand(initCmd)
}
