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

type podsModel struct {
	*GenericResourceModel
	podsInfo []k8s.PodInfo
}

func NewPods(k k8s.Client, namespace string, pods []k8s.PodInfo) (*podsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypePod,
		Title:           "Pods in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.15},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("STATUS", 0),
			components.NewColumn("RESTARTS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &podsModel{
		GenericResourceModel: genericModel,
		podsInfo:             pods,
	}

	return model, nil
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k

	if err := p.fetchData(); err != nil {
		return nil, err
	}

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

	fetchFunc := func() ([]table.Row, error) {
		if err := p.fetchData(); err != nil {
			return nil, err
		}
		return p.dataToRows(), nil
	}

	tableModel := ui.NewTable(p.config.Columns, p.config.ColumnWidths, p.dataToRows(), p.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": p.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return &autoRefreshModel{
		inner:           tableModel,
		refreshInterval: p.refreshInterval,
		k8sClient:       p.k8sClient,
	}, nil
}

func (p *podsModel) fetchData() error {
	podsInfo, err := k8s.FetchPods(*p.k8sClient, p.namespace, "")
	if err != nil {
		return err
	}
	p.podsInfo = podsInfo

	p.resourceData = make([]ResourceData, len(podsInfo))
	for i, pod := range podsInfo {
		p.resourceData[i] = PodData{&pod}
	}

	return nil
}

func (p *podsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(p.podsInfo))
	for i, pod := range p.podsInfo {
		rows[i] = table.Row{
			pod.Namespace,
			pod.Name,
			pod.Ready,
			pod.Status,
			fmt.Sprintf("%d", pod.Restarts),
			pod.Age,
		}
	}
	return rows
}
