package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/yuin/gopher-lua"
)

// ClusterContext holds cluster-specific information for plugins
type ClusterContext struct {
	ID         string
	Name       string
	Client     k8s.Client
	Settings   map[string]interface{}
	Namespace  string
	Kubeconfig string
	Tabs       []TabInfo
	Breadcrumb []string
}

// GlobalPluginManager manages a single plugin instance across multiple clusters
type GlobalPluginManager struct {
	*PluginManager // Embed the existing plugin manager
	mu             sync.RWMutex
	clusters       map[string]*ClusterContext // cluster ID -> context
	currentCluster string                     // current active cluster ID
}

// NewGlobalPluginManager creates a new global plugin manager
func NewGlobalPluginManager(pluginDir string) *GlobalPluginManager {
	pm := NewPluginManager(pluginDir)

	gpm := &GlobalPluginManager{
		PluginManager: pm,
		clusters:      make(map[string]*ClusterContext),
	}

	// Set the global manager reference in the API
	pm.SetGlobalManagerReference(gpm)

	return gpm
}

// LoadPlugins loads plugins using multi-cluster API
func (gpm *GlobalPluginManager) LoadPlugins() error {
	if gpm.pluginDir == "" {
		logger.Info("🔌 Global Plugin Manager: No plugin directory specified, skipping plugin loading")
		return nil
	}

	if _, err := os.Stat(gpm.pluginDir); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Plugin directory does not exist: %s", gpm.pluginDir))
		return nil
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Scanning for Lua plugins in: %s", gpm.pluginDir))

	var files []string
	err := filepath.Walk(gpm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error scanning directory %s: %v", path, err))
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") {
			files = append(files, path)
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Found potential plugin file: %s", path))
		}
		return nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Failed to scan plugin directory: %v", err))
		return fmt.Errorf("failed to scan plugin directory: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Found %d potential plugin files", len(files)))

	loadedCount := 0
	failedCount := 0

	for _, file := range files {
		pluginName := filepath.Base(filepath.Dir(file))
		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Attempting to load plugin: %s from %s", pluginName, file))

		if err := gpm.loadLuaPluginWithMultiClusterAPI(file); err != nil {
			logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: ❌ Failed to load plugin %s: %v", pluginName, err))
			failedCount++
			continue
		}

		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: ✅ Successfully loaded plugin: %s", pluginName))
		loadedCount++
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Plugin loading complete - %d loaded, %d failed", loadedCount, failedCount))
	return nil
}

