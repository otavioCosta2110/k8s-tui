package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type podDetailsModel struct {
	baseDetailsModel
	pod *k8s.Pod
}

type podDetailsLoadedMsg struct {
	content string
	err     error
}

func NewPodDetails(k k8s.Client, namespace, podName string) *podDetailsModel {
	return &podDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Pod: "+podName, "Loading pod details..."),
		pod:              k8s.NewPodInfo(podName, namespace, k),
	}
}

func (p *podDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.setClient(k)
	return p, nil
}

func (p *podDetailsModel) fetchPodDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*p.k8sClient)
		desc, err := api.DescribePod(p.pod.Namespace, p.pod.Name)
		return podDetailsLoadedMsg{content: desc, err: err}
	}
}

func (p *podDetailsModel) Init() tea.Cmd {
	return p.initCommon(p.fetchPodDetails())
}

func (p *podDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case podDetailsLoadedMsg:
		p.setLoadedContent(msg.content, msg.err)
		return p, nil
	default:
		cmd, _ := p.updateCommon(msg)
		return p, cmd
	}
}

func (p *podDetailsModel) View() string {
	return p.baseDetailsModel.View()
}
