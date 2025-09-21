package components

import (
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type YAMLViewer struct {
	title           string
	content         string
	originalContent string
	viewport        viewport.Model
	ready           bool
	styles          *YAMLViewerStyles
	customHelp      string
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
		Foreground(lipgloss.Color(customstyles.YAMLTitleColor)).
		Background(lipgloss.Color(customstyles.BorderColor)),
	BorderColor:   customstyles.BorderColor,
	HelpTextColor: customstyles.HelpTextColor,
}

type loadedMsg struct{}

func NewYAMLViewer(title, content string) *YAMLViewer {
	highlighted := highlightYAML(content)
	return &YAMLViewer{
		title:           title,
		content:         highlighted,
		originalContent: content,
		styles: &YAMLViewerStyles{
			TitleBar: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(customstyles.YAMLTitleColor)).
				Background(lipgloss.Color(customstyles.BorderColor)),
			BorderColor:   customstyles.BorderColor,
			HelpTextColor: customstyles.HelpTextColor,
		},
	}
}

func NewYAMLViewerWithHelp(title, content, helpText string) *YAMLViewer {
	highlighted := highlightYAML(content)
	return &YAMLViewer{
		title:           title,
		content:         highlighted,
		originalContent: content,
		styles: &YAMLViewerStyles{
			TitleBar: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(customstyles.YAMLTitleColor)).
				Background(lipgloss.Color(customstyles.BorderColor)),
			BorderColor:   customstyles.BorderColor,
			HelpTextColor: customstyles.HelpTextColor,
		},
		customHelp: helpText,
	}
}

func (m *YAMLViewer) SetCustomHelp(helpText string) {
	m.customHelp = helpText
}

func (m *YAMLViewer) GetContent() string {
	return m.content
}

func (m *YAMLViewer) GetOriginalContent() string {
	return m.originalContent
}

func (m *YAMLViewer) Init() tea.Cmd {
	contentWidth := styles.ScreenWidth
	m.viewport = viewport.New(contentWidth, styles.ScreenHeight-1)
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
		case "e":
			return m, func() tea.Msg {
				return EditMsg{Content: m.originalContent, Title: m.title}
			}
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			contentWidth := styles.ScreenWidth
			m.viewport = viewport.New(contentWidth, styles.ScreenHeight-1)
			m.ready = true
			m.viewport.SetContent(m.content)
		} else {
			m.viewport.Width = styles.ScreenWidth
			m.viewport.Height = styles.ScreenHeight - 1
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
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(m.viewport.View())

	m.viewport.Width = styles.ScreenWidth
	m.viewport.Height = styles.ScreenHeight - 1
	m.viewport.SetContent(m.content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m *YAMLViewer) headerView() string {
	title := m.styles.TitleBar.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(m.title)
	return lipgloss.PlaceHorizontal(styles.ScreenWidth, lipgloss.Left, title, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
}

func (m *YAMLViewer) footerView() string {
	helpText := "↑/↓: Scroll • q: Quit"
	if m.customHelp != "" {
		helpText = m.customHelp
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(helpText)

	return lipgloss.PlaceHorizontal(styles.ScreenWidth, lipgloss.Left, help, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
}

func highlightYAML(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	var highlighted strings.Builder

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
	plainStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor))

	width := styles.ScreenWidth

	for _, line := range lines {
		var renderedLine string
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				renderedLine = keyStyle.Render(parts[0]+":") + valueStyle.Render(parts[1])
			} else {
				renderedLine = plainStyle.Render(line)
			}
		} else {
			renderedLine = plainStyle.Render(line)
		}

		filledLine := lipgloss.PlaceHorizontal(width, lipgloss.Left, renderedLine, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
		highlighted.WriteString(filledLine + "\n")
	}
	return highlighted.String()
}