// loadLuaPluginWithMultiClusterAPI loads a plugin with multi-cluster API support
func (gpm *GlobalPluginManager) loadLuaPluginWithMultiClusterAPI(path string) error {
	pluginName := filepath.Base(filepath.Dir(path))

	logger.Debug(fmt.Sprintf("🔌 Global Plugin Manager: Creating Lua state for plugin: %s", pluginName))
	L := lua.NewState()

	content, err := os.ReadFile(path)
	if err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Failed to read Lua script %s: %v", path, err))
		return fmt.Errorf("failed to read Lua script: %v", err)
	}

	if err := L.DoString(string(content)); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Failed to execute Lua script %s: %v", path, err))
		return fmt.Errorf("failed to load Lua script: %v", err)
	}

	// Check if this is a pluginmanager-style plugin
	setupType := L.GetGlobal("Setup").Type()
	configType := L.GetGlobal("Config").Type()
	commandsType := L.GetGlobal("Commands").Type()
	hooksType := L.GetGlobal("Hooks").Type()

	isPluginmanagerStyle := setupType == lua.LTFunction ||
		configType == lua.LTFunction ||
		commandsType == lua.LTFunction ||
		hooksType == lua.LTFunction

	if isPluginmanagerStyle {
		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: 🎯 Detected pluginmanager-style plugin: %s", pluginName))
		logger.Info("🔌 Global Plugin Manager: Setting up multi-cluster k8s_tui API for pluginmanager-style plugin")

		// Set up multi-cluster Lua API for pluginmanager-style plugins
		gpm.setupMultiClusterLuaAPI(L)

		// Call Setup function if available
		if setupType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Calling Setup() for plugin: %s", pluginName))
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("Setup"),
				NRet:    1,
				Protect: true,
			}); err != nil {
				logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error calling Setup() for plugin %s: %v", pluginName, err))
				return fmt.Errorf("error calling Setup: %v", err)
			}

			// Check if Setup returned an error
			ret := L.Get(-1)
			L.Pop(1)
			if ret.Type() == lua.LTString {
				errorMsg := ret.String()
				logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Plugin %s Setup() returned error: %s", pluginName, errorMsg))
				return fmt.Errorf("%s", errorMsg)
			}
		}

		// Register commands if available
		if commandsType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Registering commands for plugin: %s", pluginName))
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("Commands"),
				NRet:    1,
				Protect: true,
			}); err != nil {
				logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error calling Commands() for plugin %s: %v", pluginName, err))
				return fmt.Errorf("error calling Commands: %v", err)
			}

			// Get the commands table
			commandsRet := L.Get(-1)
			L.Pop(1)

			if commandsRet.Type() == lua.LTTable {
				commandsTable := commandsRet.(*lua.LTable)
				commandsTable.ForEach(func(key, value lua.LValue) {
					if value.Type() == lua.LTTable {
						cmdTable := value.(*lua.LTable)
						cmdName := getStringField(cmdTable, "name")
						cmdDesc := getStringField(cmdTable, "description")
						handlerName := getStringField(cmdTable, "handler")

						if cmdName != "" && handlerName != "" {
							// Create a command handler that calls the Lua function
							handler := func(args []string) (string, error) {
								if L.GetGlobal(handlerName).Type() == lua.LTFunction {
									if err := L.CallByParam(lua.P{
										Fn:      L.GetGlobal(handlerName),
										NRet:    1,
										Protect: true,
									}); err != nil {
										return "", fmt.Errorf("error calling Lua handler %s: %v", handlerName, err)
									}

									ret := L.Get(-1)
									L.Pop(1)
									if ret.Type() == lua.LTString {
										return ret.String(), nil
									}
									return "Command executed", nil
								}
								return "", fmt.Errorf("handler function %s not found", handlerName)
							}

							gpm.api.RegisterCommand(cmdName, cmdDesc, handler)
							logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Registered command: %s", cmdName))
						}
					}
				})
			}
		}

		// Register CLI arguments if available
		cliArgsType := L.GetGlobal("CLIArguments").Type()
		if cliArgsType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Registering CLI arguments for plugin: %s", pluginName))
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("CLIArguments"),
				NRet:    1,
				Protect: true,
			}); err != nil {
				logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error calling CLIArguments() for plugin %s: %v", pluginName, err))
				return fmt.Errorf("error calling CLIArguments: %v", err)
			}

			// Get the CLI arguments table
			cliArgsRet := L.Get(-1)
			L.Pop(1)

			if cliArgsRet.Type() == lua.LTTable {
				cliArgsTable := cliArgsRet.(*lua.LTable)
				cliArgsTable.ForEach(func(key, value lua.LValue) {
					if value.Type() == lua.LTTable {
						argTable := value.(*lua.LTable)
						argName := getStringField(argTable, "name")
						argDesc := getStringField(argTable, "description")
						handlerName := getStringField(argTable, "handler")

						if argName != "" && handlerName != "" {
							// Create a CLI argument handler that calls the Lua function
							handler := func(argValue string) error {
								// Call the Lua handler function with the argument value
								if err := L.CallByParam(lua.P{
									Fn:      L.GetGlobal(handlerName),
									NRet:    2,
									Protect: true,
								}, lua.LString(argValue)); err != nil {
									return fmt.Errorf("error calling CLI handler %s: %v", handlerName, err)
								}

								// Check for errors (Lua functions can return 2 values: result, error)
								ret2 := L.Get(-1)
								_ = L.Get(-2) // First return value (result), ignore it
								L.Pop(2)

								if ret2.Type() == lua.LTString {
									// Second return value is an error
									return fmt.Errorf("%s", ret2.String())
								}

								// Success
								return nil
							}

							gpm.api.RegisterCLIArgument(argName, argDesc, handler)
							logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Registered CLI argument: %s", argName))
						}
					}
				})
			}
		}

		// Store Lua state for cleanup
		gpm.luaStates[pluginName] = L

		// Call Initialize function
		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Calling Initialize() for plugin: %s", pluginName))
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("Initialize"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error calling Initialize() for plugin %s: %v", pluginName, err))
			return fmt.Errorf("error calling Initialize: %v", err)
		}

		// Check if Initialize returned an error
		ret := L.Get(-1)
		L.Pop(1)
		if ret.Type() == lua.LTString {
			errorMsg := ret.String()
			logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Plugin %s Initialize() returned error: %s", pluginName, errorMsg))
			return fmt.Errorf("%s", errorMsg)
		}

		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Successfully initialized pluginmanager-style plugin: %s", pluginName))
		return nil
	}

	// Basic plugin style - set up multi-cluster Lua API
	gpm.setupMultiClusterLuaAPI(L)

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Available functions in %s:", pluginName))
	for _, funcName := range []string{"Name", "Version", "Description", "Initialize", "Setup", "Config", "Commands", "Hooks", "GetResourceTypes", "GetUIExtensions"} {
		funcType := L.GetGlobal(funcName).Type()
		if funcType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager:   %s: FUNCTION", funcName))
		} else {
			logger.Info(fmt.Sprintf("🔌 Global Plugin Manager:   %s: %s", funcName, funcType))
		}
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Validating required functions for plugin: %s", pluginName))

	if L.GetGlobal("Name").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Plugin %s missing required Name() function", pluginName))
		return fmt.Errorf("Lua plugin must define a Name function")
	}
	if L.GetGlobal("Initialize").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Plugin %s missing required Initialize() function", pluginName))
		return fmt.Errorf("Lua plugin must define an Initialize function")
	}

	// Store Lua state for cleanup
	gpm.luaStates[pluginName] = L

	// Call Initialize function
	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Calling Initialize() for plugin: %s", pluginName))
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("Initialize"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Error calling Initialize() for plugin %s: %v", pluginName, err))
		return fmt.Errorf("error calling Initialize: %v", err)
	}

	// Check if Initialize returned an error
	ret := L.Get(-1)
	L.Pop(1)
	if ret.Type() == lua.LTString {
		errorMsg := ret.String()
		logger.Error(fmt.Sprintf("🔌 Global Plugin Manager: Plugin %s Initialize() returned error: %s", pluginName, errorMsg))
		return fmt.Errorf("%s", errorMsg)
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Successfully initialized basic plugin: %s", pluginName))
	return nil
}

