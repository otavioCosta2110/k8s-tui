# 📋 API Reference

This document provides a comprehensive reference for the k8s-tui API, including all available functions, types, and interfaces.

## 🏗️ Core API

### Plugin Interface

```lua
-- Plugin structure
plugin = {
    name = "plugin-name",
    version = "1.0.0",
    description = "Plugin description",
    author = "Author Name",
    
    -- Lifecycle methods
    init = function() end,
    cleanup = function() end,
    
    -- Optional methods
    on_load = function() end,
    on_unload = function() end,
    on_config_change = function(key, value) end
}
```

### Resource Management API

#### List Resources

```lua
resources = k8s.list_resources(resource_type, namespace, options)

-- Parameters:
--   resource_type: string - Kubernetes resource type (pods, deployments, etc.)
--   namespace: string - Kubernetes namespace (optional, default: "default")
--   options: table - Additional options (optional)

-- Options:
--   {
--     label_selector = "app=nginx",    -- Label selector
--     field_selector = "status.phase=Running",  -- Field selector
--     limit = 100,                    -- Limit results
--     timeout = "30s"                  -- Request timeout
--   }

-- Returns:
--   table - Array of resource objects

-- Example:
local pods = k8s.list_resources("pods", "default", {
    label_selector = "app=nginx",
    limit = 50
})
```

#### Get Resource

```lua
resource = k8s.get_resource(resource_type, name, namespace, options)

-- Parameters:
--   resource_type: string - Kubernetes resource type
--   name: string - Resource name
--   namespace: string - Kubernetes namespace (optional)
--   options: table - Additional options (optional)

-- Returns:
--   table - Resource object or nil if not found

-- Example:
local pod = k8s.get_resource("pods", "my-pod", "default")
if pod then
    print("Pod status:", pod.status.phase)
end
```

#### Create Resource

```lua
result = k8s.create_resource(resource_spec, options)

-- Parameters:
--   resource_spec: table - Kubernetes resource specification
--   options: table - Additional options (optional)

-- Options:
--   {
--     dry_run = false,        -- Dry run mode
--     namespace = "default",  -- Target namespace
--     validate = true         -- Validate resource
--   }

-- Returns:
--   table - Created resource object

-- Example:
local pod_spec = {
    apiVersion = "v1",
    kind = "Pod",
    metadata = {
        name = "my-pod",
        namespace = "default"
    },
    spec = {
        containers = {
            {
                name = "nginx",
                image = "nginx:alpine"
            }
        }
    }
}

local created = k8s.create_resource(pod_spec)
```

#### Update Resource

```lua
result = k8s.update_resource(resource, options)

-- Parameters:
--   resource: table - Updated resource object
--   options: table - Additional options (optional)

-- Returns:
--   table - Updated resource object

-- Example:
local pod = k8s.get_resource("pods", "my-pod", "default")
pod.spec.containers[1].image = "nginx:latest"
local updated = k8s.update_resource(pod)
```

#### Delete Resource

```lua
success = k8s.delete_resource(resource_type, name, namespace, options)

-- Parameters:
--   resource_type: string - Kubernetes resource type
--   name: string - Resource name
--   namespace: string - Kubernetes namespace (optional)
--   options: table - Additional options (optional)

-- Options:
--   {
--     grace_period_seconds = 30,  -- Grace period for deletion
--     dry_run = false,            -- Dry run mode
--     force = false               -- Force deletion
--   }

-- Returns:
--   boolean - True if deletion was successful

-- Example:
local success = k8s.delete_resource("pods", "my-pod", "default")
if success then
    k8s.show_success("Pod deleted successfully")
end
```

### UI API

#### Show Messages

```lua
k8s.show_message(message, level, timeout)

-- Parameters:
--   message: string - Message to display
--   level: string - Message level (info, warning, error, success)
--   timeout: number - Auto-dismiss timeout in seconds (optional)

-- Examples:
k8s.show_message("Operation completed", "success")
k8s.show_warning("This action is irreversible")
k8s.show_error("Failed to create resource")
k8s.show_info("Loading data...")
```

#### Create Table

