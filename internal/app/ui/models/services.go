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

type servicesModel struct {
	*GenericResourceModel
	servicesInfo []k8s.ServiceInfo
}

func NewServices(k k8s.Client, namespace string) (*servicesModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeService,
		Title:           customstyles.ResourceIcons["Services"] + " Services in " + namespace,
		ColumnWidths:    []float64{0.3, 0.4, 0.2, 0.2, 0.5, 0.5, 0.5},
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

func (s *servicesModel) SetNamespace(namespace string) {
	s.GenericResourceModel.SetNamespace(namespace)
	// Update the title to reflect the new namespace
	s.config.Title = customstyles.ResourceIcons["Services"] + " Services in " + namespace
}

func (s *servicesModel) Help() (string, string) {
	return "Services Help", `Services expose applications running on pods.

Key Bindings:
• ↑/↓/j/k: Navigate services
• enter: View service details
• E: View service events
• f: Port forward service
• d: Delete selected services
• n: Create new service
• r: Refresh
• /: Search services
• esc: Go back

Events View Key Bindings:
• g: Go to top
• G: Go to bottom
• r: Refresh events
• esc: Go back

Service Types:
• ClusterIP: Internal cluster access
• NodePort: External access via node IP
• LoadBalancer: Cloud provider load balancer

Common Actions:
• View endpoints: Enter to see pods backing the service
• Update service: Use service details view
• Scale backing resources: Check deployments or statefulsets
• Create service: Press 'n' to open create form`
}

func (s *servicesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		serviceDetails, err := NewServiceDetails(*k, s.pluginAPI.GetCurrentNamespace(), selected).InitComponent(k)
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

	tableModel := ui.NewTable(s.config.Columns, s.config.ColumnWidths, []table.Row{}, s.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": s.createDeleteAction(tableModel),
		"f": s.createPortForwardAction(tableModel),
		"n": s.createNewServiceAction(),
		"E": s.createViewEventsAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, s.refreshInterval, s.k8sClient, "Services"), nil
}

func (s *servicesModel) createPortForwardAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(s.servicesInfo) {
			return nil
		}

		service := s.servicesInfo[selected]
		serviceName := service.Name

		return func() tea.Msg {
			return components.OpenPortForwardFormMsg{
				Form:       components.NewPortForwardForm("Port Forward", "service", serviceName, s.pluginAPI.GetCurrentNamespace()),
				Breadcrumb: serviceName + " port-forward",
			}
		}
	}
}

func (s *servicesModel) createNewServiceAction() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return components.OpenCreateFormMsg{ResourceType: "service"}
		}
	}
}

func (s *servicesModel) createViewEventsAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(s.servicesInfo) {
			return nil
		}

		service := s.servicesInfo[selected]
		serviceName := service.Name

		return func() tea.Msg {
			eventsModel := NewEventsModel(s.k8sClient, s.pluginAPI.GetCurrentNamespace(), serviceName, "Service")
			eventsScreen, err := eventsModel.InitComponent(s.k8sClient)
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *s.k8sClient,
				}
			}
			return components.NavigateMsg{
				NewScreen:  eventsScreen,
				Breadcrumb: serviceName + " events",
			}
		}
	}
}

func (s *servicesModel) fetchData() error {
	var serviceInfo []k8s.ServiceInfo
	var err error

	serviceInfo, err = s.pluginAPI.GetServices(s.pluginAPI.GetCurrentNamespace())

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
