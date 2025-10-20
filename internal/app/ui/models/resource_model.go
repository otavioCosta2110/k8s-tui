package models

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type ResourceConfig struct {
	ResourceType    k8s.ResourceType
	Title           string
	ColumnWidths    []float64
	RefreshInterval time.Duration
	Columns         []table.Column
}

type GenericResourceModel struct {
	namespace       string
	k8sClient       *k8s.Client
	pluginAPI       plugins.PluginAPI
	resourceType    k8s.ResourceType
	resourceData    []types.ResourceData
	loading         bool
	err             error
	refreshInterval time.Duration
	config          ResourceConfig
}

func NewGenericResourceModel(k k8s.Client, namespace string, config ResourceConfig) *GenericResourceModel {
	var pluginAPI plugins.PluginAPI
	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		pluginAPI = pm.GetAPI()
		pluginAPI.SetClient(k)
	}

	return &GenericResourceModel{
		namespace:       namespace,
		k8sClient:       &k,
		pluginAPI:       pluginAPI,
		resourceType:    config.ResourceType,
		loading:         false,
		err:             nil,
		refreshInterval: config.RefreshInterval,
		config:          config,
	}
}

func (g *GenericResourceModel) createDeleteAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		checked := tableModel.GetCheckedItems()
		checkedStr := make([]string, len(checked))
		for i, v := range checked {
			checkedStr[i] = fmt.Sprintf("%d", v)
		}

		if len(checked) == 0 {
			return nil
		}

		for _, idx := range checked {
			if idx < len(g.resourceData) {
				resource := g.resourceData[idx]
				_ = g.deleteResource(resource)
			}
		}

		tableModel.ClearCheckedItems()
		tableModel.Refresh()
		return nil
	}
}

