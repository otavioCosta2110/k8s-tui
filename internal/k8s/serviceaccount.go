package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountInfo struct {
	Namespace string
	Name      string
	Secrets   string
	Age       string
	Raw       *corev1.ServiceAccount
	Client    Client
}

func NewServiceAccount(name, namespace string, k Client) *ServiceAccountInfo {
	return &ServiceAccountInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchServiceAccountList(client Client, namespace string) ([]string, error) {
	sas, err := client.Clientset.CoreV1().ServiceAccounts(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch serviceaccounts: %v", err)
	}

	saNames := make([]string, 0, len(sas.Items))
	for _, sa := range sas.Items {
		saNames = append(saNames, sa.Name)
	}

	return saNames, nil
}

func GetServiceAccountsTableData(client Client, namespace string) ([]ServiceAccountInfo, error) {
	sas, err := client.Clientset.CoreV1().ServiceAccounts(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list serviceaccounts: %v", err)
	}

	var saInfos []ServiceAccountInfo
	for _, sa := range sas.Items {
		secretsCount := len(sa.Secrets)
		secretsStr := fmt.Sprintf("%d", secretsCount)

		saInfos = append(saInfos, ServiceAccountInfo{
			Namespace: sa.Namespace,
			Name:      sa.Name,
			Secrets:   secretsStr,
			Age:       utils.FormatAge(sa.CreationTimestamp.Time),
			Raw:       sa.DeepCopy(),
			Client:    client,
		})
	}

	return saInfos, nil
}

func (s *ServiceAccountInfo) Fetch() error {
	sa, err := s.Client.Clientset.CoreV1().ServiceAccounts(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get serviceaccount: %v", err)
	}
	s.Raw = sa
	return nil
}

func (s *ServiceAccountInfo) Describe() (string, error) {
	if s.Raw == nil {
		if err := s.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch serviceaccount: %v", err)
		}
	}

	var events *corev1.EventList
	var err error

	if s.Client.Clientset != nil {
		events, err = s.Client.Clientset.CoreV1().Events(s.Namespace).List(context.Background(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=ServiceAccount", s.Name, s.Namespace),
		})
		if err != nil {
			events = &corev1.EventList{Items: []corev1.Event{}}
		}
	} else {
		events = &corev1.EventList{Items: []corev1.Event{}}
	}

	data, err := s.DescribeServiceAccount(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe serviceaccount: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal serviceaccount to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (s *ServiceAccountInfo) DescribeServiceAccount(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	type SecretRef struct {
		Name string `yaml:"name"`
	}

	desc := map[string]any{
		"name":        s.Name,
		"namespace":   s.Namespace,
		"labels":      s.Raw.Labels,
		"annotations": s.Raw.Annotations,
		"created":     formatTime(s.Raw.CreationTimestamp),
	}

	if len(s.Raw.Secrets) > 0 {
		secrets := make([]SecretRef, 0, len(s.Raw.Secrets))
		for _, secret := range s.Raw.Secrets {
			secrets = append(secrets, SecretRef{Name: secret.Name})
		}
		desc["secrets"] = secrets
	}

	if len(s.Raw.ImagePullSecrets) > 0 {
		imagePullSecrets := make([]SecretRef, 0, len(s.Raw.ImagePullSecrets))
		for _, secret := range s.Raw.ImagePullSecrets {
			imagePullSecrets = append(imagePullSecrets, SecretRef{Name: secret.Name})
		}
		desc["imagePullSecrets"] = imagePullSecrets
	}

	if s.Raw.AutomountServiceAccountToken != nil {
		desc["automountServiceAccountToken"] = *s.Raw.AutomountServiceAccountToken
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

func DeleteServiceAccount(client Client, namespace string, saName string) error {
	err := client.Clientset.CoreV1().ServiceAccounts(namespace).Delete(context.Background(), saName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete serviceaccount %s: %v", saName, err)
	}
	return nil
}
