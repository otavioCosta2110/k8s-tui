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

type CloseQuickNavMsg struct{}

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
			return m, func() tea.Msg {
				return CloseQuickNavMsg{}
			}
		default:
			resourceType := m.getResourceTypeFromKey(msg.String())
			if resourceType != "" {
				return m, m.navigateToResource(resourceType)
			}
		}
	}
	return m, nil
}

func (m QuickNavModel) getResourceTypeFromKey(key string) string {
	mappings := resourceFactory.GetQuickNavMappings()

	mappings["l"] = "ResourceList"

	if resourceType, exists := mappings[key]; exists {
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
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("Quick Navigation - Press 'g' + key")

	sortedMappings := resourceFactory.GetSortedQuickNavMappings()

	mappings := []struct {
		key   string
		desc  string
		group string
	}{}

	resourceListIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(customstyles.ResourceIcons["ResourceList"])

	resourceListSpace := lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(" ")

	resourceListText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("Resource List")

	resourceListDesc := resourceListIcon + resourceListSpace + resourceListText

	mappings = append(mappings, struct {
		key   string
		desc  string
		group string
	}{"l", resourceListDesc, "Navigation"})

	resourceIcons := customstyles.ResourceIcons

	for _, mapping := range sortedMappings {
		if metadata, exists := resourceFactory.GetResourceMetadata(mapping.ResourceType); exists {
			icon := "📦"
			if resourceIcon, exists := resourceIcons[metadata.Name]; exists {
				icon = resourceIcon
			}

			iconPart := lipgloss.NewStyle().
				Foreground(lipgloss.Color(customstyles.TextColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(icon)

			spacePart := lipgloss.NewStyle().
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(" ")

			namePart := lipgloss.NewStyle().
				Foreground(lipgloss.Color(customstyles.TextColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(metadata.Name)

			descWithIcon := iconPart + spacePart + namePart
			mappings = append(mappings, struct {
				key   string
				desc  string
				group string
			}{mapping.Key, descWithIcon, metadata.Category})
		}
	}

	groups := make(map[string][]string)
	for _, mapping := range mappings {
		keyPart := lipgloss.NewStyle().
			Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(mapping.key)

		arrowPart := lipgloss.NewStyle().
			Foreground(lipgloss.Color(customstyles.TextColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(" → ")

		descPart := lipgloss.NewStyle().
			Foreground(lipgloss.Color(customstyles.TextColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(mapping.desc)

		keyDesc := keyPart + arrowPart + descPart
		groups[mapping.group] = append(groups[mapping.group], keyDesc)
	}

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

	groupOrder := []string{"Workloads", "Networking", "Configuration", "Infrastructure", "Navigation"}
	var columns []string

	colIndex := 0
	var currentColumn strings.Builder

	for _, groupName := range groupOrder {
		if items, exists := groups[groupName]; exists {
			groupIcon := "󰉋"
			if icon, exists := customstyles.ResourceIcons[groupName]; exists {
				groupIcon = icon
			}

			groupTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(customstyles.HeaderColor)).
				Width(columnWidth).
				Underline(true).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(fmt.Sprintf("%s %s", groupIcon, groupName))

			currentColumn.WriteString(groupTitle)
			currentColumn.WriteString("\n")

			for _, item := range items {
				itemLine := lipgloss.NewStyle().
					Width(columnWidth).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					PaddingLeft(2).
					Render(item)
				currentColumn.WriteString(itemLine)
				currentColumn.WriteString("\n")
			}

			spacer := lipgloss.NewStyle().
				Width(columnWidth).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render("")

			currentColumn.WriteString(spacer)
			currentColumn.WriteString("\n")

			colIndex++
			if colIndex >= numColumns {
				columns = append(columns, currentColumn.String())
				currentColumn.Reset()
				colIndex = 0
			}
		}
	}

	if currentColumn.Len() > 0 {
		columns = append(columns, currentColumn.String())
	}

	contentHeight := global.ScreenHeight + global.HeaderSize + 1

	for i := range columns {
		columns[i] = lipgloss.NewStyle().
			Width(columnWidth).
			Height(contentHeight).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(columns[i])
	}

	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Height(contentHeight).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, columns...))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Align(lipgloss.Center).
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("Press a key to navigate • esc/q to cancel")

	spacer1 := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("")

	spacer2 := lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("")

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		spacer1,
		content,
		spacer2,
		footer,
	)

	return lipgloss.NewStyle().
		Width(global.ScreenWidth).
		Height(global.ScreenHeight / 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(fullContent)
}
