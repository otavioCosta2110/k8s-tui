package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodInfo struct {
	Namespace string
	Name      string
	Ready     string
	Status    string
	Restarts  int
	Age       string
}

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

func GetPodDetails(client Client, namespace string, podName string) (PodInfo, error) {
	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(
		context.Background(),
		podName,
		metav1.GetOptions{},
	)
	if err != nil {
		return PodInfo{}, fmt.Errorf("failed to get pod details: %v", err)
	}

	readyContainers := 0
	totalContainers := len(pod.Spec.Containers)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyContainers++
		}
	}

	restarts := 0
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += int(cs.RestartCount)
	}

	age := "Unknown"
	if pod.Status.StartTime != nil {
		age = utils.FormatAge(pod.Status.StartTime.Time)
	}

	return PodInfo{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Ready:     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
		Status:    string(pod.Status.Phase),
		Restarts:  restarts,
		Age:       age,
	}, nil
}

func GetPodsTableData(client Client, namespace string, podNames []string) ([]PodInfo, error) {
	var podsInfo []PodInfo

	for _, podName := range podNames {
		info, err := GetPodDetails(client, namespace, podName)
		if err != nil {
			return nil, err
		}
		podsInfo = append(podsInfo, info)
	}

	return podsInfo, nil
}
