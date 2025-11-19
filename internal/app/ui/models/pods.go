package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
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
		ColumnWidths:    []float64{0.2, 0.3, 0.12, 0.1, 0.05, 0.1, 0.1},
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

func (p *podsModel) SetNamespace(namespace string) {
	p.GenericResourceModel.SetNamespace(namespace)
	// Update the title to reflect the new namespace
	p.config.Title = styles.ResourceIcons["Pods"] + " Pods in " + namespace
}

func (p *podsModel) Help() (string, string) {
	return "Pods Help", `Pods are the smallest deployable units in Kubernetes.

Key Bindings:
• ↑/↓/j/k: Navigate pods
• enter: View pod logs
• e: Execute command in pod
• f: Port forward pod
• v: View pod details
• E: View pod events
• t: View resource usage
• R: Restart pod
• d: Delete selected pods
• n: Create new pod
• r: Refresh
• /: Search pods
• esc: Go back

Events View Key Bindings:
• g: Go to top
• G: Go to bottom
• r: Refresh events
• esc: Go back

Pod Status:
• Running: Pod is running successfully
• Pending: Pod is being scheduled
• Failed: Pod has failed
• Succeeded: Pod completed successfully

Common Actions:
• View logs: Enter on a pod to see logs
• View details: Press 'v' to see pod details
• View resource usage: Press 't' to see CPU/memory usage
• Restart pod: Press 'R' to restart the pod (only works for managed pods)
• Delete pod: Select with space, then press 'd'
• Create pod: Press 'n' to open create form
• Refresh: Press 'r' to update the list`
}

func (p *podsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	p.k8sClient = k

	onSelect := func(selected string) tea.Msg {
		pod := k8s.NewPodInfo(selected, p.pluginAPI.GetCurrentNamespace(), *k)
		logs, err := pod.GetLogs()
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen:  components.NewYAMLViewer("Pod Logs: "+selected, logs),
			Breadcrumb: selected + " logs",
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
		"e": p.createExecAction(tableModel),
		"f": p.createPortForwardAction(tableModel),
		"n": p.createNewPodAction(),
		"t": p.createViewResourceUsageAction(tableModel),
		"v": p.createViewDetailsAction(tableModel),
		"E": p.createViewEventsAction(tableModel),
		"R": p.createRestartAction(tableModel),
	}

	if p.parentDeployment != "" {
		actions["V"] = p.createViewManifestAction(tableModel)
	}

	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, p.refreshInterval, p.k8sClient, "Pods"), nil
}

func (p *podsModel) createViewDetailsAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			podDetails, err := NewPodDetails(*p.k8sClient, p.pluginAPI.GetCurrentNamespace(), podName).InitComponent(p.k8sClient)
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}
			return components.NavigateMsg{
				NewScreen:  podDetails,
				Breadcrumb: podName,
			}
		}
	}
}

func (p *podsModel) createViewEventsAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			eventsModel := NewEventsModel(p.k8sClient, p.pluginAPI.GetCurrentNamespace(), podName, "Pod")
			eventsScreen, err := eventsModel.InitComponent(p.k8sClient)
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}
			return components.NavigateMsg{
				NewScreen:  eventsScreen,
				Breadcrumb: podName + " events",
			}
		}
	}
}

func (p *podsModel) createViewLogsAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			pod := k8s.NewPodInfo(podName, p.pluginAPI.GetCurrentNamespace(), *p.k8sClient)
			logs, err := pod.GetLogs()
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}
			return components.NavigateMsg{
				NewScreen:  components.NewYAMLViewer("Pod Logs: "+podName, logs),
				Breadcrumb: podName + " logs",
			}
		}
	}
}

func (p *podsModel) createViewResourceUsageAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			pod := k8s.NewPodInfo(podName, p.pluginAPI.GetCurrentNamespace(), *p.k8sClient)
			metrics, err := pod.GetResourceUsage()
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}

			// Format metrics as YAML for display
			metricsYAML, err := p.formatMetricsAsYAML(metrics)
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}

			return components.NavigateMsg{
				NewScreen:  components.NewYAMLViewer("Resource Usage: "+podName, metricsYAML),
				Breadcrumb: podName + " usage",
			}
		}
	}
}

