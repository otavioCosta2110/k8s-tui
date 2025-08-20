package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type cmDetailsModel struct {
	cm        *k8s.Configmap
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewConfigmapDetails(k k8s.Client, namespace, cmName string) *cmDetailsModel {
	return &cmDetailsModel{
		cm:        k8s.NewConfigmap(cmName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (c *cmDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	c.k8sClient = k

	desc, err := c.cm.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Configmap: "+c.cm.Name, desc), nil
}
