package ui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
)

func TestMultiClusterModelInvalidKubeconfig(t *testing.T) {
	// Test error handling functionality by directly setting up the state
	pluginDir := "/tmp/test-plugins"
	cfg := cli.Config{
		PluginDir: pluginDir,
	}

	model := NewMultiClusterModel(cfg)

	// Manually set up the state as if user had selected an invalid kubeconfig
	invalidKubeconfig := "/path/to/invalid/kubeconfig"
	model.pendingKubeconfig = invalidKubeconfig
	model.namespaceSelector = models.NewNamespaceSelectorModel(invalidKubeconfig)

	t.Logf("Setup state: pendingKubeconfig=%s, namespaceSelector=%v, errorScreen=%v",
		model.pendingKubeconfig, model.namespaceSelector != nil, model.errorScreen != nil)

	// Simulate NamespaceSelectedMsg - this should trigger error screen for invalid kubeconfig
	namespaceMsg := models.NamespaceSelectedMsg{Namespace: "default"}
	t.Logf("Sending NamespaceSelectedMsg: Namespace=%s", namespaceMsg.Namespace)
	updatedModel, _ := model.Update(namespaceMsg)

	multiClusterModel, ok := updatedModel.(*MultiClusterModel)
	if !ok {
		t.Fatal("Model should be MultiClusterModel")
	}

	t.Logf("After NamespaceSelectedMsg: errorScreen=%v, namespaceSelector=%v, pendingKubeconfig=%s, clusters=%d",
		multiClusterModel.errorScreen != nil, multiClusterModel.namespaceSelector != nil, multiClusterModel.pendingKubeconfig, len(multiClusterModel.clusters))

	// The error screen should be shown due to invalid kubeconfig
	if multiClusterModel.errorScreen == nil {
		t.Error("Error screen should be shown for invalid kubeconfig")
	}
	if multiClusterModel.namespaceSelector != nil {
		t.Error("Namespace selector should be cleared")
	}
	if multiClusterModel.pendingKubeconfig != "" {
		t.Error("Pending kubeconfig should be cleared")
	}
}

func TestMultiClusterModelErrorScreenDismissal(t *testing.T) {
	// Create a MultiClusterModel with an error screen
	pluginDir := "/tmp/test-plugins"
	cfg := cli.Config{
		PluginDir: pluginDir,
	}

	model := NewMultiClusterModel(cfg)

	// Create an error screen
	testError := models.NewErrorScreen(errors.New("test error"), "Test Error", "Test message")
	model.errorScreen = &testError

	// Test ESC key dismisses error screen
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(escMsg)

	multiClusterModel, ok := updatedModel.(*MultiClusterModel)
	if !ok {
		t.Fatal("Model should be MultiClusterModel")
	}
	if multiClusterModel.errorScreen != nil {
		t.Error("Error screen should be dismissed")
	}
}

func TestMultiClusterModelErrorScreenView(t *testing.T) {
	// Create a MultiClusterModel with an error screen
	pluginDir := "/tmp/test-plugins"
	cfg := cli.Config{
		PluginDir: pluginDir,
	}

	model := NewMultiClusterModel(cfg)

	// Create an error screen
	testError := models.NewErrorScreen(errors.New("test error"), "Test Error", "Test message")
	model.errorScreen = &testError

	// Test that View() returns the error screen view
	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}
	// Check that the view contains expected content (basic check)
	if len(view) < 10 {
		t.Error("View should contain substantial content")
	}
}
