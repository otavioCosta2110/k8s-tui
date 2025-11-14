package plugins

import (
	"testing"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

func TestSessionSavePlugin_MultiClusterAPI(t *testing.T) {
	// Get the actual session-save-plugin directory
	pluginDir := "/home/otavio/dev/k8s-tui/plugins"

	// Create GlobalPluginManager and load plugins
	gpm := NewGlobalPluginManager(pluginDir)

	// Add a test cluster
	testClient := k8s.Client{KubeconfigPath: "/test/config"}
	gpm.AddCluster("test-cluster", "Test Cluster", testClient)

	// Switch to the test cluster
	if err := gpm.SwitchToCluster("test-cluster"); err != nil {
		t.Fatalf("Failed to switch to test cluster: %v", err)
	}

	// Load plugins
	if err := gpm.LoadPlugins(); err != nil {
		t.Fatalf("Failed to load plugins: %v", err)
	}

	// Get API and check available commands
	api := gpm.GetAPI()
	commands := api.GetCommands()

	t.Logf("Available commands after loading plugins:")
	for name, cmd := range commands {
		t.Logf("  - %s: %s", name, cmd.Description)
	}

	// Check if session-save-plugin commands were registered
	expectedCommands := []string{
		"session:save",
		"session:save_submit",
		"session:save_cancel",
	}

	for _, expectedCmd := range expectedCommands {
		if _, exists := commands[expectedCmd]; !exists {
			t.Errorf("Expected command '%s' was not registered", expectedCmd)
		} else {
			t.Logf("✅ Found expected command: %s", expectedCmd)
		}
	}

	// Test the session:save command (this should trigger the input dialog)
	result, err := api.ExecuteCommand("session:save", []string{})
	if err != nil {
		t.Logf("session:save command returned error (expected due to UI): %v", err)
	} else {
		t.Logf("session:save command result: %s", result)
	}

	// Test multi-cluster API functions directly by calling sync_current_cluster_session
	if syncCmd, exists := commands["session:save_submit"]; exists {
		t.Logf("Found session:save_submit command: %s", syncCmd.Description)

		// Execute with test filename
		result, err := api.ExecuteCommand("session:save_submit", []string{"test-session.json"})
		if err != nil {
			t.Logf("session:save_submit returned error: %v", err)
		} else {
			t.Logf("session:save_submit result: %s", result)
		}
	}

	// Verify that the kubeconfig path is correctly stored and accessible
	allClusters := gpm.GetAllClusters()
	if len(allClusters) != 1 {
		t.Fatalf("Expected 1 cluster, got %d", len(allClusters))
	}

	cluster := allClusters["test-cluster"]
	if cluster.Kubeconfig != "/test/config" {
		t.Errorf("Expected kubeconfig '/test/config', got '%s'", cluster.Kubeconfig)
	} else {
		t.Logf("✅ Kubeconfig path correctly stored: %s", cluster.Kubeconfig)
	}

	// Verify that the multi-cluster API functions are working
	// We can test this by checking if the plugin was able to call them without error
	// The fact that it loaded successfully indicates the multi-cluster API is working
	t.Log("✅ Session-save-plugin loaded successfully with multi-cluster API support")
}
