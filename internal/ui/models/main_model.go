package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"

	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	kube      k8s.Client
	namespace string
}

func NewMainModel(k k8s.Client, namespace string) MainModel {
	return MainModel{
		kube:      k,
		namespace: namespace,
	}
}

func (m MainModel) InitComponent(k k8s.Client) (tea.Model, error) {
	if k.Namespace == "" {
		namespaces, err := NewNamespaces(k)
		if err != nil {
			return nil, err
		}
		namespacesComponent, err := namespaces.InitComponent(&k)
		if err != nil {
			return nil, err
		}
		return namespacesComponent, nil
	}
	resourceModel := NewResource(m.kube, m.namespace)
	return resourceModel.InitComponent(k), nil
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m MainModel) View() string {
	return ""
}
