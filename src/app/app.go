package app

import (
	"otaviocosta2110/k8s-tui/src/components/header"
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	stack  []tea.Model
	kube   kubernetes.KubeConfig
	header header.Model
}

func NewAppModel(k kubernetes.KubeConfig) *AppModel {
	// Set header height before creating components
	
	initialScreen := k.InitComponent(k)
	return &AppModel{
		stack:  []tea.Model{initialScreen},
		kube:   k,
		header: header.New("K8s TUI"),
	}
}

func (m *AppModel) Init() tea.Cmd {
	if len(m.stack) == 0 {
		return nil
	}
	return tea.Batch(
		m.stack[len(m.stack)-1].Init(),
		m.header.Init(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		global.ScreenWidth = msg.Width - global.Margin
		global.ScreenHeight = msg.Height - global.Margin

		global.HeaderSize = global.ScreenHeight/3 - global.Margin

		newHeader, headerCmd := m.header.Update(msg)
		m.header = newHeader.(header.Model)

		var cmds []tea.Cmd
		for i := range m.stack {
			var cmd tea.Cmd
			m.stack[i], cmd = m.stack[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(append(cmds, headerCmd)...)

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
		return m, msg.NewScreen.Init()
	}

	var cmd tea.Cmd
	current := len(m.stack) - 1
	m.stack[current], cmd = m.stack[current].Update(msg)
	return m, cmd
}

func (m *AppModel) View() string {
	if len(m.stack) == 0 {
		return ""
	}

	header := m.header.View()

	bottomPanelStyle := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(global.ScreenHeight - global.HeaderSize + global.Margin).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Blue))

	currentView := bottomPanelStyle.Render(m.stack[len(m.stack)-1].View())
	return lipgloss.JoinVertical(lipgloss.Top, header, currentView)
}
