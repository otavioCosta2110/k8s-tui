package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type daemonsetDetailsModel struct {
	daemonset  *k8s.DaemonSetInfo
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	loading    bool
	err        error
}

type daemonsetDetailsLoadedMsg struct {
	content string
	err     error
}

func NewDaemonSetDetails(k k8s.Client, namespace, daemonsetName string) *daemonsetDetailsModel {
	return &daemonsetDetailsModel{
		daemonset:  k8s.NewDaemonSet(daemonsetName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewer("DaemonSet: "+daemonsetName, "Loading..."),
		loading:    true,
		err:        nil,
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
	return tea.Batch(ds.yamlViewer.Init(), ds.fetchDaemonSetDetails())
}

func (ds *daemonsetDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case daemonsetDetailsLoadedMsg:
		ds.loading = false
		if msg.err != nil {
			ds.err = msg.err
			ds.yamlViewer.SetContent("Error loading daemonset details: " + msg.err.Error())
		} else {
			ds.yamlViewer.SetContent(msg.content)
		}
		return ds, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := ds.yamlViewer.Update(msg)
		ds.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return ds, cmd
	}
}

func (ds *daemonsetDetailsModel) View() string {
	return ds.yamlViewer.View()
}
