package kubernetes

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/src/components/list"
	yamlComponent "otaviocosta2110/k8s-tui/src/components/yaml"

	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceList struct {
	kube         KubeConfig
	namespace    string
	resourceType string
}

func NewResourceList(k KubeConfig, namespace, resourceType string) ResourceList {
	return ResourceList{
		kube:         k,
		namespace:    namespace,
		resourceType: resourceType,
	}
}

func (rl ResourceList) InitComponent(k KubeConfig) tea.Model {
	var items []string
	var onSelect func(string) tea.Msg

	switch rl.resourceType {
	case "Pods":
		pods, err := rl.kube.clientset.CoreV1().Pods(rl.namespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				items = append(items, pod.Name)
			}
			onSelect = func(selected string) tea.Msg {
				pod := NewPod(selected, rl.namespace, rl.kube.clientset, rl.kube.config)
				pod.Fetch()
				yaml, err := pod.DescribePod()
				if err != nil {
					return fmt.Errorf("failed to get pod YAML: %v", err)
				}
				viewer := yamlComponent.NewYAMLViewer(pod.Name, yaml)
				return NavigateMsg{
					NewScreen: viewer,
				}
			}
		}

	case "Deployments":
		deployments, err := rl.kube.clientset.AppsV1().Deployments(rl.namespace).List(context.Background(), metav1.ListOptions{})
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

	return list.NewList(items, rl.resourceType, onSelect)
}
