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

type replicasetsModel struct {
	*GenericResourceModel
	replicasetsInfo []k8s.ReplicaSetInfo
}

func NewReplicaSets(k k8s.Client, namespace string) (*replicasetsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeReplicaSet,
		Title:           customstyles.ResourceIcons["ReplicaSets"] + " ReplicaSets in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.12, 0.12, 0.15, 0.13},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("DESIRED", 0),
			components.NewColumn("CURRENT", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &replicasetsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (r *replicasetsModel) Help() (string, string) {
	return "ReplicaSets Help", `ReplicaSets ensure a specified number of pod replicas are running.

Key Bindings:
• ↑/↓/j/k: Navigate replicasets
• enter: View replicaset details
• R: Restart replicaset
• d: Delete selected replicasets
• r: Refresh
• /: Search replicasets
• esc: Go back

ReplicaSet Status:
• Desired: Number of desired pods
• Current: Number of current pods
• Ready: Number of ready pods

Common Actions:
• View pods: See pods managed by this replicaset
• Check owner: See which deployment owns this replicaset
• Restart replicaset: Press 'R' to trigger a rollout restart
• Manual scaling: Adjust replica count`
}

func (r *replicasetsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	r.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		replicaset := k8s.NewReplicaSetInfo(selected, r.pluginAPI.GetCurrentNamespace(), *k)
		err := replicaset.Fetch()
		if err != nil {
			return components.NavigateMsg{
				Error:   fmt.Errorf("failed to fetch replicaset: %v", err),
				Cluster: *k,
			}
		}
		selector, err := replicaset.GetLabelSelector()
		if err != nil {
			selector = fmt.Sprintf("app=%s", replicaset.Name)
		}
		pods, err := NewPodsWithParent(*k, r.pluginAPI.GetCurrentNamespace(), selected, selector)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		podsComponent, err := pods.InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}

		return components.NavigateMsg{
			NewScreen:  podsComponent,
			Breadcrumb: "Pods",
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := r.fetchData(); err != nil {
			return nil, err
		}
		return r.dataToRows(), nil
	}

	tableModel := ui.NewTable(r.config.Columns, r.config.ColumnWidths, []table.Row{}, r.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": r.createDeleteAction(tableModel),
		"R": r.createRestartAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, r.refreshInterval, r.k8sClient, "ReplicaSets"), nil
}

func (r *replicasetsModel) createRestartAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(r.replicasetsInfo) {
			return nil
		}

		replicaset := r.replicasetsInfo[selected]

		return func() tea.Msg {
			replicasetInfo := k8s.NewReplicaSetInfo(replicaset.Name, replicaset.Namespace, *r.k8sClient)
			err := replicasetInfo.Restart()
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *r.k8sClient,
				}
			}

			return components.NavigateMsg{
				NewScreen: components.NewYAMLViewer(
					"ReplicaSet Restarted",
					fmt.Sprintf("ReplicaSet %s/%s has been restarted successfully.\n\nA rollout restart has been triggered, which will recreate all pods\nmanaged by this ReplicaSet with new instances.", replicaset.Namespace, replicaset.Name),
				),
				Breadcrumb: replicaset.Name + " restarted",
			}
		}
	}
}

func (r *replicasetsModel) fetchData() error {
	var replicasetInfo []k8s.ReplicaSetInfo
	var err error

	replicasetInfo, err = r.pluginAPI.GetReplicaSets(r.pluginAPI.GetCurrentNamespace())

	if err != nil {
		return fmt.Errorf("failed to fetch replicasets: %v", err)
	}
	r.replicasetsInfo = replicasetInfo

	r.resourceData = make([]types.ResourceData, len(replicasetInfo))
	for i, replicaset := range replicasetInfo {
		r.resourceData[i] = ReplicaSetData{&replicaset}
	}

	return nil
}
