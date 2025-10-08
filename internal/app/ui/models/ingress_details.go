package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type ingressDetailsModel struct {
	ingress    *k8s.IngressInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	loading    bool
	err        error
}

type ingressDetailsLoadedMsg struct {
	content string
	err     error
}

func NewIngressDetails(k k8s.Client, namespace, ingressName string) *ingressDetailsModel {
	return &ingressDetailsModel{
		ingress:    k8s.NewIngress(ingressName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("Ingress: "+ingressName, "Loading..."),
		loading:    true,
		err:        nil,
	}
}

func (i *ingressDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	i.k8sClient = k
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
	return tea.Batch(i.yamlViewer.Init(), i.fetchIngressDetails())
}

func (i *ingressDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ingressDetailsLoadedMsg:
		i.loading = false
		if msg.err != nil {
			i.err = msg.err
			i.yamlViewer.SetContent("Error loading ingress details: " + msg.err.Error())
		} else {
			i.yamlViewer.SetContent(msg.content)
		}
		return i, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := i.yamlViewer.Update(msg)
		i.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return i, cmd
	}
}

func (i *ingressDetailsModel) View() string {
	return i.yamlViewer.View()
}
