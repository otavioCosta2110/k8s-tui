package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type servicesModel struct {
	*GenericResourceModel
	servicesInfo []k8s.ServiceInfo
}

func NewServices(k k8s.Client, namespace string) (*servicesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeService,
		Title:           customstyles.ResourceIcons["Services"] + " Services in " + namespace,
		ColumnWidths:    []float64{0.12, 0.21, 0.10, 0.15, 0.15, 0.15, 0.05},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("TYPE", 0),
			components.NewColumn("CLUSTER-IP", 0),
			components.NewColumn("EXTERNAL-IP", 0),
			components.NewColumn("PORTS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &servicesModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (s *servicesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	if err := s.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		serviceDetails, err := NewServiceDetails(*k, s.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: serviceDetails,
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

	return NewAutoRefreshModel(tableModel, s.refreshInterval, s.k8sClient, "Services"), nil
}

func (s *servicesModel) fetchData() error {
	serviceInfo, err := k8s.GetServicesTableData(*s.k8sClient, s.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch services: %v", err)
	}
	s.servicesInfo = serviceInfo

	s.resourceData = make([]types.ResourceData, len(serviceInfo))
	for idx, service := range serviceInfo {
		s.resourceData[idx] = ServiceData{&service}
	}

	return nil
}
