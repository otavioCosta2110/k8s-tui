package customstyles

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	width, _, _ = term.GetSize(int(os.Stdout.Fd()))

	AvailableWidth = width - styles.Margin
)

func NormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(TextColor)).
		Width(AvailableWidth).
		Background(lipgloss.Color(BackgroundColor))
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
		BorderForeground(lipgloss.Color(BorderColor)).
		BorderBackground(lipgloss.Color(BackgroundColor)).
		Width(AvailableWidth).
		Background(lipgloss.Color(BackgroundColor))
}

func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ErrorColor)).
		Bold(true).
		Background(lipgloss.Color(BackgroundColor))
}
