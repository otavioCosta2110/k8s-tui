package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RefreshableModel interface {
	tea.Model
	Refresh() (tea.Model, tea.Cmd)
}

type autoRefreshModel struct {
	inner           RefreshableModel
	refreshInterval time.Duration
	lastRefresh     time.Time
	k8sClient       *k8s.Client
}

func NewAutoRefreshModel(inner RefreshableModel, interval time.Duration, client *k8s.Client) *autoRefreshModel {
	return &autoRefreshModel{
		inner:           inner,
		refreshInterval: interval,
		k8sClient:       client,
	}
}

func (m *autoRefreshModel) Init() tea.Cmd {
	return tea.Batch(
		m.inner.Init(),
		m.refreshTick(),
	)
}

func (m *autoRefreshModel) refreshTick() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return components.RefreshMsg{}
	})
}

func (m *autoRefreshModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.RefreshMsg:
		var cmd tea.Cmd
		updatedModel, cmd := m.inner.Refresh()
		m.inner = updatedModel.(RefreshableModel)
		return m, tea.Batch(cmd, m.refreshTick())

	default:
		var cmd tea.Cmd
		updatedModel, cmd := m.inner.Update(msg)
		m.inner = updatedModel.(RefreshableModel)
		return m, cmd
	}
}

func (m *autoRefreshModel) View() string {
	return m.inner.View()
}

func (m *autoRefreshModel) GetCheckedItems() []int {
	if tableModel, ok := m.inner.(*ui.TableModel); ok {
		return tableModel.GetCheckedItems()
	}
	return []int{}
}
