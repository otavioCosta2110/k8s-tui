package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type podDetailsModel struct {
	pod        *k8s.Pod
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	spinner    components.SpinnerModel
	loading    bool
	err        error
}

type podDetailsLoadedMsg struct {
	content string
	err     error
}

func NewPodDetails(k k8s.Client, namespace, podName string) *podDetailsModel {
	return &podDetailsModel{
		pod:        k8s.NewPod(podName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("Pod: "+podName, "Loading..."),
		spinner:    components.NewSpinner("Loading pod details..."),
		loading:    true,
		err:        nil,
	}
}

func (p *podDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k
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
	return tea.Batch(p.yamlViewer.Init(), p.spinner.Init(), p.fetchPodDetails())
}

func (p *podDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case podDetailsLoadedMsg:
		p.loading = false
		if msg.err != nil {
			p.err = msg.err
			p.yamlViewer.SetContent("Error loading pod details: " + msg.err.Error())
		} else {
			p.yamlViewer.SetContent(msg.content)
		}
		return p, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := p.yamlViewer.Update(msg)
		p.yamlViewer = updatedViewer.(*components.YAMLViewer)

		var spinnerCmd tea.Cmd
		p.spinner, spinnerCmd = p.spinner.Update(msg)

		return p, tea.Batch(cmd, spinnerCmd)
	}
}

func (p *podDetailsModel) View() string {
	if p.loading {
		return p.spinner.View()
	}
	return p.yamlViewer.View()
}
