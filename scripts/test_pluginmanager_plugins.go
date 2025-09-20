package main

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

func main() {
	pm := plugins.NewPluginManager("./plugins")
	err := pm.LoadPlugins()
	if err != nil {
		fmt.Printf("Error loading plugins: %v\n", err)
		return
	}

	fmt.Printf("Loaded plugins successfully\n")

	// Test pluginmanager-style plugins
	pluginmanagerPlugins := pm.GetPluginmanagerPlugins()
	fmt.Printf("Found %d pluginmanager-style plugins:\n", len(pluginmanagerPlugins))

	for i, plugin := range pluginmanagerPlugins {
		fmt.Printf("  %d. %s v%s - %s\n", i+1, plugin.Name(), plugin.Version(), plugin.Description())

		// Test commands
		commands := plugin.Commands()
		if len(commands) > 0 {
			fmt.Printf("     Commands:\n")
			for _, cmd := range commands {
				fmt.Printf("       - %s: %s\n", cmd.Name, cmd.Description)
			}
		}

		// Test hooks
		hooks := plugin.Hooks()
		if len(hooks) > 0 {
			fmt.Printf("     Hooks:\n")
			for _, hook := range hooks {
				fmt.Printf("       - %s\n", hook.Event)
			}
		}

		// Test config
		config := plugin.Config()
		if len(config) > 0 {
			fmt.Printf("     Config:\n")
			for k, v := range config {
				fmt.Printf("       - %s = %v\n", k, v)
			}
		}
	}

	// Test API
	api := pm.GetAPI()
	fmt.Printf("\nPlugin API Status:\n")
	fmt.Printf("  Current namespace: %s\n", api.GetCurrentNamespace())
	fmt.Printf("  Header components: %d\n", len(api.GetHeaderComponents()))
	fmt.Printf("  Footer components: %d\n", len(api.GetFooterComponents()))
	fmt.Printf("  Registered commands: %d\n", len(api.GetCommands()))
}
