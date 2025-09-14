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

type daemonsetsModel struct {
	*GenericResourceModel
	daemonsetsInfo []k8s.DaemonSetInfo
}

func NewDaemonSets(k k8s.Client, namespace string) (*daemonsetsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeDaemonSet,
		Title:           "DaemonSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.20, 0.10, 0.10, 0.10, 0.10, 0.10, 0.10, 0.10},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("DESIRED", 0),
			components.NewColumn("CURRENT", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("UP-TO-DATE", 0),
			components.NewColumn("AVAILABLE", 0),
			components.NewColumn("NODE SELECTOR", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &daemonsetsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (ds *daemonsetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ds.k8sClient = k

	if err := ds.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		daemonsetDetails, err := NewDaemonSetDetails(*k, ds.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: daemonsetDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := ds.fetchData(); err != nil {
			return nil, err
		}
		return ds.dataToRows(), nil
	}

	tableModel := ui.NewTable(ds.config.Columns, ds.config.ColumnWidths, ds.dataToRows(), ds.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": ds.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, ds.refreshInterval, ds.k8sClient, "DaemonSets"), nil
}

func (ds *daemonsetsModel) fetchData() error {
	daemonsetInfo, err := k8s.GetDaemonSetsTableData(*ds.k8sClient, ds.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch daemonsets: %v", err)
	}
	ds.daemonsetsInfo = daemonsetInfo

	ds.resourceData = make([]ResourceData, len(daemonsetInfo))
	for idx, daemonset := range daemonsetInfo {
		ds.resourceData[idx] = DaemonSetData{&daemonset}
	}

	return nil
}

func (ds *daemonsetsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(ds.daemonsetsInfo))
	for idx, daemonset := range ds.daemonsetsInfo {
		rows[idx] = table.Row{
			daemonset.Namespace,
			daemonset.Name,
			daemonset.Desired,
			daemonset.Current,
			daemonset.Ready,
			daemonset.UpToDate,
			daemonset.Available,
			daemonset.NodeSelector,
			daemonset.Age,
		}
	}
	return rows
}
