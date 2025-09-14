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

type statefulsetsModel struct {
	*GenericResourceModel
	statefulsetsInfo []k8s.StatefulSetInfo
}

func NewStatefulSets(k k8s.Client, namespace string) (*statefulsetsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeStatefulSet,
		Title:           "StatefulSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.30, 0.26},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &statefulsetsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (ss *statefulsetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ss.k8sClient = k

	if err := ss.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		statefulsetDetails, err := NewStatefulSetDetails(*k, ss.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: statefulsetDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := ss.fetchData(); err != nil {
			return nil, err
		}
		return ss.dataToRows(), nil
	}

	tableModel := ui.NewTable(ss.config.Columns, ss.config.ColumnWidths, ss.dataToRows(), ss.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": ss.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, ss.refreshInterval, ss.k8sClient, "StatefulSets"), nil
}

func (ss *statefulsetsModel) fetchData() error {
	statefulsetInfo, err := k8s.GetStatefulSetsTableData(*ss.k8sClient, ss.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch statefulsets: %v", err)
	}
	ss.statefulsetsInfo = statefulsetInfo

	ss.resourceData = make([]ResourceData, len(statefulsetInfo))
	for idx, statefulset := range statefulsetInfo {
		ss.resourceData[idx] = StatefulSetData{&statefulset}
	}

	return nil
}

func (ss *statefulsetsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(ss.statefulsetsInfo))
	for idx, statefulset := range ss.statefulsetsInfo {
		rows[idx] = table.Row{
			statefulset.Namespace,
			statefulset.Name,
			statefulset.Ready,
			statefulset.Age,
		}
	}
	return rows
}
