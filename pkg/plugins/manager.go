package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/yuin/gopher-lua"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PluginManager struct {
	registry             *PluginRegistry
	pluginDir            string
	luaStates            map[string]*lua.LState
	api                  *PluginAPIImpl
	pluginmanagerPlugins []PluginmanagerStylePlugin
}

func NewPluginManager(pluginDir string) *PluginManager {
	api := NewPluginAPI()
	return &PluginManager{
		registry:             NewPluginRegistry(),
		pluginDir:            pluginDir,
		luaStates:            make(map[string]*lua.LState),
		api:                  api,
		pluginmanagerPlugins: make([]PluginmanagerStylePlugin, 0),
	}
}

func (pm *PluginManager) LoadPlugins() error {
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Starting plugin loading from directory: %s", pm.pluginDir))

	if pm.pluginDir == "" {
		logger.Info("🔌 Plugin Manager: No plugin directory specified, skipping plugin loading")
		return nil
	}

	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin directory does not exist: %s", pm.pluginDir))
		return nil
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Scanning for Lua plugins in: %s", pm.pluginDir))

	var files []string
	err := filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Error scanning directory %s: %v", path, err))
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") {
			files = append(files, path)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: Found potential plugin file: %s", path))
		}
		return nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to scan plugin directory: %v", err))
		return fmt.Errorf("failed to scan plugin directory: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Found %d potential plugin files", len(files)))
	for i, file := range files {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: File %d: %s", i+1, file))
	}

	loadedCount := 0
	failedCount := 0

	for _, file := range files {
		pluginName := filepath.Base(filepath.Dir(file))
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

func (pm *PluginManager) setupBasicLuaAPI(L *lua.LState) {
	print("DEBUG: Setting up basic k8s_tui API for Lua plugin")

	apiTable := L.NewTable()

	plugin := &basicLuaPlugin{api: pm.api}

	L.SetField(apiTable, "get_namespace", L.NewFunction(plugin.luaGetNamespace))
	L.SetField(apiTable, "set_namespace", L.NewFunction(plugin.luaSetNamespace))
	L.SetField(apiTable, "set_status", L.NewFunction(plugin.luaSetStatus))
	L.SetField(apiTable, "restore_tabs", L.NewFunction(plugin.luaRestoreTabs))

	L.SetField(apiTable, "log", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		logger.PluginInfo("basic", message)
		return 0
	}))

	L.SetGlobal("k8s_tui", apiTable)
	print("DEBUG: Basic k8s_tui API set up for Lua plugin")
}

type basicLuaPlugin struct {
	api *PluginAPIImpl
}

func (p *basicLuaPlugin) luaGetNamespace(L *lua.LState) int {
	namespace := p.api.GetCurrentNamespace()
	L.Push(lua.LString(namespace))
	return 1
}

func (p *basicLuaPlugin) luaSetNamespace(L *lua.LState) int {
	namespace := L.CheckString(1)
	logger.Info(fmt.Sprintf("DEBUG: basic luaSetNamespace called with: %s", namespace))
	p.api.SetCurrentNamespace(namespace)
	logger.Info("DEBUG: basic luaSetNamespace completed")
	return 0
}

func (p *basicLuaPlugin) luaSetStatus(L *lua.LState) int {
	message := L.CheckString(1)
	p.api.SetStatusMessage(message)
	return 0
}

func (p *basicLuaPlugin) luaRestoreTabs(L *lua.LState) int {
	logger.Info("DEBUG: basic luaRestoreTabs called")
	tabsTable := L.CheckTable(1)

	var tabs []TabInfo
	tabsTable.ForEach(func(key, value lua.LValue) {
		if value.Type() == lua.LTTable {
			tab := parseTabInfo(value.(*lua.LTable))
			tabs = append(tabs, tab)
		}
	})

	logger.Info(fmt.Sprintf("DEBUG: basic Parsed %d tabs, calling p.api.RestoreTabs", len(tabs)))
	err := p.api.RestoreTabs(tabs)
	if err != nil {
		logger.Info(fmt.Sprintf("DEBUG: basic p.api.RestoreTabs failed: %v", err))
		L.Push(lua.LString(fmt.Sprintf("failed to restore tabs: %v", err)))
		return 1
	}

	logger.Info("DEBUG: basic luaRestoreTabs completed successfully")
	L.Push(lua.LString("ok"))
	return 1
}

