package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	title       string
	description string
}

func (i Item) FilterValue() string { return i.title }

type Model struct {
	List list.Model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// TEM Q TER ISSO
func NewItem(title, description string) Item{
  item := Item{title:title, description:description}
  return item

}
func (i Item) Title() string {
    return i.title
}

func (i Item) Description() string {
    return i.description
}

func (m *Model)Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model)View() string {
	return m.List.View()
}

func NewList(items []list.Item, title string) *Model {
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 20, 10)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	return &Model{List: l}
}
