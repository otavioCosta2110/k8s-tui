package models

import (
	"fmt"
	"sort"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type NamespaceSelectedMsg struct {
	Namespace string
}

type NamespaceSelectorModel struct {
	list       *components.FullscreenListModel
	kubeconfig string
}

func NewNamespaceSelectorModel(kubeconfig string) *NamespaceSelectorModel {
	list := components.NewFullscreenList([]string{}, "Select Namespace", func(selected string) tea.Msg {
		return NamespaceSelectedMsg{Namespace: selected}
	})

	model := &NamespaceSelectorModel{
		list:       list,
		kubeconfig: kubeconfig,
	}

	return model
}

func (m *NamespaceSelectorModel) Init() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Fetch namespaces
		client, err := k8s.NewClient(m.kubeconfig, "")
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create client for namespace fetch: %v", err))
			return fetchedNamespacesMsg{namespaces: []string{"default"}, err: err}
		}

		namespaces, err := k8s.FetchNamespaces(*client)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to fetch namespaces: %v", err))
			return fetchedNamespacesMsg{namespaces: []string{"default"}, err: err}
		}

		names := namespaces
		sort.Strings(names)

		return fetchedNamespacesMsg{namespaces: names, err: nil}
	})
}

type fetchedNamespacesMsg struct {
	namespaces []string
	err        error
}

func (m *NamespaceSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fetchedNamespacesMsg:
		if msg.err != nil {
			// Use default if error
			m.list = components.NewFullscreenList([]string{"default"}, "Select Namespace", func(selected string) tea.Msg {
				return NamespaceSelectedMsg{Namespace: selected}
			})
		} else {
			m.list = components.NewFullscreenList(msg.namespaces, "Select Namespace", func(selected string) tea.Msg {
				return NamespaceSelectedMsg{Namespace: selected}
			})
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return nil, nil // Close selector
		}
	}

	updated, cmd := m.list.Update(msg)
	if list, ok := updated.(*components.FullscreenListModel); ok {
		m.list = list
	}
	return m, cmd
}

func (m *NamespaceSelectorModel) View() string {
	return m.list.View()
}
