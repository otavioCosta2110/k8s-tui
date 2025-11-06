# 🔧 Extension Points

This document describes the various extension points available in k8s-tui for customizing and extending functionality.

## 🎯 Overview

k8s-tui provides multiple extension points that allow plugins to integrate deeply with the application:

- **Resource Handlers** - Add support for new Kubernetes resource types
- **UI Components** - Create custom user interface components
- **Event Hooks** - Respond to application events
- **Commands** - Add custom commands and actions
- **Themes** - Customize appearance and styling
- **Data Sources** - Integrate external data sources
- **Validators** - Add custom validation logic

## 📋 Resource Handlers

Resource handlers allow plugins to add support for custom Kubernetes resources or extend existing ones.

### Handler Interface

```lua
handler = {
    -- Basic information
    name = "custom-resource",
    display_name = "Custom Resources",
    description = "Custom Kubernetes resources",
    icon = "🔧",
    category = "workloads",
    
    -- Resource type information
    api_version = "custom.example.com/v1",
    kind = "CustomResource",
    plural = "customresources",
    singular = "customresource",
    
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
    restart = function(name, namespace) end,
    port_forward = function(name, namespace, local_port, remote_port) end,
    
    -- UI customization
    list_columns = function() end,
    detail_sections = function() end,
    form_fields = function() end,
    
    -- Lifecycle hooks
    on_list = function(resources) end,
    on_get = function(resource) end,
    on_create = function(resource) end,
    on_update = function(resource) end,
    on_delete = function(name, namespace) end
}
```

### Example: Custom Resource Handler

```lua
local database_handler = {
    name = "database",
    display_name = "Databases",
    icon = "🗄️",
    api_version = "db.example.com/v1",
    kind = "Database",
    plural = "databases",
    
    -- List databases
    list = function(namespace, options)
        local databases = k8s.list_resources("databases", namespace, options)
        local rows = {}
        
        for _, db in ipairs(databases) do
            table.insert(rows, {
                db.metadata.name,
                db.spec.engine,
                db.spec.version,
                db.status.phase,
                k8s.format_time(db.metadata.creationTimestamp)
            })
        end
        
        return {
            columns = {"Name", "Engine", "Version", "Status", "Age"},
            rows = rows
        }
    end,
    
    -- Get database details
    get = function(name, namespace)
        return k8s.get_resource("databases", name, namespace)
    end,
    
    -- Create database
    create = function(spec)
        local db_spec = {
            apiVersion = "db.example.com/v1",
            kind = "Database",
            metadata = {
                name = spec.name,
                namespace = spec.namespace
            },
            spec = {
                engine = spec.engine,
                version = spec.version,
                size = spec.size,
                storage = spec.storage
            }
        }
        return k8s.create_resource(db_spec)
    end,
    
    -- Update database
    update = function(name, namespace, spec)
        local db = k8s.get_resource("databases", name, namespace)
        db.spec = spec
        return k8s.update_resource(db)
    end,
    
    -- Delete database
    delete = function(name, namespace)
        return k8s.delete_resource("databases", name, namespace)
    end,
    
    -- Custom list columns
    list_columns = function()
        return {"Name", "Engine", "Version", "Status", "Age"}
    end,
    
    -- Custom detail sections
    detail_sections = function(resource)
        return {
            {
                title = "Basic Information",
                content = {
                    {"Name", resource.metadata.name},
                    {"Namespace", resource.metadata.namespace},
                    {"Engine", resource.spec.engine},
                    {"Version", resource.spec.version},
                    {"Status", resource.status.phase}
                }
            },
            {
                title = "Configuration",
                content = {
                    {"Size", resource.spec.size},
                    {"Storage", resource.spec.storage},
                    {"Replicas", resource.spec.replicas}
                }
            },
            {
                title = "Connection",
                content = {
                    {"Host", resource.status.host},
                    {"Port", resource.status.port},
                    {"Username", resource.status.username}
                }
            }
        }
    end,
    
    -- Custom form fields
    form_fields = function()
        return {
            {
                name = "name",
                label = "Database Name",
                type = "text",
                required = true,
                validation = "^[a-z0-9-]+$"
            },
            {
                name = "engine",
                label = "Database Engine",
                type = "select",
                options = {"postgresql", "mysql", "mongodb"},
                required = true
            },
            {
                name = "version",
                label = "Version",
                type = "select",
                options = {"12", "13", "14", "15"},
                default = "14"
            },
            {
                name = "size",
                label = "Size",
                type = "select",
                options = {"small", "medium", "large"},
                default = "medium"
            },
            {
                name = "storage",
                label = "Storage (GB)",
                type = "number",
                default = 10,
                validation = "^[1-9][0-9]*$"
            }
        }
    end
}

-- Register the handler
k8s.register_resource_handler(database_handler)
```

