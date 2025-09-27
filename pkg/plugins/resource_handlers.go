package plugins

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)


type ResourceHandler interface {
	
	Get(client k8s.Client, namespace string) (interface{}, error)

	
	Delete(client k8s.Client, namespace, name string) error

	
	Describe(client k8s.Client, namespace, name string) (string, error)

	
	GetType() k8s.ResourceType
}


type DefaultResourceHandler struct {
	ResourceType k8s.ResourceType
}


func (h *DefaultResourceHandler) Get(client k8s.Client, namespace string) (interface{}, error) {
	switch h.ResourceType {
	case k8s.ResourceTypePod:
		return k8s.FetchPods(client, namespace, "")
	case k8s.ResourceTypeService:
		return k8s.GetServicesTableData(client, namespace)
	case k8s.ResourceTypeDeployment:
		return k8s.GetDeploymentsTableData(client, namespace)
	case k8s.ResourceTypeConfigMap:
		return k8s.FetchConfigmaps(client, namespace, "")
	case k8s.ResourceTypeSecret:
		return k8s.GetSecretsTableData(client, namespace)
	case k8s.ResourceTypeIngress:
		return k8s.GetIngressesTableData(client, namespace)
	case k8s.ResourceTypeJob:
		return k8s.GetJobsTableData(client, namespace)
	case k8s.ResourceTypeCronJob:
		return k8s.GetCronJobsTableData(client, namespace)
	case k8s.ResourceTypeDaemonSet:
		return k8s.GetDaemonSetsTableData(client, namespace)
	case k8s.ResourceTypeStatefulSet:
		return k8s.GetStatefulSetsTableData(client, namespace)
	case k8s.ResourceTypeReplicaSet:
		return k8s.GetReplicaSetsTableData(client, namespace)
	case k8s.ResourceTypeNode:
		return k8s.GetNodesTableData(client)
	case k8s.ResourceTypeServiceAccount:
		return k8s.GetServiceAccountsTableData(client, namespace)
	default:
		return nil, ErrResourceTypeNotSupported{ResourceType: h.ResourceType}
	}
}


func (h *DefaultResourceHandler) Delete(client k8s.Client, namespace, name string) error {
	return k8s.DeleteResource(client, h.ResourceType, namespace, name)
}


func (h *DefaultResourceHandler) Describe(client k8s.Client, namespace, name string) (string, error) {
	return k8s.DescribeResource(client, h.ResourceType, namespace, name)
}


func (h *DefaultResourceHandler) GetType() k8s.ResourceType {
	return h.ResourceType
}


type ResourceRegistry struct {
	handlers map[k8s.ResourceType]ResourceHandler
}


func NewResourceRegistry() *ResourceRegistry {
	registry := &ResourceRegistry{
		handlers: make(map[k8s.ResourceType]ResourceHandler),
	}

	
	defaultTypes := []k8s.ResourceType{
		k8s.ResourceTypePod,
		k8s.ResourceTypeService,
		k8s.ResourceTypeDeployment,
		k8s.ResourceTypeConfigMap,
		k8s.ResourceTypeSecret,
		k8s.ResourceTypeIngress,
		k8s.ResourceTypeJob,
		k8s.ResourceTypeCronJob,
		k8s.ResourceTypeDaemonSet,
		k8s.ResourceTypeStatefulSet,
		k8s.ResourceTypeReplicaSet,
		k8s.ResourceTypeNode,
		k8s.ResourceTypeServiceAccount,
	}

	for _, resourceType := range defaultTypes {
		registry.handlers[resourceType] = &DefaultResourceHandler{ResourceType: resourceType}
	}

	return registry
}


func (r *ResourceRegistry) RegisterHandler(resourceType k8s.ResourceType, handler ResourceHandler) {
	r.handlers[resourceType] = handler
}


func (r *ResourceRegistry) GetHandler(resourceType k8s.ResourceType) (ResourceHandler, bool) {
	handler, exists := r.handlers[resourceType]
	return handler, exists
}


func (r *ResourceRegistry) GetResource(client k8s.Client, resourceType k8s.ResourceType, namespace string) (interface{}, error) {
	if handler, exists := r.handlers[resourceType]; exists {
		return handler.Get(client, namespace)
	}
	return nil, ErrResourceTypeNotSupported{ResourceType: resourceType}
}


func (r *ResourceRegistry) DeleteResource(client k8s.Client, resourceType k8s.ResourceType, namespace, name string) error {
	if handler, exists := r.handlers[resourceType]; exists {
		return handler.Delete(client, namespace, name)
	}
	return ErrResourceTypeNotSupported{ResourceType: resourceType}
}


func (r *ResourceRegistry) DescribeResource(client k8s.Client, resourceType k8s.ResourceType, namespace, name string) (string, error) {
	if handler, exists := r.handlers[resourceType]; exists {
		return handler.Describe(client, namespace, name)
	}
	return "", ErrResourceTypeNotSupported{ResourceType: resourceType}
}


func (r *ResourceRegistry) GetSupportedTypes() []k8s.ResourceType {
	types := make([]k8s.ResourceType, 0, len(r.handlers))
	for resourceType := range r.handlers {
		types = append(types, resourceType)
	}
	return types
}


type ErrResourceTypeNotSupported struct {
	ResourceType k8s.ResourceType
}

func (e ErrResourceTypeNotSupported) Error() string {
	return "resource type not supported: " + string(e.ResourceType)
}
