package plugins

import (
	"testing"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

func TestPluginAPIImpl_BasicOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test namespace operations
	if api.GetCurrentNamespace() != "default" {
		t.Error("Expected default namespace to be 'default'")
	}

	api.SetCurrentNamespace("test-namespace")
	if api.GetCurrentNamespace() != "test-namespace" {
		t.Error("Expected namespace to be updated")
	}

	// Test config operations
	api.SetConfig("test-key", "test-value")
	value := api.GetConfig("test-key")
	if value != "test-value" {
		t.Error("Expected config value to be set correctly")
	}

	// Test resource type operations
	if api.GetCurrentResourceType() != "" {
		t.Error("Expected current resource type to be empty initially")
	}

	api.SetCurrentResourceType("pods")
	if api.GetCurrentResourceType() != "pods" {
		t.Error("Expected current resource type to be updated")
	}
}

func TestPluginAPIImpl_CommandOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test command registration
	called := false
	api.RegisterCommand("test-cmd", "Test command", func(args []string) (string, error) {
		called = true
		return "executed", nil
	})

	commands := api.GetCommands()
	if len(commands) != 1 {
		t.Error("Expected 1 command to be registered")
	}

	cmd, exists := commands["test-cmd"]
	if !exists {
		t.Error("Expected test-cmd to be registered")
	}

	if cmd.Description != "Test command" {
		t.Error("Expected command description to match")
	}

	// Test command execution
	result, err := api.ExecuteCommand("test-cmd", []string{})
	if err != nil {
		t.Errorf("Expected no error executing command, got: %v", err)
	}

	if result != "executed" {
		t.Error("Expected command to return 'executed'")
	}

	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestPluginAPIImpl_CLIArguments(t *testing.T) {
	api := NewPluginAPI()

	// Test CLI argument registration
	called := false
	api.RegisterCLIArgument("test-arg", "Test argument", func(value string) error {
		called = true
		return nil
	})

	arguments := api.GetCLIArguments()
	if len(arguments) != 1 {
		t.Error("Expected 1 argument to be registered")
	}

	arg, exists := arguments["test-arg"]
	if !exists {
		t.Error("Expected test-arg to be registered")
	}

	if arg.Description != "Test argument" {
		t.Error("Expected argument description to match")
	}

	if !api.HasCLIArgument("test-arg") {
		t.Error("Expected HasCLIArgument to return true")
	}

	// Test argument execution
	err := api.ExecuteCLIArgument("test-arg", "test-value")
	if err != nil {
		t.Errorf("Expected no error executing argument, got: %v", err)
	}

	if !called {
		t.Error("Expected argument handler to be called")
	}
}

func TestPluginAPIImpl_UIComponents(t *testing.T) {
	api := NewPluginAPI()

	// Test header components
	if len(api.GetHeaderComponents()) != 0 {
		t.Error("Expected no header components initially")
	}

	component := UIInjectionPoint{
		Location: "header",
		Component: DisplayComponent{
			Type: "text",
			Config: map[string]interface{}{
				"content": "test content",
			},
		},
	}

	api.AddHeaderComponent(component)
	components := api.GetHeaderComponents()
	if len(components) != 1 {
		t.Error("Expected 1 header component to be added")
	}

	// Test footer components
	api.AddFooterComponent(component)
	footerComponents := api.GetFooterComponents()
	if len(footerComponents) != 1 {
		t.Error("Expected 1 footer component to be added")
	}
}

func TestPluginAPIImpl_ClientOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test client operations
	client := MockClient()
	api.SetClient(client)

	retrievedClient := api.GetClient()
	if retrievedClient.Namespace != client.Namespace {
		t.Error("Expected client to be set and retrieved correctly")
	}
}

func TestPluginAPIImpl_HelpOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test getting help for existing resource type
	title, content := api.GetHelp("Pods")
	if title == "" {
		t.Error("Expected help title to be non-empty")
	}

	if content == "" {
		t.Error("Expected help content to be non-empty")
	}

	// Test registering custom help
	api.RegisterHelp("CustomResource", "Custom Title", "Custom Content")
	title, content = api.GetHelp("CustomResource")

	if title != "Custom Title" {
		t.Error("Expected custom help title to match")
	}

	if content != "Custom Content" {
		t.Error("Expected custom help content to match")
	}

	// Test getting help for non-existent resource type
	title, content = api.GetHelp("NonExistent")
	if title != "Help" {
		t.Error("Expected default help title")
	}
}

func TestPluginAPIImpl_EventOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test event registration and triggering
	eventTriggered := false
	api.RegisterEventHandler(EventNamespaceChanged, func(data interface{}) error {
		eventTriggered = true
		return nil
	})

	// Trigger event
	api.TriggerEvent(EventNamespaceChanged, "test-data")

	if !eventTriggered {
		t.Error("Expected event handler to be triggered")
	}
}

func TestPluginAPIImpl_TabOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test tab operations without callbacks (should return errors)
	tabs, err := api.GetTabs()
	if err == nil {
		t.Error("Expected error when getting tabs without callback")
	}

	if tabs != nil {
		t.Error("Expected nil tabs when callback not set")
	}

	testTabs := []TabInfo{
		{ID: "tab1", Title: "Test Tab 1", Namespace: "default"},
		{ID: "tab2", Title: "Test Tab 2", Namespace: "default"},
	}

	err = api.SetTabs(testTabs)
	if err == nil {
		t.Error("Expected error when setting tabs without callback")
	}
}

func TestPluginAPIImpl_BreadcrumbOperations(t *testing.T) {
	api := NewPluginAPI()

	// Test breadcrumb operations without callbacks
	breadcrumb := api.GetBreadcrumbTrail()
	if len(breadcrumb) != 0 {
		t.Error("Expected empty breadcrumb trail when no callback set")
	}

	// Set breadcrumb (should not panic)
	api.SetBreadcrumbTrail([]string{"level1", "level2"})

	// Should still return empty since no callback is set
	breadcrumb = api.GetBreadcrumbTrail()
	if len(breadcrumb) != 0 {
		t.Error("Expected empty breadcrumb trail when no callback set")
	}
}

func TestPluginAPIImpl_ResourceHandlers(t *testing.T) {
	api := NewPluginAPI()

	// Test getting supported resource types (should have defaults)
	types := api.GetSupportedResourceTypes()
	if len(types) == 0 {
		t.Error("Expected at least one supported resource type")
	}

	// Test getting existing resource handler
	retrievedHandler, exists := api.GetResourceHandler(k8s.ResourceTypePod)
	if !exists {
		t.Error("Expected pod handler to exist")
	}

	if retrievedHandler == nil {
		t.Error("Expected handler to be non-nil")
	}

	// Test resource handler type
	if retrievedHandler.GetType() != k8s.ResourceTypePod {
		t.Error("Expected handler type to be pod")
	}
}
