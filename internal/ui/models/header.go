package models

import (
	"fmt"
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HeaderRefreshInterval = 10 * time.Second
)

type HeaderRefreshMsg struct{}

type HeaderModel struct {
	content        string
	width          int
	height         int
	headerStyle    lipgloss.Style
	kubeconfig     *k8s.Client
	namespace      string
	metricsManager *MetricsManager
}

func NewHeader(headerText string, kubeconfig *k8s.Client) HeaderModel {
	return HeaderModel{
		content:    "",
		kubeconfig: kubeconfig,
		headerStyle: lipgloss.NewStyle().
			Height(global.HeaderSize),
	}
}

func (m HeaderModel) Init() tea.Cmd {
	if m.kubeconfig != nil {
		m.metricsManager = NewMetricsManager(*m.kubeconfig)
		m.updateContentFromManager()

		return tea.Tick(HeaderRefreshInterval, func(t time.Time) tea.Msg {
			return HeaderRefreshMsg{}
		})
	}
	m.height = global.HeaderSize
	m.headerStyle = m.headerStyle.Height(m.height)

	global.IsHeaderActive = true
	return nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.headerStyle = m.headerStyle.
		Height(m.height)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.headerStyle = m.headerStyle.
			Width(msg.Width).
			Height(m.height)
	case HeaderRefreshMsg:
		m.updateContentFromManager()
		if m.kubeconfig != nil {
			return m, tea.Tick(HeaderRefreshInterval, func(t time.Time) tea.Msg {
				return HeaderRefreshMsg{}
			})
		}
		return m, nil
	}
	return m, nil
}

func (m HeaderModel) View() string {
	if m.kubeconfig == nil {
		return m.headerStyle.Render("K8s TUI - No cluster connection")
	}
	return m.headerStyle.Render(m.content)
}

func (m HeaderModel) buildEnhancedHeader(metrics Metrics) string {
	clusterInfo := m.getClusterInfo()

	clusterSection := m.buildClusterSection(clusterInfo)
	metricsSection := m.buildMetricsSection(metrics)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		clusterSection,
		lipgloss.NewStyle().Width(4).Render(""),
		metricsSection,
	)
}

func (m HeaderModel) getClusterInfo() map[string]string {
	info := make(map[string]string)

	if m.kubeconfig == nil {
		return info
	}

	info["namespace"] = m.kubeconfig.Namespace
	if info["namespace"] == "" {
		info["namespace"] = "default"
	}

	if m.kubeconfig.Config != nil {
		info["server"] = m.kubeconfig.Config.Host
		if info["server"] == "" {
			info["server"] = "unknown"
		}
	}

	return info
}

func (m HeaderModel) buildClusterSection(info map[string]string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(customstyles.Foreground).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(customstyles.Foreground)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A1EFD3"))

	sectionStyle := lipgloss.NewStyle().
		Width(40).
		Padding(0, 2)

	content := []string{
		titleStyle.Render("Cluster Info"),
		labelStyle.Render("Namespace:") + " " + valueStyle.Render(info["namespace"]),
		labelStyle.Render("Server:") + " " + valueStyle.Render(info["server"]),
	}

	return sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

func (m HeaderModel) buildMetricsSection(metrics Metrics) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(customstyles.Foreground).
		Padding(0, 1)

	metricStyle := lipgloss.NewStyle().
		Foreground(customstyles.Foreground)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A1EFD3"))

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Italic(true)

	sectionStyle := lipgloss.NewStyle().
		Width(60).
		Padding(0, 2)

	formatMetric := func(label string, value int, loading bool) string {
		if loading && value == 0 {
			return metricStyle.Render(label+":") + " " + loadingStyle.Render("Loading...")
		}
		return metricStyle.Render(label+":") + " " + valueStyle.Render(fmt.Sprint(value))
	}

	content := []string{
		titleStyle.Render("Resources"),
		formatMetric("Pods", metrics.PodsNumber, metrics.Loading),
		formatMetric("Nodes", metrics.NodesNumber, metrics.Loading),
		formatMetric("Namespaces", metrics.NamespacesNumber, metrics.Loading),
		formatMetric("Deployments", metrics.DeploymentsNumber, metrics.Loading),
		formatMetric("Services", metrics.ServicesNumber, metrics.Loading),
	}

	return sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

func (m *HeaderModel) SetContent(content string) {
	m.content = content
}

func (m *HeaderModel) IsContentNil() bool {
	return m.content == ""
}

func (m *HeaderModel) SetKubeconfig(kubeconfig *k8s.Client) {
	m.kubeconfig = kubeconfig
	if kubeconfig != nil {
		m.namespace = kubeconfig.Namespace
		if m.namespace == "" {
			m.namespace = "default"
		}
		m.metricsManager = NewMetricsManager(*kubeconfig)
		m.updateContentFromManager()
	}
}

func (m *HeaderModel) SetNamespace(namespace string) {
	m.namespace = namespace
}

func (m *HeaderModel) UpdateContent() {
	m.updateContentFromManager()
}

func (m *HeaderModel) updateContentFromManager() {
	if m.metricsManager != nil {
		metrics := m.metricsManager.GetMetrics()
		m.content = m.buildEnhancedHeader(metrics)
	}
}

func (m *HeaderModel) RefreshMetrics() {
	if m.metricsManager != nil {
		m.updateContentFromManager()
	}
}

func (m *HeaderModel) Stop() {
	if m.metricsManager != nil {
		m.metricsManager.Stop()
	}
}
