package ui

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"io/ioutil"
	"strings"
)

type UIInjector struct {
	injections map[string][]plugins.UIInjectionPoint
}

func NewUIInjector() *UIInjector {
	return &UIInjector{
		injections: make(map[string][]plugins.UIInjectionPoint),
	}
}

func (ui *UIInjector) AddInjection(location string, injection plugins.UIInjectionPoint) {
	ui.injections[location] = append(ui.injections[location], injection)
}

func (ui *UIInjector) GetInjections(location string) []plugins.UIInjectionPoint {
	return ui.injections[location]
}

func (ui *UIInjector) RenderInjections(location string) string {
	injections := ui.GetInjections(location)
	if len(injections) == 0 {
		return ""
	}

	var rendered []string
	for _, injection := range injections {
		switch injection.Component.Type {
		case "text":
			if content, ok := injection.Component.Config["content"].(string); ok {
				rendered = append(rendered, content)
			}
		default:
			if content, ok := injection.Component.Config["content"].(string); ok {
				rendered = append(rendered, content)
			}
		}
	}

	return strings.Join(rendered, " | ")
}

type SessionTab struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	ResourceType string   `json:"resourceType"`
	Breadcrumb   []string `json:"breadcrumb"`
}

type SessionData struct {
	Timestamp string       `json:"timestamp"`
	Namespace string       `json:"namespace"`
	Tabs      []SessionTab `json:"tabs"`
}

type AppModel struct {
	tabManager          *models.TabManager
	header              models.HeaderModel
	kube                resources.Client
	config              config.AppConfig
	configSelected      bool
	errorPopup          *models.ErrorModel
	quickNav            tea.Model
	currentResourceType string
	breadcrumbTrail     []string
	pluginManager       *plugins.PluginManager
	uiInjector          *UIInjector
}

func loadSession(sessionFile string) (*SessionData, error) {
	if sessionFile == "" {
		return nil, nil
	}

	data, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file %s: %v", sessionFile, err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file %s: %v", sessionFile, err)
	}

	return &session, nil
}

func (m *AppModel) restoreSession(session *SessionData) error {
	if session == nil || len(session.Tabs) == 0 {
		return nil
	}

	logger.Info(fmt.Sprintf("Restoring session with %d tabs", len(session.Tabs)))

	for len(m.tabManager.GetTabsForComponent()) > 1 {
		lastTab := m.tabManager.GetTabsForComponent()[len(m.tabManager.GetTabsForComponent())-1]
		m.tabManager.Update(components.TabMsg{Action: "close", TabID: lastTab.ID})
	}

	for _, sessionTab := range session.Tabs {
		if err := m.restoreTab(sessionTab); err != nil {
			logger.Error(fmt.Sprintf("Failed to restore tab %s: %v", sessionTab.Title, err))
			continue
		}
	}

	return nil
}

func (m *AppModel) restoreTab(sessionTab SessionTab) error {
	logger.Info(fmt.Sprintf("Restoring tab: %s (%s) - Breadcrumb: %v",
		sessionTab.Title, sessionTab.ResourceType, sessionTab.Breadcrumb))

	if sessionTab.ResourceType == "ResourceList" {
		return nil
	}

	model, err := m.createModelForBreadcrumb(sessionTab.Breadcrumb)
	if err != nil {
		return fmt.Errorf("failed to create model for breadcrumb %v: %v", sessionTab.Breadcrumb, err)
	}

	if model != nil {
		_, cmd := m.tabManager.CreateNewTab(model, sessionTab.Title)
		if cmd != nil {
			model, _ = model.Update(cmd)
		}
		logger.Info(fmt.Sprintf("Successfully restored tab: %s", sessionTab.Title))
	}

	return nil
}

func (m *AppModel) createModelForBreadcrumb(breadcrumb []string) (tea.Model, error) {
	if len(breadcrumb) < 2 {
		return nil, fmt.Errorf("breadcrumb too short: %v", breadcrumb)
	}

	if breadcrumb[0] != "Resource List" {
		return nil, fmt.Errorf("invalid breadcrumb start: %s", breadcrumb[0])
	}

	resourceType := breadcrumb[1]

	if len(breadcrumb) == 2 {
		return models.NewResourceList(m.kube, m.config.DefaultNamespace, resourceType).InitComponent(m.kube)
	}

	if len(breadcrumb) >= 3 {
		resourceName := breadcrumb[2]
		return m.createResourceDetailsModel(resourceType, resourceName)
	}

	return nil, fmt.Errorf("unable to determine model type for breadcrumb: %v", breadcrumb)
}

