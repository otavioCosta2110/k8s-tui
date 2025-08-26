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
	case ResourceTypeIngress:
		return DeleteIngress(client, namespace, name)
	case ResourceTypeService:
		return DeleteService(client, namespace, name)
	case ResourceTypeSecret:
		return DeleteSecret(client, namespace, name)
	case ResourceTypeNode:
		return DeleteNode(client, name) 
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
	case ResourceTypeIngress:
		return FetchIngressList(client, namespace)
	case ResourceTypeService:
		return FetchServiceList(client, namespace)
	case ResourceTypeSecret:
		return FetchSecretList(client, namespace)
	case ResourceTypeNode:
		return FetchNodeList(client) 
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
	}
	return nil, fmt.Errorf("resource %s of type %s not found", name, resourceType)
}
