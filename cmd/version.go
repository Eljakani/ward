package cmd

import (
	"fmt"

	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/spf13/cobra"
)

const Version = "0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Render(Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
