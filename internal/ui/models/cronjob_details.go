package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

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

	var desc string
	var err error

	// Always use plugin API - resources should never bypass the plugin system
	pm := plugins.GetGlobalPluginManager()
	api := pm.GetAPI()
	api.SetClient(*k)
	desc, err = api.DescribeCronJob(cj.cronjob.Namespace, cj.cronjob.Name)

	if err != nil {
		return nil, err
	}

	return components.NewYAMLViewer("CronJob: "+cj.cronjob.Name, desc), nil
}
