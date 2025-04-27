package header

import (
	"otaviocosta2110/k8s-tui/src/global"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content     string
	width       int
	height      int
	headerStyle lipgloss.Style
}

func New(headerText string) Model {
	return Model{
		content: headerText,
		headerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(global.Colors.Blue)),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.headerStyle = m.headerStyle.
			Width(msg.Width - global.Margin).
			Height(global.HeaderSize - global.Margin) 
	}
	return m, nil
}

func (m Model) View() string {
	return m.headerStyle.Render(m.content)
}

func (m *Model) SetContent(content string) {
	m.content = content
}
