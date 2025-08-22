package models

import (
	"fmt"
	"time"

	"otaviocosta2110/k8s-tui/internal/k8s"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	"otaviocosta2110/k8s-tui/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type ResourceData interface {
	GetName() string
	GetNamespace() string
	GetColumns() table.Row
}

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
	resourceType    k8s.ResourceType
	resourceData    []ResourceData
	loading         bool
	err             error
	refreshInterval time.Duration
	config          ResourceConfig
}

func NewGenericResourceModel(k k8s.Client, namespace string, config ResourceConfig) *GenericResourceModel {
	return &GenericResourceModel{
		namespace:       namespace,
		k8sClient:       &k,
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

		if len(checked) == 0 {
			return nil
		}

		utils.WriteString("log", fmt.Sprintf("Deleting %d", checked[0]))
		for _, idx := range checked {
			utils.WriteString("log", fmt.Sprintf("Deleting %d", idx))
			if idx < len(g.resourceData) {
				resource := g.resourceData[idx]
				utils.WriteString("log", fmt.Sprintf("Deleting %s: %s/%s", g.resourceType, resource.GetNamespace(), resource.GetName()))
				_ = g.deleteResource(resource)
			}

			tableModel.Refresh()
			return nil
		}
		return nil
	}
}

func (g *GenericResourceModel) deleteResource(resource ResourceData) error {
	utils.WriteString("log", fmt.Sprintf("%s, %s, %s", g.resourceType, resource.GetNamespace(), resource.GetName()))
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