## 🎨 UI Components

Create custom UI components that integrate with the k8s-tui interface.

### Component Interface

```lua
component = {
    type = "custom-component",
    name = "my-component",
    
    -- Lifecycle methods
    init = function(config) end,
    render = function(state) end,
    update = function(event, state) end,
    cleanup = function() end,
    
    -- Event handling
    handle_key = function(key, state) end,
    handle_mouse = function(event, state) end,
    
    -- State management
    get_state = function() end,
    set_state = function(state) end,
    
    -- Configuration
    configure = function(config) end,
    get_config = function() end
}
```

### Example: Custom Metrics Component

```lua
local metrics_component = {
    type = "metrics-dashboard",
    name = "metrics-dashboard",
    
    -- Initialize component
    init = function(config)
        return {
            refresh_interval = config.refresh_interval or 30,
            metrics = {},
            last_update = 0
        }
    end,
    
    -- Render component
    render = function(state)
        local content = "📊 Metrics Dashboard\n\n"
        
        -- CPU metrics
        content = content .. "CPU Usage:\n"
        for _, metric in ipairs(state.metrics.cpu or {}) do
            content = content .. string.format("  %s: %.1f%%\n", metric.pod, metric.usage)
        end
        
        content = content .. "\nMemory Usage:\n"
        for _, metric in ipairs(state.metrics.memory or {}) do
            content = content .. string.format("  %s: %s\n", metric.pod, k8s.format_bytes(metric.usage))
        end
        
        content = content .. string.format("\nLast updated: %s", 
            k8s.format_time(state.last_update))
        
        return content
    end,
    
    -- Update component
    update = function(event, state)
        if event.type == "timer" or event.type == "refresh" then
            -- Fetch metrics
            local pods = k8s.list_resources("pods", "default")
            local cpu_metrics = {}
            local memory_metrics = {}
            
            for _, pod in ipairs(pods) do
                if pod.status.phase == "Running" then
                    -- Get metrics from metrics server
                    local metrics = k8s.get_pod_metrics(pod.metadata.name, pod.metadata.namespace)
                    if metrics then
                        table.insert(cpu_metrics, {
                            pod = pod.metadata.name,
                            usage = metrics.cpu_usage
                        })
                        table.insert(memory_metrics, {
                            pod = pod.metadata.name,
                            usage = metrics.memory_usage
                        })
                    end
                end
            end
            
            state.metrics.cpu = cpu_metrics
            state.metrics.memory = memory_metrics
            state.last_update = k8s.get_timestamp()
        end
        
        return state
    end,
    
    -- Handle key events
    handle_key = function(key, state)
        if key == "r" or key == "R" then
            -- Refresh metrics
            return {type = "refresh"}
        elseif key == "q" or key == "Q" then
            -- Quit component
            return {type = "quit"}
        end
        return nil
    end
}

-- Register component
k8s.register_component(metrics_component)
```

## 🎣 Event Hooks

Respond to application events and add custom behavior.

### Available Events

```lua
-- Application lifecycle
"on_startup"           -- Application starts
"on_shutdown"          -- Application shuts down
"on_config_change"     -- Configuration changes

-- Navigation
"on_tab_change"        -- Tab changes
"on_screen_change"     -- Screen changes
"on_cluster_change"    -- Cluster changes
"on_namespace_change"  -- Namespace changes

-- Resource operations
"on_resource_list"     -- After listing resources
"on_resource_select"   -- When resource is selected
"on_resource_create"   -- After resource creation
"on_resource_update"   -- After resource update
"on_resource_delete"   -- After resource deletion

-- UI events
"on_key_press"         -- Key press events
"on_search"            -- Search operations
"on_filter"            -- Filter operations
"on_sort"              -- Sort operations

-- Custom events
"on_timer"             -- Timer events
"on_notification"      -- Notification events
```