// setupMultiClusterLuaAPI configures the Lua API with multi-cluster support
func (gpm *GlobalPluginManager) setupMultiClusterLuaAPI(L *lua.LState) {
	apiTable := L.NewTable()

	// Basic API functions
	L.SetField(apiTable, "get_namespace", L.NewFunction(func(L *lua.LState) int {
		namespace := gpm.api.GetCurrentNamespace()
		L.Push(lua.LString(namespace))
		return 1
	}))
	L.SetField(apiTable, "set_namespace", L.NewFunction(func(L *lua.LState) int {
		namespace := L.CheckString(1)
		gpm.api.SetCurrentNamespace(namespace)
		return 0
	}))
	L.SetField(apiTable, "set_status", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		gpm.api.SetStatusMessage(message)
		return 0
	}))
	L.SetField(apiTable, "get_tabs", L.NewFunction(func(L *lua.LState) int {
		tabs, err := gpm.api.GetTabs()
		if err != nil {
			L.Push(lua.LString(fmt.Sprintf("failed to get tabs: %v", err)))
			return 1
		}

		resultTable := L.NewTable()
		for i, tab := range tabs {
			tabTable := L.NewTable()
			L.SetField(tabTable, "ID", lua.LString(tab.ID))
			L.SetField(tabTable, "Title", lua.LString(tab.Title))
			L.SetField(tabTable, "ResourceType", lua.LString(tab.ResourceType))
			L.SetField(tabTable, "CurrentIndex", lua.LNumber(tab.CurrentIndex))

			breadcrumbTable := L.NewTable()
			for j, crumb := range tab.Breadcrumb {
				L.RawSetInt(breadcrumbTable, j+1, lua.LString(crumb))
			}
			L.SetField(tabTable, "Breadcrumb", breadcrumbTable)

			metadataTable := L.NewTable()
			for k, v := range tab.Metadata {
				L.SetField(metadataTable, k, lua.LString(fmt.Sprintf("%v", v)))
			}
			L.SetField(tabTable, "Metadata", metadataTable)

			L.RawSetInt(resultTable, i+1, tabTable)
		}
		L.Push(resultTable)
		return 1
	}))
	L.SetField(apiTable, "set_tabs", L.NewFunction(func(L *lua.LState) int {
		tabsTable := L.CheckTable(1)
		tabs := make([]TabInfo, 0, tabsTable.Len())

		tabsTable.ForEach(func(key, value lua.LValue) {
			if value.Type() != lua.LTTable {
				return
			}
			tabTable := value.(*lua.LTable)

			tab := TabInfo{
				ID:           getStringField(tabTable, "ID"),
				Title:        getStringField(tabTable, "Title"),
				ResourceType: getStringField(tabTable, "ResourceType"),
				CurrentIndex: int(getNumberField(tabTable, "CurrentIndex")),
			}

			// Handle breadcrumb array
			if breadcrumbValue := tabTable.RawGetString("Breadcrumb"); breadcrumbValue.Type() == lua.LTTable {
				breadcrumbTable := breadcrumbValue.(*lua.LTable)
				var breadcrumb []string
				breadcrumbTable.ForEach(func(_, crumb lua.LValue) {
					if crumb.Type() == lua.LTString {
						breadcrumb = append(breadcrumb, crumb.String())
					}
				})
				tab.Breadcrumb = breadcrumb
			}

			// Handle metadata object
			if metadataValue := tabTable.RawGetString("Metadata"); metadataValue.Type() == lua.LTTable {
				metadataTable := metadataValue.(*lua.LTable)
				metadata := make(map[string]interface{})
				metadataTable.ForEach(func(key, val lua.LValue) {
					if key.Type() == lua.LTString && val.Type() == lua.LTString {
						metadata[key.String()] = val.String()
					}
				})
				tab.Metadata = metadata
			}

			tabs = append(tabs, tab)
		})

		err := gpm.api.SetTabs(tabs)
		if err != nil {
			L.Push(lua.LString(fmt.Sprintf("failed to set tabs: %v", err)))
			return 1
		}
		return 0
	}))
	L.SetField(apiTable, "get_breadcrumb_trail", L.NewFunction(func(L *lua.LState) int {
		breadcrumb := gpm.api.GetBreadcrumbTrail()
		resultTable := L.NewTable()
		for i, crumb := range breadcrumb {
			L.RawSetInt(resultTable, i+1, lua.LString(crumb))
		}
		L.Push(resultTable)
		return 1
	}))
	L.SetField(apiTable, "set_breadcrumb_trail", L.NewFunction(func(L *lua.LState) int {
		breadcrumbTable := L.CheckTable(1)
		breadcrumb := make([]string, 0, breadcrumbTable.Len())
		breadcrumbTable.ForEach(func(_, crumb lua.LValue) {
			if crumb.Type() == lua.LTString {
				breadcrumb = append(breadcrumb, crumb.String())
			}
		})
		gpm.api.SetBreadcrumbTrail(breadcrumb)
		return 0
	}))
	L.SetField(apiTable, "show_input_dialog", L.NewFunction(func(L *lua.LState) int {
		title := L.CheckString(1)
		placeholder := L.CheckString(2)
		submitCommand := L.CheckString(3)
		cancelCommand := L.CheckString(4)
		gpm.api.ShowInputDialog(title, placeholder, submitCommand, cancelCommand)
		return 0
	}))

	L.SetField(apiTable, "log", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		logger.PluginInfo("multicluster", message)
		return 0
	}))

	// Multi-cluster API functions
	L.SetField(apiTable, "get_all_clusters", L.NewFunction(func(L *lua.LState) int {
		allClusters := gpm.GetAllClusters()
		resultTable := L.NewTable()

		// Convert map to array for Lua
		clusterArray := make([]*ClusterContext, 0, len(allClusters))
		for _, cluster := range allClusters {
			clusterArray = append(clusterArray, cluster)
		}

		for i, cluster := range clusterArray {
			clusterTable := L.NewTable()
			L.SetField(clusterTable, "ID", lua.LString(cluster.ID))
			L.SetField(clusterTable, "Name", lua.LString(cluster.Name))
			L.SetField(clusterTable, "Namespace", lua.LString(cluster.Namespace))
			L.SetField(clusterTable, "Kubeconfig", lua.LString(cluster.Kubeconfig))
			L.SetField(clusterTable, "Index", lua.LNumber(i))
			L.SetField(clusterTable, "IsActive", lua.LBool(gpm.currentCluster == cluster.ID))

			// Add tabs if available
			if cluster.Tabs != nil {
				tabsTable := L.NewTable()
				for j, tab := range cluster.Tabs {
					tabTable := L.NewTable()
					L.SetField(tabTable, "ID", lua.LString(tab.ID))
					L.SetField(tabTable, "Title", lua.LString(tab.Title))
					L.SetField(tabTable, "ResourceType", lua.LString(tab.ResourceType))
					L.SetField(tabTable, "CurrentIndex", lua.LNumber(tab.CurrentIndex))

					// Handle breadcrumb
					if tab.Breadcrumb != nil {
						breadcrumbTable := L.NewTable()
						for k, crumb := range tab.Breadcrumb {
							L.RawSetInt(breadcrumbTable, k+1, lua.LString(crumb))
						}
						L.SetField(tabTable, "Breadcrumb", breadcrumbTable)
					}

					// Handle metadata
					if tab.Metadata != nil {
						metadataTable := L.NewTable()
						for k, v := range tab.Metadata {
							L.SetField(metadataTable, k, lua.LString(fmt.Sprintf("%v", v)))
						}
						L.SetField(tabTable, "Metadata", metadataTable)
					}

					L.RawSetInt(tabsTable, j+1, tabTable)
				}
				L.SetField(clusterTable, "Tabs", tabsTable)
			}

			// Add breadcrumb if available
			if cluster.Breadcrumb != nil {
				breadcrumbTable := L.NewTable()
				for j, crumb := range cluster.Breadcrumb {
					L.RawSetInt(breadcrumbTable, j+1, lua.LString(crumb))
				}
				L.SetField(clusterTable, "Breadcrumb", breadcrumbTable)
			}

			L.RawSetInt(resultTable, i+1, clusterTable)
		}
		L.Push(resultTable)
		return 1
	}))

	L.SetField(apiTable, "get_current_cluster", L.NewFunction(func(L *lua.LState) int {
		currentCluster := gpm.GetCurrentCluster()
		if currentCluster == nil {
			L.Push(lua.LNil)
			return 1
		}

		clusterTable := L.NewTable()
		L.SetField(clusterTable, "ID", lua.LString(currentCluster.ID))
		L.SetField(clusterTable, "Name", lua.LString(currentCluster.Name))
		L.SetField(clusterTable, "Namespace", lua.LString(currentCluster.Namespace))
		L.SetField(clusterTable, "Kubeconfig", lua.LString(currentCluster.Kubeconfig))
		L.SetField(clusterTable, "Index", lua.LNumber(0)) // Could be calculated if needed
		L.SetField(clusterTable, "IsActive", lua.LBool(true))
		L.Push(clusterTable)
		return 1
	}))

	L.SetField(apiTable, "sync_current_cluster_session", L.NewFunction(func(L *lua.LState) int {
		currentCluster := gpm.GetCurrentCluster()
		if currentCluster == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Get current tabs
		tabs, err := gpm.api.GetTabs()
		if err == nil {
			if err := gpm.SetClusterTabs(currentCluster.ID, tabs); err != nil {
				logger.Warn(fmt.Sprintf("Failed to update cluster tabs: %v", err))
			}
		}

		// Get current breadcrumb
		breadcrumb := gpm.api.GetBreadcrumbTrail()
		if len(breadcrumb) > 0 {
			if err := gpm.SetClusterBreadcrumb(currentCluster.ID, breadcrumb); err != nil {
				logger.Warn(fmt.Sprintf("Failed to update cluster breadcrumb: %v", err))
			}
		}

		// Update namespace in cluster context
		currentNamespace := gpm.api.GetCurrentNamespace()
		if err := gpm.SetClusterNamespace(currentCluster.ID, currentNamespace); err != nil {
			logger.Warn(fmt.Sprintf("Failed to update cluster namespace: %v", err))
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("k8s_tui", apiTable)
}

// AddCluster adds a new cluster context to the global plugin manager
func (gpm *GlobalPluginManager) AddCluster(id, name string, client k8s.Client) {
	gpm.AddClusterWithNamespace(id, name, client, "default")
}

// AddClusterWithNamespace adds a new cluster context with a specific namespace
func (gpm *GlobalPluginManager) AddClusterWithNamespace(id, name string, client k8s.Client, namespace string) {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	gpm.clusters[id] = &ClusterContext{
		ID:         id,
		Name:       name,
		Client:     client,
		Settings:   make(map[string]interface{}),
		Namespace:  namespace,
		Kubeconfig: client.KubeconfigPath,
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Added cluster %s (%s) with namespace %s", name, id, namespace))
}

// RemoveCluster removes a cluster context from the global plugin manager
func (gpm *GlobalPluginManager) RemoveCluster(id string) {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	if cluster, exists := gpm.clusters[id]; exists {
		delete(gpm.clusters, id)

		// If this was the current cluster, switch to first available
		if gpm.currentCluster == id {
			if len(gpm.clusters) > 0 {
				for newID := range gpm.clusters {
					gpm.SwitchToCluster(newID)
					break
				}
			} else {
				gpm.currentCluster = ""
				gpm.api.SetClient(k8s.Client{}) // Clear client
			}
		}

		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Removed cluster %s", cluster.Name))
	}
}

// SwitchToCluster switches the active cluster context for plugins
func (gpm *GlobalPluginManager) SwitchToCluster(id string) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[id]
	if !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	gpm.currentCluster = id
	gpm.api.SetClient(cluster.Client)
	gpm.api.SetCurrentNamespace(cluster.Namespace)

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Switched to cluster %s (%s) with namespace %s", cluster.Name, id, cluster.Namespace))
	return nil
}

// GetCurrentCluster returns the current active cluster context
func (gpm *GlobalPluginManager) GetCurrentCluster() *ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	if gpm.currentCluster == "" {
		return nil
	}

	return gpm.clusters[gpm.currentCluster]
}

// GetCluster returns a specific cluster context by ID
func (gpm *GlobalPluginManager) GetCluster(id string) *ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	return gpm.clusters[id]
}

// GetAllClusters returns all cluster contexts
func (gpm *GlobalPluginManager) GetAllClusters() map[string]*ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	result := make(map[string]*ClusterContext)
	for id, cluster := range gpm.clusters {
		result[id] = cluster
	}
	return result
}

