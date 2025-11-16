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
	pluginManager       *plugins.GlobalPluginManager
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
	pluginDir           string
	pluginManager       *plugins.GlobalPluginManager
	config              config.AppConfig
	width               int
	height              int
	kubeconfigSelector  *models.KubeconfigSelectorModel
	namespaceSelector   *models.NamespaceSelectorModel
	pendingKubeconfig   string
	pluginArgs          map[string]string
	sizeCheck           *models.SizeCheckModel
	errorPopup          *models.ErrorModel
}

func NewAppModel(cfg cli.Config, pluginManager *plugins.GlobalPluginManager) *AppModel {
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
		return createAppModelWithKubeClient(cfg, appConfig, pluginManager, kubeClient, true)
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
	case models.CustomResourceErrorMsg:
		// Show error popup for custom resource selection
		popup := models.NewErrorScreen(
			msg.Error,
			"Not Implemented",
			"Example resource selection is not implemented",
		)
		popup.SetDimensions(styles.ScreenWidth, styles.ScreenHeight+styles.HeaderSize)
		m.errorPopup = &popup
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

// addClusterTabFromPlugin adds a new cluster tab from a plugin request
func (m *MultiClusterModel) addClusterTabFromPlugin(kubeconfigPath, clusterName, namespace string) error {
	logger.Info(fmt.Sprintf("Adding cluster tab from plugin: %s with kubeconfig %s and namespace %s", clusterName, kubeconfigPath, namespace))

	// Create config for new cluster
	singleCfg := cli.Config{
		KubeconfigPaths: []string{kubeconfigPath},
		Namespace:       namespace,
		PluginDir:       m.pluginDir,
	}

	logger.Info(fmt.Sprintf("DEBUG: Created singleCfg with namespace: '%s'", singleCfg.Namespace))

	// Use the shared plugin manager
	newCluster := NewAppModel(singleCfg, m.pluginManager)

	// Check if cluster initialization failed
	if newCluster.errorPopup != nil {
		return fmt.Errorf("failed to initialize cluster %s", clusterName)
	}

	m.clusters = append(m.clusters, newCluster)
	newIndex := len(m.clusters) - 1

	// Get actual cluster name from client if available
	actualClusterName := clusterName
	if newCluster.kube.Clientset != nil {
		actualClusterName = newCluster.kube.GetClusterName()
	}

	// Add cluster tab
	m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), actualClusterName, "cluster")

	// Register cluster with the shared plugin manager
	clusterID := fmt.Sprintf("%d", newIndex)
	m.pluginManager.AddClusterWithNamespace(clusterID, actualClusterName, newCluster.kube, newCluster.kube.Namespace)

	logger.Info(fmt.Sprintf("Successfully added cluster tab %s (%s) from plugin", actualClusterName, clusterID))

	return nil
}

// setClusterTabsFromPlugin replaces all cluster tabs with the specified clusters from a plugin request
func (m *MultiClusterModel) setClusterTabsFromPlugin(clusters []plugins.ClusterTabConfig) error {
	logger.Info(fmt.Sprintf("Setting %d cluster tabs from plugin request", len(clusters)))

	// Clear all existing clusters
	m.clusters = make([]*AppModel, 0)

	// Clear cluster tabs
	m.clusterTabComponent.ClearTabs()

	// Clear plugin manager clusters
	if m.pluginManager != nil {
		// Get all existing cluster IDs to remove them
		existingClusters := m.pluginManager.GetAllClusters()
		for clusterID := range existingClusters {
			m.pluginManager.RemoveCluster(clusterID)
		}
	}

	// Add new clusters
	for i, clusterConfig := range clusters {
		logger.Info(fmt.Sprintf("Adding cluster %d: %s with kubeconfig %s and namespace %s", i, clusterConfig.ClusterName, clusterConfig.KubeconfigPath, clusterConfig.Namespace))

		// Create config for new cluster
		singleCfg := cli.Config{
			KubeconfigPaths: []string{clusterConfig.KubeconfigPath},
			Namespace:       clusterConfig.Namespace,
			PluginDir:       m.pluginDir,
		}

		// Use shared plugin manager
		newCluster := NewAppModel(singleCfg, m.pluginManager)

		// Check if cluster initialization failed
		if newCluster.errorPopup != nil {
			logger.Error(fmt.Sprintf("Failed to initialize cluster %s", clusterConfig.ClusterName))
			continue // Skip this cluster but continue with others
		}

		m.clusters = append(m.clusters, newCluster)
		newIndex := len(m.clusters) - 1

		// Get actual cluster name from client if available
		actualClusterName := clusterConfig.ClusterName
		if newCluster.kube.Clientset != nil {
			actualClusterName = newCluster.kube.GetClusterName()
		}

		// Add cluster tab
		m.clusterTabComponent.AddTab(fmt.Sprintf("%d", newIndex), actualClusterName, "cluster")

		// Register cluster with shared plugin manager
		clusterID := fmt.Sprintf("%d", newIndex)
		m.pluginManager.AddClusterWithNamespace(clusterID, actualClusterName, newCluster.kube, newCluster.kube.Namespace)

		logger.Info(fmt.Sprintf("Successfully added cluster tab %s (%s) from plugin", actualClusterName, clusterID))
	}

	// Set current cluster to first one if any clusters exist
	if len(m.clusters) > 0 {
		m.currentCluster = 0
		// Switch to first cluster in plugin manager
		firstClusterID := fmt.Sprintf("%d", 0)
		if err := m.pluginManager.SwitchToCluster(firstClusterID); err != nil {
			logger.Error(fmt.Sprintf("Failed to switch to first cluster: %v", err))
		}
	} else {
		m.currentCluster = -1
	}

	logger.Info(fmt.Sprintf("Successfully set %d cluster tabs from plugin", len(m.clusters)))
	return nil
}

