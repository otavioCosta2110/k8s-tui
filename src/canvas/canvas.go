package canvas

import (
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Canvas struct {
	Width  int
	Height int
	Input  string
	List   listcomponent.Model
}

func (c *Canvas) InitList() {
	items := []list.Item{
		listcomponent.NewItem("Item 1", "This is item 1"),
		listcomponent.NewItem("Item 2", "This is item 2"),
		listcomponent.NewItem("Item 3", "This is item 3"),
	}
	c.List = *listcomponent.NewList(items, "fodase", c.Width, c.Height)
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
		c.List.List.SetSize(msg.Width / 2, msg.Height - 2 )
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
		BorderForeground(lipgloss.Color("62")). 
		Width(c.Width/2 - 2). 
		Height(c.Height - 2) 

	listView := listBoxStyle.Render(c.List.View())

	rightPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Height(c.Height - 2). 
		Width(c.Width/2 - 2).
		Align(lipgloss.Center)

	rightPanel := rightPanelStyle.Render("Hello!\n(press 'q' to quit)")

	return lipgloss.JoinHorizontal(lipgloss.Left, listView, rightPanel)
}
