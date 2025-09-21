package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type kubeconfigModel struct {
	configs    []string
	k8sClient  *k8s.Client
	kubeconfig string
	loading    bool
	err        error
}

func NewKubeconfigModel() *kubeconfigModel {
	return &kubeconfigModel{
		configs:    styles.GetKubeconfigsLocations(),
		k8sClient:  nil,
		kubeconfig: "",
		loading:    true,
		err:        nil,
	}
}

func (k kubeconfigModel) InitComponent(_ *k8s.Client) (tea.Model, error) {
	styles.IsHeaderActive = false
	var items []string
	for _, configs := range styles.GetKubeconfigsLocations() {
		kubeconfigs, err := os.ReadDir(configs)
		if err != nil {
			logger.Warn(fmt.Sprintf("Warning: %v", err))
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
		k.kubeconfig = selected
		os.Setenv("KUBECONFIG", selected)
		os.Setenv("KUBERNETES_MASTER", selected)
		c, err := k8s.NewClient(selected, "")
		k.k8sClient = c
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating clientset: %v", err))
			return components.NavigateMsg{
				Error: err,
			}
		}

		namespaces, err := NewNamespaces(*k.k8sClient)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k.k8sClient,
			}
		}

		nm, err := namespaces.InitComponent(k.k8sClient)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k.k8sClient,
			}
		}
		return components.NavigateMsg{
			NewScreen: nm,
			Cluster:   *k.k8sClient,
		}
	}

	return components.NewList(items, "Kubeconfigs", onSelect), nil
}
