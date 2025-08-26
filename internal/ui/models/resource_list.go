package models

import (
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ResourceCreator func(k k8s.Client, namespace string) (ResourceModel, error)

type ResourceMetadata struct {
	Name        string
	Description string
	Category    string
	HelpText    string
}

type ResourceModel interface {
	InitComponent(k *k8s.Client) (tea.Model, error)
}

type ResourceFactory struct {
	registry   map[string]ResourceCreator
	metadata   map[string]ResourceMetadata
	validTypes []string
}

var resourceFactory *ResourceFactory

func init() {
	resourceFactory = NewResourceFactory()
}

func NewResourceFactory() *ResourceFactory {
	rf := &ResourceFactory{
		registry: make(map[string]ResourceCreator),
		metadata: make(map[string]ResourceMetadata),
	}

	rf.registerResources()

	return rf
}

func (rf *ResourceFactory) registerResources() {
	rf.registerResource("Pods", createPodsModel, ResourceMetadata{
		Name:        "Pods",
		Description: "Container workloads running in the cluster",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes pods",
	})

	rf.registerResource("Deployments", createDeploymentsModel, ResourceMetadata{
		Name:        "Deployments",
		Description: "Manage application deployments and scaling",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes deployments",
	})

	rf.registerResource("Services", createServicesModel, ResourceMetadata{
		Name:        "Services",
		Description: "Network services and load balancing",
		Category:    "Networking",
		HelpText:    "View and manage Kubernetes services",
	})

	rf.registerResource("Ingresses", createIngressesModel, ResourceMetadata{
		Name:        "Ingresses",
		Description: "HTTP routing and ingress controllers",
		Category:    "Networking",
		HelpText:    "View and manage Kubernetes ingresses",
	})

	rf.registerResource("ConfigMaps", createConfigMapsModel, ResourceMetadata{
		Name:        "ConfigMaps",
		Description: "Configuration data and environment variables",
		Category:    "Configuration",
		HelpText:    "View and manage Kubernetes configmaps",
	})

	rf.registerResource("Secrets", createSecretsModel, ResourceMetadata{
		Name:        "Secrets",
		Description: "Sensitive configuration and credentials",
		Category:    "Configuration",
		HelpText:    "View and manage Kubernetes secrets securely",
	})

	rf.registerResource("ReplicaSets", createReplicaSetsModel, ResourceMetadata{
		Name:        "ReplicaSets",
		Description: "Pod replication and scaling controllers",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes replica sets",
	})

	rf.registerResource("Nodes", createNodesModel, ResourceMetadata{
		Name:        "Nodes",
		Description: "Kubernetes cluster nodes and their resources",
		Category:    "Infrastructure",
		HelpText:    "View and manage Kubernetes cluster nodes",
	})
}

func (rf *ResourceFactory) registerResource(resourceType string, creator ResourceCreator, metadata ResourceMetadata) {
	rf.registry[resourceType] = creator
	rf.metadata[resourceType] = metadata
	rf.validTypes = append(rf.validTypes, resourceType)
}

func (rf *ResourceFactory) CreateResource(resourceType string, k k8s.Client, namespace string) (tea.Model, error) {
	creator, exists := rf.registry[resourceType]
	if !exists {
		validTypes := strings.Join(rf.validTypes, ", ")
		return nil, fmt.Errorf("unsupported resource type '%s'. Supported types: %s", resourceType, validTypes)
	}

	resourceModel, err := creator(k, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s model: %w", resourceType, err)
	}

	component, err := resourceModel.InitComponent(&k)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s component: %w", resourceType, err)
	}

	return component, nil
}

func (rf *ResourceFactory) GetValidResourceTypes() []string {
	return rf.validTypes
}

func (rf *ResourceFactory) GetResourceMetadata(resourceType string) (ResourceMetadata, bool) {
	metadata, exists := rf.metadata[resourceType]
	return metadata, exists
}

func createPodsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewPods(k, namespace, nil)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createDeploymentsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewDeployments(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createServicesModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewServices(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createIngressesModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewIngresses(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createConfigMapsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewConfigmaps(k, namespace, nil)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createSecretsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewSecrets(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createReplicaSetsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewReplicaSets(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createNodesModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewNodes(k)
	if err != nil {
		return nil, err
	}
	return model, nil
}

type ResourceList struct {
	kube         k8s.Client
	namespace    string
	resourceType string
}

func NewResourceList(k k8s.Client, namespace, resourceType string) ResourceList {
	return ResourceList{
		kube:         k,
		namespace:    namespace,
		resourceType: resourceType,
	}
}

func (rl ResourceList) InitComponent(k k8s.Client) (tea.Model, error) {
	return resourceFactory.CreateResource(rl.resourceType, rl.kube, rl.namespace)
}
