package models

import (
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeaderModel struct {
	content     string
	width       int
	height      int
	headerStyle lipgloss.Style
	kubeconfig  *k8s.Client
}

func NewHeader(headerText string, kubeconfig *k8s.Client) HeaderModel {
	return HeaderModel{
		content:    "",
		kubeconfig: kubeconfig,
		headerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Height(global.HeaderSize - global.Margin*2).
			BorderForeground(lipgloss.Color(global.Colors.Blue)),
	}
}

func (m HeaderModel) Init() tea.Cmd {
	metrics := NewMetrics(*m.kubeconfig)
	if metrics.Error != nil {
		return nil
	}
	metrics.viewMetrics()
	m.headerStyle = m.headerStyle.
		Height(global.HeaderSize - global.Margin*2)
	return nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.headerStyle = m.headerStyle.
		Height(global.HeaderSize - global.Margin*2)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.headerStyle = m.headerStyle.
			Width(msg.Width - global.Margin).
			Height(global.HeaderSize - global.Margin*2)
	}
	return m, nil
}

func (m HeaderModel) View() string {
	if m.kubeconfig == nil {
		return "No kubeconfig selected"
	}
	return m.headerStyle.Render(m.content)
}

func (m *HeaderModel) SetContent(content string) {
	m.content = content
}

func (m *HeaderModel) IsContentNil() bool {
	return m.content == ""
}

func (m *HeaderModel) SetKubeconfig(kubeconfig *k8s.Client) {
	m.kubeconfig = kubeconfig
}
