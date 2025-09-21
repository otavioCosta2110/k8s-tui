package main

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"os"
	"runtime/debug"
)

func main() {
	cfg := ui.ParseFlags()
	pluginManager := plugins.NewPluginManager(cfg.PluginDir)
	if err := pluginManager.LoadPlugins(); err != nil {
		logger.Error(fmt.Sprintf("Plugin load error: %v", err))
	} else {
		logger.Info("Plugins loaded successfully")
	}
	plugins.SetGlobalPluginManager(pluginManager)
	pluginManager.TriggerEvent(plugins.EventAppStarted, "k8s-tui started")
	resources.SetCustomResourceHandlers(
		pluginManager.GetCustomResourceData,
		pluginManager.DeleteCustomResource,
		pluginManager.GetCustomResourceInfo,
		func(resourceType string) bool {
			for _, rt := range pluginManager.GetRegistry().GetCustomResourceTypes() {
				if rt.Type == resourceType {
					return true
				}
			}
			return false
		},
	)
	m := ui.NewAppModel(cfg, pluginManager)
	p := tea.NewProgram(m, tea.WithAltScreen())
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Panic recovered: %v", r))
			debug.PrintStack()
		}
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
