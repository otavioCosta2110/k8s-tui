package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"strconv"
)

type InputDialogRequest struct {
	Title         string
	Placeholder   string
	SubmitCommand string
	CancelCommand string
}

type ClearTextInputMsg struct{}

type AppModel struct {
	tabManager          *models.TabManager
	header              models.HeaderModel
	kube                resources.Client
	config              config.AppConfig
	configSelected      bool
	errorPopup          *models.ErrorModel
	quickNav            tea.Model
	textInput           tea.Model
	helpScreen          *components.HelpModel
	pendingInputDialog  *InputDialogRequest
	currentResourceType string
	breadcrumbTrail     []string
	pluginManager       *plugins.PluginManager
	uiInjector          *UIInjector
}

type MultiClusterModel struct {
	clusters            []*AppModel
	currentCluster      int
	clusterTabComponent *components.TabComponent
	kubeconfigSelector  tea.Model
	namespaceSelector   tea.Model
	pendingKubeconfig   string
}

func NewAppModel(cfg cli.Config, pluginManager *plugins.PluginManager) *AppModel {
	appConfig := initializeAppConfigAndColors()

	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// For multi-cluster, we'll handle in MultiClusterModel
	// For now, use first kubeconfig or default
	kubeconfig := ""
	if len(cfg.KubeconfigPaths) > 0 {
		kubeconfig = cfg.KubeconfigPaths[0]
	}
	kubeClient, err := resources.NewClient(kubeconfig, cfg.Namespace)
	if err == nil && kubeClient != nil {
		return createAppModelWithKubeClient(cfg, appConfig, pluginManager, kubeClient)
	}

	_, kubeconfigErr := models.NewKubeconfigModel().InitComponent(nil)
	if kubeconfigErr != nil {
		return createAppModelWithoutKubeClient(appConfig, pluginManager, kubeconfigErr)
	}

	return createFallbackAppModel(appConfig, pluginManager)
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

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case components.NavigateMsg:
		return m.handleNavigateMsg(msg)
	case components.TabMsg:
		return m.handleTabMsg(msg)
	case models.HeaderRefreshMsg:
		return m.handleHeaderRefreshMsg(msg)
	case models.CloseQuickNavMsg:
		return m.handleCloseQuickNavMsg(msg)
	case components.CloseHelpMsg:
		return m, nil
	case components.TextInputSubmitMsg:
		return m, nil
	case components.TextInputCancelMsg:
		return m, nil
	case components.CreateSubmitMsg:
		return m.handleCreateSubmitMsg(msg)
	case components.OpenCreateFormMsg:
		return m.handleOpenCreateFormMsg(msg)
	case ClearTextInputMsg:
		m.textInput = nil
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
	if m.textInput != nil {
		return m.textInput.View()
	}

	if m.quickNav != nil {
		return m.quickNav.View()
	}

	if m.errorPopup != nil {
		return m.errorPopup.View()
	}

	if m.helpScreen != nil && m.helpScreen.IsVisible() {
		return m.helpScreen.View()
	}

	if m.tabManager == nil {
		return "Loading..."
	}

	content := m.renderContent()
	breadcrumbView := m.renderBreadcrumb()
	footerView := m.renderFooter()

	if !m.configSelected {
		return m.renderWithoutHeader(m.tabManager.View(), breadcrumbView)
	}

	if !styles.IsHeaderActive {
		if breadcrumbView != "" {
			return lipgloss.JoinVertical(lipgloss.Top, content, breadcrumbView)
		}
		return content
	}

	return m.renderWithHeader(content, breadcrumbView, footerView)
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

func NewMultiClusterModel(cfg cli.Config) *MultiClusterModel {
	// If no kubeconfigs provided, use default
	if len(cfg.KubeconfigPaths) == 0 {
		cfg.KubeconfigPaths = []string{""}
	}

	clusters := make([]*AppModel, len(cfg.KubeconfigPaths))
	clusterTabComponent := components.NewTabComponent()
	for i, kubeconfig := range cfg.KubeconfigPaths {
		// Create a copy of cfg with single kubeconfig
		singleCfg := cfg
		singleCfg.KubeconfigPaths = []string{kubeconfig}
		// Create separate plugin manager for each cluster
		pluginManager := plugins.NewPluginManager(cfg.PluginDir)
		if err := pluginManager.LoadPlugins(); err != nil {
			// Log error, but continue
			logger.Warn(fmt.Sprintf("Failed to load plugins for cluster %d: %v", i+1, err))
		}
		// Set as global for this cluster's models
		plugins.SetGlobalPluginManager(pluginManager)
		clusters[i] = NewAppModel(singleCfg, pluginManager)
		// Add cluster tab
		title := fmt.Sprintf("Cluster %d", i+1)
		clusterTabComponent.AddTab(fmt.Sprintf("%d", i), title, "cluster")
	}

	// Set active tab
	if len(cfg.KubeconfigPaths) > 0 {
		clusterTabComponent.SetActiveTab(0)
	}

	return &MultiClusterModel{
		clusters:            clusters,
		currentCluster:      0,
		clusterTabComponent: clusterTabComponent,
	}
}

func (m *MultiClusterModel) Init() tea.Cmd {
	if len(m.clusters) > 0 && m.clusters[m.currentCluster] != nil {
		plugins.SetGlobalPluginManager(m.clusters[m.currentCluster].pluginManager)
		return m.clusters[m.currentCluster].Init()
	}
	return nil
}

func (m *MultiClusterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Set width for cluster tab component
		if m.clusterTabComponent != nil {
			m.clusterTabComponent.Width = msg.Width
		}
	case tea.KeyMsg:
		// Handle cluster switching, e.g., f1, f2, f3
		if msg.String() == "f1" && len(m.clusters) > 0 {
			m.currentCluster = 0
			m.clusterTabComponent.SetActiveTab(0)
			plugins.SetGlobalPluginManager(m.clusters[0].pluginManager)
			return m, nil
		}
		if msg.String() == "f2" && len(m.clusters) > 1 {
			m.currentCluster = 1
			m.clusterTabComponent.SetActiveTab(1)
			plugins.SetGlobalPluginManager(m.clusters[1].pluginManager)
			return m, nil
		}
		if msg.String() == "f3" && len(m.clusters) > 2 {
			m.currentCluster = 2
			m.clusterTabComponent.SetActiveTab(2)
			plugins.SetGlobalPluginManager(m.clusters[2].pluginManager)
			return m, nil
		}
		if msg.String() == "ctrl+n" {
			m.kubeconfigSelector = models.NewKubeconfigSelectorModel()
			return m, m.kubeconfigSelector.Init()
		}
	// Add more as needed
	case components.TabMsg:
		// Handle cluster tab switching
		if msg.ResourceType == "cluster" && msg.Action == "switch" {
			if index, err := strconv.Atoi(msg.TabID); err == nil && index >= 0 && index < len(m.clusters) {
				m.currentCluster = index
				m.clusterTabComponent.SetActiveTab(index)
				plugins.SetGlobalPluginManager(m.clusters[index].pluginManager)
				return m, nil
			}
		}
	case models.KubeconfigSelectedMsg:
		// Store pending kubeconfig and open namespace selector
		m.pendingKubeconfig = msg.Path
		m.kubeconfigSelector = nil
		m.namespaceSelector = models.NewNamespaceSelectorModel(msg.Path)
		return m, m.namespaceSelector.Init()
	case models.NamespaceSelectedMsg:
		// Add new cluster with selected namespace
		singleCfg := cli.Config{
			KubeconfigPaths: []string{m.pendingKubeconfig},
			Namespace:       msg.Namespace,
			PluginDir:       "./plugins",
		}
		// Create separate plugin manager for the new cluster
		pluginManager := plugins.NewPluginManager(singleCfg.PluginDir)
		if err := pluginManager.LoadPlugins(); err != nil {
			// Log error, but continue
			logger.Warn(fmt.Sprintf("Failed to load plugins for new cluster: %v", err))
		}
		// Set as global for this cluster's models
		plugins.SetGlobalPluginManager(pluginManager)
		newCluster := NewAppModel(singleCfg, pluginManager)
		m.clusters = append(m.clusters, newCluster)
		newIndex := len(m.clusters) - 1
		title := fmt.Sprintf("Cluster %d", newIndex+1)
		m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), title, "cluster")
		m.currentCluster = newIndex
		m.clusterTabComponent.SetActiveTab(newIndex)
		m.namespaceSelector = nil
		m.pendingKubeconfig = ""
		return m, nil
	}

	// If selector is open, delegate to it
	if m.kubeconfigSelector != nil {
		updated, cmd := m.kubeconfigSelector.Update(msg)
		if updated == nil {
			m.kubeconfigSelector = nil // Closed
		} else {
			m.kubeconfigSelector = updated
		}
		return m, cmd
	}
	if m.namespaceSelector != nil {
		updated, cmd := m.namespaceSelector.Update(msg)
		if updated == nil {
			m.namespaceSelector = nil // Closed
			m.pendingKubeconfig = ""
		} else {
			m.namespaceSelector = updated
		}
		return m, cmd
	}

	// Delegate to current cluster
	if len(m.clusters) > 0 && m.clusters[m.currentCluster] != nil {
		updated, cmd := m.clusters[m.currentCluster].Update(msg)
		if appModel, ok := updated.(*AppModel); ok {
			m.clusters[m.currentCluster] = appModel
		}
		return m, cmd
	}
	return m, nil
}

func (m *MultiClusterModel) View() string {
	if m.kubeconfigSelector != nil {
		return m.kubeconfigSelector.View()
	}
	if m.namespaceSelector != nil {
		return m.namespaceSelector.View()
	}

	if len(m.clusters) == 0 || m.clusters[m.currentCluster] == nil {
		return "No clusters configured"
	}

	// Delegate view to current cluster
	content := m.clusters[m.currentCluster].View()

	// Render cluster tabs only if more than one cluster
	if len(m.clusters) > 1 && m.clusterTabComponent != nil {
		clusterTabView := m.clusterTabComponent.View()
		if clusterTabView != "" {
			return lipgloss.JoinVertical(lipgloss.Top, clusterTabView, content)
		}
	}

	return content
}
