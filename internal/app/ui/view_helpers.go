package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

func (m *AppModel) renderContent() string {
	currentView := m.tabManager.View()

	height := styles.ScreenHeight

	content := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(customstyles.BorderColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(currentView)

	return content
}

func (m *AppModel) renderBreadcrumb() string {
	if m.tabManager == nil {
		return ""
	}

	if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
		var breadcrumbParts []string
		breadcrumbEnd := min(activeTab.CurrentIndex+1, len(activeTab.Breadcrumb))
		for i := range breadcrumbEnd {
			crumb := activeTab.Breadcrumb[i]
			if i == activeTab.CurrentIndex && activeTab.CurrentIndex < len(activeTab.Breadcrumb) {
				breadcrumbParts = append(breadcrumbParts, lipgloss.NewStyle().
					Foreground(lipgloss.Color(customstyles.AccentColor)).
					Bold(true).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					Render(crumb))
			} else {
				breadcrumbParts = append(breadcrumbParts, lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					Render(crumb))
			}
		}
		breadCrumbArrow := lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" > ")
		breadcrumbStr := strings.Join(breadcrumbParts, breadCrumbArrow)
		breadcrumbView := lipgloss.NewStyle().
			Width(styles.ScreenWidth + styles.Margin).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(breadcrumbStr)
		return breadcrumbView
	}
	return ""
}

func (m *AppModel) renderFooter() string {
	footerInjections := m.uiInjector.RenderInjections("footer")
	if footerInjections != "" {
		return lipgloss.NewStyle().
			Width(styles.ScreenWidth).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1).
			Render(footerInjections)
	}
	return ""
}

func (m *AppModel) renderWithHeader(content, breadcrumbView, footerView string) string {
	headerView := m.header.View()

	if breadcrumbView != "" && footerView != "" {
		return lipgloss.JoinVertical(lipgloss.Top, headerView, content, breadcrumbView, footerView)
	} else if breadcrumbView != "" {
		return lipgloss.JoinVertical(lipgloss.Top, headerView, content, breadcrumbView)
	} else if footerView != "" {
		return lipgloss.JoinVertical(lipgloss.Top, headerView, content, footerView)
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, headerView, content)
	}
}

func (m *AppModel) renderWithoutHeader(content, breadcrumbView string) string {
	if breadcrumbView != "" {
		return lipgloss.JoinVertical(lipgloss.Top,
			lipgloss.NewStyle().
				Width(styles.ScreenWidth).
				Height(styles.ScreenHeight+styles.HeaderSize).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(customstyles.BorderColor)).
				BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(content),
			breadcrumbView)
	} else {
		return lipgloss.NewStyle().
			Width(styles.ScreenWidth).
			Height(styles.ScreenHeight + styles.HeaderSize).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(customstyles.BorderColor)).
			BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(content)
	}
}
