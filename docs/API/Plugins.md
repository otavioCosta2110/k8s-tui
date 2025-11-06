# 🔌 Plugins

This guide covers developing plugins for k8s-tui using the Lua plugin system.

## 🎯 Plugin Overview

k8s-tui supports extending functionality through Lua plugins. Plugins can:

- Add custom resource handlers
- Create new UI components
- Implement custom commands
- Hook into application events
- Provide custom themes and styling

## 🏗️ Plugin Architecture

```
k8s-tui Application
├── Plugin Manager
├── Lua Runtime
└── Plugin API
    ├── Resource Handlers
    ├── UI Components
    ├── Event Hooks
    └── Custom Commands
```

## 🚀 Getting Started

### Plugin Structure

```
my-plugin/
├── main.lua              # Plugin entry point
├── README.md             # Plugin documentation
├── config.json           # Plugin metadata (optional)
└── lib/                  # Additional Lua modules
    ├── utils.lua
    └── components.lua
```

### Basic Plugin Template

```lua
-- main.lua
-- Basic plugin template for k8s-tui

-- Plugin metadata
local plugin = {
    name = "my-plugin",
    version = "1.0.0",
    description = "A sample plugin for k8s-tui",
    author = "Your Name",
}

-- Initialize plugin
function plugin.init()
    print("Initializing " .. plugin.name)
    
    -- Register hooks, commands, etc.
    k8s.register_command("hello", plugin.hello_command)
    k8s.register_hook("on_pod_select", plugin.on_pod_select)
end

-- Plugin commands
function plugin.hello_command(args)
    k8s.show_message("Hello from " .. plugin.name .. "!")
end

-- Event hooks
function plugin.on_pod_select(pod)
    k8s.log("Selected pod: " .. pod.name)
end

-- Cleanup on unload
function plugin.cleanup()
    print("Cleaning up " .. plugin.name)
end

-- Return plugin object
return plugin
```

## 📋 Plugin API

### Core API Functions

```lua
-- Resource management
k8s.list_resources(resource_type, namespace)
k8s.get_resource(resource_type, name, namespace)
k8s.create_resource(resource)
k8s.update_resource(resource)
k8s.delete_resource(resource_type, name, namespace)

-- UI operations
k8s.show_message(message, level)
k8s.show_error(message)
k8s.show_success(message)
k8s.create_table(columns, rows)
k8s.create_form(fields)

-- Event handling
k8s.register_hook(event_name, handler_function)
k8s.unregister_hook(event_name, handler_function)
k8s.emit_event(event_name, data)

-- Commands
k8s.register_command(command_name, handler_function)
k8s.unregister_command(command_name)

-- Configuration
k8s.get_config(key)
k8s.set_config(key, value)
k8s.get_plugin_config(plugin_name, key)
k8s.set_plugin_config(plugin_name, key, value)

-- Logging
k8s.log(message, level)
k8s.debug(message)
k8s.info(message)
k8s.warn(message)
k8s.error(message)
```

### Resource Handlers

```lua
-- Custom resource handler
local custom_handler = {
    name = "custom-resource",
    display_name = "Custom Resources",
    icon = "🔧",
    
    -- List resources
    list = function(namespace)
        local resources = k8s.list_resources("customresources", namespace)
        local rows = {}
        
        for _, resource in ipairs(resources) do
            table.insert(rows, {
                resource.metadata.name,
                resource.spec.type,
                resource.status.phase,
                resource.metadata.creationTimestamp
            })
        end
        
        return {
            columns = {"Name", "Type", "Status", "Age"},
            rows = rows
        }
    end,
    
    -- Get resource details
    get = function(name, namespace)
        return k8s.get_resource("customresources", name, namespace)
    end,
    
    -- Create resource
    create = function(spec)
        return k8s.create_resource(spec)
    end,
    
    -- Update resource
    update = function(name, namespace, spec)
        local resource = k8s.get_resource("customresources", name, namespace)
        resource.spec = spec
        return k8s.update_resource(resource)
    end,
    
    -- Delete resource
    delete = function(name, namespace)
        return k8s.delete_resource("customresources", name, namespace)
    end
}

-- Register the handler
k8s.register_resource_handler(custom_handler)
```

### UI Components

