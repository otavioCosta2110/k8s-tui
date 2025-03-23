package canvas

import (
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Canvas struct {
	Width     int
	Height    int
	Input     string
	Component tea.Model
}

func NewCanvas(component kubernetes.ResourceInterface) *Canvas {
	newComponent := component.InitComponent()
	c := &Canvas{Component: newComponent}
	return c
}

func (c Canvas) Init() tea.Cmd {
	return nil
}

func (c *Canvas) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Width = msg.Width
		global.ScreenWidth = c.Width
		c.Height = msg.Height
		global.ScreenHeight = c.Height
		c.Component.Update(msg)
	case tea.KeyMsg:
		c.isKeyPressed(msg)
	}
	var cmd tea.Cmd
	c.Component, cmd = c.Component.Update(msg)
	return c, cmd
}

func (c *Canvas) View() string {
	listBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Blue)).
		Width(c.Width/2 - global.Margin).
		Height(c.Height - global.Margin)

	listView := listBoxStyle.Render(c.Component.View())

	rightPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Height(c.Height - global.Margin).
		Width(c.Width/2 - global.Margin).
		Align(lipgloss.Center)

	rightPanel := rightPanelStyle.Render("Hello!\n(press 'q' to quit)")

	return lipgloss.JoinHorizontal(lipgloss.Left, listView, rightPanel)
}
