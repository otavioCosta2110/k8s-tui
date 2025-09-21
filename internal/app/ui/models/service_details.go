package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type serviceDetailsModel struct {
	service   *k8s.ServiceInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewServiceDetails(k k8s.Client, namespace, serviceName string) *serviceDetailsModel {
	return &serviceDetailsModel{
		service:   k8s.NewService(serviceName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (s *serviceDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	var desc string
	var err error

	pm := plugins.GetGlobalPluginManager()
	api := pm.GetAPI()
	api.SetClient(*k)
	desc, err = api.DescribeService(s.service.Namespace, s.service.Name)

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Service: "+s.service.Name, desc), nil
}
