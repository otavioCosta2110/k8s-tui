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
	"strings"
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

// isInSearchMode checks if the current active component is in search mode
func (m *AppModel) isInSearchMode() bool {
	if m.tabManager == nil {
		return false
	}

	// Get active tab
	activeTab := m.tabManager.GetActiveTab()
	if activeTab == nil {
		return false
	}

	// Check if active model is a table component in search mode
	if tableModel, ok := activeTab.Model.(*components.TableModel); ok {
		// We need to access the unexported searchMode field
		// Since we can't access it directly, we'll use a different approach
		// We'll check if the view contains search indicators
		view := tableModel.View()
		return strings.Contains(view, "Search:") && strings.Contains(view, "█")
	}

	// Check if it's a resource model that contains a table
	if resourceModel, ok := activeTab.Model.(interface{ GetTable() *components.TableModel }); ok {
		if table := resourceModel.GetTable(); table != nil {
			view := table.View()
			return strings.Contains(view, "Search:") && strings.Contains(view, "█")
		}
	}

	return false
}

type MultiClusterModel struct {
	clusters            []*AppModel
	currentCluster      int
	clusterTabComponent *components.TabComponent
	kubeconfigSelector  tea.Model
	namespaceSelector   tea.Model
	pendingKubeconfig   string
	width               int
	height              int
	pluginDir           string
	config              config.AppConfig
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

	clusters := make([]*AppModel, 0, len(cfg.KubeconfigPaths))
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

		appModel := NewAppModel(singleCfg, pluginManager)

		// Check if the app model has an error popup (indicating initialization failure)
		if appModel.errorPopup != nil {
			// Skip this cluster but continue with others
			logger.Warn(fmt.Sprintf("Failed to initialize cluster %d: %v", i+1, kubeconfig))
			continue
		}

		clusters = append(clusters, appModel)
		// Add cluster tab with server address
		clusterName := "Unknown"
		if appModel.kube.Clientset != nil {
			clusterName = appModel.kube.GetClusterName()
		} else {
			clusterName = fmt.Sprintf("Cluster %d", i+1)
		}
		clusterTabComponent.AddTab(fmt.Sprintf("%d", len(clusters)-1), clusterName, "cluster")
	}

	// Set active tab
	if len(cfg.KubeconfigPaths) > 0 {
		clusterTabComponent.SetActiveTab(0)
	}

	return &MultiClusterModel{
		clusters:            clusters,
		currentCluster:      0,
		clusterTabComponent: clusterTabComponent,
		pluginDir:           cfg.PluginDir,
		config:              initializeAppConfigAndColors(),
	}
}

func (m *MultiClusterModel) Init() tea.Cmd {
	if m.kubeconfigSelector != nil {
		return m.kubeconfigSelector.Init()
	}
	if len(m.clusters) > 0 && m.clusters[m.currentCluster] != nil {
		plugins.SetGlobalPluginManager(m.clusters[m.currentCluster].pluginManager)
		return m.clusters[m.currentCluster].Init()
	}
	return nil
}

