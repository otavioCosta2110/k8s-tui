package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type podDetailsModel struct {
	pod       *k8s.Pod
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewPodDetails(k k8s.Client, namespace, podName string) *podDetailsModel {
	return &podDetailsModel{
		pod:       k8s.NewPod(podName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (p *podDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k

	var desc string
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if pm := plugins.GetGlobalPluginManager(); pm != nil && pm.GetAPI() != nil {
		api := pm.GetAPI()
		api.SetClient(*k)
		desc, err = api.DescribePod(p.pod.Namespace, p.pod.Name)
	} else {
		desc, err = k8s.DescribeResource(*k, k8s.ResourceTypePod, p.pod.Namespace, p.pod.Name)
	}

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Pod: "+p.pod.Name, desc), nil
}
