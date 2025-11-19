package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventInfo struct {
	Name      string
	Namespace string
	Client    *Client
}

func NewEventInfo(name, namespace string, client *Client) *EventInfo {
	return &EventInfo{
		Name:      name,
		Namespace: namespace,
		Client:    client,
	}
}

func (e *EventInfo) GetEventsForResource(resourceName, resourceType, namespace string) ([]corev1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", resourceName, resourceType)

	events, err := e.Client.Clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get events for %s %s: %v", resourceType, resourceName, err)
	}

	return events.Items, nil
}

func (e *EventInfo) GetAllEvents(namespace string) ([]corev1.Event, error) {
	events, err := e.Client.Clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to get events: %v", err)
	}

	return events.Items, nil
}

func (e *EventInfo) FormatEvents(events []corev1.Event) []map[string]interface{} {
	formattedEvents := make([]map[string]interface{}, 0)

	for _, event := range events {
		age := time.Since(event.LastTimestamp.Time).Round(time.Second)

		formattedEvent := map[string]interface{}{
			"type":      event.Type,
			"reason":    event.Reason,
			"age":       age.String(),
			"from":      event.Source.Component,
			"message":   event.Message,
			"object":    fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			"namespace": event.Namespace,
			"count":     event.Count,
			"firstSeen": event.FirstTimestamp.Time.Format(time.RFC3339),
			"lastSeen":  event.LastTimestamp.Time.Format(time.RFC3339),
		}

		formattedEvents = append(formattedEvents, formattedEvent)
	}

	return formattedEvents
}
