package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type Resource struct {
	kube         k8s.Client
	namespace    string
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
		"Ingresses",
		"ConfigMaps",
		"Secrets",
		"ServiceAccounts",
		"ReplicaSets",
		"Nodes",
		"Jobs",
		"CronJobs",
		"DaemonSets",
		"StatefulSets",
	}

	onSelect := func(selected string) tea.Msg {
		r.resourceType = selected
		newResourceList, err := NewResourceList(r.kube, r.namespace, r.resourceType).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error: err,
			}
		}
		return components.NavigateMsg{
			NewScreen: newResourceList,
		}
	}

	return components.NewList(resourceTypes, "Resource Types", onSelect)
}
