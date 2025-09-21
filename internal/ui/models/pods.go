package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type podsModel struct {
	*GenericResourceModel
	selector string
}

func NewPods(k k8s.Client, namespace string, selector ...string) (*podsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypePod,
		Title:           customstyles.ResourceIcons["Pods"] + " Pods in " + namespace,
		ColumnWidths:    []float64{1, 2, 1, 0.8, 0.5, 1},
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

	selectorStr := ""
	if len(selector) > 0 {
		selectorStr = selector[0]
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &podsModel{
		GenericResourceModel: genericModel,
		selector:             selectorStr,
	}

	return model, nil
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k

	if err := p.fetchData(p.selector); err != nil {
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
			NewScreen:  podDetails,
			Breadcrumb: selected,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := p.fetchData(p.selector); err != nil {
			return nil, err
		}
		return p.dataToRows(), nil
	}

	tableModel := ui.NewTable(p.config.Columns, p.config.ColumnWidths, p.dataToRows(), p.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": p.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, p.refreshInterval, p.k8sClient, "Pods"), nil
}

func (p *podsModel) fetchData(selector string) error {
	var podsInfo []k8s.PodInfo
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if p.pluginAPI != nil {
		podsInfo, err = p.pluginAPI.GetPods(p.namespace)
	} else {
		podsInfo, err = k8s.FetchPods(*p.k8sClient, p.namespace, selector)
	}

	if err != nil {
		return err
	}

	p.resourceData = make([]types.ResourceData, len(podsInfo))
	for i, pod := range podsInfo {
		p.resourceData[i] = PodData{&pod}
	}

	return nil
}
