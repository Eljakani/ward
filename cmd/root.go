package cmd

import (
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
	Long:  "Laravel Ward — A comprehensive security scanner for Laravel applications.",
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
