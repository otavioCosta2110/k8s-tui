package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type podsModel struct {
	list      []string
	namespace string
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewPods(k k8s.Client, namespace string, pods []string) (*podsModel, error) {
	if len(pods) == 0 {
		var err error
		pods, err = k8s.FetchPods(k, namespace, "")
		if err != nil {
			return nil, err
		}
	}

	return &podsModel{
		list:      pods,
		namespace: namespace,
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}, nil
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	onSelect := func(selected string) tea.Msg {
		podDetails, err := NewPodDetails(*k, p.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: podDetails,
		}
	}

	return ui.NewList(p.list, "Pods in "+p.namespace, onSelect), nil
}
