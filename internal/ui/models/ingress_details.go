package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type ingressDetailsModel struct {
	ingress   *k8s.IngressInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewIngressDetails(k k8s.Client, namespace, ingressName string) *ingressDetailsModel {
	return &ingressDetailsModel{
		ingress:   k8s.NewIngress(ingressName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (i *ingressDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	i.k8sClient = k

	desc, err := i.ingress.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Ingress: "+i.ingress.Name, desc), nil
}
