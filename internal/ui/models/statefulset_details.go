package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"

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

	desc, err := ss.statefulset.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("StatefulSet: "+ss.statefulset.Name, desc), nil
}
