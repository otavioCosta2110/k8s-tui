package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"maps"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type TabData struct {
	ID            string
	Title         string
	ResourceType  string
	Namespace     string
	Model         tea.Model
	ResourceModel interface{}
	Breadcrumb    []string
	ScreenStack   []tea.Model
	CurrentIndex  int
	Metadata      map[string]interface{}
}

type TabManager struct {
	tabs                 []TabData
	activeIndex          int
	kubeClient           *k8s.Client
	namespace            string
	keyBindings          map[string]string
	resourceTypeCallback func(resourceType string)
	namespaceCallback    func(namespace string)
}

type TabManagerMsg struct {
	Action       string
	TabID        string
	ResourceType string
	NewModel     tea.Model
	Breadcrumb   string
}

func NewTabManager(kubeClient *k8s.Client, namespace string, keyBindings map[string]string) *TabManager {
	if namespace == "" {
		namespace = "default"
	}
	logger.Info("DEBUG: Creating new TabManager")
	tm := &TabManager{
		tabs:        []TabData{},
		activeIndex: 0,
		kubeClient:  kubeClient,
		namespace:   namespace,
		keyBindings: keyBindings,
	}

	tm.createInitialTab()
	logger.Info(fmt.Sprintf("DEBUG: TabManager created with %d tabs", len(tm.tabs)))

	return tm
}

func (tm *TabManager) getKeyBinding(action string) string {
	// Create reverse lookup map from key->action to action->key
	reverseBindings := make(map[string]string)
	for key, cmd := range tm.keyBindings {
		reverseBindings[cmd] = key
	}

	if binding, exists := reverseBindings[action]; exists {
		return binding
	}
	defaults := map[string]string{
		"quit":    "q",
		"back":    "[",
		"forward": "]",
		"new_tab": "ctrl+t",
	}
	return defaults[action]
}

func (tm *TabManager) createInitialTab() {
	if tm.kubeClient == nil {
		logger.Error("Cannot create initial tab: kubeClient is nil")
		return
	}
	resourceModel := NewResource(*tm.kubeClient, tm.namespace)
	resourceComponent := resourceModel.InitComponent(tm.kubeClient)

	initialTab := TabData{
		ID:            "initial",
		Title:         "Resources",
		ResourceType:  "ResourceList",
		Namespace:     tm.namespace,
		Model:         resourceComponent,
		ResourceModel: resourceModel,
		Breadcrumb:    []string{"Resource List"},
		ScreenStack:   []tea.Model{resourceComponent},
		CurrentIndex:  0,
		Metadata:      make(map[string]interface{}),
	}

	tm.tabs = append(tm.tabs, initialTab)
}

func (tm *TabManager) Init() tea.Cmd {
	if len(tm.tabs) > 0 {
		return tm.tabs[0].Model.Init()
	}
	return nil
}

