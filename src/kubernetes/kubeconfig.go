package kubernetes

import (
	"log"
	"os"
	"otaviocosta2110/k8s-tui/src/global"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// must assign Kubeconfig value after
// the user selects the kubeconfig file
type KubeConfig struct {
	clientset  *kubernetes.Clientset
	Kubeconfig string
}

func NewKubeConfig(kubeconfig string) KubeConfig {
	k := KubeConfig{}
	k.setClientset()
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

func InitList() ([]string){
  var items []string
	for _, configs := range global.GetKubeconfigsLocations() {
		kubeconfigs, err := os.ReadDir(configs)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range kubeconfigs {
			if !file.IsDir() {
				items = append(items, file.Name())
			}
		}
	}

  return items
}
