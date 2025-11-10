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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update screen dimensions
		styles.ScreenWidth = msg.Width - styles.Margin
		styles.ScreenHeight = msg.Height - (styles.Margin * 2) - 1
		// Let the underlying list handle the resize as well
		updated, cmd := m.list.Update(msg)
		if list, ok := updated.(*ListModel); ok {
			m.list = list
		}
		return m, cmd
	default:
		updated, cmd := m.list.Update(msg)
		if list, ok := updated.(*ListModel); ok {
			m.list = list
		}
		return m, cmd
	}
}

func (m *FullscreenListModel) View() string {
	// Calculate available space for content
	availableWidth := styles.ScreenWidth - 4 // Account for borders

	// Ensure we have valid dimensions
	if styles.ScreenWidth <= 0 {
		styles.ScreenWidth = 80
	}
	if styles.ScreenHeight <= 0 {
		styles.ScreenHeight = 24
	}
	if availableWidth <= 0 {
		availableWidth = 76 // 80 - 4 for borders
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(customstyles.BorderColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight + styles.HeaderSize + styles.Margin + 1).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Foreground(customstyles.TextColor)

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Align(lipgloss.Center).
		Width(availableWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	// Style for help footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Align(lipgloss.Center).
		Width(availableWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	separator := strings.Repeat("─", availableWidth)
	content := m.list.View()
	helpText := "↑/↓ or j/k: Navigate • Enter: Select • esc: Close"

	fullContent := titleStyle.Render(m.title) + "\n" +
		separator + "\n" +
		content + "\n" +
		separator + "\n" +
		footerStyle.Render(helpText)

	return borderStyle.Render(fullContent)
}
