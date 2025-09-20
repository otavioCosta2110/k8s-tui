package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	global "github.com/otavioCosta2110/k8s-tui/pkg/global"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HeaderRefreshInterval = 10 * time.Second
)

type HeaderRefreshMsg struct{}

type HeaderModel struct {
	content          string
	width            int
	height           int
	headerStyle      lipgloss.Style
	kubeconfig       *k8s.Client
	namespace        string
	metricsManager   *MetricsManager
	tabComponent     *components.TabComponent
	pluginComponents []string
}

func NewHeader(headerText string, kubeconfig *k8s.Client) HeaderModel {
	return HeaderModel{
		content:      "",
		kubeconfig:   kubeconfig,
		headerStyle:  lipgloss.NewStyle().Height(global.HeaderSize).Background(lipgloss.Color(customstyles.BackgroundColor)),
		tabComponent: components.NewTabComponent(),
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
	m.headerStyle = m.headerStyle.Height(m.height).Background(lipgloss.Color(customstyles.BackgroundColor))

	global.IsHeaderActive = true
	return m.tabComponent.Init()
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.headerStyle = m.headerStyle.
		Height(m.height).Background(lipgloss.Color(customstyles.BackgroundColor))

	var tabCmd tea.Cmd
	if m.tabComponent != nil {
		updatedTab, cmd := m.tabComponent.Update(msg)
		if tab, ok := updatedTab.(*components.TabComponent); ok {
			m.tabComponent = tab
			tabCmd = cmd
		}
	}

	switch msg := msg.(type) {
	case components.TabMsg:
		return m, func() tea.Msg { return msg }
	case tea.WindowSizeMsg:
		m.headerStyle = m.headerStyle.
			Width(msg.Width).
			Height(m.height)
		if m.tabComponent != nil {
			m.tabComponent.Width = msg.Width
		}
	case HeaderRefreshMsg:
		m.updateContentFromManager()
		if m.kubeconfig != nil {
			return m, tea.Tick(HeaderRefreshInterval, func(t time.Time) tea.Msg {
				return HeaderRefreshMsg{}
			})
		}
		return m, tabCmd
	}
	return m, tabCmd
}

func (m HeaderModel) View() string {
	var headerView string
	if m.kubeconfig == nil {
		headerView = m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("K8s TUI - No cluster connection")
	} else {
		headerView = m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(m.content)
	}

	// Add plugin components to header
	if len(m.pluginComponents) > 0 {
		pluginContent := strings.Join(m.pluginComponents, " | ")
		headerView = m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(m.content + " | " + pluginContent)
	}

	if m.tabComponent != nil && m.tabComponent.GetTabCount() > 0 {
		tabView := m.tabComponent.View()
		if tabView != "" {
			return lipgloss.JoinVertical(lipgloss.Top, headerView, tabView)
		}
	}

	return headerView
}

func (m *HeaderModel) AddPluginComponent(component string) {
	m.pluginComponents = append(m.pluginComponents, component)
}

func (m HeaderModel) buildEnhancedHeader(metrics Metrics) string {
	clusterInfo := m.getClusterInfo()

	clusterSection := m.buildClusterSection(clusterInfo)
	metricsSection := m.buildMetricsSection(metrics)

	clusterLines := strings.Split(strings.TrimSuffix(clusterSection, "\n"), "\n")
	metricsLines := strings.Split(strings.TrimSuffix(metricsSection, "\n"), "\n")

	maxLines := max(len(metricsLines), len(clusterLines))

	for len(clusterLines) < maxLines {
		clusterLines = append(clusterLines, "")
	}
	for len(metricsLines) < maxLines {
		metricsLines = append(metricsLines, "")
	}

	resultLines := make([]string, maxLines)
	for i := range maxLines {
		clusterLine := lipgloss.NewStyle().
			Width(40).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(clusterLines[i])

		spacer := lipgloss.NewStyle().
			Width(4).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("    ")

		metricsLine := lipgloss.NewStyle().
			Width(60).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(metricsLines[i])

		resultLines[i] = lipgloss.JoinHorizontal(lipgloss.Left,
			clusterLine, spacer, metricsLine)
	}

	return strings.Join(resultLines, "\n")
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
		Foreground(customstyles.TextColor).
		Padding(0, 1).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	labelStyle := lipgloss.NewStyle().
		Foreground(customstyles.TextColor).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderValueColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	content := []string{
		titleStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Cluster Info"),
		lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Namespace:"),
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" "),
			valueStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(info["namespace"])),
		lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Server:"),
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" "),
			valueStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(info["server"])),
	}

	filledContent := make([]string, len(content))
	for i, line := range content {
		filledContent[i] = lipgloss.PlaceHorizontal(40, lipgloss.Left, line, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, filledContent...)
}

func (m HeaderModel) buildMetricsSection(metrics Metrics) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(customstyles.TextColor).
		Padding(0, 1).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	metricStyle := lipgloss.NewStyle().
		Foreground(customstyles.TextColor).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderValueColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderLoadingColor)).
		Italic(true).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	formatMetric := func(label string, value int, loading bool) string {
		displayLabel := label
		if icon, exists := customstyles.ResourceIcons[label]; exists {
			displayLabel = icon + " " + label
		}

		if loading && value == 0 {
			return lipgloss.JoinHorizontal(lipgloss.Left,
				metricStyle.Render(displayLabel+":"),
				metricStyle.Render(" "),
				loadingStyle.Render("Loading..."))
		}
		return lipgloss.JoinHorizontal(lipgloss.Left,
			metricStyle.Render(displayLabel+":"),
			metricStyle.Render(" "),
			valueStyle.Render(fmt.Sprint(value)))
	}

	content := []string{
		titleStyle.Render("Resources"),
		formatMetric("Pods", metrics.PodsNumber, metrics.Loading),
		formatMetric("Nodes", metrics.NodesNumber, metrics.Loading),
		formatMetric("Namespaces", metrics.NamespacesNumber, metrics.Loading),
		formatMetric("Deployments", metrics.DeploymentsNumber, metrics.Loading),
		formatMetric("Services", metrics.ServicesNumber, metrics.Loading),
	}

	filledContent := make([]string, len(content))
	for i, line := range content {
		filledContent[i] = lipgloss.PlaceHorizontal(60, lipgloss.Left, line, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, filledContent...)
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

func (m *HeaderModel) AddTab(id, title, resourceType string) {
	if m.tabComponent != nil {
		m.tabComponent.AddTab(id, title, resourceType)
	}
}

func (m *HeaderModel) RemoveTab(id string) {
	if m.tabComponent != nil {
		m.tabComponent.RemoveTab(id)
	}
}

func (m *HeaderModel) SetActiveTab(index int) {
	if m.tabComponent != nil {
		m.tabComponent.SetActiveTab(index)
	}
}

func (m *HeaderModel) GetActiveTab() *components.Tab {
	if m.tabComponent != nil {
		return m.tabComponent.GetActiveTab()
	}
	return nil
}

func (m *HeaderModel) GetTabCount() int {
	if m.tabComponent != nil {
		return m.tabComponent.GetTabCount()
	}
	return 0
}

func (m *HeaderModel) ClearTabs() {
	if m.tabComponent != nil {
		m.tabComponent.ClearTabs()
	}
}

func (m *HeaderModel) GetActiveTabIndex() int {
	if m.tabComponent != nil {
		return m.tabComponent.ActiveIndex
	}
	return 0
}
