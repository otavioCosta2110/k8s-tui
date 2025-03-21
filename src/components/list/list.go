package listcomponent

import (
	"otaviocosta2110/k8s-tui/src/global"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	title       string
	description string
}

type Model struct {
	List list.Model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func NewItem(title, description string) Item {
	item := Item{title: title, description: description}
	return item
}

// TEM Q TER ISSO
func (i Item) FilterValue() string { return i.title }

func (i Item) Title() string {
	return i.title
}

func (i Item) Description() string {
	return ""
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.List.View()
}

func NewList(items []list.Item, title string, width, height int) *Model {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(global.Colors.Pink)).
		Padding(0, 3) 

	delegate.ShowDescription = false

	l := list.New(items, delegate, width, height)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().Bold(true)

	return &Model{List: l}
}
