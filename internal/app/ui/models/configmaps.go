package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type configmapsModel struct {
	*GenericResourceModel
	cms []k8s.Configmap
}

func NewConfigmaps(k k8s.Client, namespace string) (*configmapsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeConfigMap,
		Title:           customstyles.ResourceIcons["ConfigMaps"] + " ConfigMaps in " + namespace,
		ColumnWidths:    []float64{0.30, 0.30, 0.16, 0.18},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("DATA", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &configmapsModel{
		GenericResourceModel: genericModel,
		cms:                  nil,
	}

	return model, nil
}

func (c *configmapsModel) Help() (string, string) {
	return "ConfigMaps Help", `ConfigMaps store configuration data.

Key Bindings:
• ↑/↓/j/k: Navigate configmaps
• enter: View configmap details
• d: Delete selected configmaps
• r: Refresh
• /: Search configmaps
• esc: Go back

Common Actions:
• View data: See configuration key-value pairs
• Edit values: Modify configuration data
• Check usage: See which pods use this config`
}

func (c *configmapsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	c.k8sClient = k

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

	fetchFunc := func() ([]table.Row, error) {
		if err := c.fetchData(); err != nil {
			return nil, err
		}
		return c.dataToRows(), nil
	}

	tableModel := ui.NewTable(c.config.Columns, c.config.ColumnWidths, []table.Row{}, c.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": c.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, c.refreshInterval, c.k8sClient, "ConfigMaps"), nil
}

func (c *configmapsModel) fetchData() error {
	var cms []k8s.Configmap
	var err error

	cms, err = c.pluginAPI.GetConfigMaps("")

	if err != nil {
		return err
	}
	c.cms = cms

	c.resourceData = make([]types.ResourceData, len(cms))
	for i, cm := range cms {
		c.resourceData[i] = ConfigMapData{&cm}
	}

	return nil
}
