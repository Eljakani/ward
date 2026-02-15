package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/eventbus"
	"github.com/eljakani/ward/internal/tui"
	"github.com/spf13/cobra"
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

		_ = cfg // will be passed to orchestrator when implemented

		bus := eventbus.New()
		model := tui.NewApp(bus, targetPath)

		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

		bridge := eventbus.NewBridge(bus, p)
		bridge.Start()
		defer bridge.Stop()

		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
