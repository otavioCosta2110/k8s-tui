package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type statefulsetsModel struct {
	*GenericResourceModel
	statefulsetsInfo []k8s.StatefulSetInfo
}

func NewStatefulSets(k k8s.Client, namespace string) (*statefulsetsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeStatefulSet,
		Title:           customstyles.ResourceIcons["StatefulSets"] + " StatefulSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.30, 0.26},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &statefulsetsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (ss *statefulsetsModel) Help() (string, string) {
	return "StatefulSets Help", `StatefulSets manage stateful applications with persistent storage.

Key Bindings:
• ↑/↓/j/k: Navigate statefulsets
• enter: View statefulset details
• R: Restart statefulset
• d: Delete selected statefulsets
• r: Refresh
• /: Search statefulsets
• esc: Go back

StatefulSet Status:
• Ready: Shows ready/desired replicas
• Each pod has a stable identity and storage

Common Actions:
• Scale statefulset: Change replica count
• View persistent volumes: See attached storage
• Check pod ordering: StatefulSets maintain pod identity
• Restart statefulset: Press 'R' to trigger a rollout restart`
}

func (ss *statefulsetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	ss.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		statefulsetDetails, err := NewStatefulSetDetails(*k, ss.pluginAPI.GetCurrentNamespace(), selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: statefulsetDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := ss.fetchData(); err != nil {
			return nil, err
		}
		return ss.dataToRows(), nil
	}

	tableModel := ui.NewTable(ss.config.Columns, ss.config.ColumnWidths, []table.Row{}, ss.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": ss.createDeleteAction(tableModel),
		"R": ss.createRestartAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, ss.refreshInterval, ss.k8sClient, "StatefulSets"), nil
}

func (ss *statefulsetsModel) createRestartAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(ss.statefulsetsInfo) {
			return nil
		}

		statefulset := ss.statefulsetsInfo[selected]

		return func() tea.Msg {
			statefulsetInfo := k8s.NewStatefulSetInfo(statefulset.Name, statefulset.Namespace, *ss.k8sClient)
			err := statefulsetInfo.Restart()
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *ss.k8sClient,
				}
			}

			return components.NavigateMsg{
				NewScreen: components.NewYAMLViewer(
					"StatefulSet Restarted",
					fmt.Sprintf("StatefulSet %s/%s has been restarted successfully.\n\nA rollout restart has been triggered, which will recreate all pods\nmanaged by this StatefulSet with new instances.", statefulset.Namespace, statefulset.Name),
				),
				Breadcrumb: statefulset.Name + " restarted",
			}
		}
	}
}

func (ss *statefulsetsModel) fetchData() error {
	var statefulsetInfo []k8s.StatefulSetInfo
	var err error

	statefulsetInfo, err = ss.pluginAPI.GetStatefulSets(ss.pluginAPI.GetCurrentNamespace())

	if err != nil {
		return fmt.Errorf("failed to fetch statefulsets: %v", err)
	}
	ss.statefulsetsInfo = statefulsetInfo

	ss.resourceData = make([]types.ResourceData, len(statefulsetInfo))
	for idx, statefulset := range statefulsetInfo {
		ss.resourceData[idx] = StatefulSetData{&statefulset}
	}

	return nil
}