// GetClusterNames returns all cluster names
func (gpm *GlobalPluginManager) GetClusterNames() []string {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	var names []string
	for _, cluster := range gpm.clusters {
		names = append(names, cluster.Name)
	}
	return names
}

// SetClusterNamespace sets the namespace for a specific cluster
func (gpm *GlobalPluginManager) SetClusterNamespace(clusterID, namespace string) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Namespace = namespace

	// If this is the current cluster, also update the API's namespace
	if gpm.currentCluster == clusterID {
		gpm.api.SetCurrentNamespace(namespace)
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Set namespace %s for cluster %s (%s)", namespace, cluster.Name, clusterID))
	return nil
}

// SetClusterTabs sets the tabs for a specific cluster
func (gpm *GlobalPluginManager) SetClusterTabs(clusterID string, tabs []TabInfo) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Tabs = tabs
	return nil
}

// SetClusterBreadcrumb sets the breadcrumb for a specific cluster
func (gpm *GlobalPluginManager) SetClusterBreadcrumb(clusterID string, breadcrumb []string) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Breadcrumb = breadcrumb
	return nil
}

// GetClusterNamespace gets the namespace for a specific cluster
func (gpm *GlobalPluginManager) GetClusterNamespace(clusterID string) (string, error) {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	logger.Info(fmt.Sprintf("🔍 DEBUG: GetClusterNamespace called for clusterID %s", clusterID))
	logger.Info(fmt.Sprintf("🔍 DEBUG: Available clusters: %v", gpm.getClusterIDs()))

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		logger.Error(fmt.Sprintf("🔍 DEBUG: Cluster %s not found in global manager", clusterID))
		return "", fmt.Errorf("cluster %s not found", clusterID)
	}

	logger.Info(fmt.Sprintf("🔍 DEBUG: Found cluster %s with namespace %s", clusterID, cluster.Namespace))
	return cluster.Namespace, nil
}