func parseTabInfo(tbl *lua.LTable) TabInfo {
	id := getStringField(tbl, "ID")
	title := getStringField(tbl, "Title")
	resourceType := getStringField(tbl, "ResourceType")

	var breadcrumb []string
	if breadcrumbTable := tbl.RawGetString("Breadcrumb"); breadcrumbTable.Type() == lua.LTTable {
		breadcrumbTable.(*lua.LTable).ForEach(func(key, value lua.LValue) {
			if value.Type() == lua.LTString {
				breadcrumb = append(breadcrumb, value.String())
			}
		})
	}

	return TabInfo{
		ID:           id,
		Title:        title,
		ResourceType: resourceType,
		Breadcrumb:   breadcrumb,
	}
}

func getStringField(tbl *lua.LTable, key string) string {
	if val := tbl.RawGetString(key); val.Type() == lua.LTString {
		return val.String()
	}
	return ""
}

func (pm *PluginManager) loadLuaPlugin(path string) error {
	pluginName := filepath.Base(filepath.Dir(path))

	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Creating Lua state for plugin: %s", pluginName))
	L := lua.NewState()

	var setupType, configType, commandsType, hooksType lua.LValueType
	var isNeovimStyle bool

	content, err := os.ReadFile(path)
	if err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to read Lua script %s: %v", path, err))
		return fmt.Errorf("failed to read Lua script: %v", err)
	}

	if err := L.DoString(string(content)); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to execute Lua script %s: %v", path, err))
		return fmt.Errorf("failed to load Lua script: %v", err)
	}

	print(fmt.Sprintf("DEBUG: Setting up basic k8s_tui API for plugin: %s", pluginName))
	pm.setupBasicLuaAPI(L)

	if L.GetGlobal("CLIArguments").Type() != lua.LTFunction {
		cliArgsCode := `
function CLIArguments()
    return {
        {
            name = "session",
            description = "Load session from JSON file",
            handler = "load_session_cli_handler"
        }
    }
end

function load_session_cli_handler(value)
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: load_session_cli_handler called with value: " .. tostring(value))
    end

    -- Read the session file
    local file = io.open(value, "r")
    if not file then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Failed to open session file: " .. value)
        end
        return nil, "Failed to open session file: " .. value
    end

    local content = file:read("*all")
    file:close()

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Session file content: " .. content)
    end

    -- Simple JSON parsing for namespace
    -- Look for "namespace":"value"
    local namespace_pattern = '"namespace"%s*:%s*"([^"]+)"'
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Looking for namespace with pattern: " .. namespace_pattern)
    end
    local namespace = content:match(namespace_pattern)
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Namespace match result: " .. (namespace or "nil"))
    end
    if namespace then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Found namespace: '" .. namespace .. "'")
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No namespace found in content")
        end
    end

    -- Parse tabs array with proper bracket matching
    local tabs = {}
    -- Find the position after "tabs":[
    local tabs_start = content:find('"tabs"%s*:%s*%[')
    local tabs_content = nil
    if tabs_start then
        -- Find the matching closing bracket
        local bracket_count = 0
        local in_string = false
        local escape_next = false
        local end_pos = tabs_start - 1

        for i = tabs_start, #content do
            local char = content:sub(i, i)
            if escape_next then
                escape_next = false
            elseif char == '\\' then
                escape_next = true
            elseif char == '"' then
                in_string = not in_string
            elseif not in_string then
                if char == '[' then
                    bracket_count = bracket_count + 1
                elseif char == ']' then
                    bracket_count = bracket_count - 1
                    if bracket_count == 0 then
                        end_pos = i
                        break
                    end
                end
            end
        end

        if end_pos > tabs_start then
            -- Extract content between brackets
            local start_bracket = content:find('%[', tabs_start)
            if start_bracket then
                tabs_content = content:sub(start_bracket + 1, end_pos - 1)
            end
        end
    end

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Tabs content extraction result: " .. (tabs_content or "nil"))
    end
    if tabs_content then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Found tabs_content: '" .. tabs_content .. "'")
            k8s_tui.log("DEBUG: Starting to parse individual tab objects from tabs_content")
        end
        for tab_str in tabs_content:gmatch('{%s*([^}]+)%s*}') do
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Processing tab_str: '" .. tab_str .. "'")
            end
            local tab = {}

            -- Extract ID
            local id = tab_str:match('"id"%s*:%s*"([^"]+)"') or tab_str:match('"ID"%s*:%s*"([^"]+)"')

            -- Extract title
            local title = tab_str:match('"title"%s*:%s*"([^"]+)"') or tab_str:match('"Title"%s*:%s*"([^"]+)"')

            -- Extract resourceType
            local resourceType = tab_str:match('"resourceType"%s*:%s*"([^"]+)"') or tab_str:match('"ResourceType"%s*:%s*"([^"]+)"')
            if resourceType then tab.ResourceType = resourceType end

            -- Extract breadcrumb array with proper bracket matching
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Looking for breadcrumb in tab_str")
            end
            local breadcrumb_start = tab_str:find('"breadcrumb"%s*:%s*%[') or tab_str:find('"Breadcrumb"%s*:%s*%[')
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: breadcrumb_start: " .. (breadcrumb_start or "nil"))
            end
            local breadcrumb_content = nil
            if breadcrumb_start then
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Found breadcrumb start, parsing brackets")
                end
                local bracket_count = 0
                local in_string = false
                local escape_next = false
                local end_pos = breadcrumb_start - 1

                for i = breadcrumb_start, #tab_str do
                    local char = tab_str:sub(i, i)
                    if escape_next then
                        escape_next = false
                    elseif char == '\\' then
                        escape_next = true
                    elseif char == '"' then
                        in_string = not in_string
                    elseif not in_string then
                        if char == '[' then
                            bracket_count = bracket_count + 1
                        elseif char == ']' then
                            bracket_count = bracket_count - 1
                            if bracket_count == 0 then
                                end_pos = i
                                break
                            end
                        end
                    end
                end

                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: breadcrumb end_pos: " .. end_pos)
                end
                if end_pos > breadcrumb_start then
                    local start_bracket = tab_str:find('%[', breadcrumb_start)
                    if k8s_tui and k8s_tui.log then
                        k8s_tui.log("DEBUG: breadcrumb start_bracket: " .. (start_bracket or "nil"))
                    end
                    if start_bracket then
                        breadcrumb_content = tab_str:sub(start_bracket + 1, end_pos - 1)
                        if k8s_tui and k8s_tui.log then
                            k8s_tui.log("DEBUG: breadcrumb_content: '" .. breadcrumb_content .. "'")
                        end
                    end
                end
            end

            if breadcrumb_content then
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Parsing breadcrumb crumbs")
                end
                tab.Breadcrumb = {}
                for crumb in breadcrumb_content:gmatch('"([^"]+)"') do
                    if k8s_tui and k8s_tui.log then
                        k8s_tui.log("DEBUG: Found crumb: '" .. crumb .. "'")
                    end
                    table.insert(tab.Breadcrumb, crumb)
                end
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Breadcrumb array has " .. #tab.Breadcrumb .. " elements")
                end
            else
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: No breadcrumb content found")
                end
                tab.Breadcrumb = {}
            end

            if tab.ID and tab.Title and tab.ResourceType then
                table.insert(tabs, tab)
            end
        end
    end

    -- Set namespace if found
    if namespace then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Calling k8s_tui.set_namespace with: " .. namespace)
        end
        k8s_tui.set_namespace(namespace)
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: k8s_tui.set_namespace called successfully")
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No namespace found, skipping set_namespace")
        end
    end

    -- Restore tabs if found
    if #tabs > 0 then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Calling k8s_tui.restore_tabs with " .. #tabs .. " tabs")
        end
        for i, tab in ipairs(tabs) do
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Tab " .. i .. ": ID=" .. (tab.ID or "nil") .. ", Title=" .. (tab.Title or "nil") .. ", ResourceType=" .. (tab.ResourceType or "nil"))
                if tab.Breadcrumb then
                    k8s_tui.log("DEBUG: Tab " .. i .. " Breadcrumb: " .. table.concat(tab.Breadcrumb, " > "))
                end
            end
        end
        k8s_tui.restore_tabs(tabs)
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: k8s_tui.restore_tabs called")
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No tabs found, skipping restore_tabs")
        end
    end

    -- Set status
    local status_msg = "Session loaded from " .. value
    if namespace then
        status_msg = status_msg .. " (namespace: " .. namespace .. ")"
    end
    if #tabs > 0 then
        status_msg = status_msg .. " (" .. #tabs .. " tabs restored)"
    end
    k8s_tui.set_status(status_msg)

    return "Session loaded successfully", nil
end
`
		if err := L.DoString(cliArgsCode); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to execute CLI args code for %s: %v", pluginName, err))
		}
	}

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

	setupType = L.GetGlobal("Setup").Type()
	configType = L.GetGlobal("Config").Type()
	commandsType = L.GetGlobal("Commands").Type()
	hooksType = L.GetGlobal("Hooks").Type()

	print(fmt.Sprintf("DEBUG: Plugin %s function types: Setup=%s, Config=%s, Commands=%s, Hooks=%s", pluginName, setupType, configType, commandsType, hooksType))

	isNeovimStyle = setupType == lua.LTFunction ||
		configType == lua.LTFunction ||
		commandsType == lua.LTFunction ||
		hooksType == lua.LTFunction

	print(fmt.Sprintf("DEBUG: Plugin %s isNeovimStyle: %t", pluginName, isNeovimStyle))

	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected pluginmanager-style plugin: %s", pluginName))
		logger.Info("🔌 Plugin Manager: Setting up k8s_tui API for pluginmanager-style plugin")

		apiTable := L.NewTable()

		L.SetField(apiTable, "get_namespace", L.NewFunction(func(L *lua.LState) int {
			namespace := pm.api.GetCurrentNamespace()
			L.Push(lua.LString(namespace))
			return 1
		}))
		L.SetField(apiTable, "set_namespace", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			pm.api.SetCurrentNamespace(namespace)
			return 0
		}))
		L.SetField(apiTable, "get_tabs", L.NewFunction(func(L *lua.LState) int {
			tabs, err := pm.api.GetTabs()
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
				breadcrumbTable := L.NewTable()
				for j, crumb := range tab.Breadcrumb {
					L.RawSetInt(breadcrumbTable, j+1, lua.LString(crumb))
				}
				L.SetField(tabTable, "Breadcrumb", breadcrumbTable)
				L.RawSetInt(resultTable, i+1, tabTable)
			}
			L.Push(resultTable)
			return 1
		}))
		L.SetField(apiTable, "set_tabs", L.NewFunction(func(L *lua.LState) int {
			if L.GetTop() < 1 || L.Get(1).Type() != lua.LTTable {
				L.Push(lua.LString("expected table argument"))
				return 1
			}

			tabsTable := L.CheckTable(1)
			var tabInfos []TabInfo

			tabsTable.ForEach(func(key, value lua.LValue) {
				if value.Type() == lua.LTTable {
					tabTable := value.(*lua.LTable)
					tabInfo := TabInfo{}

					if id := tabTable.RawGetString("ID"); id.Type() == lua.LTString {
						tabInfo.ID = id.String()
					}
					if title := tabTable.RawGetString("Title"); title.Type() == lua.LTString {
						tabInfo.Title = title.String()
					}
					if resourceType := tabTable.RawGetString("ResourceType"); resourceType.Type() == lua.LTString {
						tabInfo.ResourceType = resourceType.String()
					}
					if breadcrumb := tabTable.RawGetString("Breadcrumb"); breadcrumb.Type() == lua.LTTable {
						breadcrumbTable := breadcrumb.(*lua.LTable)
						var crumbs []string
						breadcrumbTable.ForEach(func(_, crumb lua.LValue) {
							if crumb.Type() == lua.LTString {
								crumbs = append(crumbs, crumb.String())
							}
						})
						tabInfo.Breadcrumb = crumbs
					}

					if tabInfo.ID != "" && tabInfo.Title != "" {
						tabInfos = append(tabInfos, tabInfo)
					}
				}
			})

			if pm.api.GetTabRestorer() != nil {
				err := pm.api.GetTabRestorer()(tabInfos)
				if err != nil {
					L.Push(lua.LString(fmt.Sprintf("failed to set tabs: %v", err)))
					return 1
				}
			}

			L.Push(lua.LString("tabs set successfully"))
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
					return "Command executed from Lua (legacy)", nil
				},
			}
			pm.api.RegisterCommand(command.Name, command.Description, command.Handler)
			return 0
		}))
		L.SetField(apiTable, "register_cli_argument", L.NewFunction(func(L *lua.LState) int {
			name := L.CheckString(1)
			description := L.CheckString(2)
			handlerName := L.CheckString(3)

			pm.api.RegisterCLIArgument(name, description, func(value string) error {
				if L.GetGlobal(handlerName).Type() == lua.LTFunction {

					if err := L.CallByParam(lua.P{
						Fn:      L.GetGlobal(handlerName),
						NRet:    1,
						Protect: true,
					}, lua.LString(value)); err != nil {
						logger.PluginError(pluginName, fmt.Sprintf("Error calling CLI argument handler %s: %v", handlerName, err))
						return err
					}
					ret := L.Get(-1)
					L.Pop(1)
					if ret.Type() == lua.LTString && ret.String() != "" {
						return fmt.Errorf("%s", ret.String())
					}
				} else {
					logger.PluginWarn(pluginName, fmt.Sprintf("CLI argument handler function %s not found", handlerName))
				}
				return nil
			})
			return 0
		}))

		L.SetField(apiTable, "get_pods", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			pods, err := k8s.FetchPods(client, namespace, "")
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch pods: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, pod := range pods {
				podTable := L.NewTable()
				L.SetField(podTable, "Name", lua.LString(pod.Name))
				L.SetField(podTable, "Namespace", lua.LString(pod.Namespace))
				L.SetField(podTable, "Ready", lua.LString(pod.Ready))
				L.SetField(podTable, "Status", lua.LString(pod.Status))
				L.SetField(podTable, "Restarts", lua.LNumber(pod.Restarts))
				L.SetField(podTable, "Age", lua.LString(pod.Age))
				L.RawSetInt(resultTable, i+1, podTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_services", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			services, err := k8s.GetServicesTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch services: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, svc := range services {
				svcTable := L.NewTable()
				L.SetField(svcTable, "Name", lua.LString(svc.Name))
				L.SetField(svcTable, "Namespace", lua.LString(svc.Namespace))
				L.SetField(svcTable, "Type", lua.LString(svc.Type))
				L.SetField(svcTable, "ClusterIP", lua.LString(svc.ClusterIP))
				L.SetField(svcTable, "ExternalIP", lua.LString(svc.ExternalIP))
				L.SetField(svcTable, "Ports", lua.LString(svc.Ports))
				L.SetField(svcTable, "Age", lua.LString(svc.Age))
				L.RawSetInt(resultTable, i+1, svcTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_deployments", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			deployments, err := k8s.GetDeploymentsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch deployments: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, dep := range deployments {
				depTable := L.NewTable()
				L.SetField(depTable, "Name", lua.LString(dep.Name))
				L.SetField(depTable, "Namespace", lua.LString(dep.Namespace))
				L.SetField(depTable, "Ready", lua.LString(dep.Ready))
				L.SetField(depTable, "UpToDate", lua.LString(dep.UpToDate))
				L.SetField(depTable, "Available", lua.LString(dep.Available))
				L.SetField(depTable, "Age", lua.LString(dep.Age))
				L.RawSetInt(resultTable, i+1, depTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_configmaps", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			configmaps, err := k8s.FetchConfigmaps(client, namespace, "")
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch configmaps: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, cm := range configmaps {
				cmTable := L.NewTable()
				L.SetField(cmTable, "Name", lua.LString(cm.Name))
				L.SetField(cmTable, "Namespace", lua.LString(cm.Namespace))
				L.SetField(cmTable, "Data", lua.LNumber(len(cm.Data)))
				L.SetField(cmTable, "Age", lua.LString(cm.Age))
				L.RawSetInt(resultTable, i+1, cmTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_secrets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			secrets, err := k8s.GetSecretsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch secrets: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, secret := range secrets {
				secretTable := L.NewTable()
				L.SetField(secretTable, "Name", lua.LString(secret.Name))
				L.SetField(secretTable, "Namespace", lua.LString(secret.Namespace))
				L.SetField(secretTable, "Type", lua.LString(secret.Type))
				L.SetField(secretTable, "Data", lua.LNumber(len(secret.Data)))
				L.SetField(secretTable, "Age", lua.LString(secret.Age))
				L.RawSetInt(resultTable, i+1, secretTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_ingresses", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			ingresses, err := k8s.GetIngressesTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch ingresses: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, ing := range ingresses {
				ingTable := L.NewTable()
				L.SetField(ingTable, "Name", lua.LString(ing.Name))
				L.SetField(ingTable, "Namespace", lua.LString(ing.Namespace))
				L.SetField(ingTable, "Class", lua.LString(ing.Class))
				L.SetField(ingTable, "Hosts", lua.LString(ing.Hosts))
				L.SetField(ingTable, "Address", lua.LString(ing.Address))
				L.SetField(ingTable, "Ports", lua.LString(ing.Ports))
				L.SetField(ingTable, "Age", lua.LString(ing.Age))
				L.RawSetInt(resultTable, i+1, ingTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_jobs", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			jobs, err := k8s.GetJobsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch jobs: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, job := range jobs {
				jobTable := L.NewTable()
				L.SetField(jobTable, "Name", lua.LString(job.Name))
				L.SetField(jobTable, "Namespace", lua.LString(job.Namespace))
				L.SetField(jobTable, "Completions", lua.LString(job.Completions))
				L.SetField(jobTable, "Duration", lua.LString(job.Duration))
				L.SetField(jobTable, "Age", lua.LString(job.Age))
				L.RawSetInt(resultTable, i+1, jobTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_cronjobs", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			cronjobs, err := k8s.GetCronJobsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch cronjobs: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, cj := range cronjobs {
				cjTable := L.NewTable()
				L.SetField(cjTable, "Name", lua.LString(cj.Name))
				L.SetField(cjTable, "Namespace", lua.LString(cj.Namespace))
				L.SetField(cjTable, "Schedule", lua.LString(cj.Schedule))
				L.SetField(cjTable, "Suspend", lua.LString(cj.Suspend))
				L.SetField(cjTable, "Active", lua.LString(cj.Active))
				L.SetField(cjTable, "LastSchedule", lua.LString(cj.LastSchedule))
				L.SetField(cjTable, "Age", lua.LString(cj.Age))
				L.RawSetInt(resultTable, i+1, cjTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_daemonsets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			daemonsets, err := k8s.GetDaemonSetsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch daemonsets: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, ds := range daemonsets {
				dsTable := L.NewTable()
				L.SetField(dsTable, "Name", lua.LString(ds.Name))
				L.SetField(dsTable, "Namespace", lua.LString(ds.Namespace))
				L.SetField(dsTable, "Desired", lua.LString(ds.Desired))
				L.SetField(dsTable, "Current", lua.LString(ds.Current))
				L.SetField(dsTable, "Ready", lua.LString(ds.Ready))
				L.SetField(dsTable, "UpToDate", lua.LString(ds.UpToDate))
				L.SetField(dsTable, "Available", lua.LString(ds.Available))
				L.SetField(dsTable, "Age", lua.LString(ds.Age))
				L.RawSetInt(resultTable, i+1, dsTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_statefulsets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			statefulsets, err := k8s.GetStatefulSetsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch statefulsets: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, sts := range statefulsets {
				stsTable := L.NewTable()
				L.SetField(stsTable, "Name", lua.LString(sts.Name))
				L.SetField(stsTable, "Namespace", lua.LString(sts.Namespace))
				L.SetField(stsTable, "Ready", lua.LString(sts.Ready))
				L.SetField(stsTable, "Age", lua.LString(sts.Age))
				L.RawSetInt(resultTable, i+1, stsTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_replicasets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			replicasets, err := k8s.GetReplicaSetsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch replicasets: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, rs := range replicasets {
				rsTable := L.NewTable()
				L.SetField(rsTable, "Name", lua.LString(rs.Name))
				L.SetField(rsTable, "Namespace", lua.LString(rs.Namespace))
				L.SetField(rsTable, "Desired", lua.LString(rs.Desired))
				L.SetField(rsTable, "Current", lua.LString(rs.Current))
				L.SetField(rsTable, "Ready", lua.LString(rs.Ready))
				L.SetField(rsTable, "Age", lua.LString(rs.Age))
				L.RawSetInt(resultTable, i+1, rsTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_nodes", L.NewFunction(func(L *lua.LState) int {
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			nodes, err := k8s.GetNodesTableData(client)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch nodes: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, node := range nodes {
				nodeTable := L.NewTable()
				L.SetField(nodeTable, "Name", lua.LString(node.Name))
				L.SetField(nodeTable, "Status", lua.LString(node.Status))
				L.SetField(nodeTable, "Roles", lua.LString(node.Roles))
				L.SetField(nodeTable, "Age", lua.LString(node.Age))
				L.SetField(nodeTable, "Version", lua.LString(node.Version))
				L.RawSetInt(resultTable, i+1, nodeTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_namespaces", L.NewFunction(func(L *lua.LState) int {
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			namespaces, err := k8s.FetchNamespaces(client)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch namespaces: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, ns := range namespaces {
				L.RawSetInt(resultTable, i+1, lua.LString(ns))
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_serviceaccounts", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			if namespace == "" {
				namespace = pm.api.GetCurrentNamespace()
			}
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			serviceaccounts, err := k8s.GetServiceAccountsTableData(client, namespace)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch serviceaccounts: %v", err)))
				return 1
			}
			resultTable := L.NewTable()
			for i, sa := range serviceaccounts {
				saTable := L.NewTable()
				L.SetField(saTable, "Name", lua.LString(sa.Name))
				L.SetField(saTable, "Namespace", lua.LString(sa.Namespace))
				L.SetField(saTable, "Secrets", lua.LString(sa.Secrets))
				L.SetField(saTable, "Age", lua.LString(sa.Age))
				L.RawSetInt(resultTable, i+1, saTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "get_tabs", L.NewFunction(func(L *lua.LState) int {
			tabs, err := pm.api.GetTabs()
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
				breadcrumbTable := L.NewTable()
				for j, crumb := range tab.Breadcrumb {
					L.RawSetInt(breadcrumbTable, j+1, lua.LString(crumb))
				}
				L.SetField(tabTable, "Breadcrumb", breadcrumbTable)
				L.RawSetInt(resultTable, i+1, tabTable)
			}
			L.Push(resultTable)
			return 1
		}))

		L.SetField(apiTable, "delete_pod", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeletePod(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete pod: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_service", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteService(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete service: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_deployment", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteDeployment(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete deployment: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_configmap", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteConfigmap(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete configmap: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_secret", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteSecret(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete secret: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_ingress", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteIngress(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete ingress: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_job", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteJob(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete job: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_cronjob", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteCronJob(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete cronjob: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_daemonset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteDaemonSet(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete daemonset: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_statefulset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteStatefulSet(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete statefulset: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_replicaset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteReplicaSet(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete replicaset: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "delete_serviceaccount", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}
			err := k8s.DeleteServiceAccount(client, namespace, name)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to delete serviceaccount: %v", err)))
				return 1
			}
			L.Push(lua.LString("ok"))
			return 1
		}))

		L.SetField(apiTable, "get_endpoints", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)

			client := pm.api.GetClient()
			if client.Clientset == nil {
				L.Push(lua.LString("no kubernetes client available"))
				return 1
			}

			endpointSlices, err := client.Clientset.DiscoveryV1().EndpointSlices(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to fetch endpoint slices: %v", err)))
				return 1
			}

			resultTable := L.NewTable()
			for i, endpointSlice := range endpointSlices.Items {

				var addresses []string
				for _, endpoint := range endpointSlice.Endpoints {
					for _, addr := range endpoint.Addresses {
						if addr != "" {

							isReady := true
							if endpoint.Conditions.Ready != nil {
								isReady = *endpoint.Conditions.Ready
							}
							if !isReady {
								addresses = append(addresses, addr+" (not ready)")
							} else {
								addresses = append(addresses, addr)
							}
						}
					}
				}
				addressesStr := strings.Join(addresses, ", ")
				if addressesStr == "" {
					addressesStr = "<none>"
				}

				var ports []string
				for _, port := range endpointSlice.Ports {
					if port.Port != nil {
						portStr := fmt.Sprintf("%d/%s", *port.Port, string(*port.Protocol))
						if port.Name != nil && *port.Name != "" {
							portStr = *port.Name + ":" + portStr
						}
						ports = append(ports, portStr)
					}
				}
				portsStr := strings.Join(ports, ", ")
				if portsStr == "" {
					portsStr = "<none>"
				}

				serviceName := "unknown"
				if endpointSlice.Labels != nil {
					if svcName, ok := endpointSlice.Labels["kubernetes.io/service-name"]; ok {
						serviceName = svcName
					}
				}

				age := "Unknown"
				if !endpointSlice.CreationTimestamp.IsZero() {
					age = format.FormatAge(endpointSlice.CreationTimestamp.Time)
				}

				endpointTable := L.NewTable()
				L.SetField(endpointTable, "Name", lua.LString(endpointSlice.Name))
				L.SetField(endpointTable, "Namespace", lua.LString(endpointSlice.Namespace))
				L.SetField(endpointTable, "Addresses", lua.LString(addressesStr))
				L.SetField(endpointTable, "Ports", lua.LString(portsStr))
				L.SetField(endpointTable, "Service", lua.LString(serviceName))
				L.SetField(endpointTable, "Age", lua.LString(age))

				L.RawSetInt(resultTable, i+1, endpointTable)
			}

			L.Push(resultTable)
			return 1
		}))

		L.SetGlobal("k8s_tui", apiTable)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: k8s_tui API set up for plugin: %s", pluginName))
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Creating plugin wrapper for: %s", pluginName))
	luaPlugin := &LuaPlugin{
		L:          L,
		pluginName: pluginName,
	}

	pluginDisplayName := luaPlugin.Name()
	pluginVersion := luaPlugin.Version()
	pluginDescription := luaPlugin.Description()

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Initializing plugin %s (%s v%s)", pluginDisplayName, pluginName, pluginVersion))

	if err := luaPlugin.Initialize(); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialization failed: %v", pluginDisplayName, err))
		return fmt.Errorf("failed to initialize Lua plugin: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialized successfully", pluginDisplayName))

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Function types for %s - Setup: %s, Config: %s, Commands: %s, Hooks: %s",
		pluginName, setupType, configType, commandsType, hooksType))

	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected pluginmanager-style plugin: %s", pluginDisplayName))

		pluginmanagerPlugin := NewPluginmanagerStyleLuaPlugin(L, pluginName, pm.api)

		pluginmanagerPlugin.SetupLuaAPI()

		defaultConfig := pluginmanagerPlugin.Config()

		if err := pluginmanagerPlugin.Setup(defaultConfig); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to setup Neovim-style plugin %s: %v", pluginDisplayName, err))
			L.Close()
			return fmt.Errorf("failed to setup Neovim-style plugin: %v", err)
		}

		commands := pluginmanagerPlugin.Commands()
		for _, cmd := range commands {
			pm.api.RegisterCommand(cmd.Name, cmd.Description, cmd.Handler)
		}

		cliArgs := pluginmanagerPlugin.CLIArguments()
		for _, arg := range cliArgs {
			pm.api.RegisterCLIArgument(arg.Name, arg.Description, arg.Handler)
		}

		hooks := pluginmanagerPlugin.Hooks()
		for _, hook := range hooks {
			pm.api.RegisterEventHandler(PluginEvent(hook.Event), hook.Handler)
		}

		pm.pluginmanagerPlugins = append(pm.pluginmanagerPlugins, pluginmanagerPlugin)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Registered pluginmanager-style plugin: %s v%s", pluginDisplayName, pluginVersion))

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

		logger.Warn(fmt.Sprintf("🔌 Plugin Manager: ⚠️  Legacy plugin detected: %s - Consider migrating to pluginmanager-style", pluginDisplayName))

		hasResourcePlugin := luaPlugin.hasResourcePlugin()
		hasUIPlugin := luaPlugin.hasUIPlugin()

		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s capabilities - Resource: %t, UI: %t", pluginDisplayName, hasResourcePlugin, hasUIPlugin))

		if hasResourcePlugin {
			pm.registry.RegisterResourcePlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Registered legacy resource plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))

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

	pm.luaStates[pluginName] = L

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎉 Plugin %s loaded and registered successfully", pluginDisplayName))

	return nil
}

func (pm *PluginManager) GetRegistry() *PluginRegistry {
	return pm.registry
}

func (pm *PluginManager) GetAPI() *PluginAPIImpl {
	return pm.api
}

func (pm *PluginManager) GetPluginmanagerPlugins() []PluginmanagerStylePlugin {
	return pm.pluginmanagerPlugins
}

func (pm *PluginManager) TriggerEvent(event PluginEvent, data interface{}) {
	pm.api.TriggerEvent(event, data)
}

func (pm *PluginManager) GetCustomResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {

	pm.api.SetClient(client)

	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.GetResourceData(client, resourceType, namespace)
			}
		}
	}
	return nil, fmt.Errorf("custom resource type %s not found", resourceType)
}

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

func (pm *PluginManager) Shutdown() error {
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Shutting down %d loaded plugins", len(pm.luaStates)))

	shutdownCount := 0
	errorCount := 0

	for name, L := range pm.luaStates {
		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Shutting down plugin: %s", name))

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
