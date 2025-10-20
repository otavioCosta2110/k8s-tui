package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
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
		Title:           customstyles.ResourceIcons["DaemonSets"] + " DaemonSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.20, 0.05, 0.05, 0.05, 0.10, 0.10, 0.10, 0.10},
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

func (ds *daemonsetsModel) Help() (string, string) {
	return "DaemonSets Help", `DaemonSets ensure that all (or some) nodes run a copy of a pod.

Key Bindings:
• ↑/↓/j/k: Navigate daemonsets
• enter: View daemonset details
• d: Delete selected daemonsets
• r: Refresh
• /: Search daemonsets
• esc: Go back

DaemonSet Status:
• Desired: Number of desired pods
• Current: Number of current pods
• Ready: Number of ready pods
• Available: Number of available pods

Common Actions:
• View pods: See pods running on each node
• Check node selectors: See which nodes run the pods
• Update daemonset: Modify pod template`
}

func (ds *daemonsetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ds.k8sClient = k

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

	tableModel := ui.NewTable(ds.config.Columns, ds.config.ColumnWidths, []table.Row{}, ds.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": ds.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, ds.refreshInterval, ds.k8sClient, "DaemonSets"), nil
}

func (ds *daemonsetsModel) fetchData() error {
	var daemonsetInfo []k8s.DaemonSetInfo
	var err error

	daemonsetInfo, err = ds.pluginAPI.GetDaemonSets("")

	if err != nil {
		return fmt.Errorf("failed to fetch daemonsets: %v", err)
	}
	ds.daemonsetsInfo = daemonsetInfo

	ds.resourceData = make([]types.ResourceData, len(daemonsetInfo))
	for idx, daemonset := range daemonsetInfo {
		ds.resourceData[idx] = DaemonSetData{&daemonset}
	}

	return nil
}
