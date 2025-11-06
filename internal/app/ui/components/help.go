package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type HelpProvider interface {
	Help() (string, string)
}

type HelpModel struct {
	viewport viewport.Model
	width    int
	height   int
	title    string
	content  string
	visible  bool
}

type CloseHelpMsg struct{}

func NewHelpModel() *HelpModel {
	vp := viewport.New(styles.ScreenWidth, styles.ScreenHeight-4)
	vp.Style = lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return &HelpModel{
		viewport: vp,
		visible:  false,
	}
}

func (m *HelpModel) SetContent(title, content string) {
	m.title = title
	contentStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(content)
	m.content = contentStyled
	m.viewport.SetContent(contentStyled)
	m.visible = true
}

func (m *HelpModel) SetContentFromProvider(provider HelpProvider) {
	title, content := provider.Help()
	m.SetContent(title, content)
}

func (m *HelpModel) IsVisible() bool {
	return m.visible
}

func (m *HelpModel) Init() tea.Cmd {
	return nil
}

func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		if m.content != "" {
			m.viewport.SetContent(m.content)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			m.visible = false
			return m, func() tea.Msg { return CloseHelpMsg{} }
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		case "pgup":
			m.viewport.HalfPageUp()
		case "pgdown":
			m.viewport.HalfPageDown()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *HelpModel) View() string {
	if !m.visible {
		return ""
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Align(lipgloss.Center).
		Width(styles.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(m.title)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Align(lipgloss.Center).
		Width(styles.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("↑/↓ or j/k: Scroll • esc/q/?: Close")

	spacer1 := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("")

	spacer2 := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("")

	contentHeight := (styles.ScreenHeight + styles.HeaderSize)
	m.viewport.Height = contentHeight

	content := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(contentHeight).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(m.viewport.View())

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		spacer1,
		content,
		spacer2,
		footer,
	)

	return lipgloss.NewStyle().
		Width(styles.ScreenWidth + styles.Margin).
		Height(styles.ScreenHeight).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(fullContent)
}

func (m *HelpModel) Close() {
	m.visible = false
}
