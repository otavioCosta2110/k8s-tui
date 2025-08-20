package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Configmap struct {
	Name      string
	Namespace string
	Data      string
	Age       string
	Raw       *corev1.ConfigMap
	Client    *kubernetes.Clientset
	Config    *rest.Config
	YAML      string
}

func NewConfigmap(name, namespace string, k Client) *Configmap {
	return &Configmap{
		Name:      name,
		Namespace: namespace,
		Client:    k.Clientset,
		Config:    k.Config,
	}
}

func (c *Configmap) Fetch() error {
	cm, err := c.Client.CoreV1().ConfigMaps(c.Namespace).Get(context.Background(), c.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configmaps: %v", err)
	}
	c.Raw = cm
	return nil
}

func FetchConfigmaps(client Client, namespace string, selector string) ([]Configmap, error) {
	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}
	cms, err := client.Clientset.CoreV1().ConfigMaps(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %v", err)
	}

	cmsInfo := make([]Configmap, 0, len(cms.Items))
	for _, cmCore := range cms.Items {
		cm, err := GetConfigmapDetails(client, namespace, &cmCore)
		if err != nil {
			return nil, err
		}
		cmsInfo = append(cmsInfo, cm)
	}

	return cmsInfo, nil
}

func GetConfigmapDetails(client Client, namespace string, cmCore *corev1.ConfigMap) (Configmap, error) {
	cm, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(
		context.Background(),
		cmCore.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return Configmap{}, fmt.Errorf("failed to get pod details: %v", err)
	}

	age := utils.FormatAge(cm.GetCreationTimestamp().Time)

	return Configmap{
		Namespace: cm.Namespace,
		Name:      cm.Name,
		Data:      fmt.Sprintf("%d", len(cm.Data)),
		Age:       age,
	}, nil
}

func (c *Configmap) Describe() (string, error) {
	if c.Raw == nil {
		if err := c.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch configmap: %v", err)
		}
	}

	events, err := c.Client.CoreV1().Events(c.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=ConfigMap", c.Name, c.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get configmap events: %v", err)
	}

	data, err := c.DescribeConfigMap(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe configmap: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configmap to YAML: %v", err)
	}

	c.YAML = string(yamlData)
	return c.YAML, nil
}

func DeleteConfigmap(client Client, namespace string, cmName string) error {
	err := client.Clientset.CoreV1().ConfigMaps(namespace).Delete(context.Background(), cmName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete configmap %s: %v", cmName, err)
	}
	return nil
}

func (cm *Configmap) DescribeConfigMap(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        cm.Name,
		"namespace":   cm.Namespace,
		"labels":      cm.Raw.Labels,
		"annotations": cm.Raw.Annotations,
		"created":     formatTime(cm.Raw.CreationTimestamp),
	}

	if len(cm.Raw.Data) > 0 {
		dataDesc := make(map[string]string)
		keys := make([]string, 0, len(cm.Raw.Data))
		for k := range cm.Raw.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := cm.Raw.Data[key]
			lines := strings.Split(value, "\n")
			if len(lines) == 1 {
				dataDesc[key] = value
			} else {
				dataDesc[key] = fmt.Sprintf("----\n%s", value)
			}
		}
		desc["data"] = dataDesc
	}

	if len(cm.Raw.BinaryData) > 0 {
		binaryDataDesc := make(map[string]string)
		keys := make([]string, 0, len(cm.Raw.BinaryData))
		for k := range cm.Raw.BinaryData {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			data := cm.Raw.BinaryData[key]
			binaryDataDesc[key] = fmt.Sprintf("[binary data, %d bytes]", len(data))
		}
		desc["binaryData"] = binaryDataDesc
	}

	if cm.Raw.Immutable != nil {
		desc["immutable"] = *cm.Raw.Immutable
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
