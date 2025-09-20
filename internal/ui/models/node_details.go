package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"

	tea "github.com/charmbracelet/bubbletea"
)

type nodeDetailsModel struct {
	node      *k8s.NodeInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewNodeDetails(k k8s.Client, nodeName string) *nodeDetailsModel {
	return &nodeDetailsModel{
		node:      k8s.NewNode(nodeName, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (n *nodeDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.k8sClient = k

	desc, err := n.node.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Node: "+n.node.Name, desc), nil
}
