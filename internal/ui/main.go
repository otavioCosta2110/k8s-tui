package ui

import (
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/cli"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"otaviocosta2110/k8s-tui/internal/ui/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	tabManager          *models.TabManager
	kube                k8s.Client
	header              models.HeaderModel
	configSelected      bool
	errorPopup          *models.ErrorModel
	quickNav            tea.Model
	currentResourceType string
	breadcrumbTrail     []string
}

func NewAppModel() *AppModel {
	cfg := cli.ParseFlags()

	if err := customstyles.InitColors(); err != nil {
		panic("Failed to initialize colors: " + err.Error())
	}

	kubeClient, err := k8s.NewClient(cfg.KubeconfigPath, cfg.Namespace)
	if err == nil && kubeClient != nil {
		header := models.NewHeader("K8s TUI", kubeClient)
		header.SetNamespace(cfg.Namespace)

		tabManager := models.NewTabManager(kubeClient, cfg.Namespace)

		appModel := &AppModel{
			tabManager:     tabManager,
			header:         header,
			kube:           *kubeClient,
			configSelected: true,
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
		return &AppModel{
			header:     models.NewHeader("K8s TUI", nil),
			errorPopup: &popup,
		}
	}

	appModel := &AppModel{
		header: models.NewHeader("K8s TUI", nil),
	}

	return appModel
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
		global.ScreenWidth = msg.Width - global.Margin
		global.ScreenHeight = msg.Height - global.Margin
		if !global.IsHeaderActive {
			global.HeaderSize = global.ScreenHeight/4 - (global.Margin * 2)
			global.IsHeaderActive = true
		}
		global.ScreenHeight -= global.HeaderSize
		global.ScreenHeight -= global.TabBarSize
		global.IsTabBarActive = true

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
			case "esc", "g":
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
		case "q":
			if m.tabManager != nil {
				updatedManager, cmd := m.tabManager.Update(msg)
				if manager, ok := updatedManager.(*models.TabManager); ok {
					m.tabManager = manager
				}
				return m, cmd
			}
			return m, nil
		case "Q":
			return m, tea.Quit
		case "g":
			if m.quickNav != nil {
				m.quickNav = nil
				return m, nil
			}
			m.quickNav = models.NewQuickNavModel(m.kube, m.kube.Namespace)
			return m, m.quickNav.Init()
		case "ctrl+t":
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
		case "ctrl+w":
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
			popup.SetDimensions(global.ScreenWidth, global.ScreenHeight)

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

	height := global.ScreenHeight

	headerView := m.header.View()

	content := lipgloss.NewStyle().
		Width(global.ScreenWidth).
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
			breadcrumbEnd := min(activeTab.CurrentIndex + 1, len(activeTab.Breadcrumb))
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
				Width(global.ScreenWidth + global.Margin).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(breadcrumbStr)
		}
	}

	if !m.configSelected {
		var finalView string
		if breadcrumbView != "" {
			finalView = lipgloss.JoinVertical(lipgloss.Top,
				lipgloss.NewStyle().
					Width(global.ScreenWidth).
					Height(global.ScreenHeight+global.HeaderSize).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color(customstyles.BorderColor)).
					BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					Render(currentView),
				breadcrumbView)
		} else {
			finalView = lipgloss.NewStyle().
				Width(global.ScreenWidth).
				Height(global.ScreenHeight + global.HeaderSize).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(customstyles.BorderColor)).
				BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Render(currentView)
		}

		return finalView
	}

	if !global.IsHeaderActive {
		var finalView string
		if breadcrumbView != "" {
			finalView = lipgloss.JoinVertical(lipgloss.Top, content, breadcrumbView)
		} else {
			finalView = content
		}

		return finalView
	}

	var finalView string
	if breadcrumbView != "" {
		finalView = lipgloss.JoinVertical(lipgloss.Top, headerView, content, breadcrumbView)
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
	if m.tabManager != nil {
		tabs := m.tabManager.GetTabsForComponent()
		m.header.ClearTabs()
		activeIndex := -1
		for i, tab := range tabs {
			m.header.AddTab(tab.ID, tab.Title, tab.ResourceType)
			if tab.IsActive {
				activeIndex = i
			}
		}
		if activeIndex >= 0 {
			m.header.SetActiveTab(activeIndex)
		}
	}
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
