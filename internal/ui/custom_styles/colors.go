package customstyles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	Blue = "#00b8ff"
  Pink = "#f29bdc"
  Purple = "#7D56F4"
  Red = "#FF0000"
  Black = "#000000"
	Foreground = lipgloss.Color(termenv.ForegroundColor().Sequence(true))
)