```lua
-- Custom table component
function plugin.create_custom_table()
    local table = k8s.create_table(
        {"Name", "Status", "Age"},
        {
            {"pod-1", "Running", "1d"},
            {"pod-2", "Pending", "5m"},
            {"pod-3", "Failed", "1h"}
        }
    )
    
    -- Set table properties
    table:set_title("My Custom Table")
    table:set_sort_column(2)  -- Sort by status
    table:set_filter("pod-")   -- Filter rows
    
    return table
end

-- Custom form component
function plugin.create_custom_form()
    local form = k8s.create_form({
        {name = "name", label = "Name", type = "text", required = true},
        {name = "namespace", label = "Namespace", type = "select", 
         options = {"default", "kube-system", "production"}},
        {name = "replicas", label = "Replicas", type = "number", default = 1},
        {name = "image", label = "Container Image", type = "text", 
         default = "nginx:alpine"},
        {name = "enabled", label = "Enable", type = "boolean", default = true}
    })
    
    form:set_title("Create Custom Resource")
    form:set_submit_handler(plugin.handle_form_submit)
    
    return form
end

function plugin.handle_form_submit(data)
    k8s.show_success("Creating resource: " .. data.name)
    -- Create resource logic here
end
```

## 🎣 Event System

### Available Events

```lua
-- Resource events
"on_resource_list"     -- After listing resources
"on_resource_select"   -- When a resource is selected
"on_resource_create"   -- After creating a resource
"on_resource_update"   -- After updating a resource
"on_resource_delete"   -- After deleting a resource

-- UI events
"on_tab_change"        -- When switching tabs
"on_screen_change"     -- When changing screens
"on_key_press"         -- When a key is pressed
"on_search"            -- When searching

-- Application events
"on_startup"           -- When application starts
"on_shutdown"          -- When application shuts down
"on_cluster_change"    -- When switching clusters
"on_namespace_change"  -- When switching namespaces
```

### Event Handler Examples

```lua
-- Handle pod selection
function plugin.on_pod_select(pod)
    if pod.status.phase == "Failed" then
        k8s.show_warning("Selected pod is in failed state")
    end
    
    -- Log selection
    k8s.log("Selected pod: " .. pod.metadata.name .. " in " .. pod.metadata.namespace)
end

-- Handle tab changes
function plugin.on_tab_change(from_tab, to_tab)
    if to_tab == "pods" then
        k8s.show_info("Switched to Pods tab")
        plugin.refresh_pod_data()
    end
end

-- Handle key presses
function plugin.on_key_press(key)
    if key == "F12" then
        plugin.show_custom_menu()
    end
end

-- Register event handlers
k8s.register_hook("on_pod_select", plugin.on_pod_select)
k8s.register_hook("on_tab_change", plugin.on_tab_change)
k8s.register_hook("on_key_press", plugin.on_key_press)
```

## ⚙️ Configuration

### Plugin Configuration

```lua
-- config.json
{
    "name": "my-plugin",
    "version": "1.0.0",
    "description": "A sample plugin",
    "author": "Your Name",
    "license": "MIT",
    "homepage": "https://github.com/user/my-plugin",
    "dependencies": [],
    "k8s-tui": {
        "min_version": "1.0.0",
        "max_version": "2.0.0"
    },
    "permissions": [
        "read:pods",
        "write:pods",
        "read:deployments"
    ],
    "config": {
        "auto_refresh": true,
        "refresh_interval": 30,
        "default_namespace": "default"
    }
}
```

### Accessing Configuration

```lua
-- Get plugin configuration
local config = k8s.get_plugin_config("my-plugin")
local auto_refresh = config.auto_refresh or false
local interval = config.refresh_interval or 60

-- Get global configuration
local global_config = k8s.get_config("global")
local theme = global_config.theme or "default"

-- Set configuration
k8s.set_plugin_config("my-plugin", "last_run", os.time())
k8s.set_config("user", "preferred_namespace", "production")
```

## 🔧 Advanced Features

### Custom Commands

```lua
-- Register custom commands
function plugin.register_commands()
    -- Simple command
    k8s.register_command("hello", function(args)
        k8s.show_message("Hello, " .. (args[1] or "World") .. "!")
    end)
    
    -- Complex command with subcommands
    k8s.register_command("myplugin", function(args)
        if #args == 0 then
            plugin.show_help()
            return
        end
        
        local subcommand = args[1]
        if subcommand == "status" then
            plugin.show_status()
        elseif subcommand == "config" then
            plugin.show_config()
        elseif subcommand == "reload" then
            plugin.reload()
        else
            k8s.show_error("Unknown subcommand: " .. subcommand)
        end
    end)
end

function plugin.show_help()
    local help = [[
My Plugin Commands:
  hello [name]     - Say hello
  myplugin status  - Show plugin status
  myplugin config  - Show configuration
  myplugin reload  - Reload plugin
    ]]
    k8s.show_message(help)
end
```

### Data Persistence

