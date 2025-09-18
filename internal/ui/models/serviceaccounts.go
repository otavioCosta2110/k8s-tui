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

type serviceaccountsModel struct {
	*GenericResourceModel
	serviceaccountsInfo []k8s.ServiceAccountInfo
}

func NewServiceAccounts(k k8s.Client, namespace string) (*serviceaccountsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeServiceAccount,
		Title:           "ServiceAccounts in " + namespace,
		ColumnWidths:    []float64{0.15, 0.50, 0.16, 0.15, 0.18},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("SECRETS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &serviceaccountsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (s *serviceaccountsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	if err := s.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		serviceaccountDetails, err := NewServiceAccountDetails(*k, s.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: serviceaccountDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := s.fetchData(); err != nil {
			return nil, err
		}
		return s.dataToRows(), nil
	}

	tableModel := ui.NewTable(s.config.Columns, s.config.ColumnWidths, s.dataToRows(), s.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": s.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, s.refreshInterval, s.k8sClient, "ServiceAccounts"), nil
}

func (s *serviceaccountsModel) fetchData() error {
	serviceaccountInfo, err := k8s.GetServiceAccountsTableData(*s.k8sClient, s.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch serviceaccounts: %v", err)
	}
	s.serviceaccountsInfo = serviceaccountInfo

	s.resourceData = make([]ResourceData, len(serviceaccountInfo))
	for idx, serviceaccount := range serviceaccountInfo {
		s.resourceData[idx] = ServiceAccountData{&serviceaccount}
	}

	return nil
}

func (s *serviceaccountsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(s.serviceaccountsInfo))
	for idx, serviceaccount := range s.serviceaccountsInfo {
		rows[idx] = table.Row{
			serviceaccount.Namespace,
			serviceaccount.Name,
			serviceaccount.Secrets,
			serviceaccount.Age,
		}
	}
	return rows
}
