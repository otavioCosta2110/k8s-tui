package components

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListItem struct {
	title       string
	description string
}

type ListModel struct {
	List        list.Model
	OnSelected  func(selected string) tea.Msg
	loading     bool
	initialized bool
	footerText  string
}

func NewItem(title, description string) ListItem {
	return ListItem{title: title, description: description}
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return "" }
func (i ListItem) FilterValue() string { return i.title }

func NewList(items []string, title string, onSelect func(selected string) tea.Msg) *ListModel {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, NewItem(item, ""))
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = customstyles.NormalStyle()

	delegate.Styles.SelectedTitle = customstyles.SelectedStyle()

	delegate.SetSpacing(0)
	delegate.ShowDescription = false

	l := list.New(listItems, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = customstyles.TitleStyle()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowTitle(false)

	l.SetShowHelp(false)

	return &ListModel{
		List:       l,
		OnSelected: onSelect,
		loading:    false,
	}
}

func NewListWithItems(items []ListItem, title string, onSelect func(selected string) tea.Msg) *ListModel {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, item)
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = customstyles.NormalStyle()

	delegate.Styles.SelectedTitle = customstyles.SelectedStyle()

	delegate.SetSpacing(0)
	delegate.ShowDescription = false

	l := list.New(listItems, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = customstyles.TitleStyle()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowTitle(false)

	l.SetShowHelp(false)

	return &ListModel{
		List:       l,
		OnSelected: onSelect,
		loading:    false,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return loadedMsg{}
	})
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		return m, nil
	case tea.WindowSizeMsg:
		m.List.SetSize(styles.ScreenWidth, styles.ScreenHeight)
	case tea.KeyMsg:
		if msg.String() == "enter" && !m.loading && m.OnSelected != nil {
			selected := m.List.SelectedItem().(ListItem).FilterValue()
			return m, func() tea.Msg {
				return m.OnSelected(selected)
			}
		}
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *ListModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Loading...")
	}
	m.List.SetSize(styles.ScreenWidth, styles.ScreenHeight-1)

	originalTitle := m.List.Title
	m.List.Title = ""
	listView := m.List.View()
	m.List.Title = originalTitle

	var view string
	if originalTitle != "" {
		titleView := customstyles.TitleStyle().Render(originalTitle)
		view = lipgloss.JoinVertical(lipgloss.Left, titleView, listView)
	} else {
		view = listView
	}

	if m.footerText != "" {
		footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			PaddingTop(1).
			Background(lipgloss.Color(customstyles.BackgroundColor))

		view = lipgloss.JoinVertical(lipgloss.Left, view, footerStyle.Render(m.footerText))
	}

	return view
}

func (m *ListModel) SetFooterText(text string) {
	m.footerText = text
}