func (m *AppModel) createResourceDetailsModel(resourceType, resourceName string) (tea.Model, error) {
	switch resourceType {
	case "Pods", "pods":
		return models.NewPodDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Services", "services":
		return models.NewServiceDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "ConfigMaps", "configmaps":
		return models.NewConfigmapDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Secrets", "secrets":
		return models.NewSecretDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Ingresses", "ingresses":
		return models.NewIngressDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Jobs", "jobs":
		return models.NewJobDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "CronJobs", "cronjobs":
		return models.NewCronJobDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "DaemonSets", "daemonsets":
		return models.NewDaemonSetDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "StatefulSets", "statefulsets":
		return models.NewStatefulSetDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "ServiceAccounts", "serviceaccounts":
		return models.NewServiceAccountDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Nodes", "nodes":
		return models.NewNodeDetails(m.kube, resourceName).InitComponent(&m.kube)
	default:
		return nil, fmt.Errorf("unsupported resource type for details: %s", resourceType)
	}
}

func NewAppModel(cfg cli.Config, pluginManager *plugins.PluginManager) *AppModel {

	appConfig, err := config.LoadAppConfig()
	if err != nil {
		panic("Failed to load app config: " + err.Error())
	}

	if err := customstyles.InitColors(); err != nil {
		panic("Failed to initialize colors: " + err.Error())
	}

	kubeClient, err := resources.NewClient(cfg.KubeconfigPath, cfg.Namespace)
	if err == nil && kubeClient != nil {
		header := models.NewHeader("K8s TUI", kubeClient)
		header.SetNamespace(cfg.Namespace)

		tabManager := models.NewTabManager(kubeClient, cfg.Namespace, appConfig.KeyBindings)

		uiInjector := NewUIInjector()

		appModel := &AppModel{
			tabManager:     tabManager,
			header:         header,
			kube:           *kubeClient,
			config:         appConfig,
			configSelected: true,
			pluginManager:  pluginManager,
			uiInjector:     uiInjector,
		}

		if pluginManager != nil {
			pluginManager.GetAPI().SetTabGetter(func() ([]plugins.TabInfo, error) {
				tabs := tabManager.GetTabsForComponent()
				var tabInfos []plugins.TabInfo
				for _, tab := range tabs {
					tabInfos = append(tabInfos, plugins.TabInfo{
						ID:           tab.ID,
						Title:        tab.Title,
						ResourceType: tab.ResourceType,
						Breadcrumb:   tab.Breadcrumb,
					})
				}
				return tabInfos, nil
			})

			pluginManager.GetAPI().SetTabRestorer(func(tabInfos []plugins.TabInfo) error {
				logger.Info(fmt.Sprintf("DEBUG: SetTabRestorer callback called with %d tabInfos", len(tabInfos)))
				for i, tabInfo := range tabInfos {
					logger.Info(fmt.Sprintf("DEBUG: TabInfo %d: ID=%s, Title=%s, ResourceType=%s, Breadcrumb=%v",
						i, tabInfo.ID, tabInfo.Title, tabInfo.ResourceType, tabInfo.Breadcrumb))
				}
				err := tabManager.RestoreTabs(tabInfos)
				if err != nil {
					logger.Error(fmt.Sprintf("DEBUG: tabManager.RestoreTabs failed: %v", err))
					return err
				}
				logger.Info("DEBUG: tabManager.RestoreTabs succeeded, calling updateHeaderTabs")
				appModel.updateHeaderTabs()
				logger.Info("DEBUG: updateHeaderTabs completed")
				return nil
			})

			pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
				logger.Info(fmt.Sprintf("DEBUG: SetNamespaceCallback called with namespace: %s", namespace))
				appModel.header.SetNamespace(namespace)
				logger.Info("DEBUG: Called header.SetNamespace, now calling UpdateContent")
				appModel.header.UpdateContent()
				logger.Info(fmt.Sprintf("Plugin changed namespace to: %s", namespace))
			})

			pluginManager.GetAPI().SetStatusCallback(func(message string) {
				logger.Info(fmt.Sprintf("Plugin status: %s", message))
			})

			appModel.loadPluginUIExtensions()

			if err := cli.HandlePluginArgs(pluginManager, cfg.PluginArgs); err != nil {
				logger.Error(fmt.Sprintf("Failed to handle plugin CLI arguments: %v", err))
			}
		}

		tabs := tabManager.GetTabsForComponent()
		activeIndex := -1
		for i, tab := range tabs {
			header.AddTab(tab.ID, tab.Title, tab.ResourceType)
			if tab.IsActive {
				activeIndex = i
			}
		}
		if activeIndex >= 0 {
			header.SetActiveTab(activeIndex)
		}

		return appModel
	}

	_, err = models.NewKubeconfigModel().InitComponent(nil)
	if err != nil {

		popup := models.NewErrorScreen(err, "Failed to initialize Kubernetes config", "")
		uiInjector := NewUIInjector()
		appModel := &AppModel{
			header:        models.NewHeader("K8s TUI", nil),
			config:        appConfig,
			errorPopup:    &popup,
			pluginManager: pluginManager,
			uiInjector:    uiInjector,
		}

		if pluginManager != nil {
			pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
				logger.Info(fmt.Sprintf("DEBUG: SetNamespaceCallback called with namespace: %s", namespace))
				appModel.header.SetNamespace(namespace)
				logger.Info("DEBUG: Called header.SetNamespace, now calling UpdateContent")
				appModel.header.UpdateContent()
				logger.Info(fmt.Sprintf("Plugin changed namespace to: %s", namespace))
			})

			pluginManager.GetAPI().SetStatusCallback(func(message string) {
				logger.Info(fmt.Sprintf("Plugin status: %s", message))
			})

			appModel.loadPluginUIExtensions()

			if err := cli.HandlePluginArgs(pluginManager, cfg.PluginArgs); err != nil {
				logger.Error(fmt.Sprintf("Failed to handle plugin CLI arguments: %v", err))
			}
		}

		return appModel
	}

	uiInjector := NewUIInjector()
	appModel := &AppModel{
		header:        models.NewHeader("K8s TUI", nil),
		config:        appConfig,
		pluginManager: pluginManager,
		uiInjector:    uiInjector,
	}

	if pluginManager != nil {
		pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
			appModel.header.SetNamespace(namespace)
			appModel.header.UpdateContent()
			logger.Info(fmt.Sprintf("Plugin changed namespace to: %s", namespace))
		})

		pluginManager.GetAPI().SetStatusCallback(func(message string) {
			logger.Info(fmt.Sprintf("Plugin status: %s", message))
		})

		appModel.loadPluginUIExtensions()

		if err := cli.HandlePluginArgs(pluginManager, cfg.PluginArgs); err != nil {
			logger.Error(fmt.Sprintf("Failed to handle plugin CLI arguments: %v", err))
		}
	}

	return appModel
}

