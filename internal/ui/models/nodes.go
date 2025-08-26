package models

import (
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type nodesModel struct {
	*GenericResourceModel
	nodesInfo []k8s.NodeInfo
}

func NewNodes(k k8s.Client) (*nodesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeNode,
		Title:           "Nodes in cluster",
		ColumnWidths:    []float64{0.25, 0.10, 0.12, 0.10, 0.05, 0.07, 0.12, 0.10},
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

	return &autoRefreshModel{
		inner:           tableModel,
		refreshInterval: n.refreshInterval,
		k8sClient:       n.k8sClient,
	}, nil
}

func (n *nodesModel) fetchData() error {
	nodeInfo, err := k8s.GetNodesTableData(*n.k8sClient)
	if err != nil {
		return fmt.Errorf("failed to fetch nodes: %v", err)
	}
	n.nodesInfo = nodeInfo

	n.resourceData = make([]ResourceData, len(nodeInfo))
	for idx, node := range nodeInfo {
		n.resourceData[idx] = NodeData{&node}
	}

	return nil
}

func (n *nodesModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(n.nodesInfo))
	for idx, node := range n.nodesInfo {
		rows[idx] = table.Row{
			node.Name,
			node.Status,
			node.Roles,
			node.Version,
			node.CPU,
			node.Memory,
			node.Pods,
			node.Age,
		}
	}
	return rows
}
