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

type secretsModel struct {
	*GenericResourceModel
	secretsInfo []k8s.SecretInfo
}

func NewSecrets(k k8s.Client, namespace string) (*secretsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeSecret,
		Title:           "Secrets in " + namespace,
		ColumnWidths:    []float64{0.13, 0.35, 0.15, 0.15, 0.17, 0.10},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("TYPE", 0),
			components.NewColumn("DATA", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &secretsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (s *secretsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	if err := s.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		secretDetails, err := NewSecretDetails(*k, s.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: secretDetails,
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

	return NewAutoRefreshModel(tableModel, s.refreshInterval, s.k8sClient, "Secrets"), nil
}

func (s *secretsModel) fetchData() error {
	secretInfo, err := k8s.GetSecretsTableData(*s.k8sClient, s.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch secrets: %v", err)
	}
	s.secretsInfo = secretInfo

	s.resourceData = make([]ResourceData, len(secretInfo))
	for idx, secret := range secretInfo {
		s.resourceData[idx] = SecretData{&secret}
	}

	return nil
}
