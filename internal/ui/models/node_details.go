package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

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

	var desc string
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if pm := plugins.GetGlobalPluginManager(); pm != nil && pm.GetAPI() != nil {
		api := pm.GetAPI()
		api.SetClient(*k)
		desc, err = api.DescribeNode(n.node.Name)
	} else {
		desc, err = k8s.DescribeResource(*k, k8s.ResourceTypeNode, "", n.node.Name)
	}

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Node: "+n.node.Name, desc), nil
}
