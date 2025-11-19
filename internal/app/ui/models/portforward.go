package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

type PortForwardModel struct {
	viewport      viewport.Model
	podName       string
	namespace     string
	localPort     int
	remotePort    int
	session       *k8s.PortForwardSession
	pluginAPI     plugins.PluginAPI
	isActive      bool
	statusMessage string
	logLines      []string
}

func NewPortForwardModel(podName, namespace string, localPort, remotePort int, pluginAPI plugins.PluginAPI) *PortForwardModel {
	vp := viewport.New(80, 20)

	model := &PortForwardModel{
		viewport:      vp,
		podName:       podName,
		namespace:     namespace,
		localPort:     localPort,
		remotePort:    remotePort,
		pluginAPI:     pluginAPI,
		isActive:      false,
		statusMessage: "Initializing port forwarding...",
		logLines:      []string{},
	}

	return model
}

func (m *PortForwardModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Start port forwarding in a goroutine
		go func() {
			client := m.pluginAPI.GetClient()
			pod := k8s.NewPodInfo(m.podName, m.namespace, client)
			session, err := pod.PortForward(m.localPort, m.remotePort)

			if err != nil {
				m.logLines = append(m.logLines, fmt.Sprintf("Error: %v", err))
				m.statusMessage = fmt.Sprintf("Failed to start port forwarding: %v", err)
				return
			}

			m.session = session
			m.isActive = true
			m.statusMessage = "Port forwarding active"
			m.logLines = append(m.logLines, "Port forwarding started successfully")
			m.logLines = append(m.logLines, fmt.Sprintf("Pod: %s/%s", m.namespace, m.podName))
			m.logLines = append(m.logLines, fmt.Sprintf("Forwarding: localhost:%d -> %d", m.localPort, m.remotePort))
			m.logLines = append(m.logLines, "")
			m.logLines = append(m.logLines, "")
			m.logLines = append(m.logLines, "Controls:")
			m.logLines = append(m.logLines, "  q: Stop port forwarding and go back")
			m.logLines = append(m.logLines, "  ↑/↓: Scroll viewport")

			// Keep the session alive
			<-session.StopChan
			m.isActive = false
			m.statusMessage = "Port forwarding stopped"
			m.logLines = append(m.logLines, "Port forwarding stopped")
		}()

		return nil
	}
}

func (m *PortForwardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = styles.ScreenHeight
		m.viewport.SetContent(m.renderContent())
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.session != nil {
				m.session.Stop()
			}
			m.isActive = false
			m.statusMessage = "Port forwarding stopped by user"
			// Return a back key message to trigger navigation back
			return m, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
			}
		case "up", "k":
			m.viewport.LineUp(1)
		case "down", "j":
			m.viewport.LineDown(1)
		case "pgup":
			m.viewport.HalfPageUp()
		case "pgdown":
			m.viewport.HalfPageDown()
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *PortForwardModel) View() string {
	if m.viewport.Height == 0 {
		m.viewport.SetContent(m.renderContent())
	}
	return m.viewport.View()
}

func (m *PortForwardModel) renderContent() string {
	style := lipgloss.NewStyle().
		Padding(1, 2).
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	content := strings.Join(m.logLines, "\n")
	if content == "" {
		content = m.statusMessage
	}

	return style.Render(content)
}

func (m *PortForwardModel) GetBreadcrumb() string {
	return fmt.Sprintf("%s port-forward", m.podName)
}
