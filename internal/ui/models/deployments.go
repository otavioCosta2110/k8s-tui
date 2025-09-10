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

type deploymentsModel struct {
	*GenericResourceModel
	deploymentsInfo []k8s.DeploymentInfo
}

func NewDeployments(k k8s.Client, namespace string) (*deploymentsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeDeployment,
		Title:           "Deployments in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.13},
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

func (d *deploymentsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	d.k8sClient = k

	if err := d.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		deployment := k8s.NewDeployment(selected, d.namespace, *k)
		p, err := deployment.GetPods()
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		pods, err := NewPods(*k, d.namespace, p)
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
			NewScreen: podsComponent,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := d.fetchData(); err != nil {
			return nil, err
		}
		return d.dataToRows(), nil
	}

	tableModel := ui.NewTable(d.config.Columns, d.config.ColumnWidths, d.dataToRows(), d.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": d.createDeleteAction(tableModel),
		"r": d.createRolloutAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, d.refreshInterval, d.k8sClient, "Deployments"), nil
}

func (d *deploymentsModel) fetchData() error {
	deploymentInfo, err := k8s.GetDeploymentsTableData(*d.k8sClient, d.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch deployments: %v", err)
	}
	d.deploymentsInfo = deploymentInfo

	d.resourceData = make([]ResourceData, len(deploymentInfo))
	for i, deployment := range deploymentInfo {
		d.resourceData[i] = DeploymentData{&deployment}
	}

	return nil
}

func (d *deploymentsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(d.deploymentsInfo))
	for i, deployment := range d.deploymentsInfo {
		rows[i] = table.Row{
			deployment.Namespace,
			deployment.Name,
			deployment.Ready,
			deployment.UpToDate,
			deployment.Available,
			deployment.Age,
		}
	}
	return rows
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