func (tm *TabManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.TabMsg:
		switch msg.Action {
		case "switch":
			return tm.switchToTab(msg.TabID)
		case "close":
			return tm.closeTab(msg.TabID)
		}

	case components.NavigateMsg:
		if msg.Error != nil {
			return tm, nil
		}

		if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
			activeTab := &tm.tabs[tm.activeIndex]

			if activeTab.CurrentIndex < len(activeTab.ScreenStack)-1 {
				activeTab.ScreenStack = activeTab.ScreenStack[:activeTab.CurrentIndex+1]
				if activeTab.CurrentIndex+1 < len(activeTab.Breadcrumb) {
					activeTab.Breadcrumb = activeTab.Breadcrumb[:activeTab.CurrentIndex+1]
				}
			}

			activeTab.ScreenStack = append(activeTab.ScreenStack, msg.NewScreen)
			activeTab.CurrentIndex = len(activeTab.ScreenStack) - 1
			activeTab.Model = msg.NewScreen

			if msg.ResourceModel != nil {
				activeTab.ResourceModel = msg.ResourceModel
			}

			if msg.Breadcrumb != "" {
				activeTab.Breadcrumb = append(activeTab.Breadcrumb, msg.Breadcrumb)
			}

			if msg.Metadata != nil {
				logger.Info(fmt.Sprintf("DEBUG: Received metadata in NavigateMsg: %v", msg.Metadata))
				if activeTab.Metadata == nil {
					activeTab.Metadata = make(map[string]interface{})
				}
				maps.Copy(activeTab.Metadata, msg.Metadata)
				logger.Info(fmt.Sprintf("DEBUG: Stored metadata in activeTab: %v", activeTab.Metadata))
			}

			if len(activeTab.Breadcrumb) > 0 {
				activeTab.Title = activeTab.Breadcrumb[len(activeTab.Breadcrumb)-1]
				activeTab.ResourceType = tm.inferResourceTypeFromBreadcrumb(activeTab.Title)
				if tm.resourceTypeCallback != nil {
					tm.resourceTypeCallback(activeTab.ResourceType)
				}
			}

			return tm, msg.NewScreen.Init()
		}

		return tm.CreateNewTab(msg.NewScreen, msg.Breadcrumb)

	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		for i := range tm.tabs {
			var cmd tea.Cmd
			tm.tabs[i].Model, cmd = tm.tabs[i].Model.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return tm, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case tm.getKeyBinding("new_tab"):
			return tm.CreateNewResourceTab()
		case tm.getKeyBinding("back"):
			return tm.navigateBack()
		case tm.getKeyBinding("forward"):
			return tm.navigateForward()
		case tm.getKeyBinding("quit"):
			// Check if current tab is in search mode before handling quit
			if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
				activeTab := &tm.tabs[tm.activeIndex]
				if tableModel, ok := activeTab.Model.(*components.TableModel); ok {
					view := tableModel.View()
					if strings.Contains(view, "Search:") && strings.Contains(view, "█") {
						// In search mode, let the table handle the key
						var cmd tea.Cmd
						activeTab.Model, cmd = activeTab.Model.Update(msg)
						return tm, cmd
					}
				}
				// Check if it's a resource model that contains a table
				if resourceModel, ok := activeTab.Model.(interface{ GetTable() *components.TableModel }); ok {
					if table := resourceModel.GetTable(); table != nil {
						view := table.View()
						if strings.Contains(view, "Search:") && strings.Contains(view, "█") {
							// In search mode, let's resource model handle the key
							var cmd tea.Cmd
							activeTab.Model, cmd = activeTab.Model.Update(msg)
							return tm, cmd
						}
					}
				}
				// Check if it's an AutoRefreshModel that contains a table
				if autoRefreshModel, ok := activeTab.Model.(interface{ GetTable() *components.TableModel }); ok {
					if table := autoRefreshModel.GetTable(); table != nil {
						view := table.View()
						if strings.Contains(view, "Search:") && strings.Contains(view, "█") {
							// In search mode, let's auto refresh model handle the key
							var cmd tea.Cmd
							activeTab.Model, cmd = activeTab.Model.Update(msg)
							return tm, cmd
						}
					}
				}
			}
			return tm.navigateBack()

		default:
			if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
				var cmd tea.Cmd
				tm.tabs[tm.activeIndex].Model, cmd = tm.tabs[tm.activeIndex].Model.Update(msg)
				return tm, cmd
			}
		}
	}

	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		var cmd tea.Cmd
		tm.tabs[tm.activeIndex].Model, cmd = tm.tabs[tm.activeIndex].Model.Update(msg)
		return tm, cmd
	}
	return tm, nil
}

