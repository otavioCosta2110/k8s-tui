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

type ingressesModel struct {
	*GenericResourceModel
	ingressesInfo []k8s.IngressInfo
}

func NewIngresses(k k8s.Client, namespace string) (*ingressesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeIngress,
		Title:           customstyles.ResourceIcons["Ingresses"] + " Ingresses in " + namespace,
		ColumnWidths:    []float64{1, 2.1, 1, 1, 1, 1, 1},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("CLASS", 0),
			components.NewColumn("HOSTS", 0),
			components.NewColumn("ADDRESS", 0),
			components.NewColumn("PORTS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &ingressesModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (i *ingressesModel) Help() (string, string) {
	return "Ingresses Help", `Ingresses manage external access to services.

Key Bindings:
• ↑/↓/j/k: Navigate ingresses
• enter: View ingress details
• d: Delete selected ingresses
• n: Create new ingress
• r: Refresh
• /: Search ingresses
• esc: Go back

Common Actions:
• View rules: See routing rules
• Check TLS: View SSL certificates
• Update rules: Modify routing configuration
• Create ingress: Press 'n' to open create form`
}

func (i *ingressesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	i.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		ingressDetails, err := NewIngressDetails(*k, i.pluginAPI.GetCurrentNamespace(), selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: ingressDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := i.fetchData(); err != nil {
			return nil, err
		}
		return i.dataToRows(), nil
	}

	tableModel := ui.NewTable(i.config.Columns, i.config.ColumnWidths, []table.Row{}, i.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": i.createDeleteAction(tableModel),
		"n": i.createNewIngressAction(),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, i.refreshInterval, i.k8sClient, "Ingresses"), nil
}

func (i *ingressesModel) createNewIngressAction() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return components.OpenCreateFormMsg{ResourceType: "ingress"}
		}
	}
}

func (i *ingressesModel) fetchData() error {
	var ingressInfo []k8s.IngressInfo
	var err error

	ingressInfo, err = i.pluginAPI.GetIngresses(i.pluginAPI.GetCurrentNamespace())

	if err != nil {
		return fmt.Errorf("failed to fetch ingresses: %v", err)
	}
	i.ingressesInfo = ingressInfo

	i.resourceData = make([]types.ResourceData, len(ingressInfo))
	for idx, ingress := range ingressInfo {
		i.resourceData[idx] = IngressData{&ingress}
	}

	return nil
}
