package components

import (
	global "otaviocosta2110/k8s-tui/internal"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type YAMLViewer struct {
	title    string
	content  string
	viewport viewport.Model
	ready    bool
	styles   *YAMLViewerStyles
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
		Background(lipgloss.Color("#7D56F4")),
	BorderColor:   "#7D56F4",
	HelpTextColor: "#757575",
}

func NewYAMLViewer(title, content string) *YAMLViewer {
	highlighted := highlightYAML(content)
	return &YAMLViewer{
		title:   title,
		content: highlighted,
		styles:  DefaultYAMLViewerStyles,
	}
}

func (m *YAMLViewer) Init() tea.Cmd {
	contentWidth := global.ScreenWidth
	m.viewport = viewport.New(contentWidth, global.ScreenHeight - global.Margin)
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return loadedMsg{}
	})
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
			contentWidth := global.ScreenWidth
			m.viewport = viewport.New(contentWidth, global.ScreenHeight)
			m.ready = true
			m.viewport.SetContent(m.content)
		} else {
			m.viewport.Width = global.ScreenWidth
			m.viewport.Height = global.ScreenHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *YAMLViewer) View() string {
	header := m.headerView()
	footer := m.footerView()
	content := lipgloss.NewStyle().
		PaddingLeft(m.styles.ContentPadding).
		Render(m.viewport.View())

	m.viewport.Width = global.ScreenWidth
	m.viewport.Height = global.ScreenHeight - global.Margin
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

func highlightYAML(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	var highlighted strings.Builder

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5E9AFF"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.NoColor{})

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				highlighted.WriteString(
					keyStyle.Render(parts[0]+":") +
						valueStyle.Render(parts[1]) + "\n",
				)
				continue
			}
		}
		highlighted.WriteString(line + "\n")
	}
	return highlighted.String()
}

