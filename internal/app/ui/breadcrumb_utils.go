package ui

import (
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
)

func (m *AppModel) getBreadcrumbTrail() string {
	if len(m.breadcrumbTrail) > 0 {
		prefix := []string{"config", "test-namespace"}
		fullTrail := append(prefix, m.breadcrumbTrail...)
		return strings.Join(fullTrail, " > ")
	}

	if m.tabManager != nil {
		if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
			return strings.Join(activeTab.Breadcrumb, " > ")
		}
	}
	return ""
}

func (m *AppModel) isCurrentScreenResourceType(resourceType string) bool {
	if len(m.breadcrumbTrail) > 0 {
		currentCrumb := m.breadcrumbTrail[len(m.breadcrumbTrail)-1]
		if currentCrumb == "Resource List" && resourceType == "ResourceList" {
			return true
		}
		return currentCrumb == resourceType
	}

	if m.tabManager != nil {
		if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
			currentCrumb := activeTab.Breadcrumb[len(activeTab.Breadcrumb)-1]
			return currentCrumb == resourceType
		}
	}
	return false
}

func (m *AppModel) initializeInitialBreadcrumb(listModel interface{}) {
	if list, ok := listModel.(*components.ListModel); ok {
		if list.List.Title == "Resource Types" {
			m.breadcrumbTrail = []string{"Resource List"}
		} else {
			m.breadcrumbTrail = []string{}
		}
	} else {
		m.breadcrumbTrail = []string{}
	}
}