// Helper method to get all cluster IDs for debugging
func (gpm *GlobalPluginManager) getClusterIDs() []string {
	var ids []string
	for id := range gpm.clusters {
		ids = append(ids, id)
	}
	return ids
}

// SetClusterSetting sets a cluster-specific setting
func (gpm *GlobalPluginManager) SetClusterSetting(clusterID, key string, value interface{}) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Settings[key] = value
	return nil
}

// GetClusterSetting gets a cluster-specific setting
func (gpm *GlobalPluginManager) GetClusterSetting(clusterID, key string) (interface{}, error) {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	return cluster.Settings[key], nil
}

// ExecuteOnCluster executes a function with a specific cluster's context
func (gpm *GlobalPluginManager) ExecuteOnCluster(clusterID string, fn func(*ClusterContext) error) error {
	gpm.mu.Lock()
	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		gpm.mu.Unlock()
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	// Save current cluster
	oldCluster := gpm.currentCluster
	oldClient := gpm.api.GetClient()

	// Switch to target cluster
	gpm.currentCluster = clusterID
	gpm.api.SetClient(cluster.Client)
	gpm.mu.Unlock()

	// Execute function
	err := fn(cluster)

	// Restore previous cluster
	gpm.mu.Lock()
	gpm.currentCluster = oldCluster
	gpm.api.SetClient(oldClient)
	gpm.mu.Unlock()

	return err
}