func (tm *TabManager) RestoreTabs(tabInfos []plugins.TabInfo) error {
	logger.Info(fmt.Sprintf("DEBUG: Restoring %d tabs", len(tabInfos)))
	tm.tabs = []TabData{}
	tm.activeIndex = 0

	for i, tabInfo := range tabInfos {
		logger.Info(fmt.Sprintf("DEBUG: Restoring tab %d: ID=%s, Title=%s, ResourceType=%s, Breadcrumb=%v",
			i, tabInfo.ID, tabInfo.Title, tabInfo.ResourceType, tabInfo.Breadcrumb))

		var screenStack []tea.Model
		var finalResourceModel any
		var finalModel tea.Model

		for j, crumb := range tabInfo.Breadcrumb {
			var model tea.Model
			var resourceModel any
			var err error

			if j == 0 && crumb == "Resource List" {
				resourceModel = NewResource(*tm.kubeClient, tm.namespace)
				model = resourceModel.(Resource).InitComponent(tm.kubeClient)
			} else {
				resourceType := crumb

				if strings.HasSuffix(strings.ToLower(crumb), " pods") && j > 0 && tabInfo.Breadcrumb[j-1] == "Deployments" {
					resourceType = "Pods"
				} else {
					if strings.Contains(strings.ToLower(crumb), "pods") && !strings.Contains(strings.ToLower(crumb), "deployments") {
						resourceType = "Pods"
					} else if strings.Contains(strings.ToLower(crumb), "deployments") {
						resourceType = "Deployments"
					} else if strings.Contains(strings.ToLower(crumb), "services") {
						resourceType = "Services"
					} else if strings.Contains(strings.ToLower(crumb), "configmaps") {
						resourceType = "ConfigMaps"
					} else if strings.Contains(strings.ToLower(crumb), "secrets") {
						resourceType = "Secrets"
					} else if strings.Contains(strings.ToLower(crumb), "ingresses") {
						resourceType = "Ingresses"
					} else if strings.Contains(strings.ToLower(crumb), "jobs") {
						resourceType = "Jobs"
					} else if strings.Contains(strings.ToLower(crumb), "cronjobs") {
						resourceType = "CronJobs"
					} else if strings.Contains(strings.ToLower(crumb), "daemonsets") {
						resourceType = "DaemonSets"
					} else if strings.Contains(strings.ToLower(crumb), "statefulsets") {
						resourceType = "StatefulSets"
					} else if strings.Contains(strings.ToLower(crumb), "replicasets") {
						resourceType = "ReplicaSets"
					} else if strings.Contains(strings.ToLower(crumb), "nodes") {
						resourceType = "Nodes"
					} else if strings.Contains(strings.ToLower(crumb), "serviceaccounts") {
						resourceType = "ServiceAccounts"
					}
				}

				if resourceType == "Pods" {
					var selector string
					var parentResource string

					if tabInfo.Metadata != nil {
						if s, ok := tabInfo.Metadata["selector"].(string); ok {
							selector = s
						}
						if p, ok := tabInfo.Metadata["parent"].(string); ok {
							parentResource = p
						}
					}

					if selector == "" && strings.HasSuffix(strings.ToLower(crumb), " pods") && j > 0 && tabInfo.Breadcrumb[j-1] == "Deployments" {
						parentResource = strings.TrimSuffix(crumb, " pods")
						logger.Info(fmt.Sprintf("DEBUG: Inferring selector for deployment %s from breadcrumb", parentResource))
						deployment := k8s.NewDeploymentInfo(parentResource, tm.namespace, *tm.kubeClient)
						if err := deployment.Fetch(); err != nil {
							logger.Error(fmt.Sprintf("Failed to fetch deployment %s for selector: %v", parentResource, err))
							continue
						}
						selector, err = deployment.GetLabelSelector()
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to get label selector for deployment %s: %v", parentResource, err))
							continue
						}
					}

					if selector != "" {
						logger.Info(fmt.Sprintf("DEBUG: Creating pods model with selector '%s' for parent '%s'", selector, parentResource))
						podsModel, err := NewPodsWithParent(*tm.kubeClient, tm.namespace, parentResource, selector)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to create pods model for crumb %s: %v", crumb, err))
							continue
						}
						model, err = podsModel.InitComponent(tm.kubeClient)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to init pods model for crumb %s: %v", crumb, err))
							continue
						}
						resourceModel = podsModel
					} else {
						resourceList := NewResourceList(*tm.kubeClient, tm.namespace, resourceType)
						model, err = resourceList.InitComponent(*tm.kubeClient)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to create model for crumb %s (type %s): %v", crumb, resourceType, err))
							continue
						}
						resourceModel = resourceList
					}
				} else {
					resourceList := NewResourceList(*tm.kubeClient, tm.namespace, resourceType)
					model, err = resourceList.InitComponent(*tm.kubeClient)
					if err != nil {
						logger.Error(fmt.Sprintf("Failed to create model for crumb %s (type %s): %v", crumb, resourceType, err))
						continue
					}
					resourceModel = resourceList
				}
			}

			screenStack = append(screenStack, model)
			if j == len(tabInfo.Breadcrumb)-1 {
				finalResourceModel = resourceModel
				finalModel = model
			}
		}

		if len(screenStack) == 0 {
			logger.Error("No models created for tab, skipping")
			continue
		}

		currentIndex := 0
		if len(tabInfo.Breadcrumb) > 0 {
			currentIndex = len(tabInfo.Breadcrumb) - 1
		}

		tabData := TabData{
			ID:            tabInfo.ID,
			Title:         tabInfo.Title,
			ResourceType:  tabInfo.ResourceType,
			Namespace:     tabInfo.Namespace,
			Model:         finalModel,
			ResourceModel: finalResourceModel,
			Breadcrumb:    tabInfo.Breadcrumb,
			ScreenStack:   screenStack,
			CurrentIndex:  currentIndex,
			Metadata:      tabInfo.Metadata,
		}

		tm.tabs = append(tm.tabs, tabData)
	}

	if len(tm.tabs) > 0 {
		tm.activeIndex = 0
		logger.Info(fmt.Sprintf("DEBUG: Set active tab to index 0, total tabs: %d", len(tm.tabs)))
	}

	return nil
}

