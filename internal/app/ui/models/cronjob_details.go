package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type cronjobDetailsModel struct {
	baseDetailsModel
	cronjob *k8s.CronJobInfo
}

type cronjobDetailsLoadedMsg struct {
	content string
	err     error
}

func NewCronJobDetails(k k8s.Client, namespace, cronjobName string) *cronjobDetailsModel {
	return &cronjobDetailsModel{
		baseDetailsModel: newBaseDetailsModel("CronJob: "+cronjobName, "Loading cronjob details..."),
		cronjob:          k8s.NewCronJobInfo(cronjobName, namespace, k),
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
	return tea.Batch(cj.yamlViewer.Init(), cj.spinner.Init(), cj.fetchCronJobDetails())
}

func (c *cronjobDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cronjobDetailsLoadedMsg:
		c.setLoadedContent(msg.content, msg.err)
		return c, nil
	default:
		cmd, _ := c.updateCommon(msg)
		return c, cmd
	}
}

func (c *cronjobDetailsModel) View() string {
	return c.baseDetailsModel.View()
}
