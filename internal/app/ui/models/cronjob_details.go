package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type cronjobDetailsModel struct {
	cronjob    *k8s.CronJobInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	loading    bool
	err        error
}

type cronjobDetailsLoadedMsg struct {
	content string
	err     error
}

func NewCronJobDetails(k k8s.Client, namespace, cronjobName string) *cronjobDetailsModel {
	return &cronjobDetailsModel{
		cronjob:    k8s.NewCronJob(cronjobName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("CronJob: "+cronjobName, "Loading..."),
		loading:    true,
		err:        nil,
	}
}

func (cj *cronjobDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	cj.k8sClient = k
	return cj, nil
}

func (cj *cronjobDetailsModel) fetchCronJobDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*cj.k8sClient)
		desc, err := api.DescribeCronJob(cj.cronjob.Namespace, cj.cronjob.Name)
		return cronjobDetailsLoadedMsg{content: desc, err: err}
	}
}

func (cj *cronjobDetailsModel) Init() tea.Cmd {
	return tea.Batch(cj.yamlViewer.Init(), cj.fetchCronJobDetails())
}

func (cj *cronjobDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cronjobDetailsLoadedMsg:
		cj.loading = false
		if msg.err != nil {
			cj.err = msg.err
			cj.yamlViewer.SetContent("Error loading cronjob details: " + msg.err.Error())
		} else {
			cj.yamlViewer.SetContent(msg.content)
		}
		return cj, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := cj.yamlViewer.Update(msg)
		cj.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return cj, cmd
	}
}

func (cj *cronjobDetailsModel) View() string {
	return cj.yamlViewer.View()
}
