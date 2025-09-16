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

type replicasetsModel struct {
	*GenericResourceModel
	replicasetsInfo []k8s.ReplicaSetInfo
}

func NewReplicaSets(k k8s.Client, namespace string) (*replicasetsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeReplicaSet,
		Title:           "ReplicaSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.12, 0.12, 0.15, 0.13},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("DESIRED", 0),
			components.NewColumn("CURRENT", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &replicasetsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (r *replicasetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	r.k8sClient = k

	if err := r.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		replicaset := k8s.NewReplicaSet(selected, r.namespace, *k)
		// p, err := replicaset.GetPods()
		// if err != nil {
		// 	return components.NavigateMsg{
		// 		Error:   err,
		// 		Cluster: *k,
		// 	}
		// }
		selector := fmt.Sprintf("app=%s", replicaset.Name)
		pods, err := NewPods(*k, r.namespace, selector)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		podsComponent, err := pods.InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		return components.NavigateMsg{
			NewScreen:  podsComponent,
			Breadcrumb: "Pods",
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := r.fetchData(); err != nil {
			return nil, err
		}
		return r.dataToRows(), nil
	}

	tableModel := ui.NewTable(r.config.Columns, r.config.ColumnWidths, r.dataToRows(), r.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": r.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, r.refreshInterval, r.k8sClient, "ReplicaSets"), nil
}

func (r *replicasetsModel) fetchData() error {
	replicasetInfo, err := k8s.GetReplicaSetsTableData(*r.k8sClient, r.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch replicasets: %v", err)
	}
	r.replicasetsInfo = replicasetInfo

	r.resourceData = make([]ResourceData, len(replicasetInfo))
	for i, replicaset := range replicasetInfo {
		r.resourceData[i] = ReplicaSetData{&replicaset}
	}

	return nil
}

func (r *replicasetsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(r.replicasetsInfo))
	for i, replicaset := range r.replicasetsInfo {
		rows[i] = table.Row{
			replicaset.Namespace,
			replicaset.Name,
			replicaset.Desired,
			replicaset.Current,
			replicaset.Ready,
			replicaset.Age,
		}
	}
	return rows
}
