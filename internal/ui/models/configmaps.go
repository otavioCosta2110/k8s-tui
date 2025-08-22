package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type configmapsModel struct {
	namespace       string
	k8sClient       *k8s.Client
	cms             []k8s.Configmap
	loading         bool
	err             error
	refreshInterval time.Duration
}

func NewConfigmaps(k k8s.Client, namespace string, cms []k8s.Configmap) (*configmapsModel, error) {
	if len(cms) == 0 {
		var err error
		cms, err = k8s.FetchConfigmaps(k, namespace, "")
		if err != nil {
			return nil, err
		}
	}
	return &configmapsModel{
		namespace:       namespace,
		k8sClient:       &k,
		loading:         false,
		err:             nil,
		refreshInterval: 5 * time.Second,
	}, nil
}

func (c *configmapsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	cms, err := k8s.FetchConfigmaps(*k, c.namespace, "")
	if err != nil {
		return nil, err
	}
	c.cms = cms

	onSelect := func(selected string) tea.Msg {
		cmDetails, err := NewConfigmapDetails(*k, c.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: cmDetails,
		}
	}

	columns := []table.Column{
		components.NewColumn("NAMESPACE", 0),
		components.NewColumn("NAME", 0),
		components.NewColumn("DATA", 0),
		components.NewColumn("AGE", 0),
	}

	colPercent := []float64{0.30, 0.30, 0.16, 0.20}

	rows := []table.Row{}
	for _, cm := range c.cms {
		rows = append(rows, table.Row{
			cm.Namespace,
			cm.Name,
			cm.Data,
			cm.Age,
		})
	}

	fetchFunc := func() ([]table.Row, error) {
		cms, err := c.fetchConfigmaps(c.k8sClient)
		if err != nil {
			return nil, err
		}

		newRows := make([]table.Row, len(cms))
		for i, cm := range cms {
			newRows[i] = table.Row{
				cm.Namespace,
				cm.Name,
				cm.Data,
				cm.Age,
			}
		}
		return newRows, nil
	}

	tableModel := ui.NewTable(columns, colPercent, rows, "ConfigMaps in "+c.namespace, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": func() tea.Cmd {
			checked := tableModel.GetCheckedItems()
			for _, idx := range checked {
				if idx < len(c.cms) {
					cm := c.cms[idx]
					err := k8s.DeleteConfigmap(*c.k8sClient, cm.Namespace, cm.Name)
					return func() tea.Msg {
						return ErrorModel{error: err}
					}
				}
			}
			tableModel.Refresh()
			return nil
		},
	}
	tableModel.SetUpdateActions(actions)

	return &autoRefreshModel{
		inner:           tableModel,
		refreshInterval: c.refreshInterval,
		k8sClient:       c.k8sClient,
	}, nil
}

func (p *configmapsModel) fetchConfigmaps(k *k8s.Client) ([]k8s.Configmap, error) {
	return k8s.FetchConfigmaps(*k, p.namespace, "")
}