### Example: Event-Driven Plugin

```lua
local event_plugin = {
    name = "event-handler",
    
    init = function()
        -- Register event handlers
        k8s.register_hook("on_resource_select", self.on_resource_select)
        k8s.register_hook("on_resource_create", self.on_resource_create)
        k8s.register_hook("on_key_press", self.on_key_press)
        k8s.register_hook("on_timer", self.on_timer)
        
        -- Start timer
        self.timer_id = k8s.start_timer(60, "cleanup_timer")
    end,
    
    -- Handle resource selection
    on_resource_select = function(resource)
        if resource.kind == "Pod" then
            -- Log pod selection
            k8s.log("Selected pod: " .. resource.metadata.name)
            
            -- Check pod health
            if resource.status.phase == "Failed" then
                k8s.show_warning("Selected pod is in failed state")
            end
            
            -- Emit custom event
            k8s.emit_event("pod_selected", {
                pod = resource.metadata.name,
                namespace = resource.metadata.namespace
            })
        end
    end,
    
    -- Handle resource creation
    on_resource_create = function(resource)
        -- Send notification
        k8s.show_success("Created " .. resource.kind .. ": " .. resource.metadata.name)
        
        -- Log to external system
        k8s.http_post("https://api.example.com/notifications", {
            type = "resource_created",
            resource = {
                kind = resource.kind,
                name = resource.metadata.name,
                namespace = resource.metadata.namespace
            },
            timestamp = k8s.get_timestamp()
        })
    end,
    
    -- Handle key presses
    on_key_press = function(key)
        if key == "F12" then
            -- Show custom menu
            self.show_custom_menu()
        elseif key == "F11" then
            -- Export current view
            self.export_current_view()
        end
    end,
    
    -- Handle timer events
    on_timer = function(timer_id)
        if timer_id == self.timer_id then
            -- Perform cleanup
            self.perform_cleanup()
        end
    end,
    
    -- Custom methods
    show_custom_menu = function()
        local menu = k8s.create_menu({
            {title = "Export Data", action = "export"},
            {title = "Refresh All", action = "refresh"},
            {title = "Settings", action = "settings"},
            {title = "About", action = "about"}
        })
        
        local choice = menu:show()
        if choice then
            self.handle_menu_choice(choice)
        end
    end,
    
    export_current_view = function()
        local current_view = k8s.get_current_view()
        local filename = "k8s-tui-export-" .. os.time() .. ".json"
        
        k8s.write_file(filename, k8s.json_encode(current_view))
        k8s.show_success("Exported to " .. filename)
    end,
    
    perform_cleanup = function()
        -- Clean up old logs
        k8s.log("Performing periodic cleanup")
        
        -- Remove old temporary files
        local temp_files = k8s.list_temp_files()
        for _, file in ipairs(temp_files) do
            if file.age > 3600 then  -- 1 hour
                k8s.remove_temp_file(file.path)
            end
        end
    end,
    
    cleanup = function()
        -- Stop timer
        if self.timer_id then
            k8s.stop_timer(self.timer_id)
        end
        
        -- Unregister hooks
        k8s.unregister_hook("on_resource_select", self.on_resource_select)
        k8s.unregister_hook("on_resource_create", self.on_resource_create)
        k8s.unregister_hook("on_key_press", self.on_key_press)
        k8s.unregister_hook("on_timer", self.on_timer)
    end
}
```

## ⌨️ Commands

Add custom commands to extend functionality.

### Command Interface

```lua
command = {
    name = "command-name",
    description = "Command description",
    usage = "command [args]",
    aliases = {"cmd", "c"},
    handler = function(args, options) end,
    completion = function(partial) end,
    validation = function(args) end
}
```

### Example: Custom Commands

