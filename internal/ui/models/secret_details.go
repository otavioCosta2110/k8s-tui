package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type secretDetailsModel struct {
	secret     *k8s.SecretInfo
	k8sClient  *k8s.Client
	loading    bool
	err        error
	showValues bool
	yamlViewer *components.YAMLViewer
}

func NewSecretDetails(k k8s.Client, namespace, secretName string) *secretDetailsModel {
	return &secretDetailsModel{
		secret:     k8s.NewSecret(secretName, namespace, k),
		k8sClient:  &k,
		loading:    false,
		err:        nil,
		showValues: false,
	}
}

func (s *secretDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	desc, err := s.secret.DescribeWithVisibility(s.showValues)
	if err != nil {
		return nil, err
	}

	title := "Secret: " + s.secret.Name
	if s.showValues {
		title += " (VALUES VISIBLE)"
	} else {
		title += " (VALUES HIDDEN)"
	}

	s.yamlViewer = components.NewYAMLViewerWithHelp(title, desc, "↑/↓: Scroll • v: Toggle Values • q: Quit")
	return s, nil
}

func (s *secretDetailsModel) Init() tea.Cmd {
	return s.yamlViewer.Init()
}

func (s *secretDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "v", "V":
			s.showValues = !s.showValues
			desc, err := s.secret.DescribeWithVisibility(s.showValues)
			if err != nil {
				s.err = err
				return s, nil
			}

			title := "Secret: " + s.secret.Name
			if s.showValues {
				title += " (VALUES VISIBLE)"
			} else {
				title += " (VALUES HIDDEN)"
			}

			s.yamlViewer = components.NewYAMLViewerWithHelp(title, desc, "↑/↓: Scroll • v: Toggle Values • q: Quit")
			return s, s.yamlViewer.Init()
		case "q", "esc":
			return s, tea.Quit
		}
	}

	var cmd tea.Cmd
	updatedModel, cmd := s.yamlViewer.Update(msg)
	if viewer, ok := updatedModel.(*components.YAMLViewer); ok {
		s.yamlViewer = viewer
	}
	return s, cmd
}

func (s *secretDetailsModel) View() string {
	if s.err != nil {
		return "Error: " + s.err.Error()
	}

	if s.yamlViewer == nil {
		return "Loading..."
	}

	return s.yamlViewer.View()
}
