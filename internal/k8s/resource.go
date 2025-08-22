package k8s

import (
	"fmt"
)

func DeleteResource(client Client, resourceType ResourceType, namespace, name string) error {
	switch resourceType {
	case ResourceTypePod:
		return DeletePod(client, namespace, name)
	case ResourceTypeDeployment:
		return DeleteDeployment(client, namespace, name)
	case ResourceTypeReplicaSet:
		return DeleteReplicaSet(client, namespace, name)
	case ResourceTypeConfigMap:
		return DeleteConfigmap(client, namespace, name)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func ListResources(client Client, resourceType ResourceType, namespace string) ([]string, error) {
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
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func GetResourceInfo(client Client, resourceType ResourceType, namespace, name string) (*ResourceInfo, error) {
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
	}
	return nil, fmt.Errorf("resource %s of type %s not found", name, resourceType)
}