func ParseFlags() cli.Config {
	return cli.ParseFlags()
}

func (m *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	if m.tabManager != nil {
		cmds = append(cmds, m.tabManager.Init())
	}

	if m.configSelected {
		cmds = append(cmds, m.header.Init())
	}

	return tea.Batch(cmds...)
}

func (m *AppModel) getKeyBinding(action string) string {
	if binding, exists := m.config.KeyBindings[action]; exists {
		return binding
	}
	defaults := map[string]string{
		"quit":      "q",
		"help":      "?",
		"refresh":   "r",
		"back":      "[",
		"forward":   "]",
		"new_tab":   "ctrl+t",
		"close_tab": "ctrl+w",
		"quick_nav": "g",
	}
	return defaults[action]
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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

	case tea.KeyMsg:
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

	case components.NavigateMsg:
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

	case components.TabMsg:
		if m.tabManager != nil {
			updatedManager, cmd := m.tabManager.Update(msg)
			if manager, ok := updatedManager.(*models.TabManager); ok {
				m.tabManager = manager
				m.updateHeaderTabs()
			}
			return m, cmd
		}
		return m, nil

	case models.HeaderRefreshMsg:
		if m.configSelected {
			newHeader, headerCmd := m.header.Update(msg)
			m.header = newHeader.(models.HeaderModel)
			return m, headerCmd
		}
		return m, nil

	case models.CloseQuickNavMsg:
		m.quickNav = nil
		return m, nil

	default:
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

func (m *AppModel) View() string {
	if m.quickNav != nil {
		return m.quickNav.View()
	}

	if m.errorPopup != nil {
		return m.errorPopup.View()
	}

	if m.tabManager == nil {
		return "Loading..."
	}

	currentView := m.tabManager.View()

	height := styles.ScreenHeight

	headerView := m.header.View()

	content := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(customstyles.BorderColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(currentView)

	var breadcrumbView string
	if m.tabManager != nil {
		if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
			var breadcrumbParts []string
			breadcrumbEnd := min(activeTab.CurrentIndex+1, len(activeTab.Breadcrumb))
			for i := range breadcrumbEnd {
				crumb := activeTab.Breadcrumb[i]
				if i == activeTab.CurrentIndex && activeTab.CurrentIndex < len(activeTab.Breadcrumb) {
					breadcrumbParts = append(breadcrumbParts, lipgloss.NewStyle().
						Foreground(lipgloss.Color(customstyles.AccentColor)).
						Bold(true).
						Background(lipgloss.Color(customstyles.BackgroundColor)).
						Render(crumb))
				} else {
					breadcrumbParts = append(breadcrumbParts, lipgloss.NewStyle().
						Foreground(lipgloss.Color("240")).
						Background(lipgloss.Color(customstyles.BackgroundColor)).
						Render(crumb))
				}
			}
			breadCrumbArrow := lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" > ")
			breadcrumbStr := strings.Join(breadcrumbParts, breadCrumbArrow)
			breadcrumbView = lipgloss.NewStyle().
				Width(styles.ScreenWidth + styles.Margin).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(breadcrumbStr)
		}
	}

	if !m.configSelected {
		var finalView string
		if breadcrumbView != "" {
			finalView = lipgloss.JoinVertical(lipgloss.Top,
				lipgloss.NewStyle().
					Width(styles.ScreenWidth).
					Height(styles.ScreenHeight+styles.HeaderSize).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color(customstyles.BorderColor)).
					BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					Render(currentView),
				breadcrumbView)
		} else {
			finalView = lipgloss.NewStyle().
				Width(styles.ScreenWidth).
				Height(styles.ScreenHeight + styles.HeaderSize).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(customstyles.BorderColor)).
				BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(currentView)
		}

		return finalView
	}

	if !styles.IsHeaderActive {
		var finalView string
		if breadcrumbView != "" {
			finalView = lipgloss.JoinVertical(lipgloss.Top, content, breadcrumbView)
		} else {
			finalView = content
		}

		return finalView
	}

	footerInjections := m.uiInjector.RenderInjections("footer")
	var footerView string
	if footerInjections != "" {
		footerView = lipgloss.NewStyle().
			Width(styles.ScreenWidth).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1).
			Render(footerInjections)
	}

	var finalView string
	if breadcrumbView != "" && footerView != "" {
		finalView = lipgloss.JoinVertical(lipgloss.Top, headerView, content, breadcrumbView, footerView)
	} else if breadcrumbView != "" {
		finalView = lipgloss.JoinVertical(lipgloss.Top, headerView, content, breadcrumbView)
	} else if footerView != "" {
		finalView = lipgloss.JoinVertical(lipgloss.Top, headerView, content, footerView)
	} else {
		finalView = lipgloss.JoinVertical(lipgloss.Top, headerView, content)
	}

	return finalView
}

