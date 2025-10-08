package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type secretDetailsModel struct {
	secret     *k8s.SecretInfo
	k8sClient  *k8s.Client
	loading    bool
	err        error
	showValues bool
	yamlViewer *components.YAMLViewer
}

type secretDetailsLoadedMsg struct {
	content string
	err     error
}

func NewSecretDetails(k k8s.Client, namespace, secretName string) *secretDetailsModel {
	return &secretDetailsModel{
		secret:     k8s.NewSecret(secretName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewerWithHelp("Secret: "+secretName+" (VALUES HIDDEN)", "Loading...", "↑/↓: Scroll • v: Toggle Values • q: Quit"),
		loading:    true,
		err:        nil,
		showValues: false,
	}
}

func (s *secretDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k
	return s, nil
}

func (s *secretDetailsModel) fetchSecretDetails() tea.Cmd {
	return func() tea.Msg {
		desc, err := s.secret.DescribeWithVisibility(s.showValues)
		return secretDetailsLoadedMsg{content: desc, err: err}
	}
}

func (s *secretDetailsModel) Init() tea.Cmd {
	return tea.Batch(s.yamlViewer.Init(), s.fetchSecretDetails())
}

func (s *secretDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case secretDetailsLoadedMsg:
		s.loading = false
		if msg.err != nil {
			s.err = msg.err
			s.yamlViewer.SetContent("Error loading secret details: " + msg.err.Error())
		} else {
			title := "Secret: " + s.secret.Name
			if s.showValues {
				title += " (VALUES VISIBLE)"
			} else {
				title += " (VALUES HIDDEN)"
			}
			s.yamlViewer = components.NewYAMLViewerWithHelp(title, msg.content, "↑/↓: Scroll • v: Toggle Values • q: Quit")
		}
		return s, s.yamlViewer.Init()
	case tea.KeyMsg:
		switch msg.String() {
		case "v", "V":
			s.showValues = !s.showValues
			return s, s.fetchSecretDetails()
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
		return lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Error: " + s.err.Error())
	}

	if s.yamlViewer == nil {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Loading...")
	}

	return s.yamlViewer.View()
}
