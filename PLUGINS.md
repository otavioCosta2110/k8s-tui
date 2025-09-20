# k8s-tui Plugin System

k8s-tui supports a powerful plugin system inspired by Neovim's architecture, allowing you to extend the application with custom functionality written in Lua.

## Plugin Types

### 1. Legacy Plugins
- Basic Lua plugins with `GetResourceTypes()`, `GetResourceData()`, etc.
- Located in `./plugins/` directory
- Example: `cluster-monitor/main.lua`, `example-plugin/main.lua`

### 2. Neovim-Style Plugins ✨
- Advanced plugins with setup functions, configuration, commands, and hooks
- Follow Neovim's plugin architecture patterns
- Example: `neovim-header/main.lua`

## Neovim-Style Plugin Structure

```lua
-- Plugin metadata
function Name()
    return "my-plugin"
end

function Version()
    return "1.0.0"
end

function Description()
    return "My awesome k8s-tui plugin"
end

-- Default configuration
function Config()
    return {
        enabled = true,
        refresh_rate = 30,
        theme = "default"
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    print("Setting up plugin with options:")
    for k, v in pairs(opts) do
        print("  " .. k .. " = " .. v)
    end
    -- Plugin initialization code here
    return nil
end

-- Initialize the plugin
function Initialize()
    print("Plugin initialized")
    k8s_tui.set_status("Plugin ready!")
    return nil
end

-- Shutdown the plugin
function Shutdown()
    print("Plugin shutting down")
    return nil
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "my-command",
            description = "Execute my custom command"
        }
    }
end

-- Hooks that this plugin registers for
function Hooks()
    return {
        {
            event = "app_started",
            handler = "on_app_started"
        },
        {
            event = "namespace_changed",
            handler = "on_namespace_changed"
        }
    }
end

-- Hook handlers
function on_app_started(data)
    print("App started event received")
    k8s_tui.add_header("🚀 Plugin Active")
end

function on_namespace_changed(data)
    print("Namespace changed to " .. data)
    k8s_tui.set_status("Switched to namespace: " .. data)
end
```

## Plugin API

Neovim-style plugins have access to the `k8s_tui` API:

### Core Functions
- `k8s_tui.get_namespace()` - Get current namespace
- `k8s_tui.set_status(message)` - Set status message
- `k8s_tui.add_header(content)` - Add content to header
- `k8s_tui.register_command(name, description, handler)` - Register a command

### Events
Plugins can register for these events:
- `app_started` - Fired when the application starts
- `app_shutdown` - Fired when the application shuts down
- `namespace_changed` - Fired when namespace changes
- `resource_selected` - Fired when a resource is selected
- `ui_update` - Fired when UI updates

## Creating a Neovim-Style Plugin

1. Create a directory in `./plugins/` (e.g., `my-plugin/`)
2. Create `main.lua` with the plugin structure above
3. Implement the required functions (`Name`, `Version`, `Description`, `Initialize`)
4. Optionally implement advanced features (`Setup`, `Config`, `Commands`, `Hooks`)

## Example Plugin: Neovim Header

The `neovim-header` plugin demonstrates:
- Configuration system
- Setup function
- Event hooks
- Header integration
- Status messages

```bash
# Test the plugin
go run cmd/main.go --plugin-dir ./plugins
```

You should see:
- "Neovim Header Plugin initialized" in logs
- Custom header content added by the plugin
- Status messages from plugin hooks

## Plugin Development Tips

1. **Error Handling**: Always return `nil` for success, or an error string for failures
2. **Logging**: Use `print()` for debug output (visible in application logs)
3. **Configuration**: Use the `Config()` function to provide sensible defaults
4. **Events**: Register for events sparingly to avoid performance issues
5. **API**: Check the k8s_tui API documentation for available functions

## Migration from Legacy Plugins

To migrate a legacy plugin to Neovim-style:

1. Add `Setup(opts)` function for configuration
2. Add `Config()` function for defaults
3. Add `Commands()` and `Hooks()` functions
4. Use the `k8s_tui` API instead of direct function calls
5. Update function signatures to match the new pattern

Legacy plugins will continue to work alongside Neovim-style plugins.