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

func (d *deploymentsModel) InitComponent(k *resources.Client) (tea.Model, error) {
	d.k8sClient = k

	if err := d.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		deployment := resources.NewDeployment(selected, d.namespace, *k)
		err := deployment.Fetch()
		if err != nil {
			return components.NavigateMsg{
				Error:   fmt.Errorf("failed to fetch deployment: %v", err),
				Cluster: *k,
			}
		}
		selector, err := deployment.GetLabelSelector()
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to get label selector for deployment %s: %v, using fallback", deployment.Name, err))
			selector = fmt.Sprintf("app=%s", deployment.Name)
		}
		logger.Debug(fmt.Sprintf("Using selector for deployment %s: %s", deployment.Name, selector))
		pods, err := NewPods(*k, d.namespace, selector)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		podsComponent, err := pods.InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		return components.NavigateMsg{
			NewScreen:  podsComponent,
			Breadcrumb: "Pods",
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := d.fetchData(); err != nil {
			return nil, err
		}
		return d.dataToRows(), nil
	}

	tableModel := ui.NewTable(d.config.Columns, d.config.ColumnWidths, d.dataToRows(), d.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": d.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, d.refreshInterval, d.k8sClient, "Deployments"), nil
}

func (d *deploymentsModel) fetchData() error {
	var deploymentInfo []resources.DeploymentInfo
	var err error

	deploymentInfo, err = d.pluginAPI.GetDeployments(d.namespace)

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