// GetAPI returns the plugin API with multi-cluster support
func (gpm *GlobalPluginManager) GetAPI() *MultiClusterPluginAPI {
	return &MultiClusterPluginAPI{
		api:           gpm.PluginManager.GetAPI(),
		globalManager: gpm,
	}
}

// MultiClusterPluginAPI extends PluginAPI with multi-cluster capabilities
type MultiClusterPluginAPI struct {
	api           *PluginAPIImpl
	globalManager *GlobalPluginManager
}

// GetCurrentClusterID returns the current active cluster ID
func (mc *MultiClusterPluginAPI) GetCurrentClusterID() string {
	return mc.globalManager.currentCluster
}

// GetCurrentClusterName returns the current active cluster name
func (mc *MultiClusterPluginAPI) GetCurrentClusterName() string {
	cluster := mc.globalManager.GetCurrentCluster()
	if cluster == nil {
		return ""
	}
	return cluster.Name
}

// SwitchToCluster switches to a different cluster
func (mc *MultiClusterPluginAPI) SwitchToCluster(clusterID string) error {
	return mc.globalManager.SwitchToCluster(clusterID)
}

// GetClusterNames returns all available cluster names
func (mc *MultiClusterPluginAPI) GetClusterNames() []string {
	return mc.globalManager.GetClusterNames()
}

// SetClusterNamespace sets the namespace for a specific cluster
func (mc *MultiClusterPluginAPI) SetClusterNamespace(clusterID, namespace string) error {
	return mc.globalManager.SetClusterNamespace(clusterID, namespace)
}

// GetClusterNamespace gets the namespace for a specific cluster
func (mc *MultiClusterPluginAPI) GetClusterNamespace(clusterID string) (string, error) {
	return mc.globalManager.GetClusterNamespace(clusterID)
}

// ExecuteOnCluster executes a function in the context of a specific cluster
func (mc *MultiClusterPluginAPI) ExecuteOnCluster(clusterID string, fn func(client k8s.Client) error) error {
	return mc.globalManager.ExecuteOnCluster(clusterID, func(cluster *ClusterContext) error {
		return fn(cluster.Client)
	})
}

// Forward all PluginAPI methods to the embedded api
func (mc *MultiClusterPluginAPI) GetCurrentNamespace() string {
	return mc.api.GetCurrentNamespace()
}

func (mc *MultiClusterPluginAPI) SetCurrentNamespace(namespace string) {
	mc.api.SetCurrentNamespace(namespace)
}

func (mc *MultiClusterPluginAPI) SetStatusMessage(message string) {
	mc.api.SetStatusMessage(message)
}

func (mc *MultiClusterPluginAPI) AddHeaderComponent(component UIInjectionPoint) {
	mc.api.AddHeaderComponent(component)
}

func (mc *MultiClusterPluginAPI) AddFooterComponent(component UIInjectionPoint) {
	mc.api.AddFooterComponent(component)
}

