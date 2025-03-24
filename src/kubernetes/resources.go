package kubernetes

import (
	listcomponent "otaviocosta2110/k8s-tui/src/components/list"

	tea "github.com/charmbracelet/bubbletea"
)

type Resources struct {
	name       string
	kubeconfig KubeConfig
}

var kubernetesResources = map[string]ResourceInterface{
	"Pod": nil,
	"ReplicationController": nil,
	"ReplicaSet": nil,
	"Deployment": nil,
	"StatefulSet": nil,
	"DaemonSet": nil,
	"Job": nil,
	"CronJob": nil,
	"Service": nil,
	"Endpoints": nil,
	"EndpointSlice": nil,
	"Ingress": nil,
	"NetworkPolicy": nil,
	"ConfigMap": nil,
	"Secret": nil,
	"PersistentVolume": nil,
	"PersistentVolumeClaim": nil,
	"StorageClass": nil,
	"VolumeSnapshot": nil,
	"VolumeSnapshotClass": nil,
  "Namespace": NewNamespaces(),
	"Node": nil,
	"ServiceAccount": nil,
	"Role": nil,
	"ClusterRole": nil,
	"RoleBinding": nil,
	"ClusterRoleBinding": nil,
	"CustomResourceDefinition": nil,
	"HorizontalPodAutoscaler": nil,
	"PodDisruptionBudget": nil,
	"LimitRange": nil,
	"ResourceQuota": nil,
	"Lease": nil,
	"CSINode": nil,
	"CSIStorageCapacity": nil,
	"MutatingWebhookConfiguration": nil,
	"ValidatingWebhookConfiguration": nil,
	"FlowSchema": nil,
	"PriorityLevelConfiguration": nil,
}

func NewResource(k KubeConfig) Resources {
  return Resources{kubeconfig: k}
}

func (r Resources) InitComponent(k KubeConfig) tea.Model {
  r.kubeconfig = k
	onSelect := func(selected string) tea.Model {
    if kubernetesResources[selected] == nil{
      return nil
    }
    return kubernetesResources[selected].InitComponent(r.kubeconfig)
	}
  var newList []string

  for item := range kubernetesResources{
    newList = append(newList, item)
  }

	list := listcomponent.NewList(newList, "Resources", onSelect)

	return list
}
