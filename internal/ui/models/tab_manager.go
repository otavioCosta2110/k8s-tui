package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"

	tea "github.com/charmbracelet/bubbletea"
)

type TabData struct {
	ID           string
	Title        string
	ResourceType string
	Model        tea.Model
	Breadcrumb   []string
	ScreenStack  []tea.Model
	CurrentIndex int
}

type TabManager struct {
	tabs        []TabData
	activeIndex int
	kubeClient  *k8s.Client
	namespace   string
	keyBindings map[string]string
}

type TabManagerMsg struct {
	Action       string
	TabID        string
	ResourceType string
	NewModel     tea.Model
	Breadcrumb   string
}

func NewTabManager(kubeClient *k8s.Client, namespace string, keyBindings map[string]string) *TabManager {
	tm := &TabManager{
		tabs:        []TabData{},
		activeIndex: 0,
		kubeClient:  kubeClient,
		namespace:   namespace,
		keyBindings: keyBindings,
	}

	tm.createInitialTab()

	return tm
}

func (tm *TabManager) getKeyBinding(action string) string {
	if binding, exists := tm.keyBindings[action]; exists {
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
	resourceModel := NewResource(*tm.kubeClient, tm.namespace)
	resourceComponent := resourceModel.InitComponent(*tm.kubeClient)

	initialTab := TabData{
		ID:           "initial",
		Title:        "Resources",
		ResourceType: "ResourceList",
		Model:        resourceComponent,
		Breadcrumb:   []string{"Resource List"},
		ScreenStack:  []tea.Model{resourceComponent},
		CurrentIndex: 0,
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

			if msg.Breadcrumb != "" {
				activeTab.Breadcrumb = append(activeTab.Breadcrumb, msg.Breadcrumb)
			}

			if len(activeTab.Breadcrumb) > 0 {
				activeTab.Title = activeTab.Breadcrumb[len(activeTab.Breadcrumb)-1]
				activeTab.ResourceType = activeTab.Title
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
			return tm.navigateBack()
		}

		if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
			var cmd tea.Cmd
			tm.tabs[tm.activeIndex].Model, cmd = tm.tabs[tm.activeIndex].Model.Update(msg)
			return tm, cmd
		}
	}

	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		var cmd tea.Cmd
		tm.tabs[tm.activeIndex].Model, cmd = tm.tabs[tm.activeIndex].Model.Update(msg)
		return tm, cmd
	}

	return tm, nil
}

func (tm *TabManager) View() string {
	if tm.activeIndex >= 0 && tm.activeIndex < len(tm.tabs) {
		return tm.tabs[tm.activeIndex].Model.View()
	}
	return "No active tab"
}

func (tm *TabManager) CreateNewTab(model tea.Model, breadcrumb string) (tea.Model, tea.Cmd) {
	resourceType := breadcrumb
	if resourceType == "" {
		resourceType = "Unknown"
	}

	tabID := fmt.Sprintf("tab-%d", len(tm.tabs)+1)

	newTab := TabData{
		ID:           tabID,
		Title:        breadcrumb,
		ResourceType: resourceType,
		Model:        model,
		Breadcrumb:   []string{breadcrumb},
		ScreenStack:  []tea.Model{model},
		CurrentIndex: 0,
	}

	tm.tabs = append(tm.tabs, newTab)
	tm.activeIndex = len(tm.tabs) - 1

	return tm, model.Init()
}

func (tm *TabManager) CreateNewResourceTab() (tea.Model, tea.Cmd) {
	resourceModel := NewResource(*tm.kubeClient, tm.namespace)
	resourceComponent := resourceModel.InitComponent(*tm.kubeClient)

	return tm.CreateNewTab(resourceComponent, "Resource List")
}

func (tm *TabManager) switchToTab(tabID string) (tea.Model, tea.Cmd) {
	for i, tab := range tm.tabs {
		if tab.ID == tabID {
			tm.activeIndex = i
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

func (tm *TabManager) SetActiveTab(index int) {
	if index >= 0 && index < len(tm.tabs) {
		tm.activeIndex = index
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
				activeTab.ResourceType = activeTab.Title
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
				activeTab.ResourceType = activeTab.Title
			}
			return tm, activeTab.Model.Init()
		}
	}
	return tm, nil
}
