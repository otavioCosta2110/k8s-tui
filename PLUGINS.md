# k8s-tui Plugin System

k8s-tui supports a powerful plugin system inspired by Neovim's architecture, allowing you to extend the application with custom functionality written in Lua.

## Plugin Architecture

The plugin system supports Lua plugins with advanced features including:
- Setup functions and configuration management
- Custom commands and CLI arguments
- Event-driven hooks
- Resource type extensions
- UI component injections

## Plugin Types

### 1. Resource Plugins
- Extend k8s-tui with custom Kubernetes resource types
- Define custom table columns and data sources
- Handle CRUD operations for custom resources
- Example: `example-plugin/`

### 2. UI Plugins
- Add custom UI components and interactions
- Register commands and key bindings
- Respond to application events
- Example: `pluginmanager-header/`, `session-save-plugin/`

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

## Creating a Plugin

1. Create a directory in `./plugins/` (e.g., `my-plugin/`)
2. Create `main.lua` with the plugin structure above
3. Implement the required functions (`Name`, `Version`, `Description`, `Initialize`)
4. Optionally implement advanced features (`Setup`, `Config`, `Commands`, `Hooks`)

## Example Plugins

### Resource Plugin: Example Plugin
The `example-plugin` demonstrates:
- Custom resource type definition
- Data fetching and display
- Basic CRUD operations

### UI Plugin: PluginManager Header
The `pluginmanager-header` plugin demonstrates:
- Configuration system with `Setup()` and `Config()`
- Event hooks for app lifecycle
- Header component injection
- Custom commands registration

### Session Save Plugin
The `session-save-plugin` demonstrates:
- Session state management
- File I/O operations
- Custom key bindings (Ctrl+S)

```bash
# Test plugins
go run cmd/main.go --plugin-dir ./plugins
```

You should see:
- Plugin initialization messages in logs
- Custom UI components added by plugins
- New commands available in the interface

## Plugin Development Tips

1. **Error Handling**: Always return `nil` for success, or an error string for failures
2. **Logging**: Use `print()` for debug output (visible in application logs at `~/.local/state/k8s-tui/logs/`)
3. **Configuration**: Use the `Config()` function to provide sensible defaults
4. **Events**: Register for events sparingly to avoid performance issues
5. **API**: Use the `k8s_tui` global API for application integration
6. **Resource Plugins**: Implement `GetResourceTypes()`, `GetResourceData()`, `DeleteResource()`, and `GetResourceInfo()`
7. **UI Plugins**: Implement `GetUIExtensions()` for custom components

## Plugin Configuration

Plugins can be configured through:
1. **Default Config**: Use the `Config()` function in your plugin
2. **User Config**: Users can override settings in `~/.config/k8s-tui/config.json`
3. **Setup Function**: The `Setup(opts)` function receives user configuration

## Available Events

Plugins can register for these events:
- `app_started` - Fired when the application starts
- `app_shutdown` - Fired when the application shuts down
- `namespace_changed` - Fired when namespace changes
- `resource_selected` - Fired when a resource is selected
- `ui_update` - Fired when UI updates

## Plugin API Reference

The `k8s_tui` global API provides:
- **Namespace Management**: `get_namespace()`, `set_namespace()`
- **UI Components**: `add_header()`, `add_footer()`, `set_status()`
- **Commands**: `register_command()`, `execute_command()`
- **Kubernetes Resources**: Access to pods, services, deployments, etc.
- **Resource Operations**: Create, read, update, delete operations
- **Session Management**: Get/set tabs, breadcrumb trail, etc.