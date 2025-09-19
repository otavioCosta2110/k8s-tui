package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/plugins"
	"otaviocosta2110/k8s-tui/internal/ui"
	"otaviocosta2110/k8s-tui/utils"
	"runtime/debug"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	utils.WriteStringNewLine("debug.log", "=== Application Starting ===")

	cfg := ui.ParseFlags()
	utils.WriteStringNewLine("debug.log", fmt.Sprintf("Parsed flags: namespace=%s, kubeconfig=%s, pluginDir=%s", cfg.Namespace, cfg.KubeconfigPath, cfg.PluginDir))

	// Initialize plugin manager
	utils.WriteStringNewLine("debug.log", "Creating plugin manager")
	pluginManager := plugins.NewPluginManager(cfg.PluginDir)
	utils.WriteStringNewLine("debug.log", "Plugin manager created")

	if err := pluginManager.LoadPlugins(); err != nil {
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Plugin load error: %v", err))
		fmt.Printf("Failed to load plugins: %v\n", err)
	} else {
		utils.WriteStringNewLine("debug.log", "Plugins loaded successfully")
	}

	// Set global plugin manager
	plugins.SetGlobalPluginManager(pluginManager)
	utils.WriteStringNewLine("debug.log", "Global plugin manager set")

	// Set up custom resource handlers
	utils.WriteStringNewLine("debug.log", "Setting up custom resource handlers")
	k8s.SetCustomResourceHandlers(
		pluginManager.GetCustomResourceData,
		pluginManager.DeleteCustomResource,
		pluginManager.GetCustomResourceInfo,
		func(resourceType string) bool {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Checking custom resource type: %s", resourceType))
			for _, rt := range pluginManager.GetRegistry().GetCustomResourceTypes() {
				utils.WriteStringNewLine("debug.log", fmt.Sprintf("Available custom resource: %s (type: %s)", rt.Name, rt.Type))
				if rt.Type == resourceType {
					utils.WriteStringNewLine("debug.log", fmt.Sprintf("Found matching custom resource: %s", rt.Name))
					return true
				}
			}
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("No matching custom resource found for: %s", resourceType))
			return false
		},
	)
	utils.WriteStringNewLine("debug.log", "Custom resource handlers set up")

	utils.WriteStringNewLine("debug.log", "Creating app model")
	m := ui.NewAppModel(cfg, pluginManager)
	utils.WriteStringNewLine("debug.log", "App model created")

	p := tea.NewProgram(m, tea.WithAltScreen())
	utils.WriteStringNewLine("debug.log", "Bubbletea program created")

	defer func() {
		utils.WriteStringNewLine("debug.log", "Entering defer function")
		if r := recover(); r != nil {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Panic recovered: %v", r))
			fmt.Println("Recovered from panic:", r)
			debug.PrintStack()
		}
		// Shutdown plugins
		if err := pluginManager.Shutdown(); err != nil {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Plugin shutdown error: %v", err))
			fmt.Printf("Error shutting down plugins: %v\n", err)
		}
		utils.WriteStringNewLine("debug.log", "Application shutting down")
	}()

	utils.WriteStringNewLine("debug.log", "Starting Bubbletea program")
	if _, err := p.Run(); err != nil {
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Bubbletea program error: %v", err))
		os.Exit(1)
	}
	utils.WriteStringNewLine("debug.log", "Application completed successfully")
}
