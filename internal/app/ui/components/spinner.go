package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type SpinnerModel struct {
	text   string
	frame  int
	ticker *time.Ticker
}

type spinnerTickMsg struct{}

func NewSpinner(text string) SpinnerModel {
	return SpinnerModel{
		text:   text,
		frame:  0,
		ticker: nil,
	}
}

func (s SpinnerModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (s SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	switch msg.(type) {
	case spinnerTickMsg:
		s.frame = (s.frame + 1) % 10
		return s, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return spinnerTickMsg{}
		})
	}
	return s, nil
}

func (s SpinnerModel) View() string {
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	char := spinnerChars[s.frame]

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(char + " " + s.text)
}

func (s SpinnerModel) CenteredView(width, height int) string {
	spinnerView := s.View()

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(spinnerView)
}

func (s SpinnerModel) CenteredScreenView() string {
	return s.CenteredView(styles.ScreenWidth, styles.ScreenHeight)
}
