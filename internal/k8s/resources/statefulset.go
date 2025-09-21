package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"time"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatefulSetInfo struct {
	Namespace string
	Name      string
	Replicas  string
	Ready     string
	Age       string
	Raw       *appsv1.StatefulSet
	Client    Client
}

func NewStatefulSet(name, namespace string, k Client) *StatefulSetInfo {
	return &StatefulSetInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchStatefulSetList(client Client, namespace string) ([]string, error) {
	statefulsets, err := client.Clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch statefulsets: %v", err)
	}

	statefulsetNames := make([]string, 0, len(statefulsets.Items))
	for _, statefulset := range statefulsets.Items {
		statefulsetNames = append(statefulsetNames, statefulset.Name)
	}

	return statefulsetNames, nil
}

func GetStatefulSetsTableData(client Client, namespace string) ([]StatefulSetInfo, error) {
	statefulsets, err := client.Clientset.AppsV1().StatefulSets(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %v", err)
	}

	var statefulsetInfos []StatefulSetInfo
	for _, statefulset := range statefulsets.Items {
		replicas := "0"
		if statefulset.Spec.Replicas != nil {
			replicas = fmt.Sprintf("%d", *statefulset.Spec.Replicas)
		}

		ready := fmt.Sprintf("%d/%d", statefulset.Status.ReadyReplicas, statefulset.Status.Replicas)

		statefulsetInfos = append(statefulsetInfos, StatefulSetInfo{
			Namespace: statefulset.Namespace,
			Name:      statefulset.Name,
			Replicas:  replicas,
			Ready:     ready,
			Age:       format.FormatAge(statefulset.CreationTimestamp.Time),
			Raw:       statefulset.DeepCopy(),
			Client:    client,
		})
	}

	return statefulsetInfos, nil
}

func (ss *StatefulSetInfo) Fetch() error {
	statefulset, err := ss.Client.Clientset.AppsV1().StatefulSets(ss.Namespace).Get(context.Background(), ss.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %v", err)
	}
	ss.Raw = statefulset
	return nil
}

func (ss *StatefulSetInfo) Describe() (string, error) {
	if ss.Raw == nil {
		if err := ss.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch statefulset: %v", err)
		}
	}

	events, err := ss.Client.Clientset.CoreV1().Events(ss.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=StatefulSet", ss.Name, ss.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get statefulset events: %v", err)
	}

	data, err := ss.DescribeStatefulSet(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe statefulset: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal statefulset to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (ss *StatefulSetInfo) DescribeStatefulSet(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        ss.Name,
		"namespace":   ss.Namespace,
		"labels":      ss.Raw.Labels,
		"annotations": ss.Raw.Annotations,
		"created":     formatTime(ss.Raw.CreationTimestamp),
	}

	desc["selector"] = ss.Raw.Spec.Selector.MatchLabels

	if ss.Raw.Spec.Replicas != nil {
		desc["replicas"] = *ss.Raw.Spec.Replicas
	}

	if ss.Raw.Spec.ServiceName != "" {
		desc["serviceName"] = ss.Raw.Spec.ServiceName
	}

	desc["status"] = map[string]any{
		"replicas":        ss.Raw.Status.Replicas,
		"readyReplicas":   ss.Raw.Status.ReadyReplicas,
		"currentReplicas": ss.Raw.Status.CurrentReplicas,
		"updatedReplicas": ss.Raw.Status.UpdatedReplicas,
	}

	if len(ss.Raw.Spec.Template.Spec.Containers) > 0 {
		containers := make([]map[string]any, 0, len(ss.Raw.Spec.Template.Spec.Containers))
		for _, container := range ss.Raw.Spec.Template.Spec.Containers {
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
					ports = append(ports, portDesc)
				}
				containerDesc["ports"] = ports
			}
			containers = append(containers, containerDesc)
		}
		desc["containers"] = containers
	}

	if len(ss.Raw.Spec.VolumeClaimTemplates) > 0 {
		volumeClaims := make([]map[string]any, 0, len(ss.Raw.Spec.VolumeClaimTemplates))
		for _, vct := range ss.Raw.Spec.VolumeClaimTemplates {
			vctDesc := map[string]any{
				"name": vct.Name,
			}
			if vct.Spec.AccessModes != nil {
				vctDesc["accessModes"] = vct.Spec.AccessModes
			}
			if vct.Spec.Resources.Requests != nil {
				if storage, ok := vct.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
					vctDesc["storage"] = storage.String()
				}
			}
			volumeClaims = append(volumeClaims, vctDesc)
		}
		desc["volumeClaimTemplates"] = volumeClaims
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

func DeleteStatefulSet(client Client, namespace string, statefulsetName string) error {
	err := client.Clientset.AppsV1().StatefulSets(namespace).Delete(context.Background(), statefulsetName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete statefulset %s: %v", statefulsetName, err)
	}
	return nil
}
