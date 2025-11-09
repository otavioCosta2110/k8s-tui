package models

import (
	"time"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type podsModel struct {
	*GenericResourceModel
	selector         string
	parentDeployment string
}

func NewPods(k k8s.Client, namespace string, selector ...string) (*podsModel, error) {
	return NewPodsWithParent(k, namespace, "", selector...)
}

func NewPodsWithParent(k k8s.Client, namespace, parentDeployment string, selector ...string) (*podsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypePod,
		Title:           styles.ResourceIcons["Pods"] + " Pods in " + namespace,
		ColumnWidths:    []float64{1, 1.5, 1.3, 0.3, 0.8, 0.5, 1},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("IMAGE", 0),
			components.NewColumn("READY", 0),
			components.NewColumn("STATUS", 0),
			components.NewColumn("RESTARTS", 0),
			components.NewColumn("AGE", 0),
		},
	}

	selectorStr := ""
	if len(selector) > 0 {
		selectorStr = selector[0]
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &podsModel{
		GenericResourceModel: genericModel,
		selector:             selectorStr,
		parentDeployment:     parentDeployment,
	}

	return model, nil
}

func (p *podsModel) Help() (string, string) {
	return "Pods Help", `Pods are the smallest deployable units in Kubernetes.

Key Bindings:
• ↑/↓/j/k: Navigate pods
• enter: View pod details
• d: Delete selected pods
• n: Create new pod
• r: Refresh
• /: Search pods
• esc: Go back

Pod Status:
• Running: Pod is running successfully
• Pending: Pod is being scheduled
• Failed: Pod has failed
• Succeeded: Pod completed successfully

Common Actions:
• View logs: Enter on a pod to see details
• Delete pod: Select with space, then press 'd'
• Create pod: Press 'n' to open create form
• Refresh: Press 'r' to update the list`
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		podDetails, err := NewPodDetails(*k, p.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen:  podDetails,
			Breadcrumb: selected,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := p.fetchData(p.selector); err != nil {
			return nil, err
		}
		return p.dataToRows(), nil
	}

	tableModel := ui.NewTable(p.config.Columns, p.config.ColumnWidths, []table.Row{}, p.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": p.createDeleteAction(tableModel),
		"n": p.createNewPodAction(),
	}

	if p.parentDeployment != "" {
		actions["v"] = p.createViewManifestAction(tableModel)
	}

	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, p.refreshInterval, p.k8sClient, "Pods"), nil
}

func (p *podsModel) createViewManifestAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		deploymentDetails, err := NewDeploymentDetails(*p.k8sClient, p.namespace, p.parentDeployment).InitComponent(p.k8sClient)
		if err != nil {
			return func() tea.Msg {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}
		}
		return func() tea.Msg {
			return components.NavigateMsg{
				NewScreen:  deploymentDetails,
				Breadcrumb: p.parentDeployment,
			}
		}
	}
}

func (p *podsModel) createNewPodAction() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return components.OpenCreateFormMsg{ResourceType: "pod"}
		}
	}
}

func (p *podsModel) fetchData(selector string) error {
	var podsInfo []k8s.PodInfo
	var err error

	podsInfo, err = p.pluginAPI.GetPods("", selector)

	if err != nil {
		return err
	}

	p.resourceData = make([]types.ResourceData, len(podsInfo))
	for i, pod := range podsInfo {
		p.resourceData[i] = PodData{&pod}
	}

	return nil
}
