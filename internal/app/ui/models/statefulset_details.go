package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type statefulsetDetailsModel struct {
	statefulset *k8s.StatefulSetInfo
	k8sClient   *k8s.Client
	yamlViewer  *components.YAMLViewer
	loading     bool
	err         error
}

type statefulsetDetailsLoadedMsg struct {
	content string
	err     error
}

func NewStatefulSetDetails(k k8s.Client, namespace, statefulsetName string) *statefulsetDetailsModel {
	return &statefulsetDetailsModel{
		statefulset: k8s.NewStatefulSet(statefulsetName, namespace, k),
		k8sClient:   &k,
		yamlViewer:  components.NewYAMLViewer("StatefulSet: "+statefulsetName, "Loading..."),
		loading:     true,
		err:         nil,
	}
}

func (ss *statefulsetDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ss.k8sClient = k
	return ss, nil
}

func (ss *statefulsetDetailsModel) fetchStatefulSetDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*ss.k8sClient)
		desc, err := api.DescribeStatefulSet(ss.statefulset.Namespace, ss.statefulset.Name)
		return statefulsetDetailsLoadedMsg{content: desc, err: err}
	}
}

func (ss *statefulsetDetailsModel) Init() tea.Cmd {
	return tea.Batch(ss.yamlViewer.Init(), ss.fetchStatefulSetDetails())
}

func (ss *statefulsetDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statefulsetDetailsLoadedMsg:
		ss.loading = false
		if msg.err != nil {
			ss.err = msg.err
			ss.yamlViewer.SetContent("Error loading statefulset details: " + msg.err.Error())
		} else {
			ss.yamlViewer.SetContent(msg.content)
		}
		return ss, nil
	default:
		var cmd tea.Cmd
		updatedViewer, cmd := ss.yamlViewer.Update(msg)
		ss.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return ss, cmd
	}
}

func (ss *statefulsetDetailsModel) View() string {
	return ss.yamlViewer.View()
}
