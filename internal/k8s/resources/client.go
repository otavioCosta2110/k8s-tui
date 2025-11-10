package k8s

import (
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceType string

const (
	ResourceTypePod                   ResourceType = "pod"
	ResourceTypeDeployment            ResourceType = "deployment"
	ResourceTypeReplicaSet            ResourceType = "replicaset"
	ResourceTypeConfigMap             ResourceType = "configmap"
	ResourceTypeService               ResourceType = "service"
	ResourceTypeServiceAccount        ResourceType = "serviceaccount"
	ResourceTypeIngress               ResourceType = "ingress"
	ResourceTypeSecret                ResourceType = "secret"
	ResourceTypeNode                  ResourceType = "node"
	ResourceTypeJob                   ResourceType = "job"
	ResourceTypeCronJob               ResourceType = "cronjob"
	ResourceTypeDaemonSet             ResourceType = "daemonset"
	ResourceTypeStatefulSet           ResourceType = "statefulset"
	ResourceTypePersistentVolume      ResourceType = "persistentvolume"
	ResourceTypePersistentVolumeClaim ResourceType = "persistentvolumeclaim"
	ResourceTypeEvent                 ResourceType = "event"
	ResourceTypeNetworkPolicy         ResourceType = "networkpolicy"
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
	Clientset      kubernetes.Interface
	Config         *rest.Config
	Namespace      string
	KubeconfigPath string
}

func NewClient(kubeconfigPath string, namespace string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		// Use in-cluster config or default kubeconfig locations
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fallback to default kubeconfig file
			config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
			if err != nil {
				return nil, err
			}
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	config.QPS = 50
	config.Burst = 100

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset:      clientset,
		Config:         config,
		Namespace:      namespace,
		KubeconfigPath: kubeconfigPath,
	}, nil
}

func (c *Client) SetNamespace(namespace string) {
	c.Namespace = namespace
}

// GetClusterName returns a truncated cluster name for display in tabs
func (c *Client) GetClusterName() string {
	if c.Config == nil || c.Config.Host == "" {
		return "Unknown"
	}

	host := c.Config.Host

	// Remove https:// prefix if present
	if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
	} else if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
	}

	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Truncate if too long (max 20 characters for tab display)
	if len(host) > 20 {
		host = host[:17] + "..."
	}

	return host
}
