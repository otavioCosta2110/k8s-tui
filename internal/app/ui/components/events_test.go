package components

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewEventsModel(t *testing.T) {
	title := "Test Events"
	model := NewEventsModel(title)

	if model.title != title {
		t.Errorf("Expected title '%s', got '%s'", title, model.title)
	}

	if !model.loading {
		t.Error("Expected loading to be true initially")
	}

	if len(model.events) != 0 {
		t.Error("Expected events to be empty initially")
	}

	if model.error != nil {
		t.Error("Expected error to be nil initially")
	}
}

func TestEventsModel_SetEvents(t *testing.T) {
	model := NewEventsModel("Test")

	events := []Event{
		{
			Type:    "Normal",
			Reason:  "Started",
			Message: "Test message",
		},
	}

	model.SetEvents(events)

	if model.loading {
		t.Error("Expected loading to be false after setting events")
	}

	if len(model.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(model.events))
	}

	if model.events[0].Type != "Normal" {
		t.Errorf("Expected event type 'Normal', got '%s'", model.events[0].Type)
	}
}

func TestEventsModel_SetError(t *testing.T) {
	model := NewEventsModel("Test")
	err := fmt.Errorf("test error")

	model.SetError(err)

	if model.loading {
		t.Error("Expected loading to be false after setting error")
	}

	if model.error == nil {
		t.Error("Expected error to be set")
	}

	if model.error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", model.error.Error())
	}
}

func TestEventsModel_SetLoading(t *testing.T) {
	model := NewEventsModel("Test")

	model.SetLoading(false)

	if model.loading {
		t.Error("Expected loading to be false")
	}

	model.SetLoading(true)

	if !model.loading {
		t.Error("Expected loading to be true")
	}

	if model.error != nil {
		t.Error("Expected error to be cleared when setting loading")
	}
}

func TestEventsModel_Update(t *testing.T) {
	model := NewEventsModel("Test")

	// Test window size update
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	updatedModel, cmd := model.Update(msg)
	if updatedModel == nil {
		t.Error("Update should return a model")
	}

	if cmd != nil {
		t.Error("Update should not return a command for window size")
	}

	eventsModel := updatedModel.(*EventsModel)
	if eventsModel.width != 100 {
		t.Errorf("Expected width 100, got %d", eventsModel.width)
	}

	if eventsModel.height != 46 { // 50 - 4
		t.Errorf("Expected height 46, got %d", eventsModel.height)
	}
}

func TestEventsModel_UpdateEventsLoaded(t *testing.T) {
	model := NewEventsModel("Test")

	events := []Event{
		{
			Type:    "Normal",
			Reason:  "Started",
			Message: "Test message",
		},
	}

	msg := EventsLoadedMsg{
		Events: events,
		Error:  nil,
	}

	updatedModel, cmd := model.Update(msg)
	if updatedModel == nil {
		t.Error("Update should return a model")
	}

	if cmd != nil {
		t.Error("Update should not return a command for events loaded")
	}

	eventsModel := updatedModel.(*EventsModel)
	if eventsModel.loading {
		t.Error("Expected loading to be false after events loaded")
	}

	if len(eventsModel.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(eventsModel.events))
	}
}

func TestEventsModel_UpdateEventsLoadedWithError(t *testing.T) {
	model := NewEventsModel("Test")

	err := fmt.Errorf("test error")
	msg := EventsLoadedMsg{
		Events: nil,
		Error:  err,
	}

	updatedModel, _ := model.Update(msg)
	eventsModel := updatedModel.(*EventsModel)

	if eventsModel.loading {
		t.Error("Expected loading to be false after error")
	}

	if eventsModel.error == nil {
		t.Error("Expected error to be set")
	}

	if eventsModel.error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", eventsModel.error.Error())
	}
}

func TestEventsModel_ViewportMethods(t *testing.T) {
	model := NewEventsModel("Test")

	// Set window size to initialize viewport
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*EventsModel)

	// Test viewport methods don't panic
	model.GotoTop()
	model.GotoBottom()
	model.LineDown(1)
	model.LineUp(1)

	if model.ViewportHeight() <= 0 {
		t.Error("Expected positive viewport height")
	}

	if model.ViewportWidth() <= 0 {
		t.Error("Expected positive viewport width")
	}
}

func TestEventsModel_RenderEmpty(t *testing.T) {
	model := NewEventsModel("Test")
	model.SetEvents([]Event{})

	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain "No events found"
	if !strings.Contains(view, "No events found") {
		t.Error("View should contain 'No events found' message")
	}
}

func TestEventsModel_RenderError(t *testing.T) {
	model := NewEventsModel("Test")
	err := fmt.Errorf("test error")
	model.SetError(err)

	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain error message
	if !strings.Contains(view, "Error loading events") {
		t.Error("View should contain error message")
	}

	if !strings.Contains(view, "test error") {
		t.Error("View should contain specific error message")
	}
}

func TestEventsModel_RenderLoading(t *testing.T) {
	model := NewEventsModel("Test")
	model.SetLoading(true)

	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain loading message
	if !strings.Contains(view, "Loading events") {
		t.Error("View should contain loading message")
	}
}

func TestEventsModel_RenderEvents(t *testing.T) {
	model := NewEventsModel("Test Events")

	// Set window size for proper rendering
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 50,
	}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*EventsModel)

	events := []Event{
		{
			Type:      "Normal",
			Reason:    "Started",
			Age:       "1m",
			From:      "kubelet",
			Message:   "Pod started successfully",
			Object:    "Pod/test-pod",
			Namespace: "default",
			Count:     1,
			FirstSeen: "2023-01-01T00:00:00Z",
			LastSeen:  "2023-01-01T00:01:00Z",
		},
		{
			Type:      "Warning",
			Reason:    "Failed",
			Age:       "30s",
			From:      "kubelet",
			Message:   "Pod failed to start",
			Object:    "Pod/test-pod-2",
			Namespace: "default",
			Count:     2,
			FirstSeen: "2023-01-01T00:02:00Z",
			LastSeen:  "2023-01-01T00:02:30Z",
		},
	}

	model.SetEvents(events)
	view := model.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain title
	if !strings.Contains(view, "Test Events") {
		t.Error("View should contain title")
	}

	// Should contain event data
	if !strings.Contains(view, "Normal") {
		t.Error("View should contain 'Normal' event type")
	}

	if !strings.Contains(view, "Warning") {
		t.Error("View should contain 'Warning' event type")
	}

	if !strings.Contains(view, "Started") {
		t.Error("View should contain 'Started' reason")
	}

	if !strings.Contains(view, "Failed") {
		t.Error("View should contain 'Failed' reason")
	}

	if !strings.Contains(view, "test-pod") {
		t.Error("View should contain pod name")
	}
}