func (mc *MultiClusterPluginAPI) GetHeaderComponents() []UIInjectionPoint {
	return mc.api.GetHeaderComponents()
}

func (mc *MultiClusterPluginAPI) GetFooterComponents() []UIInjectionPoint {
	return mc.api.GetFooterComponents()
}

func (mc *MultiClusterPluginAPI) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	mc.api.RegisterCommand(name, description, handler)
}

func (mc *MultiClusterPluginAPI) ExecuteCommand(name string, args []string) (string, error) {
	return mc.api.ExecuteCommand(name, args)
}

func (mc *MultiClusterPluginAPI) RegisterCLIArgument(name, description string, handler func(value string) error) {
	mc.api.RegisterCLIArgument(name, description, handler)
}

func (mc *MultiClusterPluginAPI) GetCLIArguments() map[string]CLIArgument {
	return mc.api.GetCLIArguments()
}

func (mc *MultiClusterPluginAPI) HasCLIArgument(name string) bool {
	return mc.api.HasCLIArgument(name)
}

func (mc *MultiClusterPluginAPI) ExecuteCLIArgument(name string, value string) error {
	return mc.api.ExecuteCLIArgument(name, value)
}

func (mc *MultiClusterPluginAPI) GetConfig(key string) any {
	return mc.api.GetConfig(key)
}

func (mc *MultiClusterPluginAPI) SetConfig(key string, value any) {
	mc.api.SetConfig(key, value)
}

func (mc *MultiClusterPluginAPI) GetClient() k8s.Client {
	return mc.api.GetClient()
}

func (mc *MultiClusterPluginAPI) SetClient(client k8s.Client) {
	mc.api.SetClient(client)
}

func (mc *MultiClusterPluginAPI) GetPods(namespace string, selector ...string) ([]k8s.PodInfo, error) {
	return mc.api.GetPods(namespace, selector...)
}

func (mc *MultiClusterPluginAPI) GetServices(namespace string) ([]k8s.ServiceInfo, error) {
	return mc.api.GetServices(namespace)
}

func (mc *MultiClusterPluginAPI) GetDeployments(namespace string) ([]k8s.DeploymentInfo, error) {
	return mc.api.GetDeployments(namespace)
}

func (mc *MultiClusterPluginAPI) GetConfigMaps(namespace string) ([]k8s.Configmap, error) {
	return mc.api.GetConfigMaps(namespace)
}

func (mc *MultiClusterPluginAPI) GetSecrets(namespace string) ([]k8s.SecretInfo, error) {
	return mc.api.GetSecrets(namespace)
}

func (mc *MultiClusterPluginAPI) GetIngresses(namespace string) ([]k8s.IngressInfo, error) {
	return mc.api.GetIngresses(namespace)
}

func (mc *MultiClusterPluginAPI) GetJobs(namespace string) ([]k8s.JobInfo, error) {
	return mc.api.GetJobs(namespace)
}

func (mc *MultiClusterPluginAPI) GetCronJobs(namespace string) ([]k8s.CronJobInfo, error) {
	return mc.api.GetCronJobs(namespace)
}

