package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

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

	desc, err := ds.daemonset.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("DaemonSet: "+ds.daemonset.Name, desc), nil
}
