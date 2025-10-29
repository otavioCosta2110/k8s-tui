package components

import (
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FullscreenListModel struct {
	list  *ListModel
	title string
}

func NewFullscreenList(items []string, title string, onSelect func(selected string) tea.Msg) *FullscreenListModel {
	list := NewList(items, "", onSelect)
	return &FullscreenListModel{
		list:  list,
		title: title,
	}
}

func (m *FullscreenListModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *FullscreenListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.list.Update(msg)
	if list, ok := updated.(*ListModel); ok {
		m.list = list
	}
	return m, cmd
}

func (m *FullscreenListModel) View() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(customstyles.BorderColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight + styles.HeaderSize + styles.Margin + 1).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Foreground(customstyles.TextColor)

	separator := strings.Repeat("─", styles.ScreenWidth-4)
	content := m.list.View()

	fullContent := m.title + "\n" + separator + "\n" + content

	return borderStyle.Render(fullContent)
}
