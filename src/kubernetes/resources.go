package kubernetes

import (
	"otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
)

type Resource struct {
	kube       KubeConfig
	namespace  string
	resourceType string
}

func NewResource(k KubeConfig, namespace string) Resource {
	return Resource{
		kube:      k,
		namespace: namespace,
	}
}

func (r Resource) InitComponent(k KubeConfig) tea.Model {
	resourceTypes := []string{
		"Pods",
		"Deployments",
		"Services",
		"ConfigMaps",
		"Secrets",
	}

	onSelect := func(selected string) tea.Msg {
		r.resourceType = selected
		return NavigateMsg{
			NewScreen: NewResourceList(r.kube, r.namespace, selected).InitComponent(k),
		}
	}

	return list.NewList(resourceTypes, "Resource Types", onSelect)
}
