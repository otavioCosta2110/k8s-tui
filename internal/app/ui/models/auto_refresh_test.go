package models

import (
	"testing"
	"time"

	ui "github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type MockRefreshableModel struct {
	refreshCount int
	initCalled   bool
	updateCalled bool
}

func (m *MockRefreshableModel) Init() tea.Cmd {
	m.initCalled = true
	return func() tea.Msg { return nil }
}

func (m *MockRefreshableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateCalled = true
	return m, func() tea.Msg { return nil }
}

func (m *MockRefreshableModel) View() string {
	return "mock view"
}

func (m *MockRefreshableModel) Refresh() (tea.Model, tea.Cmd) {
	m.refreshCount++
	return m, func() tea.Msg { return nil }
}

func TestNewAutoRefreshModel(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	interval := 5 * time.Second
	client := &k8s.Client{}
	footerText := "Test Footer"

	model := NewAutoRefreshModel(mockInner, interval, client, footerText)

	if model.inner != mockInner {
		t.Error("Inner model not set correctly")
	}

	if model.refreshInterval != interval {
		t.Error("Refresh interval not set correctly")
	}

	if model.k8sClient != client {
		t.Error("K8s client not set correctly")
	}

	if model.footerText != footerText {
		t.Error("Footer text not set correctly")
	}
}

func TestAutoRefreshModelInit(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	cmd := model.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}

	if !mockInner.initCalled {
		t.Error("Inner model's Init should have been called")
	}
}

func TestAutoRefreshModelUpdateWithRefreshMsg(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	newModel, cmd := model.Update(ui.RefreshMsg{})

	if newModel == nil {
		t.Error("Update should return a model")
	}

	if mockInner.refreshCount != 1 {
		t.Errorf("Expected refresh count to be 1, got %d", mockInner.refreshCount)
	}

	if cmd == nil {
		t.Error("Update should return a command for the next refresh tick")
	}
}

func TestAutoRefreshModelUpdateWithOtherMsg(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	testMsg := "test message"
	newModel, cmd := model.Update(testMsg)

	if newModel == nil {
		t.Error("Update should return a model")
	}

	if mockInner.refreshCount != 0 {
		t.Errorf("Expected refresh count to be 0, got %d", mockInner.refreshCount)
	}

	if cmd == nil {
		t.Error("Update should return a command from inner model")
	}
}

func TestAutoRefreshModelView(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "Test Footer")

	view := model.View()

	if view == "" {
		t.Error("View should return a non-empty string")
	}

	expectedView := mockInner.View()
	if view != expectedView {
		t.Errorf("Expected view %q, got %q", expectedView, view)
	}
}

func TestAutoRefreshModelRefresh(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	newModel, cmd := model.Refresh()

	if newModel == nil {
		t.Error("Refresh should return a model")
	}

	if mockInner.refreshCount != 1 {
		t.Errorf("Expected refresh count to be 1, got %d", mockInner.refreshCount)
	}

	if cmd == nil {
		t.Error("Refresh should return a command")
	}
}

func TestAutoRefreshModelSetFooterText(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	newFooterText := "New Footer Text"
	model.SetFooterText(newFooterText)

	if model.footerText != newFooterText {
		t.Errorf("Expected footer text %q, got %q", newFooterText, model.footerText)
	}
}

func TestRefreshTick(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 5*time.Second, nil, "")

	cmd := model.refreshTick()

	if cmd == nil {
		t.Error("refreshTick should return a command")
	}

}

func TestAutoRefreshModelFullRefreshCycle(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 100*time.Millisecond, nil, "")

	initCmd := model.Init()
	if initCmd == nil {
		t.Error("Init should return a command")
	}

	newModel, refreshCmd := model.Update(ui.RefreshMsg{})

	autoRefreshModel, ok := newModel.(*AutoRefreshModel)
	if !ok {
		t.Error("Update should return an AutoRefreshModel")
	}

	if mockInner.refreshCount != 1 {
		t.Errorf("Expected inner model to be refreshed once, got %d", mockInner.refreshCount)
	}

	if refreshCmd == nil {
		t.Error("Update should return a command for the next refresh tick")
	}

	if autoRefreshModel.inner != mockInner {
		t.Error("Inner model should remain the same instance")
	}
}

func TestAutoRefreshModelWithTableModel(t *testing.T) {
	columns := []table.Column{{Title: "Test", Width: 10}}
	rows := []table.Row{{"test data"}}

	refreshCallCount := 0
	refreshFunc := func() ([]table.Row, error) {
		refreshCallCount++
		return rows, nil
	}

	tableModel := ui.NewTable(columns, []float64{1.0}, rows, "Test Table", nil, 0, refreshFunc, nil)

	autoRefreshModel := NewAutoRefreshModel(tableModel, 5*time.Second, nil, "Test Footer")

	if autoRefreshModel.inner != tableModel {
		t.Error("TableModel should be set as inner model")
	}

	newModel, cmd := autoRefreshModel.Update(ui.RefreshMsg{})

	if newModel == nil {
		t.Error("Update should return a model")
	}

	if refreshCallCount != 1 {
		t.Errorf("Expected refresh function to be called once, got %d", refreshCallCount)
	}

	if cmd == nil {
		t.Error("Update should return a command")
	}
}

func TestAutoRefreshModelTimerScheduling(t *testing.T) {
	mockInner := &MockRefreshableModel{}
	model := NewAutoRefreshModel(mockInner, 1*time.Second, nil, "")

	initCmd := model.Init()
	if initCmd == nil {
		t.Error("Init should return a command")
	}

	if model.refreshInterval != 1*time.Second {
		t.Errorf("Expected refresh interval to be 1s, got %v", model.refreshInterval)
	}

	if model.inner != mockInner {
		t.Error("Inner model should be set correctly")
	}
}

func TestAutoRefreshModelRefreshInterval(t *testing.T) {
	testCases := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		1 * time.Minute,
	}

	for _, interval := range testCases {
		mockInner := &MockRefreshableModel{}
		model := NewAutoRefreshModel(mockInner, interval, nil, "")

		if model.refreshInterval != interval {
			t.Errorf("Expected refresh interval %v, got %v", interval, model.refreshInterval)
		}
	}
}
