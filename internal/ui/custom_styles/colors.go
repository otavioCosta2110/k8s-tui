package customstyles

import (
	"otaviocosta2110/k8s-tui/internal/config"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	BorderColor string

	AccentColor string

	HeaderColor string

	ErrorColor string

	SelectionBackground string

	SelectionForeground string

	TextColor lipgloss.Color

	BackgroundColor string
)

func InitColors() error {
	scheme, err := config.LoadColorScheme()
	if err != nil {
		return err
	}

	BorderColor = scheme.BorderColor
	AccentColor = scheme.AccentColor
	HeaderColor = scheme.HeaderColor
	ErrorColor = scheme.ErrorColor
	SelectionBackground = scheme.SelectionBackground
	SelectionForeground = scheme.SelectionForeground

	if scheme.TextColor != "" {
		TextColor = lipgloss.Color(scheme.TextColor)
	} else {
		TextColor = lipgloss.Color(termenv.ForegroundColor().Sequence(true))
	}

	if scheme.BackgroundColor != "" {
		BackgroundColor = scheme.BackgroundColor
	} else {
		BackgroundColor = "#000000"
	}

	return nil
}
