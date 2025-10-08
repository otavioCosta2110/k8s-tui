package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

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

type serviceaccountDetailsLoadedMsg struct {
	content string
	err     error
}

func NewServiceAccountDetails(k k8s.Client, namespace, serviceaccountName string) *serviceaccountDetailsModel {
	return &serviceaccountDetailsModel{
		serviceaccount: k8s.NewServiceAccount(serviceaccountName, namespace, k),
		k8sClient:      &k,
		yamlViewer:     components.NewYAMLViewerWithHelp("ServiceAccount: "+serviceaccountName, "Loading...", "↑/↓: Scroll • q: Quit"),
		loading:        true,
		err:            nil,
	}
}

func (s *serviceaccountDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	s.k8sClient = k
	return s, nil
}

func (s *serviceaccountDetailsModel) fetchServiceAccountDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*s.k8sClient)
		desc, err := api.DescribeServiceAccount(s.serviceaccount.Namespace, s.serviceaccount.Name)
		return serviceaccountDetailsLoadedMsg{content: desc, err: err}
	}
}

func (s *serviceaccountDetailsModel) Init() tea.Cmd {
	return tea.Batch(s.yamlViewer.Init(), s.fetchServiceAccountDetails())
}

func (s *serviceaccountDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case serviceaccountDetailsLoadedMsg:
		s.loading = false
		if msg.err != nil {
			s.err = msg.err
			s.yamlViewer.SetContent("Error loading serviceaccount details: " + msg.err.Error())
		} else {
			title := "ServiceAccount: " + s.serviceaccount.Name
			s.yamlViewer = components.NewYAMLViewerWithHelp(title, msg.content, "↑/↓: Scroll • q: Quit")
		}
		return s, s.yamlViewer.Init()
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
