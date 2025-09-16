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
	QuickNavKey string
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
	}, "p")

	rf.registerResource("Deployments", createDeploymentsModel, ResourceMetadata{
		Name:        "Deployments",
		Description: "Manage application deployments and scaling",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes deployments",
	}, "d")

	rf.registerResource("Services", createServicesModel, ResourceMetadata{
		Name:        "Services",
		Description: "Network services and load balancing",
		Category:    "Networking",
		HelpText:    "View and manage Kubernetes services",
	}, "s")

	rf.registerResource("Ingresses", createIngressesModel, ResourceMetadata{
		Name:        "Ingresses",
		Description: "HTTP routing and ingress controllers",
		Category:    "Networking",
		HelpText:    "View and manage Kubernetes ingresses",
	}, "i")

	rf.registerResource("ConfigMaps", createConfigMapsModel, ResourceMetadata{
		Name:        "ConfigMaps",
		Description: "Configuration data and environment variables",
		Category:    "Configuration",
		HelpText:    "View and manage Kubernetes configmaps",
	}, "c")

	rf.registerResource("Secrets", createSecretsModel, ResourceMetadata{
		Name:        "Secrets",
		Description: "Sensitive configuration and credentials",
		Category:    "Configuration",
		HelpText:    "View and manage Kubernetes secrets securely",
	}, "e")

	rf.registerResource("ServiceAccounts", createServiceAccountsModel, ResourceMetadata{
		Name:        "ServiceAccounts",
		Description: "Service accounts for API access and authentication",
		Category:    "Configuration",
		HelpText:    "View and manage Kubernetes service accounts",
	}, "a")

	rf.registerResource("ReplicaSets", createReplicaSetsModel, ResourceMetadata{
		Name:        "ReplicaSets",
		Description: "Pod replication and scaling controllers",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes replica sets",
	}, "r")

	rf.registerResource("Nodes", createNodesModel, ResourceMetadata{
		Name:        "Nodes",
		Description: "Kubernetes cluster nodes and their resources",
		Category:    "Infrastructure",
		HelpText:    "View and manage Kubernetes cluster nodes",
	}, "n")

	rf.registerResource("Jobs", createJobsModel, ResourceMetadata{
		Name:        "Jobs",
		Description: "Batch processing jobs and scheduled tasks",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes jobs",
	}, "j")

	rf.registerResource("CronJobs", createCronJobsModel, ResourceMetadata{
		Name:        "CronJobs",
		Description: "Scheduled jobs that run periodically",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes cron jobs",
	}, "k")

	rf.registerResource("DaemonSets", createDaemonSetsModel, ResourceMetadata{
		Name:        "DaemonSets",
		Description: "Pods that run on all cluster nodes",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes daemon sets",
	}, "m")

	rf.registerResource("StatefulSets", createStatefulSetsModel, ResourceMetadata{
		Name:        "StatefulSets",
		Description: "Stateful applications with persistent storage",
		Category:    "Workloads",
		HelpText:    "View and manage Kubernetes stateful sets",
	}, "t")
}

func (rf *ResourceFactory) registerResource(resourceType string, creator ResourceCreator, metadata ResourceMetadata, quickNavKey string) {
	rf.registry[resourceType] = creator
	metadata.QuickNavKey = quickNavKey
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

func (rf *ResourceFactory) GetQuickNavMappings() map[string]string {
	mappings := make(map[string]string)
	for resourceType, metadata := range rf.metadata {
		if metadata.QuickNavKey != "" {
			mappings[metadata.QuickNavKey] = resourceType
		}
	}
	return mappings
}

func (rf *ResourceFactory) GetSortedQuickNavMappings() []struct{ Key, ResourceType string } {
	var mappings []struct{ Key, ResourceType string }
	for resourceType, metadata := range rf.metadata {
		if metadata.QuickNavKey != "" {
			mappings = append(mappings, struct{ Key, ResourceType string }{metadata.QuickNavKey, resourceType})
		}
	}

	for i := 0; i < len(mappings)-1; i++ {
		for j := i + 1; j < len(mappings); j++ {
			if mappings[i].ResourceType > mappings[j].ResourceType {
				mappings[i], mappings[j] = mappings[j], mappings[i]
			}
		}
	}

	return mappings
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

func createServiceAccountsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewServiceAccounts(k, namespace)
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

func createJobsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewJobs(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createCronJobsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewCronJobs(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createDaemonSetsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewDaemonSets(k, namespace)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func createStatefulSetsModel(k k8s.Client, namespace string) (ResourceModel, error) {
	model, err := NewStatefulSets(k, namespace)
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
