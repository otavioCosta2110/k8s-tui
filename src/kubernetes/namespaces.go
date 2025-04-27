package kubernetes

import (
	"context"
	"log"
	"otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaces struct {
	kube KubeConfig
}

func NewNamespaces() *Namespaces {
	return &Namespaces{}
}

func fetchNamespaces(k KubeConfig) []string {
	if k.clientset == nil {
		log.Println("Error: clientset is nil")
		return []string{}
	}

	namespaces, err := k.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error fetching namespaces:", err)
		return []string{}
	}

	namespacesArray := make([]string, 0, len(namespaces.Items))
	for _, nm := range namespaces.Items {
		namespacesArray = append(namespacesArray, nm.Name)
	}
	return namespacesArray
}

func (n *Namespaces) InitComponent(k *KubeConfig) tea.Model {
	n.kube = *k
	namespaces := fetchNamespaces(*k)

	onSelect := func(selected string) tea.Msg {
		return NavigateMsg{
			NewScreen: NewResource(*k, selected).InitComponent(*k),
		}
	}

	return list.NewList(namespaces, "Namespaces", onSelect)
}
