package kubernetes

import (
	"context"
	"otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceList struct {
	kube       KubeConfig
	namespace  string
	resourceType string
	items      []string
}

func NewResourceList(k KubeConfig, namespace, resourceType string) ResourceList {
	return ResourceList{
		kube:        k,
		namespace:   namespace,
		resourceType: resourceType,
	}
}

func (rl ResourceList) InitComponent(k KubeConfig) tea.Model {
	var items []string
	
	switch rl.resourceType {
	case "Pods":
		pods, err := rl.kube.clientset.CoreV1().Pods(rl.namespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				items = append(items, pod.Name)
			}
		}
	case "Deployments":
		deployments, err := rl.kube.clientset.AppsV1().Deployments(rl.namespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, deploy := range deployments.Items {
				items = append(items, deploy.Name)
			}
		}
	}

	onSelect := func(selected string) tea.Msg {
		return nil
	}

	return list.NewList(items, rl.resourceType, onSelect)
}