```lua
local custom_commands = {
    -- Export command
    {
        name = "export",
        description = "Export resources to file",
        usage = "export <resource-type> [filename]",
        aliases = {"exp", "e"},
        handler = function(args, options)
            if #args < 1 then
                k8s.show_error("Usage: export <resource-type> [filename]")
                return
            end
            
            local resource_type = args[1]
            local filename = args[2] or resource_type .. "-" .. os.time() .. ".json"
            
            local resources = k8s.list_resources(resource_type)
            k8s.write_file(filename, k8s.json_encode(resources))
            k8s.show_success("Exported " .. #resources .. " " .. resource_type .. " to " .. filename)
        end,
        
        completion = function(partial)
            local resource_types = {"pods", "deployments", "services", "configmaps"}
            local matches = {}
            
            for _, type in ipairs(resource_types) do
                if string.find(type, partial, 1, true) == 1 then
                    table.insert(matches, type)
                end
            end
            
            return matches
        end
    },
    
    -- Batch operation command
    {
        name = "batch",
        description = "Perform batch operations",
        usage = "batch <operation> <resource-type> <selector>",
        aliases = {"b"},
        handler = function(args, options)
            if #args < 3 then
                k8s.show_error("Usage: batch <operation> <resource-type> <selector>")
                return
            end
            
            local operation = args[1]
            local resource_type = args[2]
            local selector = args[3]
            
            local resources = k8s.list_resources(resource_type, "default", {
                label_selector = selector
            })
            
            local success_count = 0
            local error_count = 0
            
            for _, resource in ipairs(resources) do
                local success = false
                
                if operation == "delete" then
                    success = k8s.delete_resource(resource_type, resource.metadata.name, resource.metadata.namespace)
                elseif operation == "restart" then
                    success = k8s.restart_resource(resource_type, resource.metadata.name, resource.metadata.namespace)
                end
                
                if success then
                    success_count = success_count + 1
                else
                    error_count = error_count + 1
                end
            end
            
            k8s.show_message(string.format("Batch %s completed: %d success, %d errors", 
                operation, success_count, error_count))
        end
    },
    
    -- Integration command
    {
        name = "notify",
        description = "Send notifications",
        usage = "notify <message> [level]",
        aliases = {"n"},
        handler = function(args, options)
            if #args < 1 then
                k8s.show_error("Usage: notify <message> [level]")
                return
            end
            
            local message = table.concat(args, " ", 1, #args - (args[2] and 1 or 0))
            local level = args[#args] or "info"
            
            -- Send to external notification service
            k8s.http_post("https://notify.example.com/api/notify", {
                message = message,
                level = level,
                source = "k8s-tui",
                timestamp = k8s.get_timestamp()
            })
            
            k8s.show_success("Notification sent")
        end
    }
}

-- Register commands
for _, command in ipairs(custom_commands) do
    k8s.register_command(command.name, command.handler, {
        description = command.description,
        usage = command.usage,
        aliases = command.aliases,
        completion = command.completion
    })
end
```

## 🎨 Themes

Customize the appearance of k8s-tui.

### Theme Structure

```lua
theme = {
    name = "theme-name",
    description = "Theme description",
    
    -- Color palette
    colors = {
        background = "#000000",
        foreground = "#ffffff",
        primary = "#00ff00",
        secondary = "#0000ff",
        accent = "#ff00ff",
        error = "#ff0000",
        warning = "#ffff00",
        success = "#00ff00",
        info = "#0088ff",
        
        -- Status colors
        running = "#00ff00",
        pending = "#ffff00",
        failed = "#ff0000",
        unknown = "#888888",
        
        -- Resource colors
        pod = "#00ff00",
        deployment = "#0088ff",
        service = "#ff00ff",
        configmap = "#ffff00",
        secret = "#ff8800"
    },
    
    -- Styling
    styles = {
        border = "single",           -- single, double, rounded, none
        padding = 1,                  -- padding size
        margin = 0,                   -- margin size
        spacing = 1,                  -- element spacing
        
        -- Component styles
        table = {
            header_style = "bold",
            row_style = "normal",
            selected_style = "reverse",
            border_style = "dim"
        },
        
        form = {
            label_style = "bold",
            input_style = "normal",
            error_style = "red",
            help_style = "dim"
        },
        
        statusbar = {
            style = "reverse",
            position = "bottom"       -- top, bottom
        }
    },
    
    -- Icons
    icons = {
        pod = "📦",
        deployment = "🚀",
        service = "🌐",
        configmap = "📄",
        secret = "🔐",
        ingress = "🌍",
        node = "🖥️",
        namespace = "📁",
        
        status = {
            running = "✅",
            pending = "⏳",
            failed = "❌",
            unknown = "❓"
        }
    }
}
```

