package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

func TestNewEventsModel(t *testing.T) {
	client := &k8s.Client{}
	eventsModel := NewEventsModel(client, "default", "test-pod", "Pod")

	if eventsModel.resourceName != "test-pod" {
		t.Errorf("Expected resource name 'test-pod', got '%s'", eventsModel.resourceName)
	}

	if eventsModel.resourceType != "Pod" {
		t.Errorf("Expected resource type 'Pod', got '%s'", eventsModel.resourceType)
	}

	if eventsModel.namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", eventsModel.namespace)
	}

	if eventsModel.eventInfo == nil {
		t.Error("EventInfo not initialized")
	}
}

func TestEventsModel_Init(t *testing.T) {
	client := &k8s.Client{}
	eventsModel := NewEventsModel(client, "default", "test-pod", "Pod")

	cmd := eventsModel.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestEventsModel_Update(t *testing.T) {
	client := &k8s.Client{}
	eventsModel := NewEventsModel(client, "default", "test-pod", "Pod")

	// Test with unknown message
	model, _ := eventsModel.Update(tea.KeyMsg{})
	if model == nil {
		t.Error("Update should return a model")
	}

	// Test with events loaded message
	eventsLoadedMsg := components.EventsLoadedMsg{
		Events: []components.Event{
			{
				Type:    "Normal",
				Reason:  "Started",
				Message: "Test message",
			},
		},
		Error: nil,
	}

	model, _ = eventsModel.Update(eventsLoadedMsg)
	if model == nil {
		t.Error("Update should return a model")
	}
}

func TestEventsModel_View(t *testing.T) {
	client := &k8s.Client{}
	model := NewEventsModel(client, "default", "test-pod", "Pod")

	// Set window size for proper rendering
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*eventsModel)

	view := model.View()
	if view == "" {
		t.Error("Expected view content, got empty string")
	}

	// Should contain loading message initially
	if !strings.Contains(view, "Loading events") {
		t.Error("View should contain loading message")
	}
}
