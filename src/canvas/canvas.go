package canvas

import (
	"otaviocosta2110/k8s-tui/src/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
)

type Canvas struct {
	Width  int
	Height int
	Input  string
	List   list.Model 
}

func (c *Canvas)InitList(){
  c.List = list.New([]list.Item{}, list.NewDefaultDelegate(), 0,0)
  c.List.Title = "fodase"
  c.List.SetItems([]list.Item{
		components.NewItem("Item 1",  "This is item 1" ),
		components.NewItem("Item 2",  "This is item 2" ),
		components.NewItem("Item 3",  "This is item 3" ),
  }) 
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
		c.Width = 20
		c.Height = 10
    c.List.SetSize(20, 10)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return c, tea.Quit
		}
	}
  var cmd tea.Cmd
  c.List, cmd = c.List.Update(msg)
	return c, cmd
}

func (c *Canvas) View() string {
	return c.List.View()
}
