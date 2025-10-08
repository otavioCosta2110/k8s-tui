package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type jobDetailsModel struct {
	job        *k8s.JobInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	loading    bool
	err        error
}

type jobDetailsLoadedMsg struct {
	content string
	err     error
}

func NewJobDetails(k k8s.Client, namespace, jobName string) *jobDetailsModel {
	return &jobDetailsModel{
		job:        k8s.NewJob(jobName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("Job: "+jobName, "Loading..."),
		loading:    true,
		err:        nil,
	}
}

func (j *jobDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	j.k8sClient = k
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
	return tea.Batch(j.yamlViewer.Init(), j.fetchJobDetails())
}

func (j *jobDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case jobDetailsLoadedMsg:
		j.loading = false
		if msg.err != nil {
			j.err = msg.err
			j.yamlViewer.SetContent("Error loading job details: " + msg.err.Error())
		} else {
			j.yamlViewer.SetContent(msg.content)
		}
		return j, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := j.yamlViewer.Update(msg)
		j.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return j, cmd
	}
}

func (j *jobDetailsModel) View() string {
	return j.yamlViewer.View()
}
