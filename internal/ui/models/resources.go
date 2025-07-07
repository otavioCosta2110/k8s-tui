package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type Resource struct {
	kube       k8s.Client
	namespace  string
	resourceType string
}

func NewResource(k k8s.Client, namespace string) Resource {
	return Resource{
		kube:      k,
		namespace: namespace,
	}
}

func (r Resource) InitComponent(k k8s.Client) tea.Model {
	resourceTypes := []string{
		"Pods",
		"Deployments",
		"Services",
		"ConfigMaps",
		"Secrets",
	}

	onSelect := func(selected string) tea.Msg {
		r.resourceType = selected
		return components.NavigateMsg{
			NewScreen: NewResourceList(r.kube, r.namespace, selected).InitComponent(k),
		}
	}

	return components.NewList(resourceTypes, "Resource Types", onSelect)
}

