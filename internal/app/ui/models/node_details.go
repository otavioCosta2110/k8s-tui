package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type nodeDetailsModel struct {
	node       *k8s.NodeInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	loading    bool
	err        error
}

type nodeDetailsLoadedMsg struct {
	content string
	err     error
}

func NewNodeDetails(k k8s.Client, nodeName string) *nodeDetailsModel {
	return &nodeDetailsModel{
		node:       k8s.NewNode(nodeName, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("Node: "+nodeName, "Loading..."),
		loading:    true,
		err:        nil,
	}
}

func (n *nodeDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.k8sClient = k
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
	return tea.Batch(n.yamlViewer.Init(), n.fetchNodeDetails())
}

func (n *nodeDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nodeDetailsLoadedMsg:
		n.loading = false
		if msg.err != nil {
			n.err = msg.err
			n.yamlViewer.SetContent("Error loading node details: " + msg.err.Error())
		} else {
			n.yamlViewer.SetContent(msg.content)
		}
		return n, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := n.yamlViewer.Update(msg)
		n.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return n, cmd
	}
}

func (n *nodeDetailsModel) View() string {
	return n.yamlViewer.View()
}
