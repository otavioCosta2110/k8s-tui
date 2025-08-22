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
	list            []string
	namespace       string
	k8sClient       *k8s.Client
	replicasetsInfo []k8s.ReplicaSetInfo
	loading         bool
	err             error
	refreshInterval time.Duration
}

func NewReplicaSets(k k8s.Client, namespace string) (*replicasetsModel, error) {
	replicasets, err := k8s.FetchReplicaSetList(k, namespace)
	if err != nil {
		return nil, err
	}

	return &replicasetsModel{
		list:            replicasets,
		namespace:       namespace,
		k8sClient:       &k,
		loading:         false,
		err:             nil,
		refreshInterval: 5 * time.Second,
	}, nil
}

func (r *replicasetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	r.k8sClient = k
	replicasetInfo, err := k8s.GetReplicaSetsTableData(*k, r.namespace)
	if err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		replicaset := k8s.NewReplicaSet(selected, r.namespace, *k)
		p, err := replicaset.GetPods()
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		pods, err := NewPods(*k, r.namespace, p)
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
			NewScreen: podsComponent,
		}
	}

	columns := []table.Column{
		components.NewColumn("NAMESPACE", 0),
		components.NewColumn("NAME", 0),
		components.NewColumn("DESIRED", 0),
		components.NewColumn("CURRENT", 0),
		components.NewColumn("READY", 0),
		components.NewColumn("AGE", 0),
	}

	colPercent := []float64{0.15, 0.25, 0.12, 0.12, 0.15, 0.15}

	rows := r.replicasetsToRows(replicasetInfo)

	fetchFunc := func() ([]table.Row, error) {
		rss, err := r.fetchReplicaSets(r.k8sClient)
		if err != nil {
			return nil, err
		}

		newRows := r.replicasetsToRows(rss)
		return newRows, nil
	}

	tableModel := ui.NewTable(columns, colPercent, rows, "ReplicaSets in "+r.namespace, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": func() tea.Cmd {
			checked := tableModel.GetCheckedItems()
			var lastError error
			for _, idx := range checked {
				if idx < len(r.replicasetsInfo) {
					replicaset := r.replicasetsInfo[idx]
					err := k8s.DeleteReplicaSet(*r.k8sClient, replicaset.Namespace, replicaset.Name)
					if err != nil {
						lastError = err
					}
				}
			}
			tableModel.Refresh()
			if lastError != nil {
				return func() tea.Msg {
					return ErrorModel{error: lastError}
				}
			}
			return nil
		},
	}
	tableModel.SetUpdateActions(actions)

	return &autoRefreshModel{
		inner:           tableModel,
		refreshInterval: r.refreshInterval,
		k8sClient:       r.k8sClient,
	}, nil
}

func (r *replicasetsModel) fetchReplicaSets(client *k8s.Client) ([]k8s.ReplicaSetInfo, error) {
	replicasets, err := k8s.GetReplicaSetsTableData(*client, r.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch replicasets: %v", err)
	}
	r.replicasetsInfo = replicasets
	return replicasets, nil
}

func (r *replicasetsModel) replicasetsToRows(replicasetInfo []k8s.ReplicaSetInfo) []table.Row {
	rows := []table.Row{}
	for _, replicaset := range replicasetInfo {
		rows = append(rows, table.Row{
			replicaset.Namespace,
			replicaset.Name,
			replicaset.Desired,
			replicaset.Current,
			replicaset.Ready,
			replicaset.Age,
		})
	}
	return rows
}