func (tm *TabManager) SetNamespace(namespace string) {
	tm.namespace = namespace
	// Refresh all existing resource models to use the new namespace
	tm.refreshAllResourceModels()
}

// refreshAllResourceModels refreshes all existing resource models to use the current namespace
func (tm *TabManager) refreshAllResourceModels() {
	for i := range tm.tabs {
		tab := &tm.tabs[i]
		if tab.ResourceModel != nil {
			// Update the namespace in the resource model
			if resourceModel, ok := tab.ResourceModel.(interface{ SetNamespace(string) }); ok {
				resourceModel.SetNamespace(tm.namespace)
			}
		}
	}
}

func (tm *TabManager) SetResourceTypeCallback(callback func(resourceType string)) {
	tm.resourceTypeCallback = callback
}

func (tm *TabManager) SetNamespaceCallback(callback func(namespace string)) {
	tm.namespaceCallback = callback
}

func (tm *TabManager) View() string {
	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		return tm.tabs[tm.activeIndex].Model.View()
	}
	return "No active tab"
}

func (tm *TabManager) CreateNewTab(model tea.Model, breadcrumb string) (tea.Model, tea.Cmd) {
	resourceType := tm.inferResourceTypeFromBreadcrumb(breadcrumb)
	if resourceType == "" {
		resourceType = "Unknown"
	}

	tabID := fmt.Sprintf("tab-%d", len(tm.tabs)+1)

	newTab := TabData{
		ID:            tabID,
		Title:         breadcrumb,
		ResourceType:  resourceType,
		Namespace:     tm.namespace,
		Model:         model,
		ResourceModel: nil,
		Breadcrumb:    []string{breadcrumb},
		ScreenStack:   []tea.Model{model},
		CurrentIndex:  0,
		Metadata:      make(map[string]interface{}),
	}

	tm.tabs = append(tm.tabs, newTab)
	tm.activeIndex = len(tm.tabs) - 1

	return tm, model.Init()
}

