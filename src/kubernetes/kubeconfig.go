package kubernetes

import (
	"log"
	"os"
	"otaviocosta2110/k8s-tui/src/components/list"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeConfig struct {
	clientset  *kubernetes.Clientset
	Kubeconfig string
	error      error
}

type NavigateMsg struct {
	NewScreen tea.Model
	Cluster   KubeConfig
	Error     error
}

func NewKubeConfig() KubeConfig {
	return KubeConfig{}
}

func (k *KubeConfig) setClientset() error {
	configuration, err := clientcmd.BuildConfigFromFlags("", k.Kubeconfig)
	if err != nil {
		k.error = err 
		return err
	}

	clientset, err := kubernetes.NewForConfig(configuration)
	if err != nil {
		k.error = err 
		return err
	}

	k.clientset = clientset
	return nil
}

func (k KubeConfig) InitComponent(_ *KubeConfig) (tea.Model, error) {
	var items []string
	for _, configs := range GetKubeconfigsLocations() {
		kubeconfigs, err := os.ReadDir(configs)
		if err != nil {
			log.Println("Warning:", err)
			continue
		}
		for _, file := range kubeconfigs {
			if !file.IsDir() {
				fullPath := filepath.Join(configs, file.Name())
				items = append(items, fullPath)
			}
		}
	}

	onSelect := func(selected string) tea.Msg {
		k.Kubeconfig = selected
		os.Setenv("KUBECONFIG", selected)
		os.Setenv("KUBERNETES_MASTER", selected)
		if err := k.setClientset(); err != nil {
			log.Println("Error creating clientset:", err)
			return NavigateMsg{
				Error: err,
				Cluster: k,
			}
		}

		namespaces, err := NewNamespaces().InitComponent(&k)
		if err != nil {
			return NavigateMsg{
				Error: err,
				Cluster: k,
			}
		}
		return NavigateMsg{
			NewScreen: namespaces,
			Cluster: k,
		}
	}

	return list.NewList(items, "Kubeconfigs", onSelect), nil
}
