package banner

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The main figlet banner in ANSI Shadow style.
var lines = [7]string{
	` ██╗    ██╗ █████╗ ██████╗ ██████╗ `,
	` ██║    ██║██╔══██╗██╔══██╗██╔══██╗`,
	` ██║ █╗ ██║███████║██████╔╝██║  ██║`,
	` ██║███╗██║██╔══██║██╔══██╗██║  ██║`,
	` ╚███╔███╔╝██║  ██║██║  ██║██████╔╝`,
	`  ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ `,
	``,
}

// Gradient colors from deep red/orange to purple — Laravel-native feel.
var gradient = [6]string{
	"#FF3333", // bright red
	"#FF2D55", // red-pink
	"#E93578", // pink
	"#C83DA0", // magenta
	"#A344C8", // purple
	"#7C4DFF", // deep purple
}

// Render returns the full colored banner with tagline and version.
func Render(version string) string {
	var b strings.Builder

	for i, line := range lines[:6] {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(gradient[i]))
		b.WriteString(style.Render(line))
		b.WriteByte('\n')
	}

	tagline := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"}).
		Italic(true).
		Render(fmt.Sprintf("  Laravel Security Scanner v%s", version))

	b.WriteString(tagline)
	b.WriteByte('\n')

	return b.String()
}

// RenderCompact returns a smaller single-line stylized logo for tight spaces.
func RenderCompact() string {
	w := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2D55")).Bold(true)
	a := lipgloss.NewStyle().Foreground(lipgloss.Color("#E93578")).Bold(true)
	r := lipgloss.NewStyle().Foreground(lipgloss.Color("#A344C8")).Bold(true)
	d := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C4DFF")).Bold(true)

	return w.Render("W") + a.Render("A") + r.Render("R") + d.Render("D")
}

// RenderWithBox returns the banner inside a rounded border box.
func RenderWithBox(version string) string {
	inner := Render(version)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#7C4DFF"}).
		Padding(0, 2)

	return box.Render(inner)
}

// ShieldIcon returns a small shield character for inline use.
func ShieldIcon() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C4DFF")).
		Bold(true).
		Render("🛡")
}
