package plugins

import (
	"os"
	"testing"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

func TestSessionSavePlugin_LoadSession(t *testing.T) {
	// Create GlobalPluginManager and load plugins
	gpm := NewGlobalPluginManager("/home/otavio/dev/k8s-tui/plugins")

	// Add a test cluster
	testClient := k8s.Client{KubeconfigPath: "/test/config"}
	gpm.AddCluster("test-cluster", "Test Cluster", testClient)

	// Load plugins first
	if err := gpm.LoadPlugins(); err != nil {
		t.Fatalf("Failed to load plugins: %v", err)
	}

	// Switch to test cluster
	if err := gpm.SwitchToCluster("test-cluster"); err != nil {
		t.Fatalf("Failed to switch to test cluster: %v", err)
	}

	// Get API and test loading session
	api := gpm.GetAPI()

	// Test loading the test session file via CLI argument
	cliArgs := api.GetCLIArguments()
	if len(cliArgs) == 0 {
		t.Fatal("No CLI arguments registered")
	}

	// Check if session CLI argument is registered
	sessionArg, exists := cliArgs["session"]
	if !exists {
		t.Fatal("Session CLI argument not registered")
	}

	t.Logf("Found session CLI argument: %s - %s", sessionArg.Name, sessionArg.Description)

	// Check if session file exists
	if _, err := os.Stat("/tmp/minimal-session.json"); os.IsNotExist(err) {
		t.Fatalf("Session file does not exist: /tmp/minimal-session.json")
	}

	// Test with multi-cluster format only
	t.Log("Testing with multi-cluster session format...")
	content, err := os.ReadFile("/tmp/minimal-test.json")
	if err != nil {
		t.Fatalf("Failed to read session file: %v", err)
	}
	t.Logf("Multi-cluster session file content: %s", string(content))

	err = api.ExecuteCLIArgument("session", "/tmp/minimal-test.json")
	if err != nil {
		t.Fatalf("Failed to execute session CLI argument: %v", err)
	}

	t.Log("✅ Multi-cluster format succeeded!")
}