// getClustersForPlugin returns cluster information for plugins
func (m *MultiClusterModel) getClustersForPlugin() []plugins.ClusterInfo {
	logger.Info(fmt.Sprintf("DEBUG: getClustersForPlugin called with %d clusters", len(m.clusters)))
	clusters := make([]plugins.ClusterInfo, 0, len(m.clusters))

	for i, cluster := range m.clusters {
		clusterName := "Unknown"
		if cluster.kube.Clientset != nil {
			clusterName = cluster.kube.GetClusterName()
		}

		clusterInfo := plugins.ClusterInfo{
			ID:         fmt.Sprintf("%d", i),
			Name:       clusterName,
			Namespace:  cluster.kube.Namespace,
			Kubeconfig: cluster.kube.KubeconfigPath,
		}
		clusters = append(clusters, clusterInfo)
		logger.Info(fmt.Sprintf("DEBUG: Cluster %d: %s (ID: %s, NS: %s)", i, clusterName, clusterInfo.ID, clusterInfo.Namespace))
	}

	logger.Info(fmt.Sprintf("DEBUG: getClustersForPlugin returning %d clusters", len(clusters)))
	return clusters
}

// getTabsForCluster returns tabs for a specific cluster
func (m *MultiClusterModel) getTabsForCluster(clusterID string) ([]plugins.TabInfo, error) {
	// Convert clusterID to index
	clusterIndex := -1
	if id, err := strconv.Atoi(clusterID); err == nil {
		clusterIndex = id
	} else {
		return nil, fmt.Errorf("invalid cluster ID: %s", clusterID)
	}

	// Validate cluster index
	if clusterIndex < 0 || clusterIndex >= len(m.clusters) {
		return nil, fmt.Errorf("cluster index %d out of range (0-%d)", clusterIndex, len(m.clusters)-1)
	}

	// Get tabs from the specific cluster's tab manager
	cluster := m.clusters[clusterIndex]
	if cluster.tabManager == nil {
		return nil, fmt.Errorf("cluster %s has no tab manager", clusterID)
	}

	tabs := cluster.tabManager.GetTabsInfo()
	logger.Info(fmt.Sprintf("DEBUG: getTabsForCluster for cluster %s returned %d tabs", clusterID, len(tabs)))
	for i, tab := range tabs {
		logger.Info(fmt.Sprintf("DEBUG: Tab %d: %s (%s)", i, tab.Title, tab.ResourceType))
	}

	return tabs, nil
}

