package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/types"
)

// PluginManager manages the loading and lifecycle of plugins
type PluginManager struct {
	registry      *PluginRegistry
	pluginDir     string
	loadedPlugins map[string]*plugin.Plugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	return &PluginManager{
		registry:      NewPluginRegistry(),
		pluginDir:     pluginDir,
		loadedPlugins: make(map[string]*plugin.Plugin),
	}
}

// LoadPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	if pm.pluginDir == "" {
		return nil // No plugin directory specified
	}

	// Check if plugin directory exists
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		return nil // Plugin directory doesn't exist, skip loading
	}

	// Find all .so files in the plugin directory
	files, err := filepath.Glob(filepath.Join(pm.pluginDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to scan plugin directory: %v", err)
	}

	for _, file := range files {
		if err := pm.loadPlugin(file); err != nil {
			fmt.Printf("Failed to load plugin %s: %v\n", file, err)
			continue
		}
	}

	return nil
}

// loadPlugin loads a single plugin from a .so file
func (pm *PluginManager) loadPlugin(path string) error {
	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %v", err)
	}

	// Look for the plugin symbol
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin does not export Plugin symbol: %v", err)
	}

	// Assert that the symbol is a plugin
	pluginInstance, ok := sym.(Plugin)
	if !ok {
		return fmt.Errorf("plugin symbol is not a valid Plugin interface")
	}

	// Initialize the plugin
	if err := pluginInstance.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize plugin: %v", err)
	}

	// Register the plugin based on its type
	if resourcePlugin, ok := pluginInstance.(ResourcePlugin); ok {
		pm.registry.RegisterResourcePlugin(resourcePlugin)
		fmt.Printf("Loaded resource plugin: %s v%s\n", pluginInstance.Name(), pluginInstance.Version())
	}

	if uiPlugin, ok := pluginInstance.(UIPlugin); ok {
		pm.registry.RegisterUIPlugin(uiPlugin)
		fmt.Printf("Loaded UI plugin: %s v%s\n", pluginInstance.Name(), pluginInstance.Version())
	}

	// Store the loaded plugin
	pluginName := strings.TrimSuffix(filepath.Base(path), ".so")
	pm.loadedPlugins[pluginName] = p

	return nil
}

// GetRegistry returns the plugin registry
func (pm *PluginManager) GetRegistry() *PluginRegistry {
	return pm.registry
}

// GetCustomResourceData fetches data for a custom resource type from plugins
func (pm *PluginManager) GetCustomResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.GetResourceData(client, resourceType, namespace)
			}
		}
	}
	return nil, fmt.Errorf("custom resource type %s not found", resourceType)
}

// DeleteCustomResource deletes a custom resource using the appropriate plugin
func (pm *PluginManager) DeleteCustomResource(client k8s.Client, resourceType string, namespace string, name string) error {
	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.DeleteResource(client, resourceType, namespace, name)
			}
		}
	}
	return fmt.Errorf("custom resource type %s not found", resourceType)
}

// GetCustomResourceInfo gets information about a custom resource using the appropriate plugin
func (pm *PluginManager) GetCustomResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error) {
	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.GetResourceInfo(client, resourceType, namespace, name)
			}
		}
	}
	return nil, fmt.Errorf("custom resource type %s not found", resourceType)
}

// Shutdown shuts down all loaded plugins
func (pm *PluginManager) Shutdown() error {
	for name := range pm.loadedPlugins {
		// Find the plugin instance to call Shutdown
		for _, rp := range pm.registry.resourcePlugins {
			if rp.Name() == name {
				if err := rp.Shutdown(); err != nil {
					fmt.Printf("Error shutting down plugin %s: %v\n", name, err)
				}
				break
			}
		}
		for _, up := range pm.registry.uiPlugins {
			if up.Name() == name {
				if err := up.Shutdown(); err != nil {
					fmt.Printf("Error shutting down plugin %s: %v\n", name, err)
				}
				break
			}
		}
	}
	return nil
}
