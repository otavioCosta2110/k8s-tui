package components

import (
	"otaviocosta2110/k8s-tui/internal/k8s"

	tea "github.com/charmbracelet/bubbletea"
)

type NavigateMsg struct {
	NewScreen  tea.Model
	Cluster    k8s.Client
	Error      error
	Breadcrumb string
}

type RefreshMsg struct{}

type EditMsg struct {
	Content      string
	Title        string
	ResourceType string
	ResourceName string
	Namespace    string
}
