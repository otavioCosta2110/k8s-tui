package plugins

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/yuin/gopher-lua"
)

// NeovimStyleLuaPlugin implements the NeovimStylePlugin interface for Lua plugins
type NeovimStyleLuaPlugin struct {
	L          *lua.LState
	pluginName string
	config     map[string]interface{}
	api        PluginAPI
}

// NewNeovimStyleLuaPlugin creates a new Neovim-style Lua plugin
func NewNeovimStyleLuaPlugin(L *lua.LState, pluginName string, api PluginAPI) *NeovimStyleLuaPlugin {
	return &NeovimStyleLuaPlugin{
		L:          L,
		pluginName: pluginName,
		config:     make(map[string]interface{}),
		api:        api,
	}
}

func (p *NeovimStyleLuaPlugin) Name() string {
	if err := p.L.CallByParam(lua.P{
		Fn:      p.L.GetGlobal("Name"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Name(): %v", err))
		return p.pluginName
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	return ret.String()
}

func (p *NeovimStyleLuaPlugin) Version() string {
	if p.L.GetGlobal("Version").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Version"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "1.0.0"
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		return ret.String()
	}
	return "1.0.0"
}

func (p *NeovimStyleLuaPlugin) Description() string {
	if p.L.GetGlobal("Description").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Description"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "Neovim-style Lua plugin"
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		return ret.String()
	}
	return "Neovim-style Lua plugin"
}

func (p *NeovimStyleLuaPlugin) Initialize() error {
	logger.PluginDebug(p.pluginName, "Initializing Neovim-style plugin")

	// Set up the plugin API in Lua
	p.setupLuaAPI()

	if err := p.L.CallByParam(lua.P{
		Fn:      p.L.GetGlobal("Initialize"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Initialize(): %v", err))
		return err
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	if ret.Type() == lua.LTString {
		errorMsg := ret.String()
		logger.PluginError(p.pluginName, fmt.Sprintf("Initialize() returned error: %s", errorMsg))
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

func (p *NeovimStyleLuaPlugin) Shutdown() error {
	if p.L.GetGlobal("Shutdown").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Shutdown"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return err
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTString {
			return fmt.Errorf("%s", ret.String())
		}
	}
	return nil
}

func (p *NeovimStyleLuaPlugin) Setup(opts map[string]interface{}) error {
	logger.PluginDebug(p.pluginName, "Setting up plugin with options")

	// Merge options with existing config
	for k, v := range opts {
		p.config[k] = v
	}

	// Call Lua setup function if it exists
	setupType := p.L.GetGlobal("Setup").Type()
	logger.PluginDebug(p.pluginName, fmt.Sprintf("Setup function type: %s", setupType))

	if setupType == lua.LTFunction {
		// Convert Go map to Lua table
		optsTable := p.L.NewTable()
		for k, v := range opts {
			optsTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
		}

		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Setup"),
			NRet:    1,
			Protect: true,
		}, optsTable); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Setup(): %v", err))
			return err
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTString {
			return fmt.Errorf("%s", ret.String())
		}
	}

	return nil
}

func (p *NeovimStyleLuaPlugin) Config() map[string]interface{} {
	configType := p.L.GetGlobal("Config").Type()
	logger.PluginDebug(p.pluginName, fmt.Sprintf("Config function type: %s", configType))

	if configType == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Config"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Config(): %v", err))
			return p.config
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			// Convert Lua table to Go map
			config := make(map[string]interface{})
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if key.Type() == lua.LTString {
					config[key.String()] = value.String()
				}
			})
			return config
		}
	}
	return p.config
}

func (p *NeovimStyleLuaPlugin) Commands() []PluginCommand {
	var commands []PluginCommand

	if p.L.GetGlobal("Commands").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Commands"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Commands(): %v", err))
			return commands
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if value.Type() == lua.LTTable {
					cmd := p.parsePluginCommand(value.(*lua.LTable))
					commands = append(commands, cmd)
				}
			})
		}
	}

	return commands
}

func (p *NeovimStyleLuaPlugin) Hooks() []PluginHook {
	var hooks []PluginHook

	if p.L.GetGlobal("Hooks").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Hooks"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Hooks(): %v", err))
			return hooks
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if value.Type() == lua.LTTable {
					hook := p.parsePluginHook(value.(*lua.LTable))
					hooks = append(hooks, hook)
				}
			})
		}
	}

	return hooks
}

