package app

import (
	"otaviocosta2110/k8s-tui/src/components/header"
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	stack          []tea.Model
	kube           kubernetes.KubeConfig
	header         header.Model
	configSelected bool
}

func NewAppModel() *AppModel {
	return &AppModel{
		stack:  []tea.Model{kubernetes.NewKubeConfig().InitComponent(nil)},
		header: header.New("K8s TUI", nil),
	}
}

func (m *AppModel) Init() tea.Cmd {
	if len(m.stack) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	if m.configSelected {
		cmds = append(cmds, m.header.Init())
	}
	cmds = append(cmds, m.stack[len(m.stack)-1].Init())

	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		global.ScreenWidth = msg.Width - global.Margin
		global.ScreenHeight = msg.Height - global.Margin *2

		global.HeaderSize = global.ScreenHeight/3 - global.Margin * 4

		var cmds []tea.Cmd

		if m.configSelected {
			newHeader, headerCmd := m.header.Update(msg)
			m.header = newHeader.(header.Model)
			m.header.SetKubeconfig(&m.kube)
			m.header.SetContent(kubernetes.ViewMetrics(kubernetes.NewMetrics(m.kube)))
			cmds = append(cmds, headerCmd)
		}

		for i := range m.stack {
			var cmd tea.Cmd
			m.stack[i], cmd = m.stack[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if len(m.stack) > 1 {
				m.stack = m.stack[:len(m.stack)-1]
				return m, nil
			}
			return m, tea.Quit
		}

	case kubernetes.NavigateMsg:
		m.stack = append(m.stack, msg.NewScreen)
		if !m.configSelected {
			m.configSelected = true
			m.header.SetKubeconfig(&msg.Cluster)
			m.kube = msg.Cluster
			m.header.SetContent(kubernetes.ViewMetrics(kubernetes.NewMetrics(m.kube)))
			return m, tea.Batch(
				msg.NewScreen.Init(),
				m.header.Init(),
			)
		}
		return m, msg.NewScreen.Init()
	}

	var cmd tea.Cmd
	current := len(m.stack) - 1
	m.stack[current], cmd = m.stack[current].Update(msg)
	return m, cmd
}

func (m *AppModel) View() string {
	if len(m.stack) == 0 {
		return "Loading..."
	}

	currentView := m.stack[len(m.stack)-1].View()

	headerView := m.header.View()
	contentHeight := global.ScreenHeight - lipgloss.Height(headerView) + global.Margin

	if !m.configSelected {
		if len(m.stack) > 0 {
			return lipgloss.NewStyle().
				Width(global.ScreenWidth).
				Height(contentHeight + 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(global.Colors.Blue)).
				Render(currentView)
		}
		return "Loading..."
	}
	header := m.header.View()

	content := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(contentHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Blue)).
		Render(currentView)

	return lipgloss.JoinVertical(lipgloss.Top, header, content)
}
