package canvas

import (
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"
	"otaviocosta2110/k8s-tui/src/global"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding = 2
)

type Canvas struct {
	Width  int
	Height int
	Input  string
	List   tea.Model
}

func (c *Canvas) InitList() {
	var listItems []list.Item
	for _, configs := range kubernetes.InitList() {
		k := listcomponent.NewItem(configs, "")
		listItems = append(listItems, k)
	}

	onSelect := func() {
    kubernetes.FetchNamespaces()
	}

	c.List = listcomponent.NewList(listItems, "Kubeconfigs", c.Width, c.Height, onSelect)
}

func NewCanvas() *Canvas {
	c := &Canvas{}
	c.InitList()
	return c
}

func (c Canvas) Init() tea.Cmd {
	return nil
}

func (c *Canvas) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Width = msg.Width
		c.Height = msg.Height
		c.List.Update(msg)
	case tea.KeyMsg:
		c.isKeyPressed(msg)
	}
	var cmd tea.Cmd
	c.List, cmd = c.List.Update(msg)
	return c, cmd
}

func (c *Canvas) View() string {
	listBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Blue)).
		Width(c.Width/2 - global.Margin).
		Height(c.Height - global.Margin)

	listView := listBoxStyle.Render(c.List.View())

	rightPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Height(c.Height - global.Margin).
		Width(c.Width/2 - global.Margin).
		Align(lipgloss.Center)

	rightPanel := rightPanelStyle.Render("Hello!\n(press 'q' to quit)")

	return lipgloss.JoinHorizontal(lipgloss.Left, listView, rightPanel)
}
