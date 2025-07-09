package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FetchPods(client Client, namespace string, selector string) ([]string, error) {
	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}
	pods, err := client.Clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %v", err)
	}

	podNames := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}
