package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

type eventsModel struct {
	eventsComponent *components.EventsModel
	k8sClient       *k8s.Client
	resourceName    string
	resourceType    string
	namespace       string
	eventInfo       *k8s.EventInfo
}

func NewEventsModel(k *k8s.Client, namespace, resourceName, resourceType string) *eventsModel {
	title := fmt.Sprintf("Events: %s %s", resourceType, resourceName)
	return &eventsModel{
		eventsComponent: components.NewEventsModel(title),
		k8sClient:       k,
		resourceName:    resourceName,
		resourceType:    resourceType,
		namespace:       namespace,
		eventInfo:       k8s.NewEventInfo("", namespace, k),
	}
}

func (e *eventsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	e.k8sClient = k
	e.eventInfo = k8s.NewEventInfo("", e.namespace, k)
	return e, nil
}

func (e *eventsModel) fetchEvents() tea.Cmd {
	return func() tea.Msg {
		k8sEvents, err := e.eventInfo.GetEventsForResource(e.resourceName, e.resourceType, e.namespace)
		if err != nil {
			return components.EventsLoadedMsg{Error: err}
		}

		formattedEvents := e.eventInfo.FormatEvents(k8sEvents)
		events := make([]components.Event, len(formattedEvents))

		for i, formattedEvent := range formattedEvents {
			events[i] = components.Event{
				Type:      fmt.Sprintf("%v", formattedEvent["type"]),
				Reason:    fmt.Sprintf("%v", formattedEvent["reason"]),
				Age:       fmt.Sprintf("%v", formattedEvent["age"]),
				From:      fmt.Sprintf("%v", formattedEvent["from"]),
				Message:   fmt.Sprintf("%v", formattedEvent["message"]),
				Object:    fmt.Sprintf("%v", formattedEvent["object"]),
				Namespace: fmt.Sprintf("%v", formattedEvent["namespace"]),
				Count:     formattedEvent["count"].(int32),
				FirstSeen: fmt.Sprintf("%v", formattedEvent["firstSeen"]),
				LastSeen:  fmt.Sprintf("%v", formattedEvent["lastSeen"]),
			}
		}

		return components.EventsLoadedMsg{Events: events}
	}
}

func (e *eventsModel) Init() tea.Cmd {
	return tea.Batch(e.eventsComponent.Init(), e.fetchEvents())
}

func (e *eventsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle events loaded message
	switch msg := msg.(type) {
	case components.EventsLoadedMsg:
		if msg.Error != nil {
			e.eventsComponent.SetError(msg.Error)
		} else {
			e.eventsComponent.SetEvents(msg.Events)
		}
		return e, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			e.eventsComponent.GotoTop()
			return e, nil
		case "G":
			e.eventsComponent.GotoBottom()
			return e, nil
		case "r":
			return e, e.fetchEvents()
		}
	}

	// Update the events component
	updatedModel, cmd := e.eventsComponent.Update(msg)
	e.eventsComponent = updatedModel.(*components.EventsModel)
	return e, cmd
}

func (e *eventsModel) View() string {
	return e.eventsComponent.View()
}