func (mc *MultiClusterPluginAPI) GetDaemonSets(namespace string) ([]k8s.DaemonSetInfo, error) {
	return mc.api.GetDaemonSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetStatefulSets(namespace string) ([]k8s.StatefulSetInfo, error) {
	return mc.api.GetStatefulSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetReplicaSets(namespace string) ([]k8s.ReplicaSetInfo, error) {
	return mc.api.GetReplicaSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetNodes() ([]k8s.NodeInfo, error) {
	return mc.api.GetNodes()
}

func (mc *MultiClusterPluginAPI) GetNamespaces() ([]string, error) {
	return mc.api.GetNamespaces()
}

func (mc *MultiClusterPluginAPI) GetServiceAccounts(namespace string) ([]k8s.ServiceAccountInfo, error) {
	return mc.api.GetServiceAccounts(namespace)
}

func (mc *MultiClusterPluginAPI) DeletePod(namespace, name string) error {
	return mc.api.DeletePod(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteService(namespace, name string) error {
	return mc.api.DeleteService(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteDeployment(namespace, name string) error {
	return mc.api.DeleteDeployment(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteConfigMap(namespace, name string) error {
	return mc.api.DeleteConfigMap(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteSecret(namespace, name string) error {
	return mc.api.DeleteSecret(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteIngress(namespace, name string) error {
	return mc.api.DeleteIngress(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteJob(namespace, name string) error {
	return mc.api.DeleteJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteCronJob(namespace, name string) error {
	return mc.api.DeleteCronJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteDaemonSet(namespace, name string) error {
	return mc.api.DeleteDaemonSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteStatefulSet(namespace, name string) error {
	return mc.api.DeleteStatefulSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteReplicaSet(namespace, name string) error {
	return mc.api.DeleteReplicaSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteServiceAccount(namespace, name string) error {
	return mc.api.DeleteServiceAccount(namespace, name)
}

func (mc *MultiClusterPluginAPI) GetTabs() ([]TabInfo, error) {
	return mc.api.GetTabs()
}

func (mc *MultiClusterPluginAPI) SetTabs(tabs []TabInfo) error {
	return mc.api.SetTabs(tabs)
}

func (mc *MultiClusterPluginAPI) SetTabSetterCallback(callback func()) {
	mc.api.SetTabSetterCallback(callback)
}

func (mc *MultiClusterPluginAPI) GetBreadcrumbTrail() []string {
	return mc.api.GetBreadcrumbTrail()
}

func (mc *MultiClusterPluginAPI) SetBreadcrumbTrail(breadcrumb []string) {
	mc.api.SetBreadcrumbTrail(breadcrumb)
}

func (mc *MultiClusterPluginAPI) ShowInputDialog(title, placeholder, submitCommand, cancelCommand string) {
	mc.api.ShowInputDialog(title, placeholder, submitCommand, cancelCommand)
}

func (mc *MultiClusterPluginAPI) GetCurrentResourceType() string {
	return mc.api.GetCurrentResourceType()
}

func (mc *MultiClusterPluginAPI) GetHelp(resourceType string) (title, content string) {
	return mc.api.GetHelp(resourceType)
}

func (mc *MultiClusterPluginAPI) RegisterHelp(resourceType string, title, content string) {
	mc.api.RegisterHelp(resourceType, title, content)
}

func (mc *MultiClusterPluginAPI) RegisterEventHandler(event PluginEvent, handler func(data any) error) {
	mc.api.RegisterEventHandler(event, handler)
}

func (mc *MultiClusterPluginAPI) TriggerEvent(event PluginEvent, data any) {
	mc.api.TriggerEvent(event, data)
}

func (mc *MultiClusterPluginAPI) GetCommands() map[string]PluginCommand {
	return mc.api.GetCommands()
}

func (mc *MultiClusterPluginAPI) SetNamespaceCallback(callback func(namespace string)) {
	mc.api.SetNamespaceCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetStatusCallback(callback func(message string)) {
	mc.api.SetStatusCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetBreadcrumbCallback(callback func(breadcrumb []string)) {
	mc.api.SetBreadcrumbCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetShowInputDialogCallback(callback func(title, placeholder, submitCommand, cancelCommand string)) {
	mc.api.SetShowInputDialogCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetTabGetter(getter func() ([]TabInfo, error)) {
	mc.api.SetTabGetter(getter)
}

func (mc *MultiClusterPluginAPI) SetTabSetter(setter func(tabs []TabInfo) error) {
	mc.api.SetTabSetter(setter)
}

func (mc *MultiClusterPluginAPI) GetTabSetter() func(tabs []TabInfo) error {
	return mc.api.GetTabSetter()
}

func (mc *MultiClusterPluginAPI) RegisterResourceHandler(resourceType k8s.ResourceType, handler ResourceHandler) {
	mc.api.RegisterResourceHandler(resourceType, handler)
}

func (mc *MultiClusterPluginAPI) GetSupportedResourceTypes() []k8s.ResourceType {
	return mc.api.GetSupportedResourceTypes()
}

func (mc *MultiClusterPluginAPI) GetResourceHandler(resourceType k8s.ResourceType) (ResourceHandler, bool) {
	return mc.api.GetResourceHandler(resourceType)
}

func (mc *MultiClusterPluginAPI) SetCurrentResourceType(resourceType string) {
	mc.api.SetCurrentResourceType(resourceType)
}

func (mc *MultiClusterPluginAPI) GetBreadcrumbCallback(callback func() []string) {
	mc.api.GetBreadcrumbCallback(callback)
}

func (mc *MultiClusterPluginAPI) DescribePod(namespace, name string) (string, error) {
	return mc.api.DescribePod(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeService(namespace, name string) (string, error) {
	return mc.api.DescribeService(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeDeployment(namespace, name string) (string, error) {
	return mc.api.DescribeDeployment(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeConfigMap(namespace, name string) (string, error) {
	return mc.api.DescribeConfigMap(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeSecret(namespace, name string) (string, error) {
	return mc.api.DescribeSecret(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeIngress(namespace, name string) (string, error) {
	return mc.api.DescribeIngress(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeJob(namespace, name string) (string, error) {
	return mc.api.DescribeJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeCronJob(namespace, name string) (string, error) {
	return mc.api.DescribeCronJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeDaemonSet(namespace, name string) (string, error) {
	return mc.api.DescribeDaemonSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeStatefulSet(namespace, name string) (string, error) {
	return mc.api.DescribeStatefulSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeReplicaSet(namespace, name string) (string, error) {
	return mc.api.DescribeReplicaSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeNode(name string) (string, error) {
	return mc.api.DescribeNode(name)
}

func (mc *MultiClusterPluginAPI) DescribeServiceAccount(namespace, name string) (string, error) {
	return mc.api.DescribeServiceAccount(namespace, name)
}
