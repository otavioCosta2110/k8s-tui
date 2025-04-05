package kubernetes

import (
	"log"
	"os"
	"path/filepath"
	"otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeConfig struct {
	clientset  *kubernetes.Clientset
	Kubeconfig string
}

type NavigateMsg struct {
	NewScreen tea.Model
}

func NewKubeConfig() KubeConfig {
	return KubeConfig{}
}

func (k *KubeConfig) setClientset() error {
	configuration, err := clientcmd.BuildConfigFromFlags("", k.Kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(configuration)
	if err != nil {
		return err
	}

	k.clientset = clientset
	return nil
}

func (k KubeConfig) InitComponent(_ KubeConfig) tea.Model {
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
		if err := k.setClientset(); err != nil {
			log.Println("Error creating clientset:", err)
			return nil
		}
		return NavigateMsg{NewScreen: NewNamespaces().InitComponent(k)}
	}

	return list.NewList(items, "Kubeconfigs", onSelect)
}
