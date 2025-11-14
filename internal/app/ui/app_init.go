package ui

import (
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
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

func createAppModelWithKubeClient(cfg cli.Config, appConfig config.AppConfig, pluginManager *plugins.GlobalPluginManager, kubeClient *resources.Client) *AppModel {
	var pluginAPI interface{}
	if pluginManager != nil {
		pluginAPI = pluginManager.GetAPI()
	}
	header := models.NewHeader("K8s TUI", kubeClient, pluginAPI)
	header.SetNamespace(cfg.Namespace)

	tabManager := models.NewTabManager(kubeClient, cfg.Namespace, appConfig.KeyBindings)

	if pluginManager != nil {
		pluginManager.GetAPI().SetCurrentResourceType("ResourceList")
		tabManager.SetResourceTypeCallback(func(resourceType string) {
			pluginManager.GetAPI().SetCurrentResourceType(resourceType)
		})
	}

	uiInjector := NewUIInjector()

	appModel := &AppModel{
		tabManager:     tabManager,
		header:         header,
		kube:           *kubeClient,
		config:         appConfig,
		configSelected: true,
		helpScreen:     components.NewHelpModel(),
		pluginManager:  pluginManager,
		uiInjector:     uiInjector,
	}

	setupPluginManagerForKubeClient(appModel, pluginManager, cfg, tabManager)

	initializeTabs(tabManager, header)

	return appModel
}

func setupPluginManagerForKubeClient(appModel *AppModel, pluginManager *plugins.GlobalPluginManager, cfg cli.Config, tabManager *models.TabManager) {
	if pluginManager == nil {
		return
	}

	pluginManager.GetAPI().SetCurrentNamespace(cfg.Namespace)

	pluginManager.GetAPI().SetTabGetter(func() ([]plugins.TabInfo, error) {
		return tabManager.GetTabsInfo(), nil
	})

	pluginManager.GetAPI().SetTabSetter(func(tabInfos []plugins.TabInfo) error {
		return tabManager.RestoreTabs(tabInfos)
	})

	pluginManager.GetAPI().SetTabSetterCallback(func() {
		appModel.updateHeaderTabs()
	})

	pluginManager.GetAPI().SetBreadcrumbCallback(func(breadcrumb []string) {
		appModel.breadcrumbTrail = breadcrumb
	})

	pluginManager.GetAPI().GetBreadcrumbCallback(func() []string {
		return appModel.breadcrumbTrail
	})

	pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
		logger.Info(fmt.Sprintf("DEBUG: SetNamespaceCallback called with namespace: %s for cluster", namespace))
		appModel.header.SetNamespace(namespace)
		if tabManager != nil {
			tabManager.SetNamespace(namespace)
		}
		if appModel.kube.Clientset != nil {
			appModel.kube.SetNamespace(namespace)
		}
		appModel.header.UpdateContent()
		logger.Info(fmt.Sprintf("DEBUG: Plugin changed namespace to: %s", namespace))
	})

	pluginManager.GetAPI().SetStatusCallback(func(message string) {
		logger.Info(fmt.Sprintf("Plugin status: %s", message))
	})

	pluginManager.GetAPI().SetShowInputDialogCallback(func(title, placeholder, submitCommand, cancelCommand string) {
		appModel.pendingInputDialog = &InputDialogRequest{
			Title:         title,
			Placeholder:   placeholder,
			SubmitCommand: submitCommand,
			CancelCommand: cancelCommand,
		}
	})

	appModel.loadPluginUIExtensions()

	if err := cli.HandlePluginArgs(pluginManager, cfg.PluginArgs); err != nil {
		logger.Error(fmt.Sprintf("Failed to handle plugin CLI arguments: %v", err))
	}
}

func initializeTabs(tabManager *models.TabManager, header models.HeaderModel) {
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

func createAppModelWithoutKubeClient(appConfig config.AppConfig, pluginManager *plugins.GlobalPluginManager, err error) *AppModel {
	popup := models.NewErrorScreen(err, "Failed to initialize Kubernetes config", "")
	uiInjector := NewUIInjector()
	appModel := &AppModel{
		header:         models.NewHeader("K8s TUI", nil, nil),
		config:         appConfig,
		configSelected: true,
		errorPopup:     &popup,
		helpScreen:     components.NewHelpModel(),
		pluginManager:  pluginManager,
		uiInjector:     uiInjector,
		// kube is zero value (nil Clientset)
	}

	if pluginManager != nil {
		setupPluginManagerForNoKubeClient(appModel, pluginManager)
	}

	return appModel
}

func setupPluginManagerForNoKubeClient(appModel *AppModel, pluginManager *plugins.GlobalPluginManager) {
	pluginManager.GetAPI().SetCurrentNamespace("default")

	pluginManager.GetAPI().SetNamespaceCallback(func(namespace string) {
		logger.Info(fmt.Sprintf("DEBUG: SetNamespaceCallback called with namespace: %s", namespace))
		// appModel.header.SetNamespace(namespace)
		logger.Info("DEBUG: Called header.SetNamespace, now calling UpdateContent")
		appModel.header.UpdateContent()
		logger.Info(fmt.Sprintf("Plugin changed namespace to: %s", namespace))
	})

	pluginManager.GetAPI().SetStatusCallback(func(message string) {
		logger.Info(fmt.Sprintf("Plugin status: %s", message))
	})

	pluginManager.GetAPI().SetShowInputDialogCallback(func(title, placeholder, submitCommand, cancelCommand string) {
		appModel.pendingInputDialog = &InputDialogRequest{
			Title:         title,
			Placeholder:   placeholder,
			SubmitCommand: submitCommand,
			CancelCommand: cancelCommand,
		}
	})

	appModel.loadPluginUIExtensions()
}

func createFallbackAppModel(appConfig config.AppConfig, pluginManager *plugins.GlobalPluginManager) *AppModel {
	uiInjector := NewUIInjector()
	appModel := &AppModel{
		header:         models.NewHeader("K8s TUI", nil, nil),
		config:         appConfig,
		configSelected: true,
		helpScreen:     components.NewHelpModel(),
		pluginManager:  pluginManager,
		uiInjector:     uiInjector,
		// kube is zero value (nil Clientset)
	}

	if pluginManager != nil {
		setupPluginManagerForNoKubeClient(appModel, pluginManager)
	}

	return appModel
}
