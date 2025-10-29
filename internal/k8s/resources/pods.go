package k8s

import (
	"context"
	"errors"
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	corev1 "k8s.io/api/core/v1"
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

func FetchPods(client Client, namespace string, selector string) ([]PodInfo, error) {
	if client.Clientset == nil {
		return nil, errors.New("kubernetes client not available")
	}
	logger.Debug("Fetching pods with selector: " + selector)
	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}

	pods, err := client.Clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %v", err)
	}

	podsInfo := make([]PodInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podInfo, err := processPodData(&pod)
		if err != nil {
			return nil, fmt.Errorf("failed to process pod %s: %v", pod.Name, err)
		}
		podsInfo = append(podsInfo, podInfo)
	}

	return podsInfo, nil
}

func processPodData(pod *corev1.Pod) (PodInfo, error) {
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
		age = format.FormatAge(pod.Status.StartTime.Time)
	}

	status := string(pod.Status.Phase)
	if pod.Status.Phase == corev1.PodRunning {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				status = cs.State.Waiting.Reason
				break
			}
		}
	}

	return PodInfo{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Ready:     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
		Status:    status,
		Restarts:  restarts,
		Age:       age,
	}, nil
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

	return processPodData(pod)
}

func DeletePod(client Client, namespace string, podName string) error {
	err := client.Clientset.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s: %v", podName, err)
	}
	return nil
}
