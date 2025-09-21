package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type daemonsetDetailsModel struct {
	daemonset *k8s.DaemonSetInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewDaemonSetDetails(k k8s.Client, namespace, daemonsetName string) *daemonsetDetailsModel {
	return &daemonsetDetailsModel{
		daemonset: k8s.NewDaemonSet(daemonsetName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (ds *daemonsetDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ds.k8sClient = k

	var desc string
	var err error

	// Always use plugin API - resources should never bypass the plugin system
	pm := plugins.GetGlobalPluginManager()
	api := pm.GetAPI()
	api.SetClient(*k)
	desc, err = api.DescribeDaemonSet(ds.daemonset.Namespace, ds.daemonset.Name)

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("DaemonSet: "+ds.daemonset.Name, desc), nil
}
