package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type jobDetailsModel struct {
	job       *k8s.JobInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewJobDetails(k k8s.Client, namespace, jobName string) *jobDetailsModel {
	return &jobDetailsModel{
		job:       k8s.NewJob(jobName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (j *jobDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	j.k8sClient = k

	var desc string
	var err error

	// Always use plugin API - resources should never bypass the plugin system
	pm := plugins.GetGlobalPluginManager()
	api := pm.GetAPI()
	api.SetClient(*k)
	desc, err = api.DescribeJob(j.job.Namespace, j.job.Name)

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Job: "+j.job.Name, desc), nil
}