func (m *MultiClusterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle kubeconfig selector when no clusters exist
	if m.kubeconfigSelector != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				return nil, tea.Quit
			}
		case models.KubeconfigSelectedMsg:
			// Store pending kubeconfig and open namespace selector
			m.pendingKubeconfig = msg.Path
			m.kubeconfigSelector = nil
			m.namespaceSelector = models.NewNamespaceSelectorModel(msg.Path)
			return m, m.namespaceSelector.Init()
		case models.NamespaceSelectedMsg:
			// Create first cluster with selected namespace
			singleCfg := cli.Config{
				KubeconfigPaths: []string{m.pendingKubeconfig},
				Namespace:       msg.Namespace,
				PluginDir:       m.pluginDir,
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

			// Get cluster name from client
			clusterName := "Unknown"
			if newCluster.kube.Clientset != nil {
				clusterName = newCluster.kube.GetClusterName()
			}

			m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), clusterName, "cluster")
			m.currentCluster = newIndex
			m.clusterTabComponent.SetActiveTab(newIndex)
			m.pendingKubeconfig = ""
			m.namespaceSelector = nil

			// Initialize the new cluster and update with window size if available
			initCmd := newCluster.Init()
			if m.width > 0 {
				updated, cmd := newCluster.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				if appModel, ok := updated.(*AppModel); ok {
					m.clusters[newIndex] = appModel
				}
				return m, tea.Batch(initCmd, cmd)
			}

			return m, initCmd
		}

		// Forward other messages to kubeconfig selector
		updated, cmd := m.kubeconfigSelector.Update(msg)
		if selector, ok := updated.(*models.KubeconfigSelectorModel); ok {
			m.kubeconfigSelector = selector
		}
		return m, cmd
	}

	// Handle namespace selector when it exists
	if m.namespaceSelector != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				// Go back to kubeconfig selector
				m.namespaceSelector = nil
				m.kubeconfigSelector = models.NewKubeconfigSelectorModel()
				return m, m.kubeconfigSelector.Init()
			}
		case models.NamespaceSelectedMsg:
			// Create first cluster with selected namespace
			singleCfg := cli.Config{
				KubeconfigPaths: []string{m.pendingKubeconfig},
				Namespace:       msg.Namespace,
				PluginDir:       m.pluginDir,
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

			// Get cluster name from client
			clusterName := "Unknown"
			if newCluster.kube.Clientset != nil {
				clusterName = newCluster.kube.GetClusterName()
			}

			m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), clusterName, "cluster")
			m.currentCluster = newIndex
			m.clusterTabComponent.SetActiveTab(newIndex)
			m.pendingKubeconfig = ""
			m.namespaceSelector = nil

			// Initialize the new cluster and update with window size if available
			initCmd := newCluster.Init()
			if m.width > 0 {
				updated, cmd := newCluster.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				if appModel, ok := updated.(*AppModel); ok {
					m.clusters[newIndex] = appModel
				}
				return m, tea.Batch(initCmd, cmd)
			}

			return m, initCmd
		}

		// Forward other messages to namespace selector
		updated, cmd := m.namespaceSelector.Update(msg)
		if selector, ok := updated.(*models.NamespaceSelectorModel); ok {
			m.namespaceSelector = selector
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Set width for cluster tab component
		if m.clusterTabComponent != nil {
			m.clusterTabComponent.Width = msg.Width
		}
		m.width = msg.Width
		m.height = msg.Height

		// Forward WindowSizeMsg to selectors if they exist
		if m.kubeconfigSelector != nil {
			updated, cmd := m.kubeconfigSelector.Update(msg)
			if selector, ok := updated.(*models.KubeconfigSelectorModel); ok {
				m.kubeconfigSelector = selector
			}
			return m, cmd
		}
		if m.namespaceSelector != nil {
			updated, cmd := m.namespaceSelector.Update(msg)
			if selector, ok := updated.(*models.NamespaceSelectorModel); ok {
				m.namespaceSelector = selector
			}
			return m, cmd
		}
	case tea.KeyMsg:
		// Handle cluster switching with Ctrl+Left/Ctrl+Right
		if msg.String() == "ctrl+left" {
			if len(m.clusters) > 0 {
				newCluster := (m.currentCluster - 1 + len(m.clusters)) % len(m.clusters)
				m.currentCluster = newCluster
				m.clusterTabComponent.SetActiveTab(newCluster)
				plugins.SetGlobalPluginManager(m.clusters[newCluster].pluginManager)
				if m.width > 0 {
					updated, _ := m.clusters[newCluster].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					if appModel, ok := updated.(*AppModel); ok {
						m.clusters[newCluster] = appModel
					}
				}
			}
			return m, nil
		}
		if msg.String() == "ctrl+right" {
			if len(m.clusters) > 0 {
				newCluster := (m.currentCluster + 1) % len(m.clusters)
				m.currentCluster = newCluster
				m.clusterTabComponent.SetActiveTab(newCluster)
				plugins.SetGlobalPluginManager(m.clusters[newCluster].pluginManager)
				if m.width > 0 {
					updated, _ := m.clusters[newCluster].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					if appModel, ok := updated.(*AppModel); ok {
						m.clusters[newCluster] = appModel
					}
				}
			}
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
				if m.width > 0 {
					updated, _ := m.clusters[index].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					if appModel, ok := updated.(*AppModel); ok {
						m.clusters[index] = appModel
					}
				}
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
			PluginDir:       m.pluginDir,
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

		// Get cluster name from client
		clusterName := "Unknown"
		if newCluster.kube.Clientset != nil {
			clusterName = newCluster.kube.GetClusterName()
		}

		m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), clusterName, "cluster")
		m.currentCluster = newIndex
		m.clusterTabComponent.SetActiveTab(newIndex)
		plugins.SetGlobalPluginManager(m.clusters[newIndex].pluginManager)

		// Initialize the new cluster properly
		initCmd := m.clusters[newIndex].Init()

		// Update with window size if available
		if m.width > 0 {
			updated, cmd := m.clusters[newIndex].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			if appModel, ok := updated.(*AppModel); ok {
				m.clusters[newIndex] = appModel
			}
			// Return both init and window update commands
			return m, tea.Batch(initCmd, cmd)
		}

		m.namespaceSelector = nil
		m.pendingKubeconfig = ""
		return m, initCmd
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

	// Render cluster tabs
	if m.clusterTabComponent != nil {
		clusterTabView := m.clusterTabComponent.View()
		if clusterTabView != "" {
			return lipgloss.JoinVertical(lipgloss.Top, clusterTabView, content)
		}
	}

	return content
}