```lua
table = k8s.create_table(columns, rows, options)

-- Parameters:
--   columns: table - Array of column names
--   rows: table - Array of row data
--   options: table - Table options (optional)

-- Options:
--   {
--     title = "My Table",           -- Table title
--     sortable = true,              -- Enable sorting
--     filterable = true,            -- Enable filtering
--     selectable = true,            -- Enable row selection
--     multi_select = false,         -- Enable multi-selection
--     sort_column = 1,              -- Default sort column
--     sort_desc = false,            -- Sort descending
--     max_rows = 1000               -- Maximum rows to display
--   }

-- Returns:
--   table - Table component object

-- Example:
local table = k8s.create_table(
    {"Name", "Status", "Age"},
    {
        {"pod-1", "Running", "1d"},
        {"pod-2", "Pending", "5m"},
        {"pod-3", "Failed", "1h"}
    },
    {
        title = "Pods",
        sortable = true,
        selectable = true
    }
)
```

#### Create Form

```lua
form = k8s.create_form(fields, options)

-- Parameters:
--   fields: table - Array of field definitions
--   options: table - Form options (optional)

-- Field Definition:
--   {
--     name = "field_name",           -- Field name (required)
--     label = "Field Label",         -- Display label
--     type = "text",                 -- Field type
--     required = true,               -- Required field
--     default = "default_value",     -- Default value
--     placeholder = "Enter value",  -- Placeholder text
--     options = {"opt1", "opt2"},    -- Options for select type
--     validation = "regex_pattern", -- Validation pattern
--     help = "Help text"             -- Help text
--   }

-- Field Types:
--   - text: Single line text input
--   - textarea: Multi-line text input
--   - number: Numeric input
--   - boolean: Checkbox
--   - select: Dropdown selection
--   - multiselect: Multi-select dropdown
--   - file: File selection
--   - password: Password input

-- Returns:
--   table - Form component object

-- Example:
local form = k8s.create_form({
    {
        name = "name",
        label = "Pod Name",
        type = "text",
        required = true,
        validation = "^[a-z0-9-]+$"
    },
    {
        name = "namespace",
        label = "Namespace",
        type = "select",
        options = {"default", "kube-system", "production"},
        default = "default"
    },
    {
        name = "replicas",
        label = "Replicas",
        type = "number",
        default = 1,
        validation = "^[1-9][0-9]*$"
    },
    {
        name = "enabled",
        label = "Enable",
        type = "boolean",
        default = true
    }
}, {
    title = "Create Deployment",
    submit_text = "Create",
    cancel_text = "Cancel"
})
```

### Event System API

#### Register Event Hooks

```lua
k8s.register_hook(event_name, handler_function)

-- Parameters:
--   event_name: string - Event name
--   handler_function: function - Event handler function

-- Available Events:
--   "on_startup"           - Application startup
--   "on_shutdown"          - Application shutdown
--   "on_cluster_change"    - Cluster change
--   "on_namespace_change"  - Namespace change
--   "on_tab_change"        - Tab change
--   "on_screen_change"     - Screen change
--   "on_resource_select"   - Resource selection
--   "on_resource_create"   - Resource creation
--   "on_resource_update"   - Resource update
--   "on_resource_delete"   - Resource deletion
--   "on_key_press"         - Key press
--   "on_search"            - Search action

-- Example:
k8s.register_hook("on_resource_select", function(resource)
    if resource.kind == "Pod" then
        k8s.log("Selected pod: " .. resource.metadata.name)
    end
end)
```

#### Emit Events

```lua
k8s.emit_event(event_name, data)

-- Parameters:
--   event_name: string - Event name
--   data: table - Event data (optional)

-- Example:
k8s.emit_event("custom_event", {
    message = "Something happened",
    timestamp = os.time()
})
```

### Command API

#### Register Commands

```lua
k8s.register_command(command_name, handler_function, options)

-- Parameters:
--   command_name: string - Command name
--   handler_function: function - Command handler
--   options: table - Command options (optional)

-- Options:
--   {
--     description = "Command description",
--     usage = "command [args]",
--     aliases = {"cmd", "c"},      -- Command aliases
--     hidden = false,              -- Hide from help
--     require_auth = true,        -- Require authentication
--     permissions = {"read:pods"}  -- Required permissions
--   }

-- Example:
k8s.register_command("hello", function(args)
    local name = args[1] or "World"
    k8s.show_message("Hello, " .. name .. "!")
end, {
    description = "Say hello",
    usage = "hello [name]",
    aliases = {"hi", "hey"}
})
```

