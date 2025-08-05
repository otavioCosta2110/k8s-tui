package customstyles

import (
	"os"
	global "otaviocosta2110/k8s-tui/internal"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	width, _, _ = term.GetSize(int(os.Stdout.Fd()))

	AvailableWidth = width - global.Margin

	NormalStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Black)).
			Background(lipgloss.Color(Blue)).
			Width(AvailableWidth).
			Padding(0, 1).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(Foreground)).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.Color(Purple))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Red)).
			Bold(true)
)