// setTabsForClusterFromPlugin sets tabs for a specific cluster from plugin request
func (m *MultiClusterModel) setTabsForClusterFromPlugin(clusterID string, tabs []plugins.TabInfo) error {
	logger.Info(fmt.Sprintf("Setting tabs for cluster %s from plugin request: %d tabs", clusterID, len(tabs)))

	// Convert clusterID to index
	clusterIndex := -1
	if id, err := strconv.Atoi(clusterID); err == nil {
		clusterIndex = id
	} else {
		return fmt.Errorf("invalid cluster ID: %s", clusterID)
	}

	// Validate cluster index
	if clusterIndex < 0 || clusterIndex >= len(m.clusters) {
		return fmt.Errorf("cluster index %d out of range (0-%d)", clusterIndex, len(m.clusters)-1)
	}

	// Get cluster's tab manager
	cluster := m.clusters[clusterIndex]
	if cluster.tabManager == nil {
		return fmt.Errorf("cluster %s has no tab manager", clusterID)
	}

	// Set tabs in the cluster's tab manager
	if err := cluster.tabManager.RestoreTabs(tabs); err != nil {
		return fmt.Errorf("failed to set tabs for cluster %s: %v", clusterID, err)
	}

	logger.Info(fmt.Sprintf("Successfully set %d tabs for cluster %s", len(tabs), clusterID))
	return nil
}

// restoreTabsForCluster restores tabs to a specific cluster from plugin request
func (m *MultiClusterModel) restoreTabsForCluster(clusterID string) error {
	logger.Info(fmt.Sprintf("Restoring tabs for cluster %s from plugin request", clusterID))

	// Convert clusterID to index
	clusterIndex := -1
	if id, err := strconv.Atoi(clusterID); err == nil {
		clusterIndex = id
	} else {
		return fmt.Errorf("invalid cluster ID: %s", clusterID)
	}

	// Validate cluster index
	if clusterIndex < 0 || clusterIndex >= len(m.clusters) {
		return fmt.Errorf("cluster index %d out of range (0-%d)", clusterIndex, len(m.clusters)-1)
	}

	// Get tabs from plugin manager for this cluster
	cluster := m.clusters[clusterIndex]
	if cluster.tabManager == nil {
		return fmt.Errorf("cluster %s has no tab manager", clusterID)
	}

	// Get tabs from plugin manager state
	tabs, err := m.pluginManager.GetTabsByCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed to get tabs for cluster %s: %v", clusterID, err)
	}

	// Restore tabs to the cluster's tab manager
	if err := cluster.tabManager.RestoreTabs(tabs); err != nil {
		return fmt.Errorf("failed to restore tabs for cluster %s: %v", clusterID, err)
	}

	// Update the UI to show the restored tabs
	cluster.updateHeaderTabs()

	logger.Info(fmt.Sprintf("Successfully restored %d tabs to cluster %s and updated UI", len(tabs), clusterID))
	return nil
}

// switchToClusterFromPlugin switches to a specific cluster from a plugin request
func (m *MultiClusterModel) switchToClusterFromPlugin(clusterID string) error {
	logger.Info(fmt.Sprintf("Switching to cluster %s from plugin request", clusterID))

	// Convert clusterID to index
	clusterIndex := -1
	if id, err := strconv.Atoi(clusterID); err == nil {
		clusterIndex = id
	} else {
		return fmt.Errorf("invalid cluster ID: %s", clusterID)
	}

	// Validate cluster index
	if clusterIndex < 0 || clusterIndex >= len(m.clusters) {
		return fmt.Errorf("cluster index %d out of range (0-%d)", clusterIndex, len(m.clusters)-1)
	}

	// Switch to the cluster
	m.currentCluster = clusterIndex
	m.clusterTabComponent.SetActiveTab(clusterIndex)

	// Switch to the new cluster in the plugin manager
	if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
		logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
		return err
	}

	logger.Info(fmt.Sprintf("Successfully switched to cluster %s", clusterID))

	return nil
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

			// Update header namespace from active tab
			if m.tabManager != nil {
				if activeTabData := m.tabManager.GetActiveTab(); activeTabData != nil {
					m.header.SetNamespace(activeTabData.Namespace)
					logger.Info(fmt.Sprintf("DEBUG: Updated header namespace to: %s", activeTabData.Namespace))
				}
			}
		} else {
			logger.Info("DEBUG: No active tab found")
		}
	} else {
		logger.Info("DEBUG: tabManager is nil")
	}
	logger.Info("DEBUG: updateHeaderTabs completed")
}

// getUniqueClusterName creates a unique cluster name by including namespace when needed
func getUniqueClusterName(baseName, namespace string, index int) string {
	// If we have multiple clusters (index > 0), make name unique by including namespace
	if index > 0 {
		return fmt.Sprintf("%s (%s)", baseName, namespace)
	}
	return baseName
}