func (g *GenericResourceModel) deleteResource(resource types.ResourceData) error {
	var err error
	switch g.resourceType {
	case k8s.ResourceTypePod:
		err = g.pluginAPI.DeletePod(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeService:
		err = g.pluginAPI.DeleteService(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeDeployment:
		err = g.pluginAPI.DeleteDeployment(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeConfigMap:
		err = g.pluginAPI.DeleteConfigMap(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeSecret:
		err = g.pluginAPI.DeleteSecret(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeIngress:
		err = g.pluginAPI.DeleteIngress(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeJob:
		err = g.pluginAPI.DeleteJob(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeCronJob:
		err = g.pluginAPI.DeleteCronJob(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeDaemonSet:
		err = g.pluginAPI.DeleteDaemonSet(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeStatefulSet:
		err = g.pluginAPI.DeleteStatefulSet(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeReplicaSet:
		err = g.pluginAPI.DeleteReplicaSet(resource.GetNamespace(), resource.GetName())
	case k8s.ResourceTypeServiceAccount:
		err = g.pluginAPI.DeleteServiceAccount(resource.GetNamespace(), resource.GetName())
	default:
		err = k8s.DeleteResource(*g.k8sClient, g.resourceType, resource.GetNamespace(), resource.GetName())
	}
	if err != nil {
		return fmt.Errorf("failed to delete resource %s/%s: %v", resource.GetNamespace(), resource.GetName(), err)
	}
	return nil
}

func (g *GenericResourceModel) deleteResourceK8s(resource types.ResourceData) error {
	err := k8s.DeleteResource(*g.k8sClient, g.resourceType, resource.GetNamespace(), resource.GetName())
	if err != nil {
		return fmt.Errorf("failed to delete resource %s/%s: %v", resource.GetNamespace(), resource.GetName(), err)
	}
	return fmt.Errorf("deleteResource not implemented for %s", g.resourceType)
}

func (g *GenericResourceModel) GetResourceType() k8s.ResourceType {
	return g.resourceType
}

func (g *GenericResourceModel) GetNamespace() string {
	return g.namespace
}

func (g *GenericResourceModel) Help() (string, string) {
	resourceTypeToDisplay := map[k8s.ResourceType]string{
		k8s.ResourceTypePod:                   "Pods",
		k8s.ResourceTypeDeployment:            "Deployments",
		k8s.ResourceTypeService:               "Services",
		k8s.ResourceTypeIngress:               "Ingresses",
		k8s.ResourceTypeConfigMap:             "ConfigMaps",
		k8s.ResourceTypeSecret:                "Secrets",
		k8s.ResourceTypeReplicaSet:            "ReplicaSets",
		k8s.ResourceTypeJob:                   "Jobs",
		k8s.ResourceTypeCronJob:               "CronJobs",
		k8s.ResourceTypeDaemonSet:             "DaemonSets",
		k8s.ResourceTypeStatefulSet:           "StatefulSets",
		k8s.ResourceTypeNode:                  "Nodes",
		k8s.ResourceTypePersistentVolume:      "PersistentVolumes",
		k8s.ResourceTypePersistentVolumeClaim: "PersistentVolumeClaims",
		k8s.ResourceTypeServiceAccount:        "ServiceAccounts",
	}

	displayName, exists := resourceTypeToDisplay[g.resourceType]
	if !exists {
		displayName = string(g.resourceType)
	}

	return fmt.Sprintf("%s Help", displayName), fmt.Sprintf(`Help for %s

Key Bindings:
• ↑/↓/j/k: Navigate items
• enter: View details
• d: Delete selected items (if supported)
• r: Refresh
• /: Search
• esc: Go back

For more specific help, check the resource documentation.`, displayName)
}

func (g *GenericResourceModel) dataToRows() []table.Row {
	resourceTypeToDisplay := map[k8s.ResourceType]string{
		k8s.ResourceTypePod:                   "Pods",
		k8s.ResourceTypeDeployment:            "Deployments",
		k8s.ResourceTypeService:               "Services",
		k8s.ResourceTypeIngress:               "Ingresses",
		k8s.ResourceTypeConfigMap:             "ConfigMaps",
		k8s.ResourceTypeSecret:                "Secrets",
		k8s.ResourceTypeReplicaSet:            "ReplicaSets",
		k8s.ResourceTypeJob:                   "Jobs",
		k8s.ResourceTypeCronJob:               "CronJobs",
		k8s.ResourceTypeDaemonSet:             "DaemonSets",
		k8s.ResourceTypeStatefulSet:           "StatefulSets",
		k8s.ResourceTypeNode:                  "Nodes",
		k8s.ResourceTypePersistentVolume:      "PersistentVolumes",
		k8s.ResourceTypePersistentVolumeClaim: "PersistentVolumeClaims",
		k8s.ResourceTypeServiceAccount:        "ServiceAccounts",
	}

	rows := make([]table.Row, len(g.resourceData))
	for i, rd := range g.resourceData {
		row := rd.GetColumns()
		if displayName, exists := resourceTypeToDisplay[g.config.ResourceType]; exists {
			if icon, iconExists := customstyles.ResourceIcons[displayName]; iconExists {
				nameIndex := 1
				if g.config.ResourceType == k8s.ResourceTypeNode {
					nameIndex = 0
				}
				if len(row) > nameIndex {
					isHealthy := g.isResourceHealthy(rd)
					displayIcon := icon
					if !isHealthy {
						displayIcon = "✗"
					}
					row[nameIndex] = displayIcon + " " + row[nameIndex]
				}
			}
		}
		rows[i] = row
	}
	return rows
}

func (g *GenericResourceModel) isResourceHealthy(rd types.ResourceData) bool {
	switch g.config.ResourceType {
	case k8s.ResourceTypePod:
		if podData, ok := rd.(PodData); ok {
			return podData.Status == "Running"
		}
	case k8s.ResourceTypeDeployment:
		if depData, ok := rd.(DeploymentData); ok {
			if parts := strings.Split(depData.Ready, "/"); len(parts) == 2 {
				return parts[0] == parts[1]
			}
		}
	}
	return true
}
