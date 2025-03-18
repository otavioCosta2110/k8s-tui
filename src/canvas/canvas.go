package canvas

import (
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Canvas struct {
	Width  int
	Height int
}

func (c Canvas) Init() tea.Cmd {
	return nil
}

func (c Canvas) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.Width = msg.Width
		c.Height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "q" {
			return c, tea.Quit
		}
		if msg.String() == "a" {
			return c, nil
		}
	}
	return c, nil
}

func (c Canvas) View() string {
	var boxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("212")).
		Width(c.Width-2).
		Height(c.Height/4-2).
		Align(lipgloss.Center, lipgloss.Center)

	var otherBox = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("212")).
		Width(c.Width-2).
		Height(c.Height-c.Height/4-2).
		Align(lipgloss.Center, lipgloss.Center)

	box := boxStyle.Render("Hello!")
	box2 := otherBox.Render("haha")

	boxes := lipgloss.JoinVertical(lipgloss.Center, box, box2)

	finalView := lipgloss.Place(c.Width, c.Height, lipgloss.Center, lipgloss.Center, boxes)

	return finalView
}
