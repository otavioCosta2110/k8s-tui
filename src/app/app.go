package app

import (
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	stack  []tea.Model
	kube   kubernetes.KubeConfig
}

func NewAppModel(k kubernetes.KubeConfig) *AppModel {
	initialScreen := k.InitComponent(k)
	return &AppModel{
		stack: []tea.Model{initialScreen},
		kube:  k,
	}
}

func (m *AppModel) Init() tea.Cmd {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1].Init()
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		global.ScreenWidth = msg.Width - global.Margin
		global.ScreenHeight = msg.Height - global.Margin/2
		for i := range m.stack {
			var cmd tea.Cmd
			m.stack[i], cmd = m.stack[i].Update(msg)
			if cmd != nil {
				return m, cmd
			}
		}
		return m, nil
		
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

	leftPanelStyle := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(global.ScreenHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Blue))

	currentView := leftPanelStyle.Render(m.stack[len(m.stack)-1].View())

	return currentView
}
