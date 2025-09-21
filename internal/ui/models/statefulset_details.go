package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type statefulsetDetailsModel struct {
	statefulset *k8s.StatefulSetInfo
	k8sClient   *k8s.Client
	loading     bool
	err         error
}

func NewStatefulSetDetails(k k8s.Client, namespace, statefulsetName string) *statefulsetDetailsModel {
	return &statefulsetDetailsModel{
		statefulset: k8s.NewStatefulSet(statefulsetName, namespace, k),
		k8sClient:   &k,
		loading:     false,
		err:         nil,
	}
}

func (ss *statefulsetDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ss.k8sClient = k

	var desc string
	var err error

	// Use plugin API if available, otherwise fall back to k8s client
	if pm := plugins.GetGlobalPluginManager(); pm != nil && pm.GetAPI() != nil {
		api := pm.GetAPI()
		api.SetClient(*k)
		desc, err = api.DescribeStatefulSet(ss.statefulset.Namespace, ss.statefulset.Name)
	} else {
		desc, err = k8s.DescribeResource(*k, k8s.ResourceTypeStatefulSet, ss.statefulset.Namespace, ss.statefulset.Name)
	}

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("StatefulSet: "+ss.statefulset.Name, desc), nil
}
