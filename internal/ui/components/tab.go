package components

import (
	"fmt"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	ID           string
	Title        string
	ResourceType string
	IsActive     bool
	IsModified   bool
	Breadcrumb   []string
	CurrentIndex int
}

type TabMsg struct {
	TabID        string
	Action       string
	ResourceType string
}

type TabComponent struct {
	Tabs        []Tab
	ActiveIndex int
	Width       int
	Height      int
	TabManager  interface{}
}

func NewTabComponent() *TabComponent {
	return &TabComponent{
		Tabs:        []Tab{},
		ActiveIndex: 0,
		Width:       0,
		Height:      3,
	}
}

func (t *TabComponent) Init() tea.Cmd {
	return nil
}

func (t *TabComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.Width = msg.Width
		return t, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if len(t.Tabs) > 0 {
				oldIndex := t.ActiveIndex
				t.ActiveIndex = (t.ActiveIndex - 1 + len(t.Tabs)) % len(t.Tabs)
				if oldIndex != t.ActiveIndex && t.ActiveIndex >= 0 && t.ActiveIndex < len(t.Tabs) {
					switchMsg := TabMsg{
						TabID:        t.Tabs[t.ActiveIndex].ID,
						Action:       "switch",
						ResourceType: t.Tabs[t.ActiveIndex].ResourceType,
					}
					return t, func() tea.Msg { return switchMsg }
				}
				return t, nil
			}
		case "right", "l":
			if len(t.Tabs) > 0 {
				oldIndex := t.ActiveIndex
				t.ActiveIndex = (t.ActiveIndex + 1) % len(t.Tabs)
				if oldIndex != t.ActiveIndex && t.ActiveIndex >= 0 && t.ActiveIndex < len(t.Tabs) {
					switchMsg := TabMsg{
						TabID:        t.Tabs[t.ActiveIndex].ID,
						Action:       "switch",
						ResourceType: t.Tabs[t.ActiveIndex].ResourceType,
					}
					return t, func() tea.Msg { return switchMsg }
				}
				return t, nil
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			index := int(msg.String()[0] - '1')
			if index >= 0 && index < len(t.Tabs) {
				oldIndex := t.ActiveIndex
				t.ActiveIndex = index
				if oldIndex != t.ActiveIndex {
					switchMsg := TabMsg{
						TabID:        t.Tabs[t.ActiveIndex].ID,
						Action:       "switch",
						ResourceType: t.Tabs[t.ActiveIndex].ResourceType,
					}
					return t, func() tea.Msg { return switchMsg }
				}
				return t, nil
			}
		case "x", "ctrl+w":
			if len(t.Tabs) > 1 && t.ActiveIndex >= 0 && t.ActiveIndex < len(t.Tabs) {
				closeMsg := TabMsg{
					TabID:        t.Tabs[t.ActiveIndex].ID,
					Action:       "close",
					ResourceType: t.Tabs[t.ActiveIndex].ResourceType,
				}
				return t, func() tea.Msg { return closeMsg }
			}
		}
	}
	return t, nil
}

func (t *TabComponent) View() string {
	if len(t.Tabs) == 0 || t.Width == 0 {
		return ""
	}

	var tabViews []string
	maxWidth := min(max(t.Width/len(t.Tabs), 15), 30)

	for i, tab := range t.Tabs {
		tabView := t.renderTab(tab, i == t.ActiveIndex, maxWidth)
		tabViews = append(tabViews, tabView)
	}

	totalTabsWidth := len(tabViews) * maxWidth
	if totalTabsWidth < t.Width {
		padding := strings.Repeat(" ", t.Width-totalTabsWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...) + padding
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
}

func (t *TabComponent) renderTab(tab Tab, isActive bool, maxWidth int) string {
	title := tab.Title
	if len(title) > maxWidth-4 {
		title = title[:maxWidth-7] + "..."
	}

	content := fmt.Sprintf("%s ✕", title)

	var style lipgloss.Style
	if isActive {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.SelectionBackground)).
			Foreground(lipgloss.Color(customstyles.SelectionForeground)).
			Bold(true).
			Padding(0, 1).
			Width(maxWidth)
	} else {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1).
			Width(maxWidth)
	}

	return style.Render(content)
}

func (t *TabComponent) AddTab(id, title, resourceType string) {
	newTab := Tab{
		ID:           id,
		Title:        title,
		ResourceType: resourceType,
		IsActive:     false,
		IsModified:   false,
	}

	t.Tabs = append(t.Tabs, newTab)
}

func (t *TabComponent) RemoveTab(id string) {
	for i, tab := range t.Tabs {
		if tab.ID == id {
			t.Tabs = append(t.Tabs[:i], t.Tabs[i+1:]...)
			if t.ActiveIndex >= len(t.Tabs) && len(t.Tabs) > 0 {
				t.ActiveIndex = len(t.Tabs) - 1
			} else if t.ActiveIndex > i && t.ActiveIndex > 0 {
				t.ActiveIndex--
			}
			break
		}
	}
}

func (t *TabComponent) SetActiveTab(index int) {
	if index >= 0 && index < len(t.Tabs) {
		t.ActiveIndex = index
	}
}

func (t *TabComponent) GetActiveTab() *Tab {
	if t.ActiveIndex >= 0 && t.ActiveIndex < len(t.Tabs) {
		return &t.Tabs[t.ActiveIndex]
	}
	return nil
}

func (t *TabComponent) GetTabCount() int {
	return len(t.Tabs)
}

func (t *TabComponent) ClearTabs() {
	t.Tabs = []Tab{}
	t.ActiveIndex = 0
}
