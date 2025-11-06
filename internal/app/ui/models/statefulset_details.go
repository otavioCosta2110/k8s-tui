package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

type statefulsetDetailsModel struct {
	baseDetailsModel
	statefulset *k8s.StatefulSetInfo
}

type statefulsetDetailsLoadedMsg struct {
	content string
	err     error
}

func NewStatefulSetDetails(k k8s.Client, namespace, statefulsetName string) *statefulsetDetailsModel {
	return &statefulsetDetailsModel{
		baseDetailsModel: newBaseDetailsModel("StatefulSet: "+statefulsetName, "Loading statefulset details..."),
		statefulset:      k8s.NewStatefulSetInfo(statefulsetName, namespace, k),
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
	return tea.Batch(ss.yamlViewer.Init(), ss.spinner.Init(), ss.fetchStatefulSetDetails())
}

func (s *statefulsetDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statefulsetDetailsLoadedMsg:
		s.setLoadedContent(msg.content, msg.err)
		return s, nil
	default:
		cmd, _ := s.updateCommon(msg)
		return s, cmd
	}
}

func (s *statefulsetDetailsModel) View() string {
	return s.baseDetailsModel.View()
}