func NewMultiClusterModel(cfg cli.Config) *MultiClusterModel {
	// If no kubeconfigs provided, use default
	if len(cfg.KubeconfigPaths) == 0 {
		cfg.KubeconfigPaths = []string{""}
	}

	// Create ONE shared plugin manager for all clusters
	sharedPluginManager := plugins.NewGlobalPluginManager(cfg.PluginDir)
	if err := sharedPluginManager.LoadPlugins(); err != nil {
		logger.Warn(fmt.Sprintf("Failed to load plugins: %v", err))
	}
	// Set as global
	plugins.SetGlobalPluginManager(sharedPluginManager)

	// Set up cluster management callbacks for plugins
	// These will be set after the MultiClusterModel is created
	// We'll store them temporarily and set them in the returned model

	clusters := make([]*AppModel, 0, len(cfg.KubeconfigPaths))
	clusterTabComponent := components.NewTabComponent()
	for i, kubeconfig := range cfg.KubeconfigPaths {
		// Create a copy of cfg with single kubeconfig
		singleCfg := cfg
		singleCfg.KubeconfigPaths = []string{kubeconfig}

		appModel := NewAppModel(singleCfg, sharedPluginManager)

		// Check if the app model has an error popup (indicating initialization failure)
		if appModel.errorPopup != nil {
			// Skip this cluster but continue with others
			logger.Warn(fmt.Sprintf("Failed to initialize cluster %d: %v", i+1, kubeconfig))
			continue
		}

		clusters = append(clusters, appModel)
		clusterIndex := len(clusters) - 1

		// Add cluster tab with server address
		clusterName := "Unknown"
		if appModel.kube.Clientset != nil {
			baseName := appModel.kube.GetClusterName()
			// Make cluster name unique by including namespace if multiple clusters use same kubeconfig
			uniqueName := getUniqueClusterName(baseName, appModel.kube.Namespace, i)
			clusterName = uniqueName
		} else {
			clusterName = fmt.Sprintf("Cluster %d", i+1)
		}
		clusterTabComponent.AddTab(fmt.Sprintf("%d", clusterIndex), clusterName, "cluster")

		// Register cluster with the shared plugin manager
		clusterID := fmt.Sprintf("%d", clusterIndex)
		sharedPluginManager.AddClusterWithNamespace(clusterID, clusterName, appModel.kube, appModel.kube.Namespace)
		logger.Info(fmt.Sprintf("Registered cluster %s (%s) with plugin manager", clusterName, clusterID))
	}

	// Set active tab and active cluster in plugin manager
	if len(clusters) > 0 {
		clusterTabComponent.SetActiveTab(0)
		if err := sharedPluginManager.SwitchToCluster("0"); err != nil {
			logger.Warn(fmt.Sprintf("Failed to switch to cluster 0: %v", err))
		}
	}

	model := &MultiClusterModel{
		clusters:            clusters,
		currentCluster:      0,
		clusterTabComponent: clusterTabComponent,
		pluginDir:           cfg.PluginDir,
		pluginManager:       sharedPluginManager,
		config:              initializeAppConfigAndColors(),
		pluginArgs:          cfg.PluginArgs,
		sizeCheck:           models.NewSizeCheckModel(),
	}

	// Set up cluster management callbacks for plugins
	logger.Info("DEBUG: Setting up cluster management callbacks for plugins")
	sharedPluginManager.GetAPI().SetAddClusterTabCallback(model.addClusterTabFromPlugin)
	sharedPluginManager.GetAPI().SetClusterTabsCallback(model.setClusterTabsFromPlugin)
	sharedPluginManager.GetAPI().SetGetClustersCallback(model.getClustersForPlugin)
	sharedPluginManager.GetAPI().SetSwitchToClusterCallback(model.switchToClusterFromPlugin)
	sharedPluginManager.GetAPI().SetGetTabsForClusterCallback(model.getTabsForCluster)
	sharedPluginManager.GetAPI().SetSetTabsForClusterCallback(model.setTabsForClusterFromPlugin)

	// Add callback for restoring tabs to individual clusters
	sharedPluginManager.GetAPI().SetRestoreTabsForClusterCallback(model.restoreTabsForCluster)
	logger.Info("DEBUG: Cluster management callbacks set")

	return model
}

