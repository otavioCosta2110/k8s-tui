package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

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

	var desc string
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if pm := plugins.GetGlobalPluginManager(); pm != nil && pm.GetAPI() != nil {
		api := pm.GetAPI()
		api.SetClient(*k)
		desc, err = api.DescribeIngress(i.ingress.Namespace, i.ingress.Name)
	} else {
		desc, err = k8s.DescribeResource(*k, k8s.ResourceTypeIngress, i.ingress.Namespace, i.ingress.Name)
	}

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Ingress: "+i.ingress.Name, desc), nil
}
