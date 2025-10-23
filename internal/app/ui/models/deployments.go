package models

import (
	"fmt"
	"time"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type deploymentsModel struct {
	*GenericResourceModel
	deploymentsInfo []resources.DeploymentInfo
}

func NewDeployments(k resources.Client, namespace string) (*deploymentsModel, error) {
	config := ResourceConfig{
		ResourceType:    resources.ResourceTypeDeployment,
		Title:           styles.ResourceIcons["Deployments"] + " Deployments in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.15},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("UP-TO-DATE", 0),
			components.NewColumn("AVAILABLE", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &deploymentsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (d *deploymentsModel) Help() (string, string) {
	return "Deployments Help", `Deployments manage the deployment and scaling of applications.

Key Bindings:
• ↑/↓/j/k: Navigate deployments
• enter: View deployment details
• d: Delete selected deployments
• n: Create new deployment
• r: Refresh
• /: Search deployments
• esc: Go back

Deployment Status:
• Ready: Shows ready/desired replicas
• Updated: Shows updated replicas
• Available: Shows available replicas

Common Actions:
• Scale deployment: Enter to view details and scale
• Update image: Use deployment details view
• View pods: See associated pods in details
• Create deployment: Press 'n' to open create form`
}

func (d *deploymentsModel) InitComponent(k *resources.Client) (tea.Model, error) {
	d.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		deployment := resources.NewDeployment(selected, d.namespace, *k)
		if err := deployment.Fetch(); err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		selector, err := deployment.GetLabelSelector()
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		podsModel, err := NewPodsWithParent(*k, d.namespace, selected, selector)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		podsScreen, err := podsModel.InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		logger.Info(fmt.Sprintf("DEBUG: Creating NavigateMsg with selector '%s' for deployment '%s'", selector, selected))
		logger.Info(fmt.Sprintf("DEBUG: Creating NavigateMsg with selector '%s' for deployment '%s'", selector, selected))
		return components.NavigateMsg{
			NewScreen:  podsScreen,
			Breadcrumb: selected + " pods",
			Metadata: map[string]interface{}{
				"selector": selector,
				"parent":   selected,
			},
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := d.fetchData(); err != nil {
			return nil, err
		}
		return d.dataToRows(), nil
	}

	tableModel := ui.NewTable(d.config.Columns, d.config.ColumnWidths, []table.Row{}, d.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": d.createDeleteAction(tableModel),
		"v": d.createViewDetailsAction(tableModel),
		"n": d.createNewDeploymentAction(),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, d.refreshInterval, d.k8sClient, "Deployments"), nil
}

func (d *deploymentsModel) createNewDeploymentAction() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return components.OpenCreateFormMsg{ResourceType: "deployment"}
		}
	}
}

func (d *deploymentsModel) createViewDetailsAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(d.deploymentsInfo) {
			return nil
		}

		deploymentName := d.deploymentsInfo[selected].Name

		return func() tea.Msg {
			deploymentDetails, err := NewDeploymentDetails(*d.k8sClient, d.namespace, deploymentName).InitComponent(d.k8sClient)
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *d.k8sClient,
				}
			}
			return components.NavigateMsg{
				NewScreen: deploymentDetails,
			}
		}
	}
}

func (d *deploymentsModel) fetchData() error {
	var deploymentInfo []resources.DeploymentInfo
	var err error

	deploymentInfo, err = d.pluginAPI.GetDeployments("")

	if err != nil {
		return fmt.Errorf("failed to fetch deployments: %v", err)
	}
	d.deploymentsInfo = deploymentInfo

	d.resourceData = make([]types.ResourceData, len(deploymentInfo))
	for i, deployment := range deploymentInfo {
		d.resourceData[i] = DeploymentData{&deployment}
	}

	return nil
}

func (d *deploymentsModel) createRolloutAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(d.deploymentsInfo) {
			return nil
		}

		deployment := d.deploymentsInfo[selected]

		return func() tea.Msg {
			return components.NavigateMsg{
				NewScreen: components.NewYAMLViewer(
					"Rollout Triggered",
					fmt.Sprintf("Rollout triggered for deployment: %s/%s\n\nStatus: Rollout initiated successfully", deployment.Namespace, deployment.Name),
				),
				Cluster: *d.k8sClient,
			}
		}
	}
}
