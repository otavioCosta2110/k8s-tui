package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Resource struct {
	kube         k8s.Client
	namespace    string
	resourceType string
}

func NewResource(k k8s.Client, namespace string) Resource {
	return Resource{
		kube:      k,
		namespace: namespace,
	}
}

func (r Resource) InitComponent(k k8s.Client) tea.Model {
	resourceTypes := resourceFactory.GetValidResourceTypes()

	var listItems []components.ListItem
	for _, resourceType := range resourceTypes {
		if icon, exists := customstyles.ResourceIcons[resourceType]; exists {
			logger.Debug(fmt.Sprintf("🔌 UI: Using built-in icon for %s: %s", resourceType, icon))
			listItems = append(listItems, components.NewItem(icon+" "+resourceType, ""))
		} else {
			icon := ""
			if pm := plugins.GetGlobalPluginManager(); pm != nil {
				logger.Debug(fmt.Sprintf("🔌 UI: Plugin manager available for icon lookup, checking %d custom types", len(pm.GetRegistry().GetCustomResourceTypes())))
				for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
					logger.Debug(fmt.Sprintf("🔌 UI: Checking custom resource %s for icon", rt.Name))
					if rt.Name == resourceType {
						icon = rt.Icon
						logger.Info(fmt.Sprintf("🔌 UI: Found plugin icon for %s: '%s'", resourceType, icon))
						break
					}
				}
			} else {
				logger.Warn("🔌 UI: Plugin manager not available for icon lookup")
			}

			if icon != "" {
				listItems = append(listItems, components.NewItem(icon+" "+resourceType, ""))
			} else {
				logger.Debug(fmt.Sprintf("🔌 UI: No icon found for %s", resourceType))
				listItems = append(listItems, components.NewItem(resourceType, ""))
			}
		}
	}

	onSelect := func(selected string) tea.Msg {
		resourceType := selected

		for _, icon := range customstyles.ResourceIcons {
			if strings.HasPrefix(selected, icon+" ") {
				resourceType = strings.TrimPrefix(selected, icon+" ")
				break
			}
		}

		if resourceType == selected { 
			if pm := plugins.GetGlobalPluginManager(); pm != nil {
				for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
					iconWithSpace := rt.Icon + " "
					if strings.HasPrefix(selected, iconWithSpace) {
						resourceType = strings.TrimPrefix(selected, iconWithSpace)
						logger.Debug(fmt.Sprintf("🔌 UI: Stripped plugin icon '%s' from '%s', result: '%s'", rt.Icon, selected, resourceType))
						break
					}
				}
			}
		}

		r.resourceType = resourceType
		newResourceList, err := NewResourceList(r.kube, r.namespace, resourceType).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error: err,
			}
		}
		return components.NavigateMsg{
			NewScreen:  newResourceList,
			Breadcrumb: resourceType,
		}
	}

	return components.NewListWithItems(listItems, customstyles.ResourceIcons["ResourceList"]+" Resource Types", onSelect)
}
