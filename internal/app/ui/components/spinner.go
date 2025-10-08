package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type SpinnerModel struct {
	spinner spinner.Model
	text    string
}

func NewSpinner(text string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor))

	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
	textWithStyle := textStyle.Render(text)

	return SpinnerModel{
		spinner: s,
		text:    textWithStyle,
	}
}

func (s SpinnerModel) Init() tea.Cmd {
	return s.spinner.Tick
}

func (s SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s SpinnerModel) View() string {
	whiteSpace:= lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ")
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Render(s.spinner.View() + whiteSpace + s.text)
}
