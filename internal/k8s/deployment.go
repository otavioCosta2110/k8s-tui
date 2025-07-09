package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment struct {
	Name      string
	Namespace string
	Raw       *appsv1.Deployment
	Client    Client
}

func NewDeployment(name, namespace string, k Client) *Deployment {
	return &Deployment{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchDeploymentList(client Client, namespace string) ([]string, error) {
	ds, err := client.Clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %v", err)
	}

	deploymentNames := make([]string, 0, len(ds.Items))
	for _, deployment := range ds.Items {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	return deploymentNames, nil
}

func (d *Deployment) GetPods() ([]string, error) {
	if d.Raw == nil {
		if err := d.Fetch(); err != nil {
			return nil, fmt.Errorf("failed to fetch deployment: %v", err)
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(d.Raw.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("failed to create selector: %v", err)
	}

	pods, err := FetchPods(d.Client, d.Namespace, selector.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	return pods, nil
}

func (d *Deployment) Fetch() error {
	deployment, err := d.Client.Clientset.AppsV1().Deployments(d.Namespace).Get(context.Background(), d.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}
	d.Raw = deployment
	return nil
}
