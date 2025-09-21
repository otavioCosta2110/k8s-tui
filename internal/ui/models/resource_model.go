package models

import (
	"fmt"
	"time"

	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"

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
	// Try to get plugin API from global plugin manager
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
	// Always use plugin API - resources should never bypass the plugin system
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
		// Use plugin API for custom resources
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

func (g *GenericResourceModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(g.resourceData))
	for i, rd := range g.resourceData {
		rows[i] = rd.GetColumns()
	}
	return rows
}