### Example: Custom Theme

```lua
local dark_theme = {
    name = "dark-pro",
    description = "Professional dark theme",
    
    colors = {
        background = "#1a1a1a",
        foreground = "#e0e0e0",
        primary = "#00d4aa",
        secondary = "#7c3aed",
        accent = "#f59e0b",
        error = "#ef4444",
        warning = "#f59e0b",
        success = "#10b981",
        info = "#3b82f6",
        
        running = "#10b981",
        pending = "#f59e0b",
        failed = "#ef4444",
        unknown = "#6b7280",
        
        pod = "#10b981",
        deployment = "#3b82f6",
        service = "#8b5cf6",
        configmap = "#f59e0b",
        secret = "#ef4444"
    },
    
    styles = {
        border = "rounded",
        padding = 1,
        margin = 0,
        spacing = 1,
        
        table = {
            header_style = "bold",
            row_style = "normal",
            selected_style = "reverse",
            border_style = "dim"
        },
        
        form = {
            label_style = "bold",
            input_style = "normal",
            error_style = "red",
            help_style = "dim"
        },
        
        statusbar = {
            style = "reverse",
            position = "bottom"
        }
    },
    
    icons = {
        pod = "🟢",
        deployment = "🚀",
        service = "🌐",
        configmap = "📋",
        secret = "🔒",
        ingress = "🌍",
        node = "🖥️",
        namespace = "📁",
        
        status = {
            running = "✅",
            pending = "⏳",
            failed = "❌",
            unknown = "❓"
        }
    }
}

-- Register theme
k8s.register_theme(dark_theme)
```

## 📊 Data Sources

Integrate external data sources with k8s-tui.

### Data Source Interface

```lua
data_source = {
    name = "data-source-name",
    description = "Data source description",
    
    -- Data retrieval
    fetch = function(query, options) end,
    stream = function(query, callback, options) end,
    
    -- Metadata
    get_schema = function() end,
    get_capabilities = function() end,
    
    -- Configuration
    configure = function(config) end,
    test_connection = function() end
}
```

### Example: Monitoring Data Source

```lua
local monitoring_source = {
    name = "prometheus",
    description = "Prometheus monitoring data",
    
    -- Fetch metrics from Prometheus
    fetch = function(query, options)
        local url = "http://prometheus:9090/api/v1/query"
        local params = "query=" .. k8s.url_encode(query)
        
        local response = k8s.http_get(url .. "?" .. params)
        if response.status ~= 200 then
            return nil, "Failed to fetch metrics: " .. response.status
        end
        
        local data = k8s.json_decode(response.body)
        if data.status ~= "success" then
            return nil, "Prometheus query failed: " .. (data.error or "Unknown error")
        end
        
        return data.data.result
    end,
    
    -- Stream metrics
    stream = function(query, callback, options)
        local url = "http://prometheus:9090/api/v1/query_range"
        local params = string.format(
            "query=%s&start=%d&end=%d&step=%d",
            k8s.url_encode(query),
            options.start_time or (k8s.get_timestamp() - 300),
            options.end_time or k8s.get_timestamp(),
            options.step or 15
        )
        
        local response = k8s.http_get(url .. "?" .. params)
        if response.status == 200 then
            local data = k8s.json_decode(response.body)
            if data.status == "success" then
                callback(data.data.result)
            end
        end
    end,
    
    -- Get available metrics schema
    get_schema = function()
        return {
            metrics = {
                "cpu_usage",
                "memory_usage",
                "network_throughput",
                "disk_io",
                "request_rate",
                "error_rate"
            },
            labels = {
                "pod",
                "namespace",
                "deployment",
                "service",
                "node"
            }
        }
    end,
    
    -- Test connection
    test_connection = function()
        local response = k8s.http_get("http://prometheus:9090/api/v1/status/config")
        return response.status == 200
    end
}

-- Register data source
k8s.register_data_source(monitoring_source)
```