func (m *AppModel) getResourceTypeFromKey(key string) string {
	resourceMap := map[string]string{
		"p": "Pods",
		"d": "Deployments",
		"s": "Services",
		"i": "Ingresses",
		"c": "ConfigMaps",
		"e": "Secrets",
		"a": "ServiceAccounts",
		"r": "ReplicaSets",
		"n": "Nodes",
		"j": "Jobs",
		"k": "CronJobs",
		"m": "DaemonSets",
		"t": "StatefulSets",
		"l": "ResourceList",
	}

	if resourceType, exists := resourceMap[key]; exists {
		return resourceType
	}
	return ""
}

func (m *AppModel) updateHeaderTabs() {
	logger.Info("DEBUG: updateHeaderTabs called")
	if m.tabManager != nil {
		tabs := m.tabManager.GetTabsForComponent()
		logger.Info(fmt.Sprintf("DEBUG: Got %d tabs from tabManager", len(tabs)))
		for i, tab := range tabs {
			logger.Info(fmt.Sprintf("DEBUG: Tab %d: ID=%s, Title=%s, ResourceType=%s, IsActive=%v",
				i, tab.ID, tab.Title, tab.ResourceType, tab.IsActive))
		}
		m.header.ClearTabs()
		logger.Info("DEBUG: Cleared header tabs")
		activeIndex := -1
		for i, tab := range tabs {
			logger.Info(fmt.Sprintf("DEBUG: Adding tab %d to header: ID=%s, Title=%s, ResourceType=%s",
				i, tab.ID, tab.Title, tab.ResourceType))
			m.header.AddTab(tab.ID, tab.Title, tab.ResourceType)
			if tab.IsActive {
				activeIndex = i
				logger.Info(fmt.Sprintf("DEBUG: Tab %d is active", i))
			}
		}
		if activeIndex >= 0 {
			logger.Info(fmt.Sprintf("DEBUG: Setting active tab to index %d", activeIndex))
			m.header.SetActiveTab(activeIndex)
		} else {
			logger.Info("DEBUG: No active tab found")
		}
	} else {
		logger.Info("DEBUG: tabManager is nil")
	}
	logger.Info("DEBUG: updateHeaderTabs completed")
}