func (m *MultiClusterModel) Init() tea.Cmd {
	// Process plugin CLI arguments here after callbacks are set
	logger.Info("DEBUG: MultiClusterModel.Init() called")
	logger.Info(fmt.Sprintf("DEBUG: Plugin args to process: %d", len(m.pluginArgs)))

	if len(m.pluginArgs) > 0 {
		logger.Info("DEBUG: About to process plugin args")
		if err := cli.HandlePluginArgs(m.pluginManager, m.pluginArgs); err != nil {
			logger.Error(fmt.Sprintf("Failed to handle plugin CLI arguments in MultiClusterModel.Init(): %v", err))
		}
		// Clear plugin args after processing
		m.pluginArgs = make(map[string]string)
		logger.Info("DEBUG: Plugin args processed and cleared")
	}

	var cmds []tea.Cmd

	// Initialize size checker
	if m.sizeCheck != nil {
		cmds = append(cmds, m.sizeCheck.Init())
	}

	if m.kubeconfigSelector != nil {
		cmds = append(cmds, m.kubeconfigSelector.Init())
	} else if len(m.clusters) > 0 && m.clusters[m.currentCluster] != nil {
		// The shared plugin manager is already set in NewMultiClusterModel
		// Just ensure it's set as global
		if m.pluginManager != nil {
			plugins.SetGlobalPluginManager(m.pluginManager)
		}
		cmds = append(cmds, m.clusters[m.currentCluster].Init())
	}

	return tea.Batch(cmds...)
}

