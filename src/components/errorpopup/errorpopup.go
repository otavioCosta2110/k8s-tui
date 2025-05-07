package errorpopup

import (
	"otaviocosta2110/k8s-tui/src/global"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	error   error
	title   string
	message string
	width   int
	height  int
}

func New(err error, title, message string) Model {
	return Model{
		error:   err,
		title:   title,
		message: message,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	errorStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height + global.Margin).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width) 

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Align(lipgloss.Center).
		Width(m.width)

	content := titleStyle.Render(m.title) + "\n\n"
	content += messageStyle.Render(m.message) + "\n\n"
	if m.error != nil {
		content += messageStyle.Render("Error details: " + m.error.Error())
	}
	content += "\n\n" + messageStyle.Render("Press ESC or Q to dismiss")

	return errorStyle.Render(content)
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}
