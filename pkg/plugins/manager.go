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
	registry      *PluginRegistry
	pluginDir     string
	luaStates     map[string]*lua.LState
	api           *PluginAPIImpl
	neovimPlugins []NeovimStylePlugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	api := NewPluginAPI()
	return &PluginManager{
		registry:      NewPluginRegistry(),
		pluginDir:     pluginDir,
		luaStates:     make(map[string]*lua.LState),
		api:           api,
		neovimPlugins: make([]NeovimStylePlugin, 0),
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

	// Variables for Neovim-style detection
	var setupType, configType, commandsType, hooksType lua.LValueType
	var isNeovimStyle bool

	// Load the Lua script
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Loading Lua script: %s", path))
	if err := L.DoFile(path); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to execute Lua script %s: %v", path, err))
		return fmt.Errorf("failed to load Lua script: %v", err)
	}
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Successfully loaded Lua script: %s", path))

	// Debug: Check what functions are available
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Available functions in %s:", pluginName))
	for _, funcName := range []string{"Name", "Version", "Description", "Initialize", "Setup", "Config", "Commands", "Hooks", "GetResourceTypes", "GetUIExtensions"} {
		funcType := L.GetGlobal(funcName).Type()
		if funcType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Plugin Manager:   %s: FUNCTION", funcName))
		} else {
			logger.Info(fmt.Sprintf("🔌 Plugin Manager:   %s: %s", funcName, funcType))
		}
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Validating required functions for plugin: %s", pluginName))

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

	// Check if this is a Neovim-style plugin
	setupType = L.GetGlobal("Setup").Type()
	configType = L.GetGlobal("Config").Type()
	commandsType = L.GetGlobal("Commands").Type()
	hooksType = L.GetGlobal("Hooks").Type()

	isNeovimStyle = setupType == lua.LTFunction ||
		configType == lua.LTFunction ||
		commandsType == lua.LTFunction ||
		hooksType == lua.LTFunction

	// For Neovim-style plugins, set up the k8s_tui API before initialization
	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected Neovim-style plugin: %s", pluginName))
		logger.Info("🔌 Plugin Manager: Setting up k8s_tui API for Neovim-style plugin")

		// Create API table
		apiTable := L.NewTable()

		// Add API functions
		L.SetField(apiTable, "get_namespace", L.NewFunction(func(L *lua.LState) int {
			namespace := pm.api.GetCurrentNamespace()
			L.Push(lua.LString(namespace))
			return 1
		}))
		L.SetField(apiTable, "set_status", L.NewFunction(func(L *lua.LState) int {
			message := L.CheckString(1)
			pm.api.SetStatusMessage(message)
			return 0
		}))
		L.SetField(apiTable, "add_header", L.NewFunction(func(L *lua.LState) int {
			content := L.CheckString(1)
			component := UIInjectionPoint{
				Location: "header",
				Position: "right",
				Priority: 10,
				Component: DisplayComponent{
					Type: "text",
					Config: map[string]any{
						"content": content,
						"style":   "info",
					},
				},
				DataSource:     "static",
				UpdateInterval: 0,
			}
			pm.api.AddHeaderComponent(component)
			return 0
		}))
		L.SetField(apiTable, "register_command", L.NewFunction(func(L *lua.LState) int {
			name := L.CheckString(1)
			description := L.CheckString(2)
			command := PluginCommand{
				Name:        name,
				Description: description,
				Handler: func(args []string) (string, error) {
					return "Command executed from Lua", nil
				},
			}
			pm.api.RegisterCommand(command.Name, command.Description, command.Handler)
			return 0
		}))

		// Set the API in the global environment
		L.SetGlobal("k8s_tui", apiTable)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: k8s_tui API set up for plugin: %s", pluginName))
	}

	// Create a LuaPlugin wrapper
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Creating plugin wrapper for: %s", pluginName))
	luaPlugin := &LuaPlugin{
		L:          L,
		pluginName: pluginName,
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

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Function types for %s - Setup: %s, Config: %s, Commands: %s, Hooks: %s",
		pluginName, setupType, configType, commandsType, hooksType))

	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected Neovim-style plugin: %s", pluginDisplayName))

		// Create Neovim-style plugin wrapper
		neovimPlugin := NewNeovimStyleLuaPlugin(L, pluginName, pm.api)

		// Setup the plugin with default config
		defaultConfig := neovimPlugin.Config()
		if err := neovimPlugin.Setup(defaultConfig); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to setup Neovim-style plugin %s: %v", pluginDisplayName, err))
			L.Close()
			return fmt.Errorf("failed to setup Neovim-style plugin: %v", err)
		}

		// Register commands
		commands := neovimPlugin.Commands()
		for _, cmd := range commands {
			pm.api.RegisterCommand(cmd.Name, cmd.Description, cmd.Handler)
		}

		// Register hooks
		hooks := neovimPlugin.Hooks()
		for _, hook := range hooks {
			pm.api.RegisterEventHandler(PluginEvent(hook.Event), hook.Handler)
		}

		pm.neovimPlugins = append(pm.neovimPlugins, neovimPlugin)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Registered Neovim-style plugin: %s v%s", pluginDisplayName, pluginVersion))

		// Also register as legacy plugin for backward compatibility
		hasResourcePlugin := luaPlugin.hasResourcePlugin()
		hasUIPlugin := luaPlugin.hasUIPlugin()

		if hasResourcePlugin {
			pm.registry.RegisterResourcePlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Also registered as legacy resource plugin: %s", pluginDisplayName))
		}

		if hasUIPlugin {
			pm.registry.RegisterUIPlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎨 Also registered as legacy UI plugin: %s", pluginDisplayName))
		}
	} else {
		// Handle legacy plugins (deprecated)
		logger.Warn(fmt.Sprintf("🔌 Plugin Manager: ⚠️  Legacy plugin detected: %s - Consider migrating to Neovim-style", pluginDisplayName))

		// Check plugin capabilities
		hasResourcePlugin := luaPlugin.hasResourcePlugin()
		hasUIPlugin := luaPlugin.hasUIPlugin()

		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s capabilities - Resource: %t, UI: %t", pluginDisplayName, hasResourcePlugin, hasUIPlugin))

		// Register based on capabilities
		if hasResourcePlugin {
			pm.registry.RegisterResourcePlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Registered legacy resource plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))

			// Log resource types
			resourceTypes := luaPlugin.GetResourceTypes()
			for _, rt := range resourceTypes {
				logger.Info(fmt.Sprintf("🔌 Plugin Manager:   └─ Resource type: %s (%s)", rt.Name, rt.Type))
			}
		}

		if hasUIPlugin {
			pm.registry.RegisterUIPlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎨 Registered legacy UI plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))
		}
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

// GetAPI returns the plugin API
func (pm *PluginManager) GetAPI() *PluginAPIImpl {
	return pm.api
}

// GetNeovimPlugins returns all loaded Neovim-style plugins
func (pm *PluginManager) GetNeovimPlugins() []NeovimStylePlugin {
	return pm.neovimPlugins
}

// TriggerEvent triggers an event for all registered plugins
func (pm *PluginManager) TriggerEvent(event PluginEvent, data interface{}) {
	pm.api.TriggerEvent(event, data)
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