func (m *MultiClusterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle size checking first
	if m.sizeCheck != nil {
		updated, cmd := m.sizeCheck.Update(msg)
		if sizeCheck, ok := updated.(*models.SizeCheckModel); ok {
			m.sizeCheck = sizeCheck
		}
		// If size is invalid, don't process other messages
		if !m.sizeCheck.IsSizeValid() {
			return m, cmd
		}
	}

	// Handle error popup
	if m.errorPopup != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" {
				m.errorPopup = nil
				return m, nil
			}
		}
		updatedModel, cmd := m.errorPopup.Update(msg)
		m.errorPopup = &updatedModel
		return m, cmd
	}

	// Handle kubeconfig selector when no clusters exist
	if m.kubeconfigSelector != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				return nil, tea.Quit
			}
			if msg.String() == "q" {
				// If we have existing clusters, cancel selector and go back to main view
				if len(m.clusters) > 0 {
					m.kubeconfigSelector = nil
					return m, nil
				}
				// If no clusters exist, quit the app
				return nil, tea.Quit
			}
		case models.KubeconfigErrorMsg:
			// Show error popup for invalid kubeconfig
			popup := models.NewErrorScreen(
				msg.Error,
				"Invalid Kubeconfig",
				"The selected kubeconfig file is not valid",
			)
			popup.SetDimensions(m.width, m.height+styles.HeaderSize)
			m.errorPopup = &popup
			return m, nil
		case models.CustomResourceErrorMsg:
			// Show error popup for custom resource selection
			popup := models.NewErrorScreen(
				msg.Error,
				"Not Implemented",
				"Example resource selection is not implemented",
			)
			popup.SetDimensions(m.width, m.height+styles.HeaderSize)
			m.errorPopup = &popup
			return m, nil
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
			// Use the shared plugin manager
			newCluster := NewAppModel(singleCfg, m.pluginManager)
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

			// Register cluster with the shared plugin manager
			clusterID := fmt.Sprintf("%d", newIndex)
			m.pluginManager.AddClusterWithNamespace(clusterID, clusterName, newCluster.kube, newCluster.kube.Namespace)
			if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
				logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
			}
			logger.Info(fmt.Sprintf("Registered and switched to cluster %s (%s)", clusterName, clusterID))

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
			if msg.String() == "esc" || msg.String() == "q" {
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
			// Use the shared plugin manager
			newCluster := NewAppModel(singleCfg, m.pluginManager)
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

			// Register cluster with the shared plugin manager
			clusterID := fmt.Sprintf("%d", newIndex)
			m.pluginManager.AddClusterWithNamespace(clusterID, clusterName, newCluster.kube, newCluster.kube.Namespace)
			if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
				logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
			}
			logger.Info(fmt.Sprintf("Registered and switched to cluster %s (%s)", clusterName, clusterID))

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

		// Forward WindowSizeMsg to size checker
		if m.sizeCheck != nil {
			updated, cmd := m.sizeCheck.Update(msg)
			if sizeCheck, ok := updated.(*models.SizeCheckModel); ok {
				m.sizeCheck = sizeCheck
			}
			// If size is invalid, don't process further
			if !m.sizeCheck.IsSizeValid() {
				return m, cmd
			}
		}

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

				// Switch to the new cluster in the plugin manager
				clusterID := fmt.Sprintf("%d", newCluster)
				if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
					logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
				}

				// Update header namespace to match current cluster's namespace
				currentClusterModel := m.clusters[newCluster]
				if currentClusterModel.kube.Clientset != nil {
					currentClusterModel.header.SetNamespace(currentClusterModel.kube.Namespace)
					currentClusterModel.header.UpdateContent()
				}

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

				// Switch to the new cluster in the plugin manager
				clusterID := fmt.Sprintf("%d", newCluster)
				if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
					logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
				}

				// Update header namespace to match current cluster's namespace
				currentClusterModel := m.clusters[newCluster]
				if currentClusterModel.kube.Clientset != nil {
					currentClusterModel.header.SetNamespace(currentClusterModel.kube.Namespace)
					currentClusterModel.header.UpdateContent()
				}

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

				// Switch to the new cluster in the plugin manager
				clusterID := fmt.Sprintf("%d", index)
				if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
					logger.Warn(fmt.Sprintf("Failed to switch to cluster %s: %v", clusterID, err))
				}

				// Update plugin manager's current namespace to match cluster's namespace
				if m.clusters[index].kube.Clientset != nil {
					m.pluginManager.GetAPI().SetCurrentNamespace(m.clusters[index].kube.Namespace)
					m.clusters[index].header.SetNamespace(m.clusters[index].kube.Namespace)
					m.clusters[index].header.UpdateContent()
				}

				if m.width > 0 {
					updated, _ := m.clusters[index].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					if appModel, ok := updated.(*AppModel); ok {
						m.clusters[index] = appModel
					}
				}
				return m, nil
			}
		}
	case models.KubeconfigErrorMsg:
		// Show error popup for invalid kubeconfig
		popup := models.NewErrorScreen(
			msg.Error,
			"Invalid Kubeconfig",
			"The selected kubeconfig file is not valid",
		)
		popup.SetDimensions(m.width, m.height+styles.HeaderSize)
		m.errorPopup = &popup
		return m, nil
	case models.CustomResourceErrorMsg:
		// Show error popup for custom resource selection
		popup := models.NewErrorScreen(
			msg.Error,
			"Not Implemented",
			"Example resource selection is not implemented",
		)
		popup.SetDimensions(m.width, m.height+styles.HeaderSize)
		m.errorPopup = &popup
		return m, nil
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
		pluginManager := plugins.NewGlobalPluginManager(singleCfg.PluginDir)
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
		// Switch to the cluster in the shared plugin manager
		clusterID := fmt.Sprintf("%d", newIndex)
		if err := m.pluginManager.SwitchToCluster(clusterID); err != nil {
			logger.Warn(fmt.Sprintf("Failed to switch to cluster %s in plugin manager: %v", clusterID, err))
		}

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
		} else if selector, ok := updated.(*models.KubeconfigSelectorModel); ok {
			m.kubeconfigSelector = selector
		}
		return m, cmd
	}
	if m.namespaceSelector != nil {
		updated, cmd := m.namespaceSelector.Update(msg)
		if updated == nil {
			m.namespaceSelector = nil // Closed
		} else if selector, ok := updated.(*models.NamespaceSelectorModel); ok {
			m.namespaceSelector = selector
		}
		return m, cmd
	}
	if m.namespaceSelector != nil {
		updated, cmd := m.namespaceSelector.Update(msg)
		if updated == nil {
			m.namespaceSelector = nil // Closed
			m.pendingKubeconfig = ""
		} else if selector, ok := updated.(*models.NamespaceSelectorModel); ok {
			m.namespaceSelector = selector
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
	// Show size check if terminal is too small
	if m.sizeCheck != nil && !m.sizeCheck.IsSizeValid() {
		return m.sizeCheck.View()
	}

	// Show error popup if present
	if m.errorPopup != nil {
		return m.errorPopup.View()
	}

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
