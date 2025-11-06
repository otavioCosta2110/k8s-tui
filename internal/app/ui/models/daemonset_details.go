package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type daemonsetDetailsModel struct {
	baseDetailsModel
	daemonset *k8s.DaemonSetInfo
}

type daemonsetDetailsLoadedMsg struct {
	content string
	err     error
}

func NewDaemonSetDetails(k k8s.Client, namespace, daemonsetName string) *daemonsetDetailsModel {
	return &daemonsetDetailsModel{
		baseDetailsModel: newBaseDetailsModel("DaemonSet: "+daemonsetName, "Loading daemonset details..."),
		daemonset:        k8s.NewDaemonSetInfo(daemonsetName, namespace, k),
	}
}

func (ds *daemonsetDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ds.k8sClient = k
	return ds, nil
}

func (ds *daemonsetDetailsModel) fetchDaemonSetDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*ds.k8sClient)
		desc, err := api.DescribeDaemonSet(ds.daemonset.Namespace, ds.daemonset.Name)
		return daemonsetDetailsLoadedMsg{content: desc, err: err}
	}
}

func (ds *daemonsetDetailsModel) Init() tea.Cmd {
	return tea.Batch(ds.yamlViewer.Init(), ds.spinner.Init(), ds.fetchDaemonSetDetails())
}

func (d *daemonsetDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case daemonsetDetailsLoadedMsg:
		d.setLoadedContent(msg.content, msg.err)
		return d, nil
	default:
		cmd, _ := d.updateCommon(msg)
		return d, cmd
	}
}

func (d *daemonsetDetailsModel) View() string {
	return d.baseDetailsModel.View()
}
