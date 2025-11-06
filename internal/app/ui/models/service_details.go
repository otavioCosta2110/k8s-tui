package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type serviceDetailsModel struct {
	baseDetailsModel
	service *k8s.ServiceInfo
}

type serviceDetailsLoadedMsg struct {
	content string
	err     error
}

func NewServiceDetails(k k8s.Client, namespace, serviceName string) *serviceDetailsModel {
	return &serviceDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Service: "+serviceName, "Loading service details..."),
		service:          k8s.NewServiceInfo(serviceName, namespace, k),
	}
}

func (s *serviceDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.setClient(k)
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
	return s.initCommon(s.fetchServiceDetails())
}

func (s *serviceDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case serviceDetailsLoadedMsg:
		s.setLoadedContent(msg.content, msg.err)
		return s, nil
	default:
		cmd, _ := s.updateCommon(msg)
		return s, cmd
	}
}

func (s *serviceDetailsModel) View() string {
	return s.baseDetailsModel.View()
}
