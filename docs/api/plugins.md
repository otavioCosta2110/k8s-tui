# Plugin System

k8s-tui supports extensibility through a Lua-based plugin system that allows developers to add custom functionality without modifying the core application.

## Overview

Plugins enable:
- Custom resource viewers
- Enhanced navigation
- Automated operations
- Integration with external tools
- Custom themes and UI modifications

## Plugin Architecture

### Plugin Manager

The plugin system consists of:

- **Plugin Manager**: Loads and manages plugin lifecycle
- **Lua Runtime**: Executes plugin code in isolated environment
- **API Bridge**: Provides safe access to k8s-tui internals
- **Security Sandbox**: Restricts plugin capabilities

### Plugin Structure

```
plugins/
├── my-plugin/
│   └── main.lua          # Plugin entry point
└── another-plugin/
    ├── main.lua          # Plugin code
    └── config.lua        # Plugin configuration
```

## Writing Plugins

### Basic Plugin Structure

```lua
-- main.lua
local plugin = {}

-- Plugin metadata
plugin.name = "My Custom Plugin"
plugin.version = "1.0.0"
plugin.description = "Adds custom functionality to k8s-tui"

-- Plugin initialization
function plugin.init()
    print("My plugin initialized!")
end

-- Plugin cleanup
function plugin.cleanup()
    print("My plugin cleaned up!")
end

-- Export plugin
return plugin
```

### Plugin Lifecycle

1. **Discovery**: Plugin manager scans `plugins/` directory
2. **Loading**: Lua code is loaded into isolated runtime
3. **Initialization**: `init()` function called
4. **Execution**: Plugin responds to events and hooks
5. **Cleanup**: `cleanup()` called on shutdown

## Plugin API

### Core Functions

#### Logging

```lua
-- Log messages at different levels
log.info("Information message")
log.warn("Warning message")
log.error("Error message")
log.debug("Debug message")
```

#### UI Interaction

```lua
-- Display notification
ui.notify("Operation completed successfully")

-- Show confirmation dialog
local confirmed = ui.confirm("Are you sure you want to delete this resource?")

-- Open URL in browser
ui.open_url("https://kubernetes.io/docs")
```

#### Kubernetes Operations

```lua
-- Get current context
local context = k8s.get_context()

-- List resources
local pods = k8s.list_pods("default")

-- Execute kubectl command
local result = k8s.exec("get pods -o json")

-- Watch resources
k8s.watch_pods("default", function(event, pod)
    log.info("Pod " .. pod.metadata.name .. " " .. event)
end)
```

### Event Hooks

Plugins can hook into k8s-tui events:

```lua
function plugin.on_resource_selected(resource_type, resource_name, namespace)
    log.info("Selected " .. resource_type .. ": " .. resource_name)
end

function plugin.on_tab_changed(tab_name)
    log.info("Switched to tab: " .. tab_name)
end

function plugin.on_key_pressed(key)
    -- Handle custom key bindings
    if key == "ctrl+x" then
        -- Custom action
    end
end
```

### Custom Commands

```lua
-- Register custom command
commands.register("my-command", function(args)
    log.info("Executing my command with args: " .. table.concat(args, " "))
end)

-- Command can be executed from k8s-tui command line
-- :my-command arg1 arg2
```

## Advanced Features

### Custom Resource Views

```lua
-- Register custom resource viewer
views.register("my-resource", {
    list = function(namespace)
        -- Return list of custom resources
        return {
            {name = "resource1", status = "active"},
            {name = "resource2", status = "inactive"}
        }
    end,

    detail = function(namespace, name)
        -- Return detailed view of resource
        return {
            name = name,
            status = "active",
            created = "2024-01-01"
        }
    end,

    actions = {
        "restart",
        "scale"
    }
})
```

### Theme Customization

```lua
-- Modify color scheme
theme.set_color("accent", "#ff0000")
theme.set_color("background", "#000000")

-- Add custom styles
theme.add_style("my-style", {
    foreground = "#ffffff",
    background = "#333333",
    bold = true
})
```

### HTTP Client

```lua
-- Make HTTP requests
local response = http.get("https://api.example.com/status")
local data = json.decode(response.body)

-- Post data
local result = http.post("https://api.example.com/webhook", {
    headers = {["Content-Type"] = "application/json"},
    body = json.encode({event = "resource_created"})
})
```

## Plugin Configuration

### Configuration Files

```lua
-- config.lua
return {
    api_endpoint = "https://api.example.com",
    timeout = 30,
    debug = false
}
```

### Environment Variables

