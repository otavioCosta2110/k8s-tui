package listcomponent

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

// loading works but is not implemented anywhere
type Model struct {
	List       list.Model
	OnSelected []func(selected string) tea.Model
	loading    bool
}

func (m *Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return loadedMsg{}
	})
}

type loadedMsg struct{}

func NewItem(title, description string) Item {
	item := Item{title: title, description: description}
	return item
}

// MUST HAVE THIS
func (i Item) FilterValue() string { return i.title }

func (i Item) Title() string {
	return i.title
}

func (i Item) Description() string {
	return ""
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		return m, nil
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width/2, msg.Height-global.Margin)
	case tea.KeyMsg:
		if msg.String() == "enter" && !m.loading {
			for _, fn := range m.OnSelected {
				selectedItem := m.List.SelectedItem().(Item).FilterValue()
				newModel := fn(selectedItem)
				if newModel != nil {
					return newModel, nil
				}
			}
			return m, nil
		}
	}
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.loading {
		loadingStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(global.Colors.Pink))
		return loadingStyle.Render("Loading...")
	}

	m.List.SetSize(global.ScreenWidth/2, global.ScreenHeight-global.Margin)
	return m.List.View()
}

// onSelect should return the next resource component
func NewList(items []string, title string, onSelect ...func(selected string) tea.Model) tea.Model {
	var listItems []list.Item
	for _, configs := range items {
		k := NewItem(configs, "")
		listItems = append(listItems, k)
	}

	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(global.Colors.Pink)).
		SetString(">").
		Bold(true).
		Padding(0, 0)

	delegate.ShowDescription = false

	l := list.New(listItems, delegate, global.ScreenWidth, global.ScreenHeight)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().Bold(true)

	return &Model{List: l, OnSelected: onSelect}
}
