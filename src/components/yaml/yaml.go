package yaml

import (
	"otaviocosta2110/k8s-tui/src/global"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type YAMLViewer struct {
	title       string
	content     string
	viewport    viewport.Model
	ready       bool
	styles      *YAMLViewerStyles
}

type YAMLViewerStyles struct {
	TitleBar       lipgloss.Style
	BorderColor    string
	HelpTextColor  string
	ContentPadding int
}

var DefaultYAMLViewerStyles = &YAMLViewerStyles{
	TitleBar: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1),
	BorderColor:    "#7D56F4",
	HelpTextColor:  "#757575",
	ContentPadding: 2,
}

func NewYAMLViewer(title, content string) *YAMLViewer {
	return &YAMLViewer{
		title:    title,
		content:  content,
		styles:   DefaultYAMLViewerStyles,
	}
}

func (m *YAMLViewer) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		func() tea.Msg {
			return tea.WindowSizeMsg{Width: global.ScreenWidth + global.Margin, Height: global.ScreenHeight + global.Margin / 2} 
		},
	)
}

func (m *YAMLViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			contentHeight := msg.Height - global.Margin
			contentWidth := msg.Width - global.Margin
			m.viewport = viewport.New(contentWidth, contentHeight)
			m.ready = true
			m.viewport.SetContent(m.content)
		} else {
			m.viewport.Width = msg.Width - global.Margin
			m.viewport.Height = msg.Height - 10
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *YAMLViewer) View() string {
	// if !m.ready {
	// 	return "Loading..."
	// }

	header := m.headerView()
	footer := m.footerView()
	content := lipgloss.NewStyle().
		PaddingLeft(m.styles.ContentPadding).
		Render(m.viewport.View())

	m.viewport.Width = global.ScreenWidth - global.Margin
	m.viewport.Height = global.ScreenHeight -global.Margin/2
	m.viewport.SetContent(m.content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m *YAMLViewer) headerView() string {
	title := m.styles.TitleBar.Render(m.title)
	return title
}

func (m *YAMLViewer) footerView() string {
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.HelpTextColor)).
		Render("↑/↓: Scroll • q: Quit")

	return help
}
