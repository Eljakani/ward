package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags:
//
//	go build -ldflags "-X github.com/eljakani/ward/cmd.Version=v0.3.0 \
//	  -X github.com/eljakani/ward/cmd.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/eljakani/ward/cmd.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func init() {
	// When installed via `go install @v0.3.0`, ldflags aren't set, but Go
	// embeds the module version in the binary. Read it as a fallback.
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	if Commit == "none" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" && len(s.Value) >= 7 {
					Commit = s.Value[:7]
					break
				}
			}
		}
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Render(Version))

		dim := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"})
		accent := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#B388FF"}).
			Bold(true)

		fmt.Println(accent.Render("  Version: ") + dim.Render(Version))
		fmt.Println(accent.Render("  Commit:  ") + dim.Render(Commit))
		fmt.Println(accent.Render("  Built:   ") + dim.Render(Date))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