### Configuration API

#### Get Configuration

```lua
value = k8s.get_config(key, default_value)

-- Parameters:
--   key: string - Configuration key
--   default_value: any - Default value if key not found (optional)

-- Returns:
--   any - Configuration value

-- Example:
local theme = k8s.get_config("theme", "default")
local auto_refresh = k8s.get_config("auto_refresh", false)
```

#### Set Configuration

```lua
k8s.set_config(key, value)

-- Parameters:
--   key: string - Configuration key
--   value: any - Configuration value

-- Example:
k8s.set_config("theme", "dark")
k8s.set_config("auto_refresh", true)
```

#### Plugin Configuration

```lua
value = k8s.get_plugin_config(plugin_name, key, default_value)
k8s.set_plugin_config(plugin_name, key, value)

-- Parameters:
--   plugin_name: string - Plugin name
--   key: string - Configuration key
--   value: any - Configuration value
--   default_value: any - Default value (optional)

-- Example:
local debug = k8s.get_plugin_config("my-plugin", "debug", false)
k8s.set_plugin_config("my-plugin", "last_run", os.time())
```

### Logging API

```lua
k8s.log(message, level)
k8s.debug(message)
k8s.info(message)
k8s.warn(message)
k8s.error(message)

-- Parameters:
--   message: string - Log message
--   level: string - Log level (debug, info, warn, error)

-- Log Levels:
--   debug: Detailed debugging information
--   info: General information messages
--   warn: Warning messages
--   error: Error messages

-- Example:
k8s.debug("Starting operation")
k8s.info("Operation completed successfully")
k8s.warn("Deprecated API used")
k8s.error("Failed to connect to cluster")
```

### HTTP API

```lua
response = k8s.http_get(url, options)
response = k8s.http_post(url, data, options)
response = k8s.http_put(url, data, options)
response = k8s.http_delete(url, options)

-- Parameters:
--   url: string - Request URL
--   data: table - Request data (for POST/PUT)
--   options: table - Request options (optional)

-- Options:
--   {
--     headers = {["Header"] = "Value"},  -- Request headers
--     timeout = 30,                      -- Timeout in seconds
--     auth = {type = "bearer", token = "token"},  -- Authentication
--     verify_ssl = true                  -- SSL verification
--   }

-- Response Object:
--   {
--     status = 200,           -- HTTP status code
--     headers = {},           -- Response headers
--     body = "response body"  -- Response body
--   }

-- Example:
local response = k8s.http_get("https://api.example.com/data", {
    headers = {["Accept"] = "application/json"},
    timeout = 10
})

if response.status == 200 then
    local data = k8s.json_decode(response.body)
    k8s.info("Data retrieved successfully")
else
    k8s.error("Failed to retrieve data: " .. response.status)
end
```

### Utility API

#### JSON Operations

```lua
json_string = k8s.json_encode(data)
data = k8s.json_decode(json_string)

-- Parameters:
--   data: table - Data to encode
--   json_string: string - JSON string to decode

-- Returns:
--   string - JSON encoded data
--   table - Decoded data

-- Example:
local data = {name = "test", value = 123}
local json = k8s.json_encode(data)
local decoded = k8s.json_decode(json)
```

#### File Operations

```lua
content = k8s.read_file(path)
success = k8s.write_file(path, content)
exists = k8s.file_exists(path)

-- Parameters:
--   path: string - File path
--   content: string - File content

-- Returns:
--   string - File content
--   boolean - Operation success
--   boolean - File exists

-- Example:
if k8s.file_exists("/tmp/data.json") then
    local content = k8s.read_file("/tmp/data.json")
    local data = k8s.json_decode(content)
end
```

#### String Operations

