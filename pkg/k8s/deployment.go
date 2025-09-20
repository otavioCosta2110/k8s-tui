package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentInfo struct {
	Namespace string
	Name      string
	Ready     string
	UpToDate  string
	Available string
	Age       string
	Raw       *appsv1.Deployment
	Client    Client
}

func NewDeployment(name, namespace string, k Client) *DeploymentInfo {
	return &DeploymentInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchDeploymentList(client Client, namespace string) ([]string, error) {
	ds, err := client.Clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployments: %v", err)
	}

	deploymentNames := make([]string, 0, len(ds.Items))
	for _, deployment := range ds.Items {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	return deploymentNames, nil
}

func GetDeploymentsTableData(client Client, namespace string) ([]DeploymentInfo, error) {
	deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		freshDeployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(
			context.Background(),
			deployment.Name,
			metav1.GetOptions{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get fresh status for deployment %s: %v", deployment.Name, err)
		}

		status := freshDeployment.Status
		spec := freshDeployment.Spec

		var desiredReplicas int32
		if spec.Replicas != nil {
			desiredReplicas = *spec.Replicas
		}

		readyStr := fmt.Sprintf("%d/%d", status.ReadyReplicas, desiredReplicas)

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Namespace: freshDeployment.Namespace,
			Name:      freshDeployment.Name,
			Ready:     readyStr,
			UpToDate:  fmt.Sprintf("%d", status.UpdatedReplicas),
			Available: fmt.Sprintf("%d", status.AvailableReplicas),
			Age:       utils.FormatAge(freshDeployment.CreationTimestamp.Time),
			Raw:       freshDeployment.DeepCopy(),
			Client:    client,
		})
	}

	return deploymentInfos, nil
}

func (d *DeploymentInfo) GetPods() ([]PodInfo, error) {
	pods, err := FetchPods(d.Client, d.Namespace, fmt.Sprintf("app=%s", d.Name))
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func DeleteDeployment(client Client, namespace string, deploymentName string) error {
	err := client.Clientset.AppsV1().Deployments(namespace).Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s: %v", deploymentName, err)
	}
	return nil
}
