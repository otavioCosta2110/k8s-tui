package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"github.com/yuin/gopher-lua"
)

// PluginManager manages the loading and lifecycle of plugins
type PluginManager struct {
	registry  *PluginRegistry
	pluginDir string
	luaStates map[string]*lua.LState
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	return &PluginManager{
		registry:  NewPluginRegistry(),
		pluginDir: pluginDir,
		luaStates: make(map[string]*lua.LState),
	}
}

// LoadPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Starting plugin loading from directory: %s", pm.pluginDir))

	if pm.pluginDir == "" {
		logger.Info("🔌 Plugin Manager: No plugin directory specified, skipping plugin loading")
		return nil
	}

	// Check if plugin directory exists
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin directory does not exist: %s", pm.pluginDir))
		return nil
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Scanning for Lua plugins in: %s", pm.pluginDir))

	// Find all .lua files in the plugin directory and subdirectories
	var files []string
	err := filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Error scanning directory %s: %v", path, err))
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") {
			files = append(files, path)
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Found potential plugin file: %s", path))
		}
		return nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to scan plugin directory: %v", err))
		return fmt.Errorf("failed to scan plugin directory: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Found %d potential plugin files", len(files)))

	loadedCount := 0
	failedCount := 0

	for _, file := range files {
		pluginName := strings.TrimSuffix(filepath.Base(file), ".lua")
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: Attempting to load plugin: %s from %s", pluginName, file))

		if err := pm.loadLuaPlugin(file); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: ❌ Failed to load plugin %s: %v", pluginName, err))
			failedCount++
			continue
		}

		logger.Info(fmt.Sprintf("🔌 Plugin Manager: ✅ Successfully loaded plugin: %s", pluginName))
		loadedCount++
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin loading complete - %d loaded, %d failed", loadedCount, failedCount))

	return nil
}

// loadLuaPlugin loads a single plugin from a .lua file
func (pm *PluginManager) loadLuaPlugin(path string) error {
	pluginName := strings.TrimSuffix(filepath.Base(path), ".lua")

	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Creating Lua state for plugin: %s", pluginName))
	L := lua.NewState()

	// Load the Lua script
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Loading Lua script: %s", path))
	if err := L.DoFile(path); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to execute Lua script %s: %v", path, err))
		return fmt.Errorf("failed to load Lua script: %v", err)
	}

	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Validating required functions for plugin: %s", pluginName))

	// Check if required functions exist
	if L.GetGlobal("Name").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s missing required Name() function", pluginName))
		return fmt.Errorf("Lua plugin must define a Name function")
	}
	if L.GetGlobal("Initialize").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s missing required Initialize() function", pluginName))
		return fmt.Errorf("Lua plugin must define an Initialize function")
	}

	// Create a LuaPlugin wrapper
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Creating plugin wrapper for: %s", pluginName))
	luaPlugin := &LuaPlugin{
		L: L,
	}

	// Get plugin metadata for logging
	pluginDisplayName := luaPlugin.Name()
	pluginVersion := luaPlugin.Version()
	pluginDescription := luaPlugin.Description()

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Initializing plugin %s (%s v%s)", pluginDisplayName, pluginName, pluginVersion))

	// Initialize the plugin
	if err := luaPlugin.Initialize(); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialization failed: %v", pluginDisplayName, err))
		return fmt.Errorf("failed to initialize Lua plugin: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialized successfully", pluginDisplayName))

	// Check plugin capabilities
	hasResourcePlugin := luaPlugin.hasResourcePlugin()
	hasUIPlugin := luaPlugin.hasUIPlugin()

	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s capabilities - Resource: %t, UI: %t", pluginDisplayName, hasResourcePlugin, hasUIPlugin))

	// Register based on capabilities
	if hasResourcePlugin {
		pm.registry.RegisterResourcePlugin(luaPlugin)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Registered resource plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))

		// Log resource types
		resourceTypes := luaPlugin.GetResourceTypes()
		for _, rt := range resourceTypes {
			logger.Info(fmt.Sprintf("🔌 Plugin Manager:   └─ Resource type: %s (%s)", rt.Name, rt.Type))
		}
	}

	if hasUIPlugin {
		pm.registry.RegisterUIPlugin(luaPlugin)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎨 Registered UI plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))
	}

	// Store the Lua state
	pm.luaStates[pluginName] = L

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎉 Plugin %s loaded and registered successfully", pluginDisplayName))

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
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Shutting down %d loaded plugins", len(pm.luaStates)))

	shutdownCount := 0
	errorCount := 0

	for name, L := range pm.luaStates {
		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Shutting down plugin: %s", name))

		// Call Shutdown if defined
		if L.GetGlobal("Shutdown").Type() == lua.LTFunction {
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Calling Shutdown() for plugin: %s", name))
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("Shutdown"),
				NRet:    1,
				Protect: true,
			}); err != nil {
				logger.Error(fmt.Sprintf("🔌 Plugin Manager: Error calling Shutdown() for plugin %s: %v", name, err))
				errorCount++
			} else {
				// Check for error return
				ret := L.Get(-1)
				L.Pop(1)
				if ret.Type() == lua.LTString {
					logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s shutdown returned error: %s", name, ret.String()))
					errorCount++
				} else {
					logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s shutdown completed successfully", name))
					shutdownCount++
				}
			}
		} else {
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s has no Shutdown() function, skipping", name))
		}

		L.Close()
		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Lua state closed for plugin: %s", name))
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Shutdown complete - %d plugins shut down, %d errors", shutdownCount, errorCount))

	return nil
}