```lua
-- Access environment variables
local api_key = os.getenv("API_KEY")
local debug_mode = os.getenv("DEBUG") == "true"
```

## Security Considerations

### Sandboxing

Plugins run in a restricted Lua environment with limited access to:
- File system operations
- Network access (controlled)
- System commands
- k8s-tui internals (API-only)

### Permission Model

Plugins must declare required permissions:

```lua
plugin.permissions = {
    "read_pods",
    "write_configmaps",
    "http_access"
}
```

### Code Review

All plugins should be reviewed for:
- Security vulnerabilities
- Performance issues
- API abuse
- Malicious intent

## Example Plugins

### Resource Health Checker

```lua
local health_checker = {}

function health_checker.init()
    log.info("Health checker plugin loaded")
end

function health_checker.on_resource_selected(resource_type, name, namespace)
    if resource_type == "pod" then
        local pod = k8s.get_pod(namespace, name)
        if pod.status.phase ~= "Running" then
            ui.notify("Pod " .. name .. " is not healthy", "warning")
        end
    end
end

return health_checker
```

### Custom Dashboard

```lua
local dashboard = {}

function dashboard.init()
    -- Register custom tab
    tabs.register("dashboard", {
        title = "My Dashboard",
        render = function()
            local pods = k8s.list_pods("default")
            local deployments = k8s.list_deployments("default")

            return {
                "Pods: " .. #pods,
                "Deployments: " .. #deployments,
                "Cluster healthy: " .. tostring(#pods > 0)
            }
        end
    })
end

return dashboard
```

### Integration Plugin

```lua
local slack_integration = {}

function slack_integration.init()
    -- Hook into resource creation events
    hooks.register("resource_created", function(resource_type, name, namespace)
        http.post(os.getenv("SLACK_WEBHOOK"), {
            body = json.encode({
                text = "New " .. resource_type .. " created: " .. name .. " in " .. namespace
            })
        })
    end)
end

return slack_integration
```

## Plugin Development Tools

### Plugin Template

```bash
# Create new plugin from template
cp -r plugins/example-plugin plugins/my-plugin
cd plugins/my-plugin

# Edit main.lua with your logic
vim main.lua
```

### Testing Plugins

```lua
-- test_plugin.lua
local plugin = require("my-plugin")

-- Mock k8s-tui API
_G.k8s = {
    list_pods = function() return {} end,
    get_pod = function() return {status = {phase = "Running"}} end
}

_G.log = {
    info = print,
    warn = print,
    error = print
}

-- Test plugin
plugin.init()
plugin.on_resource_selected("pod", "test-pod", "default")
```

### Debugging

```lua
-- Enable debug logging
log.set_level("debug")

-- Inspect plugin state
plugin.inspect = function()
    return {
        name = plugin.name,
        version = plugin.version,
        hooks = plugin.registered_hooks or {}
    }
end
```

## Distribution

### Plugin Registry

Plugins can be distributed through:
- GitHub repositories
- k8s-tui plugin registry (future)
- Direct file sharing

### Installation

```bash
# Clone plugin repository
git clone https://github.com/user/my-plugin.git plugins/my-plugin

# Or download and extract
wget https://github.com/user/my-plugin/archive/main.zip
unzip main.zip -d plugins/my-plugin
```

### Updates

```bash
# Update plugin
cd plugins/my-plugin
git pull

# Restart k8s-tui to reload plugins
```

## Best Practices

### Performance

- Avoid blocking operations in event handlers
- Cache expensive operations
- Use asynchronous operations when possible

### Error Handling

```lua
function safe_operation()
    local success, result = pcall(function()
        -- Potentially failing operation
        return k8s.list_pods("default")
    end)

    if not success then
        log.error("Operation failed: " .. result)
        return nil
    end

    return result
end
```

### Code Organization

```lua
-- Split large plugins into modules
local utils = require("utils")
local api = require("api")

local plugin = {}

function plugin.init()
    utils.setup()
    api.connect()
end

return plugin
```

### Documentation

```lua
--- My Custom Plugin
-- This plugin adds custom functionality to k8s-tui
--
-- @usage
-- Place this plugin in the plugins/ directory and restart k8s-tui
--
-- @config
-- api_endpoint: API endpoint URL
-- timeout: Request timeout in seconds
local plugin = {}
```

This plugin system makes k8s-tui highly extensible while maintaining security and performance.</content>
</xai:function_call">  

Assistant: Now API reference. 

<xai:function_call name="write">
<parameter name="filePath">docs/api/api-reference.md