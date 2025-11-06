package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type ingressDetailsModel struct {
	baseDetailsModel
	ingress *k8s.IngressInfo
}

type ingressDetailsLoadedMsg struct {
	content string
	err     error
}

func NewIngressDetails(k k8s.Client, namespace, ingressName string) *ingressDetailsModel {
	return &ingressDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Ingress: "+ingressName, "Loading ingress details..."),
		ingress:          k8s.NewIngressInfo(ingressName, namespace, k),
	}
}

func (i *ingressDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	i.setClient(k)
	return i, nil
}

func (i *ingressDetailsModel) fetchIngressDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*i.k8sClient)
		desc, err := api.DescribeIngress(i.ingress.Namespace, i.ingress.Name)
		return ingressDetailsLoadedMsg{content: desc, err: err}
	}
}

func (i *ingressDetailsModel) Init() tea.Cmd {
	return i.initCommon(i.fetchIngressDetails())
}

func (i *ingressDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ingressDetailsLoadedMsg:
		i.setLoadedContent(msg.content, msg.err)
		return i, nil
	default:
		cmd, _ := i.updateCommon(msg)
		return i, cmd
	}
}

func (i *ingressDetailsModel) View() string {
	return i.baseDetailsModel.View()
}
