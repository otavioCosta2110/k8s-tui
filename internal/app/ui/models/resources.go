package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
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

func (r Resource) InitComponent(k *k8s.Client) tea.Model {
	resourceTypes := resourceFactory.GetValidResourceTypes()

	var listItems []components.ListItem
	for _, resourceType := range resourceTypes {
		if icon, exists := customstyles.ResourceIcons[resourceType]; exists {
			listItems = append(listItems, components.NewItem(icon+" "+resourceType, ""))
		} else {
			icon := ""
			if pm := plugins.GetGlobalPluginManager(); pm != nil {
				for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
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
						break
					}
				}
			}
		}

		r.resourceType = resourceType
		resourceList := NewResourceList(r.kube, r.namespace, resourceType)
		newResourceList, err := resourceList.InitComponent(*k)
		if err != nil {
			return components.NavigateMsg{
				Error: err,
			}
		}
		return components.NavigateMsg{
			NewScreen:     newResourceList,
			ResourceModel: resourceList,
			Breadcrumb:    resourceType,
		}
	}

	return components.NewListWithItems(listItems, customstyles.ResourceIcons["ResourceList"]+" Resource Types", onSelect)
}

func (r Resource) Help() (string, string) {
	return "Resource List Help", `Resource List shows available Kubernetes resources.

Key Bindings:
• ↑/↓/j/k: Navigate resources
• enter: Select resource type
• /: Search resources
• esc: Go back

Resource Categories:
• Workloads: Pods, Deployments, Jobs, etc.
• Networking: Services, Ingresses
• Configuration: ConfigMaps, Secrets
• Infrastructure: Nodes

Quick Navigation:
• p: Pods
• d: Deployments
• s: Services
• i: Ingresses
• c: ConfigMaps
• e: Secrets
• n: Nodes
• j: Jobs
• k: CronJobs
• m: DaemonSets
• t: StatefulSets
• r: ReplicaSets
• a: ServiceAccounts`
}
