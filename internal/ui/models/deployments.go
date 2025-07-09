package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type deploymentsModel struct {
	list      []string
	namespace string
	k8sClient *k8s.Client
	loading   bool
	err       error
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
	}, nil
}

func (d *deploymentsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	d.k8sClient = k
	deployments, err := k8s.FetchDeploymentList(*k, d.namespace)
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

	return ui.NewList(deployments, "Deployments in "+d.namespace, onSelect), nil
}
