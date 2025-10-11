package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type TextInputModel struct {
	textInput textinput.Model
	title     string
	onSubmit  func(value string) tea.Msg
	onCancel  func() tea.Msg
}

type TextInputSubmitMsg struct {
	Value string
}

type TextInputCancelMsg struct{}

func NewTextInput(title string, placeholder string, onSubmit func(value string) tea.Msg, onCancel func() tea.Msg) *TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.TextStyle = lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Foreground(lipgloss.Color(customstyles.TextColor))
	ti.CharLimit = 100
	ti.Width = 50

	return &TextInputModel{
		textInput: ti,
		title:     title,
		onSubmit:  onSubmit,
		onCancel:  onCancel,
	}
}

func (m *TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			value := m.textInput.Value()
			if m.onSubmit != nil {
				return m, func() tea.Msg {
					return m.onSubmit(value)
				}
			}
		case "esc":
			if m.onCancel != nil {
				return m, func() tea.Msg {
					return m.onCancel()
				}
			}
		default:
			m.textInput, cmd = m.textInput.Update(msg)
		}
	}

	return m, cmd
}

func (m *TextInputModel) View() string {
	W := styles.ScreenWidth + styles.Margin
	H := styles.ScreenHeight + styles.HeaderSize + styles.TabBarSize + 1
	halfW := W / 2

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(halfW).
		PaddingTop(H/2 - 4).
		Render(m.title)

	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		BorderForeground(lipgloss.Color(customstyles.BorderColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Padding(1).
		Width(halfW - 2).
		Render(m.textInput.View())

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		PaddingTop(1).
		Width(halfW).
		Render("Enter: Submit • Esc: Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		input,
		help,
	)

	return lipgloss.NewStyle().
		Width(W).
		Height(H).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(content)
}
