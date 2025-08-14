package components

import (
	"otaviocosta2110/k8s-tui/internal/k8s"

	tea "github.com/charmbracelet/bubbletea"
)

type NavigateMsg struct {
	NewScreen tea.Model
	Cluster   k8s.Client
	Error     error
}

type RefreshMsg struct { }
