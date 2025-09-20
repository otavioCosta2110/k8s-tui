package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type serviceaccountDetailsModel struct {
	serviceaccount *k8s.ServiceAccountInfo
	k8sClient      *k8s.Client
	loading        bool
	err            error
	yamlViewer     *components.YAMLViewer
}

func NewServiceAccountDetails(k k8s.Client, namespace, serviceaccountName string) *serviceaccountDetailsModel {
	return &serviceaccountDetailsModel{
		serviceaccount: k8s.NewServiceAccount(serviceaccountName, namespace, k),
		k8sClient:      &k,
		loading:        false,
		err:            nil,
	}
}

func (s *serviceaccountDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k

	desc, err := s.serviceaccount.Describe()
	if err != nil {
		return nil, err
	}

	title := "ServiceAccount: " + s.serviceaccount.Name

	s.yamlViewer = components.NewYAMLViewerWithHelp(title, desc, "↑/↓: Scroll • q: Quit")
	return s, nil
}

func (s *serviceaccountDetailsModel) Init() tea.Cmd {
	return s.yamlViewer.Init()
}

func (s *serviceaccountDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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

func (s *serviceaccountDetailsModel) View() string {
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
