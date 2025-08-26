package ui

import (
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/cli"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"otaviocosta2110/k8s-tui/internal/ui/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	stack          []tea.Model
	kube           k8s.Client
	header         models.HeaderModel
	configSelected bool
	errorPopup     *models.ErrorModel
}

func NewAppModel() *AppModel {
	cfg := cli.ParseFlags()

	kubeClient, err := k8s.NewClient(cfg.KubeconfigPath, cfg.Namespace)
	if err == nil && kubeClient != nil {
		mainModel, err := models.NewMainModel(*kubeClient, cfg.Namespace).InitComponent(*kubeClient)
		if err != nil {
			popup := models.NewErrorScreen(err, "Failed to initialize main view", "")
			return &AppModel{
				errorPopup: &popup,
			}
		}

		header := models.NewHeader("K8s TUI", kubeClient)
		header.SetNamespace(cfg.Namespace)

		return &AppModel{
			stack:          []tea.Model{mainModel},
			header:         header,
			kube:           *kubeClient,
			configSelected: true,
		}
	}

	initialModel, err := models.NewKubeconfigModel().InitComponent(nil)
	if err != nil {
		popup := models.NewErrorScreen(err, "Failed to initialize Kubernetes config", "")
		return &AppModel{
			stack:      []tea.Model{initialModel},
			header:     models.NewHeader("K8s TUI", nil),
			errorPopup: &popup,
		}
	}

	return &AppModel{
		stack:  []tea.Model{initialModel},
		header: models.NewHeader("K8s TUI", nil),
	}
}

func (m *AppModel) Init() tea.Cmd {
	if len(m.stack) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, m.stack[len(m.stack)-1].Init())

	if m.configSelected {
		cmds = append(cmds, m.header.Init())
	}

	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		global.ScreenWidth = msg.Width - global.Margin
		global.ScreenHeight = msg.Height - global.Margin
		if !global.IsHeaderActive {
			global.HeaderSize = global.ScreenHeight/4 - ((global.Margin * 2) - 1)
			global.IsHeaderActive = true
		}
		global.ScreenHeight -= global.HeaderSize

		var cmds []tea.Cmd
		if m.configSelected {
			newHeader, headerCmd := m.header.Update(msg)
			m.header = newHeader.(models.HeaderModel)
			m.header.SetKubeconfig(&m.kube)
			m.header.UpdateContent()
			cmds = append(cmds, headerCmd)
		}

		for i := range m.stack {
			var cmd tea.Cmd
			m.stack[i], cmd = m.stack[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.errorPopup != nil {
				m.errorPopup = nil
				return m, nil
			}
			if len(m.stack) > 1 {
				m.stack = m.stack[:len(m.stack)-1]
				return m, nil
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			current := len(m.stack) - 1
			m.stack[current], cmd = m.stack[current].Update(msg)
			return m, cmd
		}

	case components.NavigateMsg:
		if msg.Error != nil {
			popup := models.NewErrorScreen(
				msg.Error,
				"Kubernetes Connection Error",
				"Failed to connect to the Kubernetes cluster",
			)
			popup.SetDimensions(global.ScreenWidth, global.ScreenHeight)

			return &AppModel{
				stack:      m.stack,
				header:     m.header,
				kube:       msg.Cluster,
				errorPopup: &popup,
			}, nil
		}

		m.stack = append(m.stack, msg.NewScreen)
		if !m.configSelected {
			m.configSelected = true
			m.header.SetKubeconfig(&msg.Cluster)
			m.kube = msg.Cluster
			m.header.UpdateContent()

			return m, tea.Batch(
				msg.NewScreen.Init(),
				m.header.Init(),
			)
		}
		return m, msg.NewScreen.Init()

	case models.HeaderRefreshMsg:
		if m.configSelected {
			newHeader, headerCmd := m.header.Update(msg)
			m.header = newHeader.(models.HeaderModel)
			return m, headerCmd
		}
		return m, nil

	default:
		var cmd tea.Cmd
		current := len(m.stack) - 1
		m.stack[current], cmd = m.stack[current].Update(msg)
		return m, cmd
	}
}

func (m *AppModel) View() string {
	if m.errorPopup != nil {
		return m.errorPopup.View()
	}

	if len(m.stack) == 0 {
		return "Loading..."
	}

	currentView := m.stack[len(m.stack)-1].View()

	var height int
	if !global.IsHeaderActive {
		height = global.ScreenHeight + global.HeaderSize
	} else {
		height = global.ScreenHeight
	}

	headerView := m.header.View()

	content := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(customstyles.Blue)).
		Render(currentView)

	// If no kubeconfig is selected, show content fullscreen without header
	if !m.configSelected {
		return lipgloss.NewStyle().
			Width(global.ScreenWidth).
			Height(global.ScreenHeight + global.HeaderSize).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(customstyles.Blue)).
			Render(currentView)
	}

	if !global.IsHeaderActive {
		return content
	}
	return lipgloss.JoinVertical(lipgloss.Top, headerView, content)
}
