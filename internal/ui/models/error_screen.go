package models

import (
	global "github.com/otavioCosta2110/k8s-tui/internal"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ErrorModel struct {
	error   error
	title   string
	message string
	width   int
	height  int
}

func NewErrorScreen(err error, title, message string) ErrorModel {
	return ErrorModel{
		error:   err,
		title:   title,
		message: message,
	}
}

func (m ErrorModel) Init() tea.Cmd {
	return nil
}

func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	return m, nil
}

func (m ErrorModel) View() string {
	errorStyle := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(global.ScreenHeight+global.HeaderSize).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000")).
		Padding(1, 2).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	content := titleStyle.Render(m.title) + "\n\n"
	content += messageStyle.Render(m.message) + "\n\n"
	if m.error != nil {
		content += messageStyle.Render("Error details: " + m.error.Error())
	}
	content += "\n\n" + messageStyle.Render("Press ESC or Q to dismiss")

	return errorStyle.Render(content)
}

func (m *ErrorModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}
