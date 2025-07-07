package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type namespacesModel struct {
	list      []string
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewNamespaces(k k8s.Client) (*namespacesModel, error) {
	namespaces, err := k8s.FetchNamespaces(k)
	if err != nil {
		return nil, err
	}

	return &namespacesModel{
		list:      namespaces,
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}, nil
}

func (n *namespacesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.k8sClient = k
	namespaces, err := k8s.FetchNamespaces(*k)
	if err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		return components.NavigateMsg{
			NewScreen: NewResource(*k, selected).InitComponent(*k),
		}
	}

	return ui.NewList(namespaces, "Namespaces", onSelect), nil
}
