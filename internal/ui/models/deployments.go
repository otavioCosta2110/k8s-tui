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
	list      []string
	namespace string
	k8sClient *k8s.Client
	deploymentsInfo []k8s.DeploymentInfo
	loading   bool
	err       error
	refreshInterval time.Duration
}

func NewDeployments(k k8s.Client, namespace string) (*deploymentsModel, error) {
	deployments, err := k8s.FetchDeploymentList(k, namespace)
	if err != nil {
		return nil, err
	}

	return &deploymentsModel{
		list:      deployments,
		namespace: namespace,
		k8sClient: &k,
		loading:   false,
		err:       nil,
		refreshInterval: 5 * time.Second,
	}, nil
}

func (d *deploymentsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	d.k8sClient = k
	deploymentInfo, err := k8s.GetDeploymentsTableData(*k, d.namespace)
	if err != nil {
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

	columns := []table.Column{
		components.NewColumn("NAMESPACE", 0),
		components.NewColumn("NAME", 0),
		components.NewColumn("READY", 0),
		components.NewColumn("UP-TO-DATE", 0),
		components.NewColumn("AVAILABLE", 0),
		components.NewColumn("AGE", 0),
	}

	colPercent := []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.15}

	rows := d.deploymentsToRows(deploymentInfo)

	fetchFunc := func() ([]table.Row, error) {
		deps, err := d.fetchDeps(d.k8sClient)
		if err != nil {
			return nil, err
		}

		newRows := d.deploymentsToRows(deps)
		return newRows, nil
	}

	tableModel := ui.NewTable(columns, colPercent, rows, "Deployments in "+d.namespace, onSelect, 1, fetchFunc)

	return &autoRefreshModel{
		inner:           tableModel,
		refreshInterval: d.refreshInterval,
		k8sClient:       d.k8sClient,
	}, nil
}

func (d *deploymentsModel) fetchDeps(client *k8s.Client) ([]k8s.DeploymentInfo, error) {
	deployments, err := k8s.GetDeploymentsTableData(*client, d.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployments: %v", err)
	}
	return deployments, nil
}

func (d *deploymentsModel) deploymentsToRows(deploymentInfo []k8s.DeploymentInfo) []table.Row {
	rows := []table.Row{}
	for _, deployment := range deploymentInfo {
		rows = append(rows, table.Row{
			deployment.Namespace,
			deployment.Name,
			deployment.Ready,
			deployment.UpToDate,
			deployment.Available,
			deployment.Age,
		})
	}
	return rows
}
