package models

import (
	"context"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (rl ResourceList) InitComponent(k k8s.Client) tea.Model {
	var items []string
	var onSelect func(string) tea.Msg

	switch rl.resourceType {
	case "Pods":
		pods, _ := NewPods(rl.kube, rl.namespace)

		c, _ := pods.InitComponent(&k)
		return c

	case "Deployments":
		deployments, err := rl.kube.Clientset.AppsV1().Deployments(rl.namespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, deploy := range deployments.Items {
				items = append(items, deploy.Name)
			}
			onSelect = func(selected string) tea.Msg {
				return nil
			}
		}
	}

	if onSelect == nil {
		onSelect = func(selected string) tea.Msg {
			return nil
		}
	}

	return components.NewList(items, rl.resourceType, onSelect)
}
