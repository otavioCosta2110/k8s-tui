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

func (c *configmapsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	c.k8sClient = k

	if err := c.fetchData(); err != nil {
		return nil, err
	}

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

	tableModel := ui.NewTable(c.config.Columns, c.config.ColumnWidths, c.dataToRows(), c.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": c.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, c.refreshInterval, c.k8sClient, "ConfigMaps"), nil
}

func (c *configmapsModel) fetchData() error {
	var cms []k8s.Configmap
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if c.pluginAPI != nil {
		cms, err = c.pluginAPI.GetConfigMaps(c.namespace)
	} else {
		cms, err = k8s.FetchConfigmaps(*c.k8sClient, c.namespace, "")
	}

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