func (tm *TabManager) CreateNewResourceTab() (tea.Model, tea.Cmd) {
	resourceModel := NewResource(*tm.kubeClient, tm.namespace)
	resourceComponent := resourceModel.InitComponent(tm.kubeClient)

	tm.tabs = append(tm.tabs, TabData{
		ID:            fmt.Sprintf("tab-%d", len(tm.tabs)+1),
		Title:         "Resource List",
		ResourceType:  "ResourceList",
		Namespace:     tm.namespace,
		Model:         resourceComponent,
		ResourceModel: resourceModel,
		Breadcrumb:    []string{"Resource List"},
		ScreenStack:   []tea.Model{resourceComponent},
		CurrentIndex:  0,
		Metadata:      make(map[string]interface{}),
	})
	tm.activeIndex = len(tm.tabs) - 1

	return tm, resourceComponent.Init()
}

func (tm *TabManager) switchToTab(tabID string) (tea.Model, tea.Cmd) {
	for i, tab := range tm.tabs {
		if tab.ID == tabID {
			tm.activeIndex = i
			if tm.resourceTypeCallback != nil {
				tm.resourceTypeCallback(tm.tabs[i].ResourceType)
			}
			return tm, nil
		}
	}
	return tm, nil
}

func (tm *TabManager) closeTab(tabID string) (tea.Model, tea.Cmd) {
	if len(tm.tabs) <= 1 {
		return tm, nil
	}

	for i, tab := range tm.tabs {
		if tab.ID == tabID {
			tm.tabs = append(tm.tabs[:i], tm.tabs[i+1:]...)
			if tm.activeIndex >= len(tm.tabs) {
				tm.activeIndex = len(tm.tabs) - 1
			} else if tm.activeIndex > i && tm.activeIndex > 0 {
				tm.activeIndex--
			}
			break
		}
	}
	return tm, nil
}

func (tm *TabManager) findTabByResourceType(resourceType string) int {
	for i, tab := range tm.tabs {
		if tab.ResourceType == resourceType {
			return i
		}
	}
	return -1
}

func (tm *TabManager) updateBreadcrumb(tabIndex int, breadcrumb string) {
	if tabIndex >= 0 && tabIndex < len(tm.tabs) {
		for _, existing := range tm.tabs[tabIndex].Breadcrumb {
			if existing == breadcrumb {
				return
			}
		}
		tm.tabs[tabIndex].Breadcrumb = append(tm.tabs[tabIndex].Breadcrumb, breadcrumb)
	}
}

func (tm *TabManager) GetActiveTab() *TabData {
	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		return &tm.tabs[tm.activeIndex]
	}
	return nil
}

func (tm *TabManager) GetTabCount() int {
	return len(tm.tabs)
}

func (tm *TabManager) GetActiveTabIndex() int {
	return tm.activeIndex
}

func (tm *TabManager) GetTabsForComponent() []components.Tab {
	var tabs []components.Tab
	for i, tab := range tm.tabs {
		tabs = append(tabs, components.Tab{
			ID:           tab.ID,
			Title:        tab.Title,
			ResourceType: tab.ResourceType,
			IsActive:     i == tm.activeIndex,
			IsModified:   false,
			Breadcrumb:   tab.Breadcrumb,
			CurrentIndex: tab.CurrentIndex,
		})
	}
	return tabs
}

func (tm *TabManager) GetTabsInfo() []plugins.TabInfo {
	var tabs []plugins.TabInfo
	for _, tab := range tm.tabs {
		logger.Info(fmt.Sprintf("DEBUG: GetTabsInfo - tab %s has metadata: %v", tab.ID, tab.Metadata))
		tabs = append(tabs, plugins.TabInfo{
			ID:           tab.ID,
			Title:        tab.Title,
			ResourceType: tab.ResourceType,
			Namespace:    tab.Namespace,
			Breadcrumb:   tab.Breadcrumb,
			CurrentIndex: tab.CurrentIndex,
			Metadata:     tab.Metadata,
		})
	}
	return tabs
}

