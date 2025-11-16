package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
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
	pluginAPI        any // Plugin API interface for getting current namespace
}

func NewHeader(headerText string, kubeconfig *k8s.Client, pluginAPI interface{}) HeaderModel {
	return HeaderModel{
		content:      headerText,
		kubeconfig:   kubeconfig,
		pluginAPI:    pluginAPI,
		headerStyle:  lipgloss.NewStyle().Height(styles.HeaderSize).Background(lipgloss.Color(customstyles.BackgroundColor)),
		tabComponent: components.NewTabComponent(),
		height:       styles.HeaderSize,
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
	m.height = styles.HeaderSize
	m.headerStyle = m.headerStyle.Height(m.height).Background(lipgloss.Color(customstyles.BackgroundColor))

	styles.IsHeaderActive = true
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
		m.width = msg.Width
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
	var left, right string

	if m.kubeconfig == nil {
		left = "K8s TUI - No cluster connection"
	} else {
		left = m.content
	}

	if len(m.pluginComponents) > 0 {
		right = strings.Join(m.pluginComponents, " | ")
	}

	leftLines := strings.Split(left, "\n")

	// Handle very small terminals with better responsiveness
	if m.width < 60 {
		// For very small terminals, show only essential info
		if len(leftLines) > 0 {
			// Truncate the first line to fit
			essentialInfo := leftLines[0]
			if len(essentialInfo) > m.width-2 && m.width > 5 {
				essentialInfo = essentialInfo[:m.width-5] + "..."
			}
			line1 := lipgloss.NewStyle().
				Width(m.width).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(essentialInfo)
			return m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(line1)
		}
	}

	rightWidth := lipgloss.Width(right)

	// Ensure we have enough space for both left and right content
	if rightWidth > m.width/3 {
		rightWidth = m.width / 3
		// Truncate right content if needed
		if len(right) > rightWidth-3 && rightWidth > 3 {
			right = right[:rightWidth-3] + "..."
		}
	}

	leftWidth := m.width - rightWidth // Leave some spacing

	line1Left := lipgloss.NewStyle().
		Width(leftWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(leftLines[0])
	line1Right := lipgloss.NewStyle().
		Width(rightWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Align(lipgloss.Right).
		Render(right)
	line1 := lipgloss.JoinHorizontal(lipgloss.Left, line1Left, line1Right)

	otherLines := leftLines[1:]

	paddedOtherLines := make([]string, len(otherLines))
	for i, line := range otherLines {
		// Truncate lines that are too long
		if lipgloss.Width(line) > m.width-2 && m.width > 5 {
			line = line[:m.width-5] + "..."
		}
		paddedOtherLines[i] = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, line, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
	}

	allLines := append([]string{line1}, paddedOtherLines...)

	headerView := lipgloss.JoinVertical(lipgloss.Left, allLines...)

	if m.tabComponent != nil {
		tabView := m.tabComponent.View()
		if tabView != "" {
			return lipgloss.JoinVertical(lipgloss.Top, m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(headerView), tabView)
		}
	}

	return m.headerStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(headerView)
}

func (m *HeaderModel) AddPluginComponent(component string) {
	m.pluginComponents = append(m.pluginComponents, component)
}

func (m HeaderModel) buildEnhancedHeader(metrics Metrics) string {
	clusterInfo := m.getClusterInfo()

	// For very narrow terminals, use a compact layout
	if m.width < 80 {
		return m.buildCompactHeader(clusterInfo, metrics)
	}

	clusterSection := m.buildClusterSection(clusterInfo)
	metricsSection := m.buildMetricsSection(metrics)

	clusterLines := strings.Split(clusterSection, "\n")
	metricsLines := strings.Split(metricsSection, "\n")

	maxLines := max(len(metricsLines), len(clusterLines))

	for len(clusterLines) < maxLines {
		clusterLines = append(clusterLines, "")
	}
	for len(metricsLines) < maxLines {
		metricsLines = append(metricsLines, "")
	}

	// Calculate dynamic widths based on available space
	clusterWidth := min(40, m.width/3)
	metricsWidth := min(60, m.width-clusterWidth-4)

	resultLines := make([]string, maxLines)
	for i := range maxLines {
		clusterLine := lipgloss.NewStyle().
			Width(clusterWidth).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(clusterLines[i])

		spacer := lipgloss.NewStyle().
			Width(4).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("    ")

		metricsLine := lipgloss.NewStyle().
			Width(metricsWidth).
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

	// Use the plugin API to get the current namespace instead of the internal field
	if m.pluginAPI != nil {
		// Try to call GetCurrentNamespace on the plugin API
		if api, ok := m.pluginAPI.(interface{ GetCurrentNamespace() string }); ok {
			info["namespace"] = api.GetCurrentNamespace()
		} else {
			// Fallback to internal namespace field
			info["namespace"] = m.namespace
		}
	} else {
		// Fallback to internal namespace field
		info["namespace"] = m.namespace
	}

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

	serverValue := info["server"]
	// Adjust truncation based on available width
	maxServerLength := 25
	if m.width < 100 {
		maxServerLength = 15
	}
	if len(serverValue) > maxServerLength {
		serverValue = serverValue[:maxServerLength-3] + "..."
	}

	content := []string{
		titleStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Cluster Info"),
		lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Namespace:"),
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" "),
			valueStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(info["namespace"])),
		lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Server:"),
			labelStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" "),
			valueStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render(serverValue)),
	}

	// Use dynamic width based on available space
	sectionWidth := min(40, m.width/3)
	filledContent := make([]string, len(content))
	for i, line := range content {
		filledContent[i] = lipgloss.PlaceHorizontal(sectionWidth, lipgloss.Left, line, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
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
		titleStyle.Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Resources"),
		formatMetric("Pods", metrics.PodsNumber, metrics.Loading),
		formatMetric("Nodes", metrics.NodesNumber, metrics.Loading),
		formatMetric("Namespaces", metrics.NamespacesNumber, metrics.Loading),
		formatMetric("Deployments", metrics.DeploymentsNumber, metrics.Loading),
		formatMetric("Services", metrics.ServicesNumber, metrics.Loading),
	}

	// Use dynamic width based on available space
	sectionWidth := min(60, m.width*2/3)
	filledContent := make([]string, len(content))
	for i, line := range content {
		filledContent[i] = lipgloss.PlaceHorizontal(sectionWidth, lipgloss.Left, line, lipgloss.WithWhitespaceBackground(lipgloss.Color(customstyles.BackgroundColor)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, filledContent...)
}

func (m HeaderModel) buildCompactHeader(clusterInfo map[string]string, metrics Metrics) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(customstyles.TextColor).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderValueColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderLoadingColor)).
		Italic(true).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	// Truncate server URL for compact display
	serverValue := clusterInfo["server"]
	if len(serverValue) > 15 {
		serverValue = serverValue[:12] + "..."
	}

	// Build compact lines
	lines := []string{
		titleStyle.Render("Cluster: " + clusterInfo["namespace"] + "@" + serverValue),
	}

	// Add metrics in compact format
	metricItems := []string{}
	if metrics.Loading {
		metricItems = append(metricItems, loadingStyle.Render("Loading..."))
	} else {
		if metrics.PodsNumber > 0 {
			metricItems = append(metricItems, valueStyle.Render(fmt.Sprintf("P:%d", metrics.PodsNumber)))
		}
		if metrics.NodesNumber > 0 {
			metricItems = append(metricItems, valueStyle.Render(fmt.Sprintf("N:%d", metrics.NodesNumber)))
		}
		if metrics.DeploymentsNumber > 0 {
			metricItems = append(metricItems, valueStyle.Render(fmt.Sprintf("D:%d", metrics.DeploymentsNumber)))
		}
		if metrics.ServicesNumber > 0 {
			metricItems = append(metricItems, valueStyle.Render(fmt.Sprintf("S:%d", metrics.ServicesNumber)))
		}
	}

	if len(metricItems) > 0 {
		lines = append(lines, strings.Join(metricItems, " "))
	}

	return strings.Join(lines, "\n")
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
	m.updateContent()
}

func (m *HeaderModel) updateContent() {
	var metrics Metrics
	if m.metricsManager != nil {
		metrics = m.metricsManager.GetMetrics()
	}
	m.content = m.buildEnhancedHeader(metrics)
}

func (m *HeaderModel) updateContentFromManager() {
	if m.metricsManager != nil {
		metrics := m.metricsManager.GetMetrics()
		m.content = m.buildEnhancedHeader(metrics)
	}
}

func (m *HeaderModel) RefreshMetrics() {
	if m.metricsManager != nil {
		m.updateContent()
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
