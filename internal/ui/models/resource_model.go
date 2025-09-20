package models

import (
	"fmt"
	"time"

	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"

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
	resourceType    k8s.ResourceType
	resourceData    []types.ResourceData
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