func (p *NeovimStyleLuaPlugin) parsePluginCommand(tbl *lua.LTable) PluginCommand {
	return PluginCommand{
		Name:        p.getStringField(tbl, "name"),
		Description: p.getStringField(tbl, "description"),
		Handler: func(args []string) (string, error) {
			// This would need to be implemented to call Lua functions
			return "Command executed", nil
		},
	}
}

func (p *NeovimStyleLuaPlugin) parsePluginHook(tbl *lua.LTable) PluginHook {
	event := p.getStringField(tbl, "event")
	handlerName := p.getStringField(tbl, "handler")

	return PluginHook{
		Event: event,
		Handler: func(data interface{}) error {
			// Call the Lua handler function
			if p.L.GetGlobal(handlerName).Type() == lua.LTFunction {
				if err := p.L.CallByParam(lua.P{
					Fn:      p.L.GetGlobal(handlerName),
					NRet:    1,
					Protect: true,
				}, lua.LString(fmt.Sprintf("%v", data))); err != nil {
					logger.PluginError(p.pluginName, fmt.Sprintf("Error calling hook handler %s: %v", handlerName, err))
					return err
				}
				ret := p.L.Get(-1)
				p.L.Pop(1)
				if ret.Type() == lua.LTString {
					errorMsg := ret.String()
					logger.PluginError(p.pluginName, fmt.Sprintf("Hook handler %s returned error: %s", handlerName, errorMsg))
					return fmt.Errorf("%s", errorMsg)
				}
				logger.PluginDebug(p.pluginName, fmt.Sprintf("Hook handler %s called successfully", handlerName))
			} else {
				logger.PluginWarn(p.pluginName, fmt.Sprintf("Hook handler function %s not found", handlerName))
			}
			return nil
		},
	}
}

func (p *NeovimStyleLuaPlugin) getStringField(tbl *lua.LTable, key string) string {
	if val := tbl.RawGetString(key); val.Type() == lua.LTString {
		return val.String()
	}
	return ""
}

func (p *NeovimStyleLuaPlugin) setupLuaAPI() {
	// Create the k8s_tui API table
	apiTable := p.L.NewTable()

	// Add API functions
	p.L.SetField(apiTable, "get_namespace", p.L.NewFunction(p.luaGetNamespace))
	p.L.SetField(apiTable, "set_status", p.L.NewFunction(p.luaSetStatus))
	p.L.SetField(apiTable, "add_header", p.L.NewFunction(p.luaAddHeader))
	p.L.SetField(apiTable, "register_command", p.L.NewFunction(p.luaRegisterCommand))

	// Set the API in the global environment
	p.L.SetGlobal("k8s_tui", apiTable)
}

func (p *NeovimStyleLuaPlugin) luaGetNamespace(L *lua.LState) int {
	namespace := p.api.GetCurrentNamespace()
	L.Push(lua.LString(namespace))
	return 1
}

func (p *NeovimStyleLuaPlugin) luaSetStatus(L *lua.LState) int {
	message := L.CheckString(1)
	p.api.SetStatusMessage(message)
	return 0
}

func (p *NeovimStyleLuaPlugin) luaAddHeader(L *lua.LState) int {
	content := L.CheckString(1)
	component := UIInjectionPoint{
		Location: "header",
		Position: "right",
		Priority: 10,
		Component: DisplayComponent{
			Type: "text",
			Config: map[string]interface{}{
				"content": content,
				"style":   "info",
			},
		},
		DataSource:     "static",
		UpdateInterval: 0,
	}
	p.api.AddHeaderComponent(component)
	return 0
}

func (p *NeovimStyleLuaPlugin) luaRegisterCommand(L *lua.LState) int {
	name := L.CheckString(1)
	description := L.CheckString(2)
	// Store the command for later execution
	command := PluginCommand{
		Name:        name,
		Description: description,
		Handler: func(args []string) (string, error) {
			// This would call the Lua function
			return "Command executed from Lua", nil
		},
	}
	p.api.RegisterCommand(command.Name, command.Description, command.Handler)
	return 0
}
