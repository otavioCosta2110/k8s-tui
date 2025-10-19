package k8s

import (
	"context"
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"gopkg.in/yaml.v3"

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

func (d *DeploymentInfo) Fetch() error {
	deployment, err := d.Client.Clientset.AppsV1().Deployments(d.Namespace).Get(
		context.Background(),
		d.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}
	d.Raw = deployment
	return nil
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
		status := deployment.Status
		spec := deployment.Spec

		var desiredReplicas int32
		if spec.Replicas != nil {
			desiredReplicas = *spec.Replicas
		}

		readyStr := fmt.Sprintf("%d/%d", status.ReadyReplicas, desiredReplicas)

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
			Ready:     readyStr,
			UpToDate:  fmt.Sprintf("%d", status.UpdatedReplicas),
			Available: fmt.Sprintf("%d", status.AvailableReplicas),
			Age:       format.FormatAge(deployment.CreationTimestamp.Time),
			Raw:       deployment.DeepCopy(),
			Client:    client,
		})
	}

	return deploymentInfos, nil
}

func (d *DeploymentInfo) GetPods() ([]PodInfo, error) {
	selector, err := d.GetLabelSelector()
	if err != nil {
		return nil, err
	}
	pods, err := FetchPods(d.Client, d.Namespace, selector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (d *DeploymentInfo) GetLabelSelector() (string, error) {
	if d.Raw == nil {
		return "", fmt.Errorf("deployment raw data not available")
	}

	if d.Raw.Spec.Selector == nil {
		return "", fmt.Errorf("deployment has no selector")
	}

	requirements, err := metav1.LabelSelectorAsSelector(d.Raw.Spec.Selector)
	if err != nil {
		return "", fmt.Errorf("failed to convert label selector: %v", err)
	}

	return requirements.String(), nil
}

func (d *DeploymentInfo) Describe() (string, error) {
	if d.Raw == nil {
		if err := d.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch deployment: %v", err)
		}
	}

	yamlData, err := yaml.Marshal(d.Raw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (d *DeploymentInfo) Apply(yamlContent string) error {
	var deployment appsv1.Deployment
	if err := yaml.Unmarshal([]byte(yamlContent), &deployment); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	// Ensure the name and namespace match the current deployment
	deployment.Name = d.Name
	deployment.Namespace = d.Namespace

	_, err := d.Client.Clientset.AppsV1().Deployments(d.Namespace).Update(
		context.Background(),
		&deployment,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}

func DeleteDeployment(client Client, namespace string, deploymentName string) error {
	err := client.Clientset.AppsV1().Deployments(namespace).Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s: %v", deploymentName, err)
	}
	return nil
}
