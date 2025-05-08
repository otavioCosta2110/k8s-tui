package list

import (
	"otaviocosta2110/k8s-tui/src/global"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	title       string
	description string
}

type Model struct {
	List        list.Model
	OnSelected  func(selected string) tea.Msg
	loading     bool
	initialized bool
}

type loadedMsg struct{}

func NewItem(title, description string) Item {
	return Item{title: title, description: description}
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return "" }
func (i Item) FilterValue() string { return i.title }

func NewList(items []string, title string, onSelect func(selected string) tea.Msg) *Model {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, NewItem(item, ""))
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Height(2).
		SetString(" ").
		Padding(0, 3)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Margin(0, 1).
		Height(2).
		Foreground(lipgloss.Color(global.Colors.Pink)).
		SetString("➤ ").
		Bold(true).
		Padding(0, 1)

	delegate.SetSpacing(0)
	delegate.ShowDescription = false
	delegate.SetHeight(1)

	l := list.New(listItems, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().Bold(true)

	l.Styles.PaginationStyle = lipgloss.NewStyle().
		MarginTop(1)
	l.Styles.HelpStyle = lipgloss.NewStyle().
		MarginTop(1)

	return &Model{
		List:       l,
		OnSelected: onSelect,
		loading:    false,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return loadedMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.List.SetSize(global.ScreenWidth - global.Margin, global.ScreenHeight - len(m.List.Items()) - global.Margin * 5)
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		return m, nil
	case tea.WindowSizeMsg:
		m.List.SetSize(global.ScreenWidth - global.Margin, global.ScreenHeight - global.Margin)
	case tea.KeyMsg:
		if msg.String() == "enter" && !m.loading && m.OnSelected != nil {
			selected := m.List.SelectedItem().(Item).FilterValue()
			return m, func() tea.Msg {
				return m.OnSelected(selected)
			}
		}
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	// if m.loading {
	// 	return lipgloss.NewStyle().
	// 		Align(lipgloss.Center, lipgloss.Center).
	// 		Foreground(lipgloss.Color(global.Colors.Pink)).
	// 		Width(global.ScreenWidth - global.Margin).
	// 		Height(global.ScreenHeight - global.Margin).
	// 		Render("Loading...")
	// }
	// m.List.SetSize(global.ScreenWidth - global.Margin, global.ScreenHeight - len(m.List.Items()))
	if m.loading {
		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading...")
	}
	m.List.SetSize(global.ScreenWidth - global.Margin, global.ScreenHeight - len(m.List.Items()) - global.Margin * 2)
	return m.List.View()
}
