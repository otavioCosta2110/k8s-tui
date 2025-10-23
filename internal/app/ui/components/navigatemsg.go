package components

import (
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

	tea "github.com/charmbracelet/bubbletea"
)

type NavigateMsg struct {
	NewScreen     tea.Model
	ResourceModel interface{}
	Cluster       resources.Client
	Error         error
	Breadcrumb    string
	Metadata      map[string]interface{}
}

type RefreshMsg struct{}

type EditMsg struct {
	Content      string
	Title        string
	ResourceType string
	ResourceName string
	Namespace    string
}

type OpenCreateFormMsg struct {
	ResourceType string
}
