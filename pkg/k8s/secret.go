package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/utils"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretInfo struct {
	Namespace string
	Name      string
	Type      string
	Data      string
	Age       string
	Raw       *corev1.Secret
	Client    Client
}

func NewSecret(name, namespace string, k Client) *SecretInfo {
	return &SecretInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchSecretList(client Client, namespace string) ([]string, error) {
	secrets, err := client.Clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secrets: %v", err)
	}

	secretNames := make([]string, 0, len(secrets.Items))
	for _, secret := range secrets.Items {
		secretNames = append(secretNames, secret.Name)
	}

	return secretNames, nil
}

func GetSecretsTableData(client Client, namespace string) ([]SecretInfo, error) {
	secrets, err := client.Clientset.CoreV1().Secrets(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}

	var secretInfos []SecretInfo
	for _, secret := range secrets.Items {
		secretType := string(secret.Type)

		dataCount := len(secret.Data)
		dataStr := fmt.Sprintf("%d", dataCount)

		secretInfos = append(secretInfos, SecretInfo{
			Namespace: secret.Namespace,
			Name:      secret.Name,
			Type:      secretType,
			Data:      dataStr,
			Age:       utils.FormatAge(secret.CreationTimestamp.Time),
			Raw:       secret.DeepCopy(),
			Client:    client,
		})
	}

	return secretInfos, nil
}

func (s *SecretInfo) Fetch() error {
	secret, err := s.Client.Clientset.CoreV1().Secrets(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %v", err)
	}
	s.Raw = secret
	return nil
}

func (s *SecretInfo) Describe() (string, error) {
	return s.DescribeWithVisibility(false)
}

func (s *SecretInfo) DescribeWithVisibility(showValues bool) (string, error) {
	if s.Raw == nil {
		if err := s.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch secret: %v", err)
		}
	}

	var events *corev1.EventList
	var err error

	if s.Client.Clientset != nil {
		events, err = s.Client.Clientset.CoreV1().Events(s.Namespace).List(context.Background(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Secret", s.Name, s.Namespace),
		})
		if err != nil {
			events = &corev1.EventList{Items: []corev1.Event{}}
		}
	} else {
		events = &corev1.EventList{Items: []corev1.Event{}}
	}

	data, err := s.DescribeSecretWithVisibility(events, showValues)
	if err != nil {
		return "", fmt.Errorf("failed to describe secret: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (s *SecretInfo) DescribeSecret(events *corev1.EventList) (map[string]any, error) {
	return s.DescribeSecretWithVisibility(events, false)
}

func (s *SecretInfo) DescribeSecretWithVisibility(events *corev1.EventList, showValues bool) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        s.Name,
		"namespace":   s.Namespace,
		"labels":      s.Raw.Labels,
		"annotations": s.Raw.Annotations,
		"created":     formatTime(s.Raw.CreationTimestamp),
	}

	desc["type"] = string(s.Raw.Type)

	if len(s.Raw.Data) > 0 {
		if showValues {
			dataMap := make(map[string]string)
			for key, value := range s.Raw.Data {
				dataMap[key] = string(value)
			}
			desc["data"] = dataMap
		} else {
			dataKeys := make([]string, 0, len(s.Raw.Data))
			for key := range s.Raw.Data {
				dataKeys = append(dataKeys, key)
			}
			sort.Strings(dataKeys)
			desc["dataKeys"] = dataKeys
		}
	}

	if len(s.Raw.StringData) > 0 {
		if showValues {
			desc["stringData"] = s.Raw.StringData
		} else {
			stringDataKeys := make([]string, 0, len(s.Raw.StringData))
			for key := range s.Raw.StringData {
				stringDataKeys = append(stringDataKeys, key)
			}
			sort.Strings(stringDataKeys)
			desc["stringDataKeys"] = stringDataKeys
		}
	}

	if s.Raw.Immutable != nil {
		desc["immutable"] = *s.Raw.Immutable
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

func DeleteSecret(client Client, namespace string, secretName string) error {
	err := client.Clientset.CoreV1().Secrets(namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret %s: %v", secretName, err)
	}
	return nil
}
