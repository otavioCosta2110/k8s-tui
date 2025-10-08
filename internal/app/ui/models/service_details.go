package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type serviceDetailsModel struct {
	service    *k8s.ServiceInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	spinner    components.SpinnerModel
	loading    bool
	err        error
}

type serviceDetailsLoadedMsg struct {
	content string
	err     error
}

func NewServiceDetails(k k8s.Client, namespace, serviceName string) *serviceDetailsModel {
	return &serviceDetailsModel{
		service:    k8s.NewService(serviceName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("Service: "+serviceName, "Loading..."),
		spinner:    components.NewSpinner("Loading service details..."),
		loading:    true,
		err:        nil,
	}
}

func (s *serviceDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k
	return s, nil
}

func (s *serviceDetailsModel) fetchServiceDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*s.k8sClient)
		desc, err := api.DescribeService(s.service.Namespace, s.service.Name)
		return serviceDetailsLoadedMsg{content: desc, err: err}
	}
}

func (s *serviceDetailsModel) Init() tea.Cmd {
	return tea.Batch(s.yamlViewer.Init(), s.spinner.Init(), s.fetchServiceDetails())
}

func (s *serviceDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case serviceDetailsLoadedMsg:
		s.loading = false
		if msg.err != nil {
			s.err = msg.err
			s.yamlViewer.SetContent("Error loading service details: " + msg.err.Error())
		} else {
			s.yamlViewer.SetContent(msg.content)
		}
		return s, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := s.yamlViewer.Update(msg)
		s.yamlViewer = updatedViewer.(*components.YAMLViewer)

		var spinnerCmd tea.Cmd
		s.spinner, spinnerCmd = s.spinner.Update(msg)

		return s, tea.Batch(cmd, spinnerCmd)
	}
}

func (s *serviceDetailsModel) View() string {
	if s.loading {
		return s.spinner.View()
	}
	return s.yamlViewer.View()
}
