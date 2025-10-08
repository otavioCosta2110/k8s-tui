package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type nodeDetailsModel struct {
	baseDetailsModel
	node *k8s.NodeInfo
}

type nodeDetailsLoadedMsg struct {
	content string
	err     error
}

func NewNodeDetails(k k8s.Client, nodeName string) *nodeDetailsModel {
	return &nodeDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Node: "+nodeName, "Loading node details..."),
		node:             k8s.NewNode(nodeName, k),
	}
}

func (n *nodeDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.setClient(k)
	return n, nil
}

func (n *nodeDetailsModel) fetchNodeDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*n.k8sClient)
		desc, err := api.DescribeNode(n.node.Name)
		return nodeDetailsLoadedMsg{content: desc, err: err}
	}
}

func (n *nodeDetailsModel) Init() tea.Cmd {
	return n.initCommon(n.fetchNodeDetails())
}

func (n *nodeDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nodeDetailsLoadedMsg:
		n.setLoadedContent(msg.content, msg.err)
		return n, nil
	default:
		cmd, _ := n.updateCommon(msg)
		return n, cmd
	}
}

func (n *nodeDetailsModel) View() string {
	return n.baseDetailsModel.View()
}