```lua
formatted = k8s.format_string(template, ...)
trimmed = k8s.trim_string(string)
split = k8s.split_string(string, delimiter)
joined = k8s.join_strings(array, delimiter)

-- Examples:
local formatted = k8s.format_string("Hello %s!", "World")
local trimmed = k8s.trim_string("  hello world  ")
local parts = k8s.split_string("a,b,c", ",")
local joined = k8s.join_strings({"a", "b", "c"}, ",")
```

#### Time Operations

```lua
timestamp = k8s.get_timestamp()
formatted = k8s.format_time(timestamp, format)
parsed = k8s.parse_time(time_string)

-- Parameters:
--   timestamp: number - Unix timestamp
--   time_string: string - Time string
--   format: string - Time format (optional, default: "2006-01-02 15:04:05")

-- Returns:
--   number - Current timestamp
--   string - Formatted time
--   number - Parsed timestamp

-- Example:
local now = k8s.get_timestamp()
local formatted = k8s.format_time(now, "2006-01-02")
local parsed = k8s.parse_time("2023-12-25")
```

## 🔧 Advanced API

### Resource Handlers

```lua
handler = {
    name = "custom-resource",
    display_name = "Custom Resources",
    icon = "🔧",
    
    -- Required methods
    list = function(namespace, options) end,
    get = function(name, namespace) end,
    create = function(spec) end,
    update = function(name, namespace, spec) end,
    delete = function(name, namespace) end,
    
    -- Optional methods
    describe = function(name, namespace) end,
    logs = function(name, namespace, options) end,
    exec = function(name, namespace, command) end,
    scale = function(name, namespace, replicas) end,
    restart = function(name, namespace) end
}

k8s.register_resource_handler(handler)
```

### Custom UI Components

```lua
component = {
    type = "custom-component",
    render = function(state) end,
    update = function(event, state) end,
    init = function() end,
    cleanup = function() end
}

k8s.register_component(component)
```

### Theme API

```lua
theme = {
    name = "custom-theme",
    colors = {
        background = "#000000",
        foreground = "#ffffff",
        primary = "#00ff00",
        secondary = "#0000ff",
        error = "#ff0000",
        warning = "#ffff00",
        success = "#00ff00",
        info = "#0088ff"
    },
    styles = {
        border = "single",
        padding = 1,
        margin = 0
    }
}

k8s.register_theme(theme)
k8s.set_theme("custom-theme")
```

## 📊 Data Types

### Resource Object

```lua
resource = {
    apiVersion = "v1",
    kind = "Pod",
    metadata = {
        name = "pod-name",
        namespace = "default",
        uid = "unique-id",
        creationTimestamp = "2023-01-01T00:00:00Z",
        labels = {app = "nginx"},
        annotations = {description = "My pod"}
    },
    spec = {
        -- Resource specification
    },
    status = {
        -- Resource status
    }
}
```

### Event Object

```lua
event = {
    type = "resource_select",
    timestamp = 1672531200,
    source = "ui",
    data = {
        resource = resource_object,
        context = "pods-tab"
    }
}
```

### Configuration Object

```lua
config = {
    theme = "dark",
    auto_refresh = true,
    refresh_interval = 30,
    default_namespace = "default",
    clusters = {
        {
            name = "production",
            kubeconfig = "/path/to/config",
            namespace = "default"
        }
    },
    plugins = {
        ["my-plugin"] = {
            enabled = true,
            config = {
                debug = false,
                auto_refresh = true
            }
        }
    }
}
```

## 🚨 Error Handling

### Error Types

```lua
-- API errors
{
    type = "api_error",
    message = "Failed to list pods",
    code = 403,
    details = "Access denied"
}

-- Validation errors
{
    type = "validation_error",
    message = "Invalid field value",
    field = "name",
    value = "invalid-name"
}

-- Network errors
{
    type = "network_error",
    message = "Connection timeout",
    url = "https://api.example.com",
    timeout = 30
}
```

### Error Handling Pattern

```lua
local success, result = pcall(function()
    return k8s.list_resources("pods", "default")
end)

if not success then
    k8s.error("Failed to list pods: " .. result.message)
    return nil
end

return result
```

This comprehensive API reference provides all the tools needed to extend and customize k8s-tui functionality through plugins.