package models

import (
	"context"
	"otaviocosta2110/k8s-tui/internal/k8s"

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
		pods, err := rl.kube.CoreV1().Pods(rl.namespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				items = append(items, pod.Name)
			}
			onSelect = func(selected string) tea.Msg {
				pod := NewPod(selected, rl.namespace, rl.kube.clientset, rl.kube.config)
				pod.Fetch()

				terminal := terminal.NewTerminal("Pod Terminal", "kubectl", "exec", "-it", selected, "-n", rl.namespace, "--", "sh")
				return NavigateMsg{
					NewScreen: terminal,
					Cluster: rl.kube,
				}
			}

			// return nil
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

