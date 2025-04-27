package header

import (
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content     string
	width       int
	height      int
	headerStyle lipgloss.Style
	kubeconfig  *kubernetes.KubeConfig
}

func New(headerText string, kubeconfig *kubernetes.KubeConfig) Model {
	return Model{
		content:    "",
		kubeconfig: kubeconfig,
		headerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Height(global.HeaderSize - global.Margin).
			BorderForeground(lipgloss.Color(global.Colors.Blue)),
	}
}

func (m Model) Init() tea.Cmd {
	metrics := kubernetes.NewMetrics(*m.kubeconfig)
	kubernetes.ViewMetrics(metrics)
	m.headerStyle = m.headerStyle.
		Height(global.HeaderSize - global.Margin)
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.headerStyle = m.headerStyle.
		Height(global.HeaderSize - global.Margin)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.headerStyle = m.headerStyle.
			Width(msg.Width - global.Margin).
			Height(global.HeaderSize - global.Margin)
	}
	return m, nil
}

func (m Model) View() string {
	if m.kubeconfig == nil {
		return "No kubeconfig selected"
	}
	return m.headerStyle.Render(m.content)
}

func (m *Model) SetContent(content string) {
	m.content = content
}

func (m *Model) SetKubeconfig(kubeconfig *kubernetes.KubeConfig) {
	m.kubeconfig = kubeconfig
}
