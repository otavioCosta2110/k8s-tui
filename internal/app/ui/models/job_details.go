package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type jobDetailsModel struct {
	baseDetailsModel
	job *k8s.JobInfo
}

type jobDetailsLoadedMsg struct {
	content string
	err     error
}

func NewJobDetails(k k8s.Client, namespace, jobName string) *jobDetailsModel {
	return &jobDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Job: "+jobName, "Loading job details..."),
		job:              k8s.NewJobInfo(jobName, namespace, k),
	}
}

func (j *jobDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	j.setClient(k)
	return j, nil
}

func (j *jobDetailsModel) fetchJobDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*j.k8sClient)
		desc, err := api.DescribeJob(j.job.Namespace, j.job.Name)
		return jobDetailsLoadedMsg{content: desc, err: err}
	}
}

func (j *jobDetailsModel) Init() tea.Cmd {
	return j.initCommon(j.fetchJobDetails())
}

func (j *jobDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case jobDetailsLoadedMsg:
		j.setLoadedContent(msg.content, msg.err)
		return j, nil
	default:
		cmd, _ := j.updateCommon(msg)
		return j, cmd
	}
}

func (j *jobDetailsModel) View() string {
	return j.baseDetailsModel.View()
}
