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

type ingressesModel struct {
	*GenericResourceModel
	ingressesInfo []k8s.IngressInfo
}

func NewIngresses(k k8s.Client, namespace string) (*ingressesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeIngress,
		Title:           "Ingresses in " + namespace,
		ColumnWidths:    []float64{0.13, 0.23, 0.13, 0.13, 0.13, 0.13, 0.05},
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

func (i *ingressesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	i.k8sClient = k

	if err := i.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		ingressDetails, err := NewIngressDetails(*k, i.namespace, selected).InitComponent(k)
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

	tableModel := ui.NewTable(i.config.Columns, i.config.ColumnWidths, i.dataToRows(), i.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": i.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, i.refreshInterval, i.k8sClient, "Ingresses"), nil
}

func (i *ingressesModel) fetchData() error {
	ingressInfo, err := k8s.GetIngressesTableData(*i.k8sClient, i.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch ingresses: %v", err)
	}
	i.ingressesInfo = ingressInfo

	i.resourceData = make([]ResourceData, len(ingressInfo))
	for idx, ingress := range ingressInfo {
		i.resourceData[idx] = IngressData{&ingress}
	}

	return nil
}

func (i *ingressesModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(i.ingressesInfo))
	for idx, ingress := range i.ingressesInfo {
		rows[idx] = table.Row{
			ingress.Namespace,
			ingress.Name,
			ingress.Class,
			ingress.Hosts,
			ingress.Address,
			ingress.Ports,
			ingress.Age,
		}
	}
	return rows
}
