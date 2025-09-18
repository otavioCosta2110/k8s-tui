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
)

func NormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1)
}

func SelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(SelectionForeground)).
		Background(lipgloss.Color(SelectionBackground)).
		Width(AvailableWidth).
		Padding(0, 1).
		Bold(true)
}

func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color(HeaderColor))
}

func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ErrorColor)).
		Bold(true)
}
