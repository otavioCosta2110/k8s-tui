package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type nodesModel struct {
	*GenericResourceModel
	nodesInfo []k8s.NodeInfo
}

func NewNodes(k k8s.Client, namespace string) (*nodesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeNode,
		Title:           customstyles.ResourceIcons["Nodes"] + " Nodes in cluster",
		ColumnWidths:    []float64{0.5, 0.3, 0.5, 0.3, 0.3, 0.3, 0.3, 0.5},
		RefreshInterval: 10 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAME", 0),
			components.NewColumn("STATUS", 0),
			components.NewColumn("ROLES", 0),
			components.NewColumn("VERSION", 0),
			components.NewColumn("CPU", 0),
			components.NewColumn("MEMORY", 0),
			components.NewColumn("PODS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, "", config)

	model := &nodesModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (n *nodesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.k8sClient = k

	if err := n.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		nodeDetails, err := NewNodeDetails(*k, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: nodeDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := n.fetchData(); err != nil {
			return nil, err
		}
		return n.dataToRows(), nil
	}

	tableModel := ui.NewTable(n.config.Columns, n.config.ColumnWidths, n.dataToRows(), n.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": n.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, n.refreshInterval, n.k8sClient, "Nodes"), nil
}

func (n *nodesModel) fetchData() error {
	var nodeInfo []k8s.NodeInfo
	var err error

	// Always use plugin API - resources should never bypass the plugin system
	nodeInfo, err = n.pluginAPI.GetNodes()

	if err != nil {
		return fmt.Errorf("failed to fetch nodes: %v", err)
	}
	n.nodesInfo = nodeInfo

	n.resourceData = make([]types.ResourceData, len(nodeInfo))
	for idx, node := range nodeInfo {
		n.resourceData[idx] = NodeData{&node}
	}

	return nil
}
