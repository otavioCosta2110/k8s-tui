package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

func (m *AppModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	styles.ScreenWidth = msg.Width - styles.Margin
	styles.ScreenHeight = msg.Height - 1
	if !styles.IsHeaderActive {
		styles.HeaderSize = styles.ScreenHeight/4 - (styles.Margin * 2)
		styles.IsHeaderActive = true
	}
	styles.ScreenHeight -= styles.HeaderSize
	styles.ScreenHeight -= styles.TabBarSize
	styles.IsTabBarActive = true

	var cmds []tea.Cmd
	if m.configSelected {
		newHeader, headerCmd := m.header.Update(msg)
		m.header = newHeader.(models.HeaderModel)
		m.header.SetKubeconfig(&m.kube)
		m.header.UpdateContent()
		cmds = append(cmds, headerCmd)
	}

	if m.tabManager != nil {
		var cmd tea.Cmd
		updatedManager, cmd := m.tabManager.Update(msg)
		if manager, ok := updatedManager.(*models.TabManager); ok {
			m.tabManager = manager
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if m.quickNav != nil {
		var cmd tea.Cmd
		m.quickNav, cmd = m.quickNav.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.textInput != nil {
		switch msg.String() {
		case "esc":
			m.textInput = nil
			return m, nil
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	if m.quickNav != nil {
		switch msg.String() {
		case "esc", m.getKeyBinding("quick_nav"):
			m.quickNav = nil
			return m, nil
		default:
			var cmd tea.Cmd
			m.quickNav, cmd = m.quickNav.Update(msg)
			return m, cmd
		}
	}

	if m.helpScreen != nil && m.helpScreen.IsVisible() {
		switch msg.String() {
		case "esc", "q", "?":
			m.helpScreen.Close()
			return m, nil
		default:
			var cmd tea.Cmd
			updatedHelp, cmd := m.helpScreen.Update(msg)
			if help, ok := updatedHelp.(*components.HelpModel); ok {
				m.helpScreen = help
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "esc":
		if m.errorPopup != nil {
			m.errorPopup = nil
			return m, nil
		}
		return m, tea.Quit
	case m.getKeyBinding("quit"), m.getKeyBinding("back"), m.getKeyBinding("forward"):
		if m.tabManager != nil {
			updatedManager, cmd := m.tabManager.Update(msg)
			if manager, ok := updatedManager.(*models.TabManager); ok {
				m.tabManager = manager
				m.updateHeaderTabs()
			}
			return m, cmd
		}
		return m, nil
	case "Q":
		return m, tea.Quit
	case m.getKeyBinding("quick_nav"):
		if m.quickNav != nil {
			m.quickNav = nil
			return m, nil
		}
		m.quickNav = models.NewQuickNavModel(m.kube, m.kube.Namespace)
		return m, m.quickNav.Init()
	case m.getKeyBinding("help"):
		if m.helpScreen != nil && m.pluginManager != nil {
			resourceType := m.pluginManager.GetAPI().GetCurrentResourceType()
			title, content := m.pluginManager.GetAPI().GetHelp(resourceType)
			m.helpScreen.SetContent(title, content)
		}
		return m, nil
	case m.getKeyBinding("new_tab"):
		if m.tabManager != nil {
			updatedManager, cmd := m.tabManager.Update(msg)
			if manager, ok := updatedManager.(*models.TabManager); ok {
				m.tabManager = manager
				m.updateHeaderTabs()
			}
			return m, cmd
		}
		return m, nil

	case "left", "right":
		if m.header.GetTabCount() > 0 {
			newHeader, headerCmd := m.header.Update(msg)
			if header, ok := newHeader.(models.HeaderModel); ok {
				m.header = header
				if m.tabManager != nil {
					m.tabManager.SetActiveTab(m.header.GetActiveTabIndex())
				}
				return m, headerCmd
			}
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if m.header.GetTabCount() > 0 {
			newHeader, headerCmd := m.header.Update(msg)
			if header, ok := newHeader.(models.HeaderModel); ok {
				m.header = header
				if m.tabManager != nil {
					m.tabManager.SetActiveTab(m.header.GetActiveTabIndex())
				}
				return m, headerCmd
			}
		}
		return m, nil
	case m.getKeyBinding("close_tab"):
		if m.header.GetTabCount() > 1 {
			newHeader, headerCmd := m.header.Update(msg)
			if header, ok := newHeader.(models.HeaderModel); ok {
				m.header = header
				m.updateHeaderTabs()
				return m, headerCmd
			}
		}
		return m, nil

	default:
		if command, exists := m.config.KeyBindings[msg.String()]; exists {
			if m.pluginManager != nil {
				result, err := m.pluginManager.GetAPI().ExecuteCommand(command, []string{})
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to execute command %s: %v", command, err))
				} else {
					logger.Info(fmt.Sprintf("Executed command %s: %s", command, result))
				}
				if m.pendingInputDialog != nil {
					request := m.pendingInputDialog
					m.pendingInputDialog = nil

					onSubmit := func(value string) tea.Msg {
						if m.pluginManager != nil {
							m.pluginManager.GetAPI().ExecuteCommand(request.SubmitCommand, []string{value})
						}
						return ClearTextInputMsg{}
					}

					onCancel := func() tea.Msg {
						if m.pluginManager != nil && request.CancelCommand != "" {
							m.pluginManager.GetAPI().ExecuteCommand(request.CancelCommand, []string{})
						}
						return ClearTextInputMsg{}
					}

					m.textInput = components.NewTextInput(request.Title, "", onSubmit, onCancel)
					return m, m.textInput.Init()
				}
				return m, nil
			}
		}

		if m.tabManager != nil {
			updatedManager, cmd := m.tabManager.Update(msg)
			if manager, ok := updatedManager.(*models.TabManager); ok {
				m.tabManager = manager
			}
			return m, cmd
		}
		return m, nil
	}
}

func (m *AppModel) handleNavigateMsg(msg components.NavigateMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		popup := models.NewErrorScreen(
			msg.Error,
			"Kubernetes Connection Error",
			"Failed to connect to the Kubernetes cluster",
		)
		popup.SetDimensions(styles.ScreenWidth, styles.ScreenHeight)

		return &AppModel{
			tabManager: m.tabManager,
			header:     m.header,
			kube:       msg.Cluster,
			errorPopup: &popup,
			quickNav:   nil,
		}, nil
	}

	m.quickNav = nil

	if m.tabManager != nil {
		updatedManager, cmd := m.tabManager.Update(msg)
		if manager, ok := updatedManager.(*models.TabManager); ok {
			m.tabManager = manager
		}

		m.updateHeaderTabs()

		if !m.configSelected {
			m.configSelected = true
			m.header.SetKubeconfig(&msg.Cluster)
			m.kube = msg.Cluster
			m.header.UpdateContent()

			return m, tea.Batch(
				cmd,
				m.header.Init(),
			)
		}
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) handleTabMsg(msg components.TabMsg) (tea.Model, tea.Cmd) {
	if m.tabManager != nil {
		updatedManager, cmd := m.tabManager.Update(msg)
		if manager, ok := updatedManager.(*models.TabManager); ok {
			m.tabManager = manager
			m.updateHeaderTabs()
		}
		return m, cmd
	}
	return m, nil
}

func (m *AppModel) handleHeaderRefreshMsg(msg models.HeaderRefreshMsg) (tea.Model, tea.Cmd) {
	if m.configSelected {
		newHeader, headerCmd := m.header.Update(msg)
		m.header = newHeader.(models.HeaderModel)
		return m, headerCmd
	}
	return m, nil
}

func (m *AppModel) handleCloseQuickNavMsg(msg models.CloseQuickNavMsg) (tea.Model, tea.Cmd) {
	m.quickNav = nil
	return m, nil
}

func (m *AppModel) handleCreateSubmitMsg(msg components.CreateSubmitMsg) (tea.Model, tea.Cmd) {
	var err error
	switch msg.ResourceType {
	case "pod":
		pod := resources.NewPod("", "", m.kube)
		err = pod.Create(msg.Values["name"], msg.Values["image"], msg.Values["namespace"])
	case "deployment":
		deployment := resources.NewDeployment("", "", m.kube)
		err = deployment.Create(msg.Values["name"], msg.Values["image"], msg.Values["replicas"], msg.Values["namespace"])
	case "service":
		service := resources.NewService("", "", m.kube)
		err = service.Create(msg.Values["name"], msg.Values["type"], msg.Values["port"], msg.Values["targetPort"], msg.Values["namespace"])
	case "configmap":
		configmap := resources.NewConfigmap("", "", m.kube)
		err = configmap.Create(msg.Values["name"], msg.Values["namespace"])
	case "secret":
		secret := resources.NewSecret("", "", m.kube)
		err = secret.Create(msg.Values["name"], msg.Values["type"], msg.Values["namespace"])
	case "ingress":
		ingress := resources.NewIngress("", "", m.kube)
		err = ingress.Create(msg.Values["name"], msg.Values["host"], msg.Values["path"], msg.Values["serviceName"], msg.Values["servicePort"], msg.Values["namespace"])
	case "job":
		job := resources.NewJob("", "", m.kube)
		err = job.Create(msg.Values["name"], msg.Values["image"], msg.Values["backoffLimit"], msg.Values["namespace"])
	case "cronjob":
		cronjob := resources.NewCronJob("", "", m.kube)
		err = cronjob.Create(msg.Values["name"], msg.Values["image"], msg.Values["schedule"], msg.Values["suspend"], msg.Values["namespace"])
	case "daemonset":
		daemonset := resources.NewDaemonSet("", "", m.kube)
		err = daemonset.Create(msg.Values["name"], msg.Values["image"], msg.Values["namespace"])
	case "statefulset":
		statefulset := resources.NewStatefulSet("", "", m.kube)
		err = statefulset.Create(msg.Values["name"], msg.Values["image"], msg.Values["replicas"], msg.Values["namespace"])
	case "namespace":
		namespace := resources.NewNamespaces(m.kube)
		err = namespace.Create(msg.Values["name"])
	case "node":
		err = fmt.Errorf("creating nodes is not supported")
	default:
		err = fmt.Errorf("unknown resource type: %s", msg.ResourceType)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create %s: %v", msg.ResourceType, err))
	}
	m.textInput = nil
	return m, nil
}

func (m *AppModel) handleOpenCreateFormMsg(msg components.OpenCreateFormMsg) (tea.Model, tea.Cmd) {
	var fields []components.Field
	switch msg.ResourceType {
	case "pod":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 2},
		}
	case "deployment":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "replicas", Placeholder: "Replicas", DefaultValue: "1", Row: 1},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 2},
		}
	case "service":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "type", Placeholder: "Type", DefaultValue: "ClusterIP", Row: 1},
			{Name: "port", Placeholder: "Port", DefaultValue: "", Row: 2},
			{Name: "targetPort", Placeholder: "Target Port", DefaultValue: "", Row: 2},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 3},
		}
	case "configmap":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 1},
		}
	case "secret":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "type", Placeholder: "Type", DefaultValue: "Opaque", Row: 1},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 2},
		}
	case "ingress":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "host", Placeholder: "Host", DefaultValue: "", Row: 1},
			{Name: "path", Placeholder: "Path", DefaultValue: "/", Row: 2},
			{Name: "serviceName", Placeholder: "Service Name", DefaultValue: "", Row: 3},
			{Name: "servicePort", Placeholder: "Service Port", DefaultValue: "", Row: 3},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 4},
		}
	case "job":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "backoffLimit", Placeholder: "Backoff Limit", DefaultValue: "6", Row: 2},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 3},
		}
	case "cronjob":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "schedule", Placeholder: "Schedule", DefaultValue: "*/5 * * * *", Row: 2},
			{Name: "suspend", Placeholder: "Suspend", DefaultValue: "false", Row: 3},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 4},
		}
	case "daemonset":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 2},
		}
	case "statefulset":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "image", Placeholder: "Image", DefaultValue: "", Row: 1},
			{Name: "replicas", Placeholder: "Replicas", DefaultValue: "1", Row: 1},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 2},
		}
	case "namespace":
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
		}
	default:
		fields = []components.Field{
			{Name: "name", Placeholder: "Name", DefaultValue: "", Row: 0},
			{Name: "namespace", Placeholder: "Namespace", DefaultValue: m.kube.Namespace, Row: 1},
		}
	}
	title := "Create New " + msg.ResourceType
	m.textInput = components.NewCreateForm(title, msg.ResourceType, fields)
	return m, m.textInput.Init()
}