## ✅ Validators

Add custom validation logic for resources and forms.

### Validator Interface

```lua
validator = {
    name = "validator-name",
    description = "Validator description",
    
    -- Validation methods
    validate_resource = function(resource, context) end,
    validate_field = function(field, value, context) end,
    validate_form = function(form_data, context) end,
    
    -- Configuration
    get_rules = function() end,
    add_rule = function(rule) end
}
```

### Example: Custom Validators

```lua
local custom_validators = {
    -- Resource naming validator
    {
        name = "naming-convention",
        description = "Enforce naming conventions",
        
        validate_resource = function(resource, context)
            local errors = {}
            
            -- Validate resource name
            if not string.match(resource.metadata.name, "^[a-z0-9-]+$") then
                table.insert(errors, {
                    field = "metadata.name",
                    message = "Name must contain only lowercase letters, numbers, and hyphens"
                })
            end
            
            -- Validate namespace
            if resource.metadata.namespace == "default" then
                table.insert(errors, {
                    field = "metadata.namespace",
                    message = "Avoid using default namespace"
                })
            end
            
            return #errors == 0, errors
        end,
        
        validate_field = function(field, value, context)
            if field == "image" and type(value) == "string" then
                -- Validate image tag
                if not string.find(value, ":", 1, true) then
                    return false, "Image should specify a tag"
                end
                
                -- Validate image registry
                local allowed_registries = {"docker.io", "gcr.io", "quay.io"}
                local registry = string.match(value, "^([^/]+)")
                if registry and not k8s.contains(allowed_registries, registry) then
                    return false, "Registry not in allowed list"
                end
            end
            
            return true, nil
        end
    },
    
    -- Security validator
    {
        name = "security-policy",
        description = "Security policy validation",
        
        validate_resource = function(resource, context)
            local errors = {}
            
            if resource.kind == "Pod" or resource.kind == "Deployment" then
                local containers = resource.spec.containers or {}
                
                for _, container in ipairs(containers) do
                    -- Check for privileged containers
                    if container.securityContext and container.securityContext.privileged then
                        table.insert(errors, {
                            field = "spec.containers." .. container.name .. ".securityContext.privileged",
                            message = "Privileged containers are not allowed"
                        })
                    end
                    
                    -- Check for root containers
                    if container.securityContext and 
                       container.securityContext.runAsUser == 0 then
                        table.insert(errors, {
                            field = "spec.containers." .. container.name .. ".securityContext.runAsUser",
                            message = "Running as root is not allowed"
                        })
                    end
                    
                    -- Check for host network
                    if container.securityContext and 
                       container.securityContext.hostNetwork then
                        table.insert(errors, {
                            field = "spec.containers." .. container.name .. ".securityContext.hostNetwork",
                            message = "Host network is not allowed"
                        })
                    end
                end
            end
            
            return #errors == 0, errors
        end
    }
}

-- Register validators
for _, validator in ipairs(custom_validators) do
    k8s.register_validator(validator)
end
```

## 🔧 Best Practices

### Extension Development

1. **Follow naming conventions** - Use consistent naming for extensions
2. **Handle errors gracefully** - Provide meaningful error messages
3. **Validate inputs** - Validate all user inputs and API responses
4. **Use async operations** - Don't block the UI for long operations
5. **Provide feedback** - Show progress and status to users
6. **Document your extensions** - Provide clear documentation
7. **Test thoroughly** - Test in various environments and scenarios
8. **Handle permissions** - Request only necessary permissions

### Performance Considerations

1. **Cache data** - Cache frequently accessed data
2. **Limit API calls** - Minimize unnecessary API requests
3. **Use streaming** - Use streaming for real-time data
4. **Optimize rendering** - Efficient UI rendering
5. **Background processing** - Use background tasks for heavy operations

### Security Considerations

1. **Validate inputs** - Sanitize all user inputs
2. **Secure communications** - Use HTTPS for external calls
3. **Handle secrets** - Don't expose sensitive data
4. **Principle of least privilege** - Request minimal permissions
5. **Audit logging** - Log important operations

These extension points provide powerful ways to customize and extend k8s-tui functionality while maintaining stability and performance.