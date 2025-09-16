package models

import (
	"fmt"
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuickNavModel struct {
	width     int
	height    int
	kube      k8s.Client
	namespace string
}

func NewQuickNavModel(k k8s.Client, namespace string) QuickNavModel {
	return QuickNavModel{
		kube:      k,
		namespace: namespace,
	}
}

func (m QuickNavModel) Init() tea.Cmd {
	return nil
}

func (m QuickNavModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			// Return to previous screen instead of quitting
			return m, nil
		default:
			// Check if the pressed key corresponds to a resource
			resourceType := m.getResourceTypeFromKey(msg.String())
			if resourceType != "" {
				return m, m.navigateToResource(resourceType)
			}
		}
	}
	return m, nil
}

func (m QuickNavModel) getResourceTypeFromKey(key string) string {
	resourceMap := map[string]string{
		"p": "Pods",
		"d": "Deployments",
		"s": "Services",
		"i": "Ingresses",
		"c": "ConfigMaps",
		"e": "Secrets",
		"a": "ServiceAccounts",
		"r": "ReplicaSets",
		"n": "Nodes",
		"j": "Jobs",
		"k": "CronJobs",
		"m": "DaemonSets",
		"t": "StatefulSets",
		"l": "ResourceList",
	}

	if resourceType, exists := resourceMap[key]; exists {
		return resourceType
	}
	return ""
}

func (m QuickNavModel) navigateToResource(resourceType string) tea.Cmd {
	return func() tea.Msg {
		if resourceType == "ResourceList" {
			resourceScreen := NewResource(m.kube, m.namespace)
			resourceComponent := resourceScreen.InitComponent(m.kube)
			return components.NavigateMsg{
				NewScreen:  resourceComponent,
				Breadcrumb: "Resource List",
			}
		}

		resourceList, err := NewResourceList(m.kube, m.namespace, resourceType).InitComponent(m.kube)
		if err != nil {
			return components.NavigateMsg{
				Error: err,
			}
		}

		return components.NavigateMsg{
			NewScreen:  resourceList,
			Breadcrumb: resourceType,
		}
	}
}

func (m QuickNavModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.Blue)).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Render("Quick Navigation - Press 'g' + key")

	// Define the key mappings
	mappings := []struct {
		key   string
		desc  string
		group string
	}{
		{"p", "Pods", "Workloads"},
		{"d", "Deployments", "Workloads"},
		{"r", "ReplicaSets", "Workloads"},
		{"j", "Jobs", "Workloads"},
		{"k", "CronJobs", "Workloads"},
		{"m", "DaemonSets", "Workloads"},
		{"t", "StatefulSets", "Workloads"},
		{"s", "Services", "Networking"},
		{"i", "Ingresses", "Networking"},
		{"c", "ConfigMaps", "Configuration"},
		{"e", "Secrets", "Configuration"},
		{"a", "ServiceAccounts", "Configuration"},
		{"n", "Nodes", "Infrastructure"},
		{"l", "Resource List", "Navigation"},
	}

	// Group mappings by category
	groups := make(map[string][]string)
	for _, mapping := range mappings {
		keyDesc := fmt.Sprintf("%s → %s", mapping.key, mapping.desc)
		groups[mapping.group] = append(groups[mapping.group], keyDesc)
	}

	// Calculate column layout
	screenWidth := global.ScreenWidth
	numColumns := 3
	columnWidth := screenWidth / numColumns
	if columnWidth < 20 {
		numColumns = 2
		columnWidth = screenWidth / numColumns
	}
	if columnWidth < 15 {
		numColumns = 1
		columnWidth = screenWidth
	}

	// Create columns
	groupOrder := []string{"Workloads", "Networking", "Configuration", "Infrastructure", "Navigation"}
	var columns []string

	colIndex := 0
	var currentColumn strings.Builder

	for _, groupName := range groupOrder {
		if items, exists := groups[groupName]; exists {
			groupTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(customstyles.Purple)).
				Render(fmt.Sprintf("%s", groupName))

			currentColumn.WriteString(groupTitle)
			currentColumn.WriteString("\n")

			for _, item := range items {
				currentColumn.WriteString(item)
				currentColumn.WriteString("\n")
			}
			currentColumn.WriteString("\n")

			colIndex++
			if colIndex >= numColumns {
				columns = append(columns, currentColumn.String())
				currentColumn.Reset()
				colIndex = 0
			}
		}
	}

	// Add remaining column if any
	if currentColumn.Len() > 0 {
		columns = append(columns, currentColumn.String())
	}

	// Pad columns to equal width
	for i := range columns {
		columns[i] = lipgloss.NewStyle().
			Width(columnWidth).
			Height(global.ScreenHeight - 4). // Leave space for title and footer
			Render(columns[i])
	}

	// Join columns horizontally
	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, columns...))

	footer := lipgloss.NewStyle().
		Faint(true).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Render("Press a key to navigate • esc/q to cancel")

	// Use full screen
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
		"",
		footer,
	)

	return lipgloss.NewStyle().
		Width(screenWidth).
		Height(global.ScreenHeight + 1).
		Align(lipgloss.Center).
		Render(fullContent)
}
