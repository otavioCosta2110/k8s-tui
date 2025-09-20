package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"

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

	desc, err := j.job.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("Job: "+j.job.Name, desc), nil
}