func (m *AppModel) getBreadcrumbTrail() string {
	if len(m.breadcrumbTrail) > 0 {
		prefix := []string{"config", "test-namespace"}
		fullTrail := append(prefix, m.breadcrumbTrail...)
		return strings.Join(fullTrail, " > ")
	}

	if m.tabManager != nil {
		if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
			return strings.Join(activeTab.Breadcrumb, " > ")
		}
	}
	return ""
}

func (m *AppModel) isCurrentScreenResourceType(resourceType string) bool {
	if len(m.breadcrumbTrail) > 0 {
		currentCrumb := m.breadcrumbTrail[len(m.breadcrumbTrail)-1]
		if currentCrumb == "Resource List" && resourceType == "ResourceList" {
			return true
		}
		return currentCrumb == resourceType
	}

	if m.tabManager != nil {
		if activeTab := m.tabManager.GetActiveTab(); activeTab != nil && len(activeTab.Breadcrumb) > 0 {
			currentCrumb := activeTab.Breadcrumb[len(activeTab.Breadcrumb)-1]
			return currentCrumb == resourceType
		}
	}
	return false
}

func (m *AppModel) initializeInitialBreadcrumb(listModel interface{}) {
	if list, ok := listModel.(*components.ListModel); ok {
		if list.List.Title == "Resource Types" {
			m.breadcrumbTrail = []string{"Resource List"}
		} else {
			m.breadcrumbTrail = []string{}
		}
	} else {
		m.breadcrumbTrail = []string{}
	}
}

func (m *AppModel) loadPluginUIExtensions() {
	if m.pluginManager == nil {
		return
	}

	registry := m.pluginManager.GetRegistry()
	if registry == nil {
		return
	}

	for _, plugin := range registry.GetUIPlugins() {
		extensions := plugin.GetUIExtensions()
		for _, ext := range extensions {
			for _, injection := range ext.InjectionPoints {
				m.uiInjector.AddInjection(injection.Location, injection)
			}
		}
	}

	api := m.pluginManager.GetAPI()
	if api != nil {
		headerComponents := api.GetHeaderComponents()
		for _, component := range headerComponents {
			if content, ok := component.Component.Config["content"].(string); ok {
				m.header.AddPluginComponent(content)
			}
		}

		footerComponents := api.GetFooterComponents()
		for _, component := range footerComponents {
			if _, ok := component.Component.Config["content"].(string); ok {
				m.uiInjector.AddInjection("footer", component)
			}
		}
	}
}
