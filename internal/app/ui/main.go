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
	pendingInputDialog  *InputDialogRequest
	currentResourceType string
	breadcrumbTrail     []string
	pluginManager       *plugins.PluginManager
	uiInjector          *UIInjector
}

func NewAppModel(cfg cli.Config, pluginManager *plugins.PluginManager) *AppModel {
	appConfig := initializeAppConfigAndColors()

	kubeClient, err := resources.NewClient(cfg.KubeconfigPath, cfg.Namespace)
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
	case components.TextInputSubmitMsg:
		return m, nil
	case components.TextInputCancelMsg:
		return m, nil
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