func (tm *TabManager) SetActiveTab(index int) {
	if index >= 0 && index < len(tm.tabs) {
		tm.activeIndex = index
		// Update namespace from the active tab if different
		activeTab := tm.tabs[index]
		if activeTab.Namespace != tm.namespace && tm.namespaceCallback != nil {
			tm.namespace = activeTab.Namespace
			tm.namespaceCallback(activeTab.Namespace)
		}
	}
}

func (tm *TabManager) navigateBack() (tea.Model, tea.Cmd) {
	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		activeTab := &tm.tabs[tm.activeIndex]
		if activeTab.CurrentIndex > 0 && activeTab.CurrentIndex < len(activeTab.ScreenStack) {
			activeTab.CurrentIndex--
			activeTab.Model = activeTab.ScreenStack[activeTab.CurrentIndex]
			if len(activeTab.Breadcrumb) > 0 && activeTab.CurrentIndex < len(activeTab.Breadcrumb) {
				activeTab.Title = activeTab.Breadcrumb[activeTab.CurrentIndex]
				activeTab.ResourceType = tm.inferResourceTypeFromBreadcrumb(activeTab.Title)
			}
			return tm, activeTab.Model.Init()
		}
	}
	return tm, nil
}

func (tm *TabManager) navigateForward() (tea.Model, tea.Cmd) {
	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		activeTab := &tm.tabs[tm.activeIndex]
		if activeTab.CurrentIndex < len(activeTab.ScreenStack)-1 && activeTab.CurrentIndex >= 0 {
			activeTab.CurrentIndex++
			activeTab.Model = activeTab.ScreenStack[activeTab.CurrentIndex]
			if len(activeTab.Breadcrumb) > 0 && activeTab.CurrentIndex < len(activeTab.Breadcrumb) {
				activeTab.Title = activeTab.Breadcrumb[activeTab.CurrentIndex]
				activeTab.ResourceType = tm.inferResourceTypeFromBreadcrumb(activeTab.Title)
				if tm.resourceTypeCallback != nil {
					tm.resourceTypeCallback(activeTab.ResourceType)
				}
			}
			return tm, activeTab.Model.Init()
		}
	}
	return tm, nil
}

func (tm *TabManager) inferResourceTypeFromBreadcrumb(breadcrumb string) string {
	if strings.HasSuffix(strings.ToLower(breadcrumb), " pods") {
		return "Pods"
	}

	if strings.Contains(strings.ToLower(breadcrumb), "deployments") {
		return "Deployments"
	} else if strings.Contains(strings.ToLower(breadcrumb), "services") {
		return "Services"
	} else if strings.Contains(strings.ToLower(breadcrumb), "configmaps") {
		return "ConfigMaps"
	} else if strings.Contains(strings.ToLower(breadcrumb), "secrets") {
		return "Secrets"
	} else if strings.Contains(strings.ToLower(breadcrumb), "ingresses") {
		return "Ingresses"
	} else if strings.Contains(strings.ToLower(breadcrumb), "jobs") {
		return "Jobs"
	} else if strings.Contains(strings.ToLower(breadcrumb), "cronjobs") {
		return "CronJobs"
	} else if strings.Contains(strings.ToLower(breadcrumb), "daemonsets") {
		return "DaemonSets"
	} else if strings.Contains(strings.ToLower(breadcrumb), "statefulsets") {
		return "StatefulSets"
	} else if strings.Contains(strings.ToLower(breadcrumb), "replicasets") {
		return "ReplicaSets"
	} else if strings.Contains(strings.ToLower(breadcrumb), "nodes") {
		return "Nodes"
	} else if strings.Contains(strings.ToLower(breadcrumb), "serviceaccounts") {
		return "ServiceAccounts"
	}

	return breadcrumb
}
