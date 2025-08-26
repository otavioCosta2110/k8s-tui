package k8s

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceType string

const (
	ResourceTypePod            ResourceType = "pod"
	ResourceTypeDeployment     ResourceType = "deployment"
	ResourceTypeReplicaSet     ResourceType = "replicaset"
	ResourceTypeConfigMap      ResourceType = "configmap"
	ResourceTypeService        ResourceType = "service"
	ResourceTypeServiceAccount ResourceType = "serviceaccount"
	ResourceTypeIngress        ResourceType = "ingress"
	ResourceTypeSecret         ResourceType = "secret"
	ResourceTypeNode           ResourceType = "node"
)

type ResourceInfo struct {
	Name      string
	Namespace string
	Kind      ResourceType
	Age       string
	CreatedAt time.Time
}

type ResourceManager interface {
	GetName() string
	GetNamespace() string
	GetKind() ResourceType
	Delete() error
	GetPods() ([]PodInfo, error)
}

type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Namespace string
}

func NewClient(kubeconfigPath string, namespace string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		return nil, nil
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
		Namespace: namespace,
	}, nil
}
