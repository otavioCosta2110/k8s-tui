package k8s

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
)

func DeleteResource(client Client, resourceType ResourceType, namespace, name string) error {
	if IsCustomResourceType(resourceType) {
		return DeleteCustomResource(client, resourceType, namespace, name)
	}

	switch resourceType {
	case ResourceTypePod:
		return DeletePod(client, namespace, name)
	case ResourceTypeDeployment:
		return DeleteDeployment(client, namespace, name)
	case ResourceTypeReplicaSet:
		return DeleteReplicaSet(client, namespace, name)
	case ResourceTypeConfigMap:
		return DeleteConfigmap(client, namespace, name)
	case ResourceTypeIngress:
		return DeleteIngress(client, namespace, name)
	case ResourceTypeService:
		return DeleteService(client, namespace, name)
	case ResourceTypeServiceAccount:
		return DeleteServiceAccount(client, namespace, name)
	case ResourceTypeSecret:
		return DeleteSecret(client, namespace, name)
	case ResourceTypeNode:
		return DeleteNode(client, name)
	case ResourceTypeJob:
		return DeleteJob(client, namespace, name)
	case ResourceTypeCronJob:
		return DeleteCronJob(client, namespace, name)
	case ResourceTypeDaemonSet:
		return DeleteDaemonSet(client, namespace, name)
	case ResourceTypeStatefulSet:
		return DeleteStatefulSet(client, namespace, name)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func ListResources(client Client, resourceType ResourceType, namespace string) ([]string, error) {
	if IsCustomResourceType(resourceType) {
		data, err := GetCustomResourceData(client, resourceType, namespace)
		if err != nil {
			return nil, err
		}
		names := make([]string, len(data))
		for i, item := range data {
			names[i] = item.GetName()
		}
		return names, nil
	}

	switch resourceType {
	case ResourceTypePod:
		pods, err := FetchPods(client, namespace, "")
		if err != nil {
			return nil, err
		}
		names := make([]string, len(pods))
		for i, pod := range pods {
			names[i] = pod.Name
		}
		return names, nil
	case ResourceTypeDeployment:
		return FetchDeploymentList(client, namespace)
	case ResourceTypeReplicaSet:
		return FetchReplicaSetList(client, namespace)
	case ResourceTypeConfigMap:
		cms, err := FetchConfigmaps(client, namespace, "")
		if err != nil {
			return nil, err
		}
		names := make([]string, len(cms))
		for i, cm := range cms {
			names[i] = cm.Name
		}
		return names, nil
	case ResourceTypeIngress:
		return FetchIngressList(client, namespace)
	case ResourceTypeService:
		return FetchServiceList(client, namespace)
	case ResourceTypeServiceAccount:
		return FetchServiceAccountList(client, namespace)
	case ResourceTypeSecret:
		return FetchSecretList(client, namespace)
	case ResourceTypeNode:
		return FetchNodeList(client)
	case ResourceTypeJob:
		return FetchJobList(client, namespace)
	case ResourceTypeCronJob:
		return FetchCronJobList(client, namespace)
	case ResourceTypeDaemonSet:
		return FetchDaemonSetList(client, namespace)
	case ResourceTypeStatefulSet:
		return FetchStatefulSetList(client, namespace)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func GetResourceInfo(client Client, resourceType ResourceType, namespace, name string) (*ResourceInfo, error) {
	if IsCustomResourceType(resourceType) {
		return GetCustomResourceInfo(client, resourceType, namespace, name)
	}

	switch resourceType {
	case ResourceTypePod:
		pods, err := FetchPods(client, namespace, "")
		if err != nil {
			return nil, err
		}
		for _, pod := range pods {
			if pod.Name == name {
				return &ResourceInfo{
					Name:      pod.Name,
					Namespace: pod.Namespace,
					Kind:      ResourceTypePod,
					Age:       pod.Age,
				}, nil
			}
		}
	case ResourceTypeDeployment:
		deployments, err := GetDeploymentsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, deployment := range deployments {
			if deployment.Name == name {
				return &ResourceInfo{
					Name:      deployment.Name,
					Namespace: deployment.Namespace,
					Kind:      ResourceTypeDeployment,
					Age:       deployment.Age,
				}, nil
			}
		}
	case ResourceTypeReplicaSet:
		replicasets, err := GetReplicaSetsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, rs := range replicasets {
			if rs.Name == name {
				return &ResourceInfo{
					Name:      rs.Name,
					Namespace: rs.Namespace,
					Kind:      ResourceTypeReplicaSet,
					Age:       rs.Age,
				}, nil
			}
		}
	case ResourceTypeConfigMap:
		cms, err := FetchConfigmaps(client, namespace, "")
		if err != nil {
			return nil, err
		}
		for _, cm := range cms {
			if cm.Name == name {
				return &ResourceInfo{
					Name:      cm.Name,
					Namespace: cm.Namespace,
					Kind:      ResourceTypeConfigMap,
					Age:       cm.Age,
				}, nil
			}
		}
	case ResourceTypeIngress:
		ingresses, err := GetIngressesTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, ingress := range ingresses {
			if ingress.Name == name {
				return &ResourceInfo{
					Name:      ingress.Name,
					Namespace: ingress.Namespace,
					Kind:      ResourceTypeIngress,
					Age:       ingress.Age,
				}, nil
			}
		}
	case ResourceTypeService:
		services, err := GetServicesTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, service := range services {
			if service.Name == name {
				return &ResourceInfo{
					Name:      service.Name,
					Namespace: service.Namespace,
					Kind:      ResourceTypeService,
					Age:       service.Age,
				}, nil
			}
		}
	case ResourceTypeServiceAccount:
		serviceaccounts, err := GetServiceAccountsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, sa := range serviceaccounts {
			if sa.Name == name {
				return &ResourceInfo{
					Name:      sa.Name,
					Namespace: sa.Namespace,
					Kind:      ResourceTypeServiceAccount,
					Age:       sa.Age,
				}, nil
			}
		}
	case ResourceTypeSecret:
		secrets, err := GetSecretsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, secret := range secrets {
			if secret.Name == name {
				return &ResourceInfo{
					Name:      secret.Name,
					Namespace: secret.Namespace,
					Kind:      ResourceTypeSecret,
					Age:       secret.Age,
				}, nil
			}
		}
	case ResourceTypeNode:
		nodes, err := GetNodesTableData(client)
		if err != nil {
			return nil, err
		}
		for _, node := range nodes {
			if node.Name == name {
				return &ResourceInfo{
					Name:      node.Name,
					Namespace: "",
					Kind:      ResourceTypeNode,
					Age:       node.Age,
				}, nil
			}
		}
	case ResourceTypeJob:
		jobs, err := GetJobsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, job := range jobs {
			if job.Name == name {
				return &ResourceInfo{
					Name:      job.Name,
					Namespace: job.Namespace,
					Kind:      ResourceTypeJob,
					Age:       job.Age,
				}, nil
			}
		}
	case ResourceTypeCronJob:
		cronjobs, err := GetCronJobsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, cronjob := range cronjobs {
			if cronjob.Name == name {
				return &ResourceInfo{
					Name:      cronjob.Name,
					Namespace: cronjob.Namespace,
					Kind:      ResourceTypeCronJob,
					Age:       cronjob.Age,
				}, nil
			}
		}
	case ResourceTypeDaemonSet:
		daemonsets, err := GetDaemonSetsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, daemonset := range daemonsets {
			if daemonset.Name == name {
				return &ResourceInfo{
					Name:      daemonset.Name,
					Namespace: daemonset.Namespace,
					Kind:      ResourceTypeDaemonSet,
					Age:       daemonset.Age,
				}, nil
			}
		}
	case ResourceTypeStatefulSet:
		statefulsets, err := GetStatefulSetsTableData(client, namespace)
		if err != nil {
			return nil, err
		}
		for _, statefulset := range statefulsets {
			if statefulset.Name == name {
				return &ResourceInfo{
					Name:      statefulset.Name,
					Namespace: statefulset.Namespace,
					Kind:      ResourceTypeStatefulSet,
					Age:       statefulset.Age,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("resource %s of type %s not found", name, resourceType)
}

var (
	GetCustomResourceDataFunc func(client Client, resourceType string, namespace string) ([]types.ResourceData, error)
	DeleteCustomResourceFunc  func(client Client, resourceType string, namespace string, name string) error
	GetCustomResourceInfoFunc func(client Client, resourceType string, namespace string, name string) (*ResourceInfo, error)
	IsCustomResourceTypeFunc  func(resourceType string) bool
)

func SetCustomResourceHandlers(
	getDataFunc func(Client, string, string) ([]types.ResourceData, error),
	deleteFunc func(Client, string, string, string) error,
	getInfoFunc func(Client, string, string, string) (*ResourceInfo, error),
	isCustomFunc func(string) bool,
) {
	GetCustomResourceDataFunc = getDataFunc
	DeleteCustomResourceFunc = deleteFunc
	GetCustomResourceInfoFunc = getInfoFunc
	IsCustomResourceTypeFunc = isCustomFunc
}

func IsCustomResourceType(resourceType ResourceType) bool {
	if IsCustomResourceTypeFunc != nil {
		return IsCustomResourceTypeFunc(string(resourceType))
	}
	return false
}

func GetCustomResourceData(client Client, resourceType ResourceType, namespace string) ([]types.ResourceData, error) {
	if GetCustomResourceDataFunc != nil {
		return GetCustomResourceDataFunc(client, string(resourceType), namespace)
	}
	return nil, fmt.Errorf("custom resource handler not set")
}

func DeleteCustomResource(client Client, resourceType ResourceType, namespace string, name string) error {
	if DeleteCustomResourceFunc != nil {
		return DeleteCustomResourceFunc(client, string(resourceType), namespace, name)
	}
	return fmt.Errorf("custom resource handler not set")
}

func GetCustomResourceInfo(client Client, resourceType ResourceType, namespace string, name string) (*ResourceInfo, error) {
	if GetCustomResourceInfoFunc != nil {
		return GetCustomResourceInfoFunc(client, string(resourceType), namespace, name)
	}
	return nil, fmt.Errorf("custom resource handler not set")
}

func DescribeResource(client Client, resourceType ResourceType, namespace, name string) (string, error) {
	if IsCustomResourceType(resourceType) {
		return "", fmt.Errorf("custom resource description not implemented")
	}

	switch resourceType {
	case ResourceTypePod:
		pod := NewPod(name, namespace, client)
		return pod.Describe()
	case ResourceTypeService:
		service := NewService(name, namespace, client)
		return service.Describe()
	case ResourceTypeConfigMap:
		configmap := NewConfigmap(name, namespace, client)
		return configmap.Describe()
	case ResourceTypeSecret:
		secret := NewSecret(name, namespace, client)
		return secret.Describe()
	case ResourceTypeIngress:
		ingress := NewIngress(name, namespace, client)
		return ingress.Describe()
	case ResourceTypeJob:
		job := NewJob(name, namespace, client)
		return job.Describe()
	case ResourceTypeCronJob:
		cronjob := NewCronJob(name, namespace, client)
		return cronjob.Describe()
	case ResourceTypeDaemonSet:
		daemonset := NewDaemonSet(name, namespace, client)
		return daemonset.Describe()
	case ResourceTypeStatefulSet:
		statefulset := NewStatefulSet(name, namespace, client)
		return statefulset.Describe()
	case ResourceTypeNode:
		node := NewNode(name, client)
		return node.Describe()
	case ResourceTypeServiceAccount:
		serviceaccount := NewServiceAccount(name, namespace, client)
		return serviceaccount.Describe()
	default:
		return "", fmt.Errorf("unsupported resource type for description: %s", resourceType)
	}
}

func GetResourceLogs(client Client, resourceType ResourceType, namespace, name string) (string, error) {
	if IsCustomResourceType(resourceType) {
		return "", fmt.Errorf("custom resource logs not implemented")
	}

	switch resourceType {
	case ResourceTypePod:
		pod := NewPod(name, namespace, client)
		return pod.GetLogs()
	default:
		return "", fmt.Errorf("logs not supported for resource type: %s", resourceType)
	}
}

func ExecResource(client Client, resourceType ResourceType, namespace, name string, command []string) (string, string, error) {
	if IsCustomResourceType(resourceType) {
		return "", "", fmt.Errorf("custom resource exec not implemented")
	}

	switch resourceType {
	case ResourceTypePod:
		pod := NewPod(name, namespace, client)
		return pod.Exec(command)
	default:
		return "", "", fmt.Errorf("exec not supported for resource type: %s", resourceType)
	}
}
