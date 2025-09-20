package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type cronjobDetailsModel struct {
	cronjob   *k8s.CronJobInfo
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewCronJobDetails(k k8s.Client, namespace, cronjobName string) *cronjobDetailsModel {
	return &cronjobDetailsModel{
		cronjob:   k8s.NewCronJob(cronjobName, namespace, k),
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}
}

func (cj *cronjobDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	cj.k8sClient = k

	desc, err := cj.cronjob.Describe()
	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("CronJob: "+cj.cronjob.Name, desc), nil
}
