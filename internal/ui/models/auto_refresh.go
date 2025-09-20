package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RefreshableModel interface {
	tea.Model
	Refresh() (tea.Model, tea.Cmd)
}

type AutoRefreshModel struct {
	inner           RefreshableModel
	refreshInterval time.Duration
	lastRefresh     time.Time
	k8sClient       *k8s.Client
	footerText      string
}

func NewAutoRefreshModel(inner RefreshableModel, interval time.Duration, client *k8s.Client, footerText string) *AutoRefreshModel {
	// Prevent refresh intervals that are too small to avoid infinite loops
	if interval < time.Second {
		interval = 10 * time.Second // Default to 10 seconds minimum
	}

	return &AutoRefreshModel{
		inner:           inner,
		refreshInterval: interval,
		k8sClient:       client,
		footerText:      footerText,
	}
}

func (m *AutoRefreshModel) Init() tea.Cmd {
	return tea.Batch(
		m.inner.Init(),
		m.refreshTick(),
	)
}

func (m *AutoRefreshModel) refreshTick() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return components.RefreshMsg{}
	})
}

func (m *AutoRefreshModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *AutoRefreshModel) View() string {
	return m.inner.View()
}

func (m *AutoRefreshModel) GetCheckedItems() []int {
	if tableModel, ok := m.inner.(*ui.TableModel); ok {
		return tableModel.GetCheckedItems()
	}
	return []int{}
}

func (m *AutoRefreshModel) SetFooterText(text string) {
	m.footerText = text
}

func (m *AutoRefreshModel) Refresh() (tea.Model, tea.Cmd) {
	if m.inner == nil {
		return m, nil
	}

	updatedModel, cmd := m.inner.Refresh()
	m.inner = updatedModel.(RefreshableModel)
	return m, cmd
}
