package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/utils"
	"time"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSetInfo struct {
	Namespace    string
	Name         string
	Desired      string
	Current      string
	Ready        string
	UpToDate     string
	Available    string
	NodeSelector string
	Age          string
	Raw          *appsv1.DaemonSet
	Client       Client
}

func NewDaemonSet(name, namespace string, k Client) *DaemonSetInfo {
	return &DaemonSetInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchDaemonSetList(client Client, namespace string) ([]string, error) {
	daemonsets, err := client.Clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daemonsets: %v", err)
	}

	daemonsetNames := make([]string, 0, len(daemonsets.Items))
	for _, daemonset := range daemonsets.Items {
		daemonsetNames = append(daemonsetNames, daemonset.Name)
	}

	return daemonsetNames, nil
}

func GetDaemonSetsTableData(client Client, namespace string) ([]DaemonSetInfo, error) {
	daemonsets, err := client.Clientset.AppsV1().DaemonSets(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %v", err)
	}

	var daemonsetInfos []DaemonSetInfo
	for _, daemonset := range daemonsets.Items {
		desired := fmt.Sprintf("%d", daemonset.Status.DesiredNumberScheduled)
		current := fmt.Sprintf("%d", daemonset.Status.CurrentNumberScheduled)
		ready := fmt.Sprintf("%d", daemonset.Status.NumberReady)
		upToDate := fmt.Sprintf("%d", daemonset.Status.UpdatedNumberScheduled)
		available := fmt.Sprintf("%d", daemonset.Status.NumberAvailable)

		nodeSelector := "<none>"
		if len(daemonset.Spec.Template.Spec.NodeSelector) > 0 {
			for key, value := range daemonset.Spec.Template.Spec.NodeSelector {
				nodeSelector = fmt.Sprintf("%s=%s", key, value)
				break
			}
		}

		daemonsetInfos = append(daemonsetInfos, DaemonSetInfo{
			Namespace:    daemonset.Namespace,
			Name:         daemonset.Name,
			Desired:      desired,
			Current:      current,
			Ready:        ready,
			UpToDate:     upToDate,
			Available:    available,
			NodeSelector: nodeSelector,
			Age:          utils.FormatAge(daemonset.CreationTimestamp.Time),
			Raw:          daemonset.DeepCopy(),
			Client:       client,
		})
	}

	return daemonsetInfos, nil
}

func (ds *DaemonSetInfo) Fetch() error {
	daemonset, err := ds.Client.Clientset.AppsV1().DaemonSets(ds.Namespace).Get(context.Background(), ds.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonset: %v", err)
	}
	ds.Raw = daemonset
	return nil
}

func (ds *DaemonSetInfo) Describe() (string, error) {
	if ds.Raw == nil {
		if err := ds.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch daemonset: %v", err)
		}
	}

	events, err := ds.Client.Clientset.CoreV1().Events(ds.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=DaemonSet", ds.Name, ds.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get daemonset events: %v", err)
	}

	data, err := ds.DescribeDaemonSet(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe daemonset: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal daemonset to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (ds *DaemonSetInfo) DescribeDaemonSet(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        ds.Name,
		"namespace":   ds.Namespace,
		"labels":      ds.Raw.Labels,
		"annotations": ds.Raw.Annotations,
		"created":     formatTime(ds.Raw.CreationTimestamp),
	}

	desc["selector"] = ds.Raw.Spec.Selector.MatchLabels

	desc["status"] = map[string]any{
		"desiredNumberScheduled": ds.Raw.Status.DesiredNumberScheduled,
		"currentNumberScheduled": ds.Raw.Status.CurrentNumberScheduled,
		"numberReady":            ds.Raw.Status.NumberReady,
		"updatedNumberScheduled": ds.Raw.Status.UpdatedNumberScheduled,
		"numberAvailable":        ds.Raw.Status.NumberAvailable,
		"numberUnavailable":      ds.Raw.Status.NumberUnavailable,
	}

	if len(ds.Raw.Spec.Template.Spec.Containers) > 0 {
		containers := make([]map[string]any, 0, len(ds.Raw.Spec.Template.Spec.Containers))
		for _, container := range ds.Raw.Spec.Template.Spec.Containers {
			containerDesc := map[string]any{
				"name":  container.Name,
				"image": container.Image,
			}
			if len(container.Command) > 0 {
				containerDesc["command"] = container.Command
			}
			if len(container.Args) > 0 {
				containerDesc["args"] = container.Args
			}
			if len(container.Ports) > 0 {
				ports := make([]map[string]any, 0, len(container.Ports))
				for _, port := range container.Ports {
					portDesc := map[string]any{
						"name":          port.Name,
						"containerPort": port.ContainerPort,
						"protocol":      string(port.Protocol),
					}
					if port.HostPort != 0 {
						portDesc["hostPort"] = port.HostPort
					}
					ports = append(ports, portDesc)
				}
				containerDesc["ports"] = ports
			}
			containers = append(containers, containerDesc)
		}
		desc["containers"] = containers
	}

	if len(ds.Raw.Spec.Template.Spec.NodeSelector) > 0 {
		desc["nodeSelector"] = ds.Raw.Spec.Template.Spec.NodeSelector
	}

	if len(events.Items) > 0 {
		eventList := make([]Event, 0)
		for _, event := range events.Items {
			age := time.Since(event.LastTimestamp.Time).Round(time.Second)
			eventList = append(eventList, Event{
				Type:    event.Type,
				Reason:  event.Reason,
				Age:     age.String(),
				From:    event.Source.Component,
				Message: event.Message,
			})
		}
		desc["events"] = eventList
	}

	return desc, nil
}

func DeleteDaemonSet(client Client, namespace string, daemonsetName string) error {
	err := client.Clientset.AppsV1().DaemonSets(namespace).Delete(context.Background(), daemonsetName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete daemonset %s: %v", daemonsetName, err)
	}
	return nil
}
