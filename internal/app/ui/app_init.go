package ui

import (
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

func initializeAppConfigAndColors() config.AppConfig {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		panic("Failed to load app config: " + err.Error())
	}

	if err := customstyles.InitColors(); err != nil {
		panic("Failed to initialize colors: " + err.Error())
	}

	return appConfig
}

func createAppModelWithKubeClient(cfg cli.Config, appConfig config.AppConfig, pluginManager *plugins.PluginManager, kubeClient *resources.Client) *AppModel {
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

	setupPluginManagerForKubeClient(appModel, pluginManager, cfg, tabManager, header)

	initializeTabs(appModel, tabManager, header)

	return appModel
}

func setupPluginManagerForKubeClient(appModel *AppModel, pluginManager *plugins.PluginManager, cfg cli.Config, tabManager *models.TabManager, header models.HeaderModel) {
	if pluginManager == nil {
		return
	}

	pluginManager.GetAPI().SetCurrentNamespace(cfg.Namespace)

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

	pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
		logger.Info(fmt.Sprintf("DEBUG: SetNamespaceCallback called with namespace: %s", namespace))
		appModel.header.SetNamespace(namespace)
		if tabManager != nil {
			tabManager.SetNamespace(namespace)
		}
		if appModel.kube.Clientset != nil {
			appModel.kube.SetNamespace(namespace)
		}
		logger.Info("DEBUG: Called header.SetNamespace, tabManager.SetNamespace, and kube.SetNamespace, now calling UpdateContent")
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

func initializeTabs(appModel *AppModel, tabManager *models.TabManager, header models.HeaderModel) {
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
}

func createAppModelWithoutKubeClient(appConfig config.AppConfig, pluginManager *plugins.PluginManager, err error) *AppModel {
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
		setupPluginManagerForNoKubeClient(appModel, pluginManager)
	}

	return appModel
}

func setupPluginManagerForNoKubeClient(appModel *AppModel, pluginManager *plugins.PluginManager) {
	pluginManager.GetAPI().SetCurrentNamespace("default")

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
}

func createFallbackAppModel(appConfig config.AppConfig, pluginManager *plugins.PluginManager) *AppModel {
	uiInjector := NewUIInjector()
	appModel := &AppModel{
		header:        models.NewHeader("K8s TUI", nil),
		config:        appConfig,
		pluginManager: pluginManager,
		uiInjector:    uiInjector,
	}

	if pluginManager != nil {
		setupPluginManagerForNoKubeClient(appModel, pluginManager)
	}

	return appModel
}
