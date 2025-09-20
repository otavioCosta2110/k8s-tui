package main

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/logger"
	"github.com/otavioCosta2110/k8s-tui/internal/plugins"
	"github.com/otavioCosta2110/k8s-tui/internal/ui"
	"os"
	"runtime/debug"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	logger.Debug("=== Application Starting ===")

	cfg := ui.ParseFlags()
	logger.Debug(fmt.Sprintf("Parsed flags: namespace=%s, kubeconfig=%s, pluginDir=%s", cfg.Namespace, cfg.KubeconfigPath, cfg.PluginDir))

	// Initialize plugin manager
	logger.Debug("Creating plugin manager")
	pluginManager := plugins.NewPluginManager(cfg.PluginDir)
	logger.Debug("Plugin manager created")

	if err := pluginManager.LoadPlugins(); err != nil {
		logger.Error(fmt.Sprintf("Plugin load error: %v", err))
	} else {
		logger.Info("Plugins loaded successfully")
	}

	// Set global plugin manager
	plugins.SetGlobalPluginManager(pluginManager)
	logger.Debug("Global plugin manager set")

	// Set up custom resource handlers
	logger.Debug("Setting up custom resource handlers")
	k8s.SetCustomResourceHandlers(
		pluginManager.GetCustomResourceData,
		pluginManager.DeleteCustomResource,
		pluginManager.GetCustomResourceInfo,
		func(resourceType string) bool {
			logger.Debug(fmt.Sprintf("Checking custom resource type: %s", resourceType))
			for _, rt := range pluginManager.GetRegistry().GetCustomResourceTypes() {
				logger.Debug(fmt.Sprintf("Available custom resource: %s (type: %s)", rt.Name, rt.Type))
				if rt.Type == resourceType {
					logger.Debug(fmt.Sprintf("Found matching custom resource: %s", rt.Name))
					return true
				}
			}
			logger.Debug(fmt.Sprintf("No matching custom resource found for: %s", resourceType))
			return false
		},
	)
	logger.Debug("Custom resource handlers set up")

	logger.Debug("Creating app model")
	m := ui.NewAppModel(cfg, pluginManager)
	logger.Debug("App model created")

	p := tea.NewProgram(m, tea.WithAltScreen())
	logger.Debug("Bubbletea program created")

	defer func() {
		logger.Debug("Entering defer function")
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Panic recovered: %v", r))
			debug.PrintStack()
		}
		// Shutdown plugins
		if err := pluginManager.Shutdown(); err != nil {
			logger.Error(fmt.Sprintf("Plugin shutdown error: %v", err))
		}
		logger.Info("Application shutting down")
		logger.Close()
	}()

	logger.Info("Starting Bubbletea program")
	if _, err := p.Run(); err != nil {
		logger.Error(fmt.Sprintf("Bubbletea program error: %v", err))
		os.Exit(1)
	}
	logger.Info("Application completed successfully")
}
