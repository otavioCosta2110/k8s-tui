package kubernetes

import (
	"context"
	"errors"
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

func fetchNamespaces(k KubeConfig)  ([]string, error) {
	if k.clientset == nil {
		return []string{}, errors.New("clientset is nil")
	}

	namespaces, err := k.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}

	namespacesArray := make([]string, 0, len(namespaces.Items))
	for _, nm := range namespaces.Items {
		namespacesArray = append(namespacesArray, nm.Name)
	}
	return namespacesArray, nil
}

func (n *Namespaces) InitComponent(k *KubeConfig) (tea.Model, error) {
	n.kube = *k
	namespaces, err := fetchNamespaces(*k)
	if err!=nil{
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		return NavigateMsg{
			NewScreen: NewResource(*k, selected).InitComponent(*k),
		}
	}

	return list.NewList(namespaces, "Namespaces", onSelect), nil
}
