package kubernetes

import tea "github.com/charmbracelet/bubbletea"

type ResourceInterface interface {
  InitComponent(*KubeConfig) (tea.Model, error)
	SetSize(width, height int)
}
