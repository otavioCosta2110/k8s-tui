package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type ResourceList struct {
	kube         k8s.Client
	namespace    string
	resourceType string
}

func NewResourceList(k k8s.Client, namespace, resourceType string) ResourceList {
	return ResourceList{
		kube:         k,
		namespace:    namespace,
		resourceType: resourceType,
	}
}

func (rl ResourceList) InitComponent(k k8s.Client) (tea.Model, error) {
	var items []string
	var onSelect func(string) tea.Msg

	switch rl.resourceType {
	case "Pods":
		pods, err := NewPods(rl.kube, rl.namespace, nil)
		if err != nil {
			return nil, err
		}

		c, err := pods.InitComponent(&k)
		if err != nil {
			return nil, err
		}
		return c, nil

	case "Deployments":
		deployments, err := NewDeployments(rl.kube, rl.namespace)
		if err != nil {
			return nil, err
		}
		deploymentsComponent, err := deployments.InitComponent(&k)
		if err != nil {
			return nil, err
		}
		return deploymentsComponent, nil
	
	case "ConfigMaps":
		configMaps, err := NewConfigmaps(rl.kube, rl.namespace, nil)
		if err != nil {
			return nil, err
		}
		configMapsComponent, err := configMaps.InitComponent(&k)
		if err != nil {
			return nil, err
		}
		return configMapsComponent, nil
	}

	return components.NewList(items, rl.resourceType, onSelect), nil
}
