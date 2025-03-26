package kubernetes

import (
	"context"
	"log"
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaces struct {
	items []string
}

func NewNamespaces() Namespaces {
	return Namespaces{}
}

func fetchNamespaces(k KubeConfig) []string {
	namespaces, err := k.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		log.Fatal(err)
	}
	var namespacesArray []string

	for _, nm := range namespaces.Items {
		namespacesArray = append(namespacesArray, nm.Name)
	}

	return namespacesArray
}

func (n Namespaces) InitComponent(k KubeConfig) tea.Model {
	namespaces := fetchNamespaces(k)

	onSelect := func(selected string) tea.Model {
		r := NewResource(k)
		return r.InitComponent(k)
	}

	n.items = namespaces

	list := listcomponent.NewList(n.items, "Namespaces", onSelect)

	return list
}
