package kubernetes

import (
	"log"
	"os"
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"
	"otaviocosta2110/k8s-tui/src/global"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// must assign Kubeconfig value after
// the user selects the kubeconfig file
type KubeConfig struct {
	clientset  *kubernetes.Clientset
	Kubeconfig string
}

func NewKubeConfig() KubeConfig {
	k := KubeConfig{}
	return k
}

func (k *KubeConfig) setClientset() {
	configuration, err := clientcmd.BuildConfigFromFlags("", k.Kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(configuration)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	k.clientset = clientset
}

func (k KubeConfig) InitComponent(_ KubeConfig) (tea.Model){
  var items []string
	for _, configs := range global.GetKubeconfigsLocations() {
		kubeconfigs, err := os.ReadDir(configs)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range kubeconfigs {
			if !file.IsDir() {
        fullPath := filepath.Join(configs, file.Name())
				items = append(items, fullPath)
			}
		}
	}

	onSelect := func(selected string)(tea.Model) {
    k.Kubeconfig = selected
    k.setClientset()
    // n := NewNamespaces()
    r := NewResource(k)
    return r.InitComponent(k)
	}

  list := listcomponent.NewList(items, "Kubeconfigs", onSelect)

  return list
}
