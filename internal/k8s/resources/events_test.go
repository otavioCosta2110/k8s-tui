package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewEventInfo(t *testing.T) {
	client := &Client{}
	eventInfo := NewEventInfo("test-event", "default", client)

	if eventInfo.Name != "test-event" {
		t.Errorf("Expected event name 'test-event', got '%s'", eventInfo.Name)
	}

	if eventInfo.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", eventInfo.Namespace)
	}

	if eventInfo.Client != client {
		t.Error("Client not set correctly")
	}
}

func TestEventInfo_GetEventsForResource(t *testing.T) {
	// Create a fake clientset with some events
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-event-1",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "test-pod",
			},
			Type:    "Normal",
			Reason:  "Started",
			Message: "Pod started successfully",
		},
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-event-2",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Service",
				Name: "test-service",
			},
			Type:    "Warning",
			Reason:  "Failed",
			Message: "Service failed to start",
		},
	)

	client := &Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	eventInfo := NewEventInfo("", "default", client)

	// Test getting events for a specific pod
	events, err := eventInfo.GetEventsForResource("test-pod", "Pod", "default")
	if err != nil {
		t.Fatalf("Error getting events: %v", err)
	}

	// Note: The fake client doesn't support field selectors, so it returns all events
	// We'll just verify we get some events and the pod event is included
	if len(events) == 0 {
		t.Error("Expected at least one event")
	}

	// Find the pod event
	foundPodEvent := false
	for _, event := range events {
		if event.InvolvedObject.Name == "test-pod" && event.InvolvedObject.Kind == "Pod" {
			foundPodEvent = true
			if event.Type != "Normal" {
				t.Errorf("Expected event type 'Normal', got '%s'", event.Type)
			}
			break
		}
	}

	if !foundPodEvent {
		t.Error("Expected to find pod event in results")
	}
}

func TestEventInfo_FormatEvents(t *testing.T) {
	eventInfo := &EventInfo{}

	events := []corev1.Event{
		{
			Type:    "Normal",
			Reason:  "Started",
			Message: "Pod started",
			Source: corev1.EventSource{
				Component: "kubelet",
			},
			LastTimestamp: metav1.Now(),
			InvolvedObject: corev1.ObjectReference{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
			},
			Count: 1,
		},
		{
			Type:    "Warning",
			Reason:  "Failed",
			Message: "Pod failed to start",
			Source: corev1.EventSource{
				Component: "kubelet",
			},
			LastTimestamp: metav1.Now(),
			InvolvedObject: corev1.ObjectReference{
				Kind:      "Pod",
				Name:      "test-pod-2",
				Namespace: "default",
			},
			Count: 2,
		},
	}

	formattedEvents := eventInfo.FormatEvents(events)

	if len(formattedEvents) != 2 {
		t.Errorf("Expected 2 formatted events, got %d", len(formattedEvents))
	}

	// Check first event
	event1 := formattedEvents[0]
	if event1["type"] != "Normal" {
		t.Errorf("Expected event type 'Normal', got '%v'", event1["type"])
	}

	if event1["reason"] != "Started" {
		t.Errorf("Expected reason 'Started', got '%v'", event1["reason"])
	}

	if event1["from"] != "kubelet" {
		t.Errorf("Expected from 'kubelet', got '%v'", event1["from"])
	}

	if event1["object"] != "Pod/test-pod" {
		t.Errorf("Expected object 'Pod/test-pod', got '%v'", event1["object"])
	}

	if event1["count"] != int32(1) {
		t.Errorf("Expected count 1, got '%v'", event1["count"])
	}

	// Check second event
	event2 := formattedEvents[1]
	if event2["type"] != "Warning" {
		t.Errorf("Expected event type 'Warning', got '%v'", event2["type"])
	}

	if event2["count"] != int32(2) {
		t.Errorf("Expected count 2, got '%v'", event2["count"])
	}
}
