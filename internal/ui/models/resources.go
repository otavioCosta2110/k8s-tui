package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"
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
			listItems = append(listItems, components.NewItem(icon+" "+resourceType, ""))
		} else {
			listItems = append(listItems, components.NewItem(resourceType, ""))
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