```lua
-- Store plugin data
function plugin.save_data(key, value)
    local data = plugin.load_data()
    data[key] = value
    k8s.set_plugin_config("my-plugin", "data", data)
end

-- Load plugin data
function plugin.load_data()
    return k8s.get_plugin_config("my-plugin", "data") or {}
end

-- Example: Save last selected pod
function plugin.on_pod_select(pod)
    plugin.save_data("last_pod", {
        name = pod.metadata.name,
        namespace = pod.metadata.namespace,
        timestamp = os.time()
    })
end
```

### HTTP Requests

```lua
-- Make HTTP requests (if allowed)
function plugin.fetch_external_data()
    local response = k8s.http_get("https://api.example.com/data")
    if response.status == 200 then
        local data = k8s.json_decode(response.body)
        plugin.process_data(data)
    else
        k8s.show_error("Failed to fetch data: " .. response.status)
    end
end

-- POST request
function plugin.send_data(data)
    local response = k8s.http_post("https://api.example.com/submit", {
        headers = {["Content-Type"] = "application/json"},
        body = k8s.json_encode(data)
    })
    
    if response.status == 200 then
        k8s.show_success("Data submitted successfully")
    else
        k8s.show_error("Failed to submit data")
    end
end
```

## 🧪 Testing Plugins

### Unit Testing

```lua
-- test/my-plugin_test.lua
local plugin = require("main")

-- Mock k8s API
local mock_k8s = {
    show_message = function(msg) print("MESSAGE: " .. msg) end,
    show_error = function(msg) print("ERROR: " .. msg) end,
    log = function(msg) print("LOG: " .. msg) end,
    get_config = function(key) return {} end,
    set_config = function(key, value) end
}

-- Replace global k8s with mock
_G.k8s = mock_k8s

-- Test plugin initialization
function test_init()
    plugin.init()
    assert(plugin.name == "my-plugin")
    print("✓ Plugin initialization test passed")
end

-- Test command registration
function test_command_registration()
    plugin.register_commands()
    -- Verify commands are registered (implementation dependent)
    print("✓ Command registration test passed")
end

-- Run tests
test_init()
test_command_registration()
print("All tests passed!")
```

### Integration Testing

```bash
# Test plugin with k8s-tui
k8s-tui --plugin-dir ./my-plugin --test-mode

# Test with specific configuration
k8s-tui --plugin-dir ./my-plugin --config test-config.json
```

## 📦 Distribution

### Plugin Package Structure

```
my-plugin-1.0.0.tar.gz
├── main.lua
├── README.md
├── config.json
├── lib/
│   ├── utils.lua
│   └── components.lua
└── assets/
    ├── icon.png
    └── theme.json
```

### Installation

```bash
# Install from file
k8s-tui plugin install my-plugin-1.0.0.tar.gz

# Install from URL
k8s-tui plugin install https://github.com/user/my-plugin/releases/download/v1.0.0/my-plugin-1.0.0.tar.gz

# Install from registry
k8s-tui plugin install my-plugin

# List installed plugins
k8s-tui plugin list

# Uninstall plugin
k8s-tui plugin uninstall my-plugin
```

## 📚 Best Practices

### DO ✅

1. **Handle errors gracefully** - Use try/catch patterns
2. **Validate inputs** - Check parameters before using them
3. **Use descriptive names** - For functions and variables
4. **Document your code** - Add comments and README
5. **Test thoroughly** - Unit and integration tests
6. **Follow Lua conventions** - Use proper Lua idioms
7. **Handle permissions** - Request only necessary permissions
8. **Provide feedback** - Show progress and status messages

### DON'T ❌

1. **Don't block the UI** - Use async operations for long tasks
2. **Don't ignore errors** - Always handle error conditions
3. **Don't use global variables** - Keep state encapsulated
4. **Don't hardcode values** - Use configuration
5. **Don't make unnecessary API calls** - Cache when possible
6. **Don't modify core functionality** - Extend, don't replace
7. **Don't assume resources exist** - Check before using
8. **Don't ignore version compatibility** - Check k8s-tui version

## 🔍 Debugging

### Debug Logging

```lua
-- Enable debug mode
local debug = k8s.get_plugin_config("my-plugin", "debug") or false

function plugin.debug_log(message)
    if debug then
        k8s.debug("[my-plugin] " .. message)
    end
end

-- Usage
plugin.debug_log("Starting operation")
local result = some_operation()
plugin.debug_log("Operation result: " .. tostring(result))
```

### Error Handling

```lua
function plugin.safe_operation()
    local success, result = pcall(function()
        -- Potentially failing operation
        return risky_operation()
    end)
    
    if not success then
        k8s.show_error("Operation failed: " .. result)
        return nil
    end
    
    return result
end
```

This comprehensive plugin system allows you to extend k8s-tui's functionality while maintaining stability and performance.