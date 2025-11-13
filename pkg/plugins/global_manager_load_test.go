package plugins

import (
	"os"
	"path/filepath"
	"testing"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

func TestGlobalPluginManager_LoadPlugins(t *testing.T) {
	// Create a temporary plugin directory
	tempDir, err := os.MkdirTemp("", "test-plugins-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test plugin directory
	testPluginDir := filepath.Join(tempDir, "session-save-plugin")
	if err := os.MkdirAll(testPluginDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Copy the session-save-plugin main.lua to test directory
	pluginContent := `
function Name()
  return "session-save-plugin"
end

function Version()
  return "1.0.0"
end

function Description()
  return "Test plugin for multi-cluster API"
end

function Initialize()
  return nil
end

function Setup(opts)
  return nil
end

function Config()
  return {
    enabled = true
  }
end

function Commands()
  return {
    {
      name = "test:sync",
      description = "Test sync function",
      handler = "test_sync"
    }
  }
end

function test_sync()
  if not k8s_tui then
    return "ERROR: k8s_tui not available"
  end

  local result = {}
  
  -- Test multi-cluster functions
  if k8s_tui.get_all_clusters then
    local clusters = k8s_tui.get_all_clusters()
    result.clusters_available = true
    result.cluster_count = #clusters
  else
    result.clusters_available = false
  end

  if k8s_tui.get_current_cluster then
    local current = k8s_tui.get_current_cluster()
    result.current_available = true
    result.current_nil = (current == nil)
  else
    result.current_available = false
  end

  if k8s_tui.sync_current_cluster_session then
    local sync_result = k8s_tui.sync_current_cluster_session()
    result.sync_available = true
    result.sync_result = sync_result
  else
    result.sync_available = false
  end

  return "Multi-cluster API test completed: " .. tostring(result.clusters_available) .. ", " .. tostring(result.current_available) .. ", " .. tostring(result.sync_available)
end
`

	pluginFile := filepath.Join(testPluginDir, "main.lua")
	if err := os.WriteFile(pluginFile, []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to write plugin file: %v", err)
	}

	// Create GlobalPluginManager and load plugins
	gpm := NewGlobalPluginManager(tempDir)

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

	// Test that the plugin was loaded by checking if we can access its commands
	api := gpm.GetAPI()
	commands := api.GetCommands()

	t.Logf("Available commands after loading plugins:")
	for name, cmd := range commands {
		t.Logf("  - %s: %s", name, cmd.Description)
	}

	if len(commands) == 0 {
		t.Fatal("No commands were registered by the plugin")
	}

	// Check if our test command was registered
	testCmd, exists := commands["test:sync"]
	if !exists {
		t.Fatal("test:sync command was not registered")
	}

	if testCmd.Name != "test:sync" {
		t.Errorf("Expected command name 'test:sync', got '%s'", testCmd.Name)
	}

	// Execute the test command to verify multi-cluster API functions work
	result, err := api.ExecuteCommand("test:sync", []string{})
	if err != nil {
		t.Fatalf("Failed to execute test command: %v", err)
	}

	expectedResult := "Multi-cluster API test completed: true, true, true"
	if result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, result)
	}

	t.Logf("Command result: %s", result)
}
