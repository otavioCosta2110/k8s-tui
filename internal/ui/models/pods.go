package models

import (
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type podsModel struct {
	namespace string
	k8sClient *k8s.Client
	podsInfo  []k8s.PodInfo
	loading   bool
	err       error
}

func NewPods(k k8s.Client, namespace string, pods []k8s.PodInfo) (*podsModel, error) {
	if len(pods) == 0 {
		var err error
		pods, err = k8s.FetchPods(k, namespace, "")
		if err != nil {
			return nil, err
		}
	}
	return &podsModel{
		namespace: namespace,
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}, nil
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	podsInfo, err := k8s.FetchPods(*k, p.namespace, "")
	if err != nil {
		return nil, err
	}
	p.podsInfo = podsInfo

	onSelect := func(selected string) tea.Msg {
		podDetails, err := NewPodDetails(*k, p.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: podDetails,
		}
	}

	columns := []table.Column{
		components.NewColumn("NAMESPACE", 0),
		components.NewColumn("NAME", 0),
		components.NewColumn("READY", 0),
		components.NewColumn("STATUS", 0),
		components.NewColumn("RESTARTS", 0),
		components.NewColumn("AGE", 0),
	}

	colPercent := []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.15}

	rows := []table.Row{}
	for _, pod := range p.podsInfo {
		rows = append(rows, table.Row{
			pod.Namespace,
			pod.Name,
			pod.Ready,
			pod.Status,
			fmt.Sprintf("%d", pod.Restarts),
			pod.Age,
		})
	}

	return ui.NewTable(columns, colPercent, rows, "Pods in "+p.namespace, onSelect, 1), nil
}
