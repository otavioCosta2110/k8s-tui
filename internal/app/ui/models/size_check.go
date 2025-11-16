package models

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"time"
)

const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

type SizeCheckModel struct {
	width       int
	height      int
	lastCheck   time.Time
	checkPeriod time.Duration
}

func NewSizeCheckModel() *SizeCheckModel {
	return &SizeCheckModel{
		checkPeriod: 500 * time.Millisecond,
		lastCheck:   time.Now(),
	}
}

func (m *SizeCheckModel) Init() tea.Cmd {
	return tea.Tick(m.checkPeriod, func(t time.Time) tea.Msg {
		return SizeCheckMsg{}
	})
}

func (m *SizeCheckModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Handle quit keys when terminal is too small
		if !m.isSizeValid() {
			switch msg.String() {
			case "esc", "q", "Q":
				return m, tea.Quit
			}
		}
	case SizeCheckMsg:
		m.lastCheck = time.Now()
		return m, tea.Tick(m.checkPeriod, func(t time.Time) tea.Msg {
			return SizeCheckMsg{}
		})
	}
	return m, nil
}

func (m *SizeCheckModel) View() string {
	if m.isSizeValid() {
		return ""
	}

	title := "Terminal Size Too Small"
	message := fmt.Sprintf("Current: %dx%d | Required: %dx%d",
		m.width, m.height, MinTerminalWidth, MinTerminalHeight)
	instruction := "Please resize your terminal window to continue"
	quitOption := "Press Esc or Q to quit"

	// Use colorscheme colors
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.ErrorColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(m.width).
		Align(lipgloss.Center)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(m.width).
		Align(lipgloss.Center)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(m.width).
		MarginTop(1)

	quitStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(m.width).
		MarginTop(1)

	// Create content without full-height container
	titleText := titleStyle.Render(title)
	messageText := messageStyle.Render(message)
	instructionText := instructionStyle.Render(instruction)
	quitText := quitStyle.Render(quitOption)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleText,
		messageText,
		instructionText,
		quitText,
	)

	contentStyled := lipgloss.NewStyle().
	Background(lipgloss.Color(customstyles.BackgroundColor)).
	Render(content)

	// Center content in terminal without filling entire height
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		contentStyled,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)),
	)
}

func (m *SizeCheckModel) isSizeValid() bool {
	return m.width >= MinTerminalWidth && m.height >= MinTerminalHeight
}

func (m *SizeCheckModel) IsSizeValid() bool {
	return m.isSizeValid()
}

type SizeCheckMsg struct{}

func CheckTerminalSize(width, height int) bool {
	return width >= MinTerminalWidth && height >= MinTerminalHeight
}