func (p *podsModel) formatMetricsAsYAML(metrics *k8s.PodMetrics) (string, error) {
	type ContainerUsage struct {
		Name   string `yaml:"name"`
		CPU    string `yaml:"cpu"`
		Memory string `yaml:"memory"`
	}

	type PodUsage struct {
		Name       string           `yaml:"name"`
		Namespace  string           `yaml:"namespace"`
		Containers []ContainerUsage `yaml:"containers"`
	}

	usage := PodUsage{
		Name:       metrics.Name,
		Namespace:  metrics.Namespace,
		Containers: make([]ContainerUsage, len(metrics.Containers)),
	}

	for i, container := range metrics.Containers {
		usage.Containers[i] = ContainerUsage{
			Name:   container.Name,
			CPU:    container.CPU,
			Memory: container.Memory,
		}
	}

	yamlData, err := yaml.Marshal(usage)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (p *podsModel) createViewManifestAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		deploymentDetails, err := NewDeploymentDetails(*p.k8sClient, p.pluginAPI.GetCurrentNamespace(), p.parentDeployment).InitComponent(p.k8sClient)
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

func (p *podsModel) createExecAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			onSubmit := func(value string) tea.Msg {
				logger.Info("Pods onSubmit called with value: " + value)
				// Split command by spaces, simple parsing
				command := strings.Fields(value)
				if len(command) == 0 {
					logger.Info("Pods onSubmit: empty command")
					return components.NavigateMsg{
						Error:   fmt.Errorf("empty command"),
						Cluster: *p.k8sClient,
					}
				}

				logger.Info("Pods onSubmit: executing command: " + fmt.Sprintf("%v", command))
				pod := k8s.NewPodInfo(podName, p.pluginAPI.GetCurrentNamespace(), *p.k8sClient)
				stdout, stderr, err := pod.Exec(command)
				logger.Info("Pods onSubmit: exec returned, err: " + fmt.Sprintf("%v", err))
				logger.Info("Pods onSubmit: stdout: " + stdout)
				logger.Info("Pods onSubmit: stderr: " + stderr)
				if err != nil {
					logger.Info("Pods onSubmit: exec error: " + err.Error())
					return components.NavigateMsg{
						Error:   err,
						Cluster: *p.k8sClient,
					}
				}

				logger.Info("Pods onSubmit: exec successful, stdout len: " + fmt.Sprintf("%d", len(stdout)) + ", stderr len: " + fmt.Sprintf("%d", len(stderr)))
				output := stdout
				if stderr != "" {
					if output != "" {
						output += "\n--- STDERR ---\n" + stderr
					} else {
						output = stderr
					}
				}

				logger.Info("Pods onSubmit: returning NavigateMsg with output")
				return components.NavigateMsg{
					NewScreen:  components.NewYAMLViewer("Command Output: "+podName, output),
					Breadcrumb: podName + " exec",
				}
			}

			onCancel := func() tea.Msg {
				return nil // Stay on current screen
			}

			return components.NavigateMsg{
				NewScreen:  components.NewTextInput("Execute command in "+podName, "", onSubmit, onCancel),
				Breadcrumb: podName + " exec input",
			}
		}
	}
}

func (p *podsModel) createRestartAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			pod := k8s.NewPodInfo(podName, p.pluginAPI.GetCurrentNamespace(), *p.k8sClient)
			err := pod.Restart()
			if err != nil {
				return components.NavigateMsg{
					Error:   err,
					Cluster: *p.k8sClient,
				}
			}

			return components.NavigateMsg{
				NewScreen: components.NewYAMLViewer(
					"Pod Restarted",
					fmt.Sprintf("Pod %s has been restarted successfully.\n\nIf this pod is managed by a controller (Deployment, ReplicaSet, etc.),\nit will be automatically recreated with a new instance.", podName),
				),
				Breadcrumb: podName + " restarted",
			}
		}
	}
}

func (p *podsModel) createPortForwardAction(tableModel *ui.TableModel) func() tea.Cmd {
	return func() tea.Cmd {
		if tableModel == nil {
			return nil
		}

		selected := tableModel.Table.Cursor()
		if selected < 0 || selected >= len(p.resourceData) {
			return nil
		}

		podData := p.resourceData[selected].(PodData)
		podName := podData.Name

		return func() tea.Msg {
			return components.OpenPortForwardFormMsg{
				Form:       components.NewPortForwardForm("Port Forward", "pod", podName, p.pluginAPI.GetCurrentNamespace()),
				Breadcrumb: podName + " port-forward",
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

	podsInfo, err = p.pluginAPI.GetPods(p.pluginAPI.GetCurrentNamespace(), selector)

	if err != nil {
		return err
	}

	p.resourceData = make([]types.ResourceData, len(podsInfo))
	for i, pod := range podsInfo {
		p.resourceData[i] = PodData{&pod}
	}

	return nil
}
