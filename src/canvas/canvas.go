package canvas

import (
	"log"
	"os"
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"
	"otaviocosta2110/k8s-tui/src/global"

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
	List   listcomponent.Model
}

func (c *Canvas) InitList() {
	var listItems []list.Item
	for _, configs := range global.GetKubeconfigsLocations() {
		kubeconfigs, err := os.ReadDir(configs)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range kubeconfigs {
			if !file.IsDir() {
				k := listcomponent.NewItem(file.Name(), "")
				listItems = append(listItems, k)
			}
		}
	}

	c.List = *listcomponent.NewList(listItems, "Kubeconfigs", c.Width, c.Height)
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
		c.List.List.SetSize(msg.Width/2, msg.Height-global.Margin)
	case tea.KeyMsg:
		c.isKeyPressed(msg)
	}
	var cmd tea.Cmd
	c.List.List, cmd = c.List.List.Update(msg)
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
