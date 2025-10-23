# Extension Points

k8s-tui provides multiple extension points for developers to customize and extend functionality without modifying core code.

## Plugin System

### Lua Plugin Architecture

k8s-tui uses Lua as its plugin language, providing a safe, sandboxed environment for extensions.

#### Plugin Loading Process

1. **Discovery**: Scan `plugins/` directory for `main.lua` files
2. **Validation**: Check plugin metadata and permissions
3. **Isolation**: Load plugin in separate Lua state
4. **Initialization**: Call plugin `init()` function
5. **Registration**: Register plugin hooks and commands

#### Plugin Lifecycle

```lua
local plugin = {}

function plugin.init()
    -- Plugin initialization
    log.info("Plugin loaded")
end

function plugin.cleanup()
    -- Plugin cleanup
    log.info("Plugin unloaded")
end

return plugin
```

### Hook System

Plugins can hook into k8s-tui events:

#### UI Hooks

```lua
function plugin.on_tab_changed(tab_name)
    -- React to tab switches
end

function plugin.on_resource_selected(resource_type, name, namespace)
    -- React to resource selection
end

function plugin.on_form_submitted(form_data)
    -- React to form submissions
end
```

#### Resource Hooks

```lua
function plugin.on_resource_created(resource_type, name, namespace)
    -- React to resource creation
end

function plugin.on_resource_updated(resource_type, name, namespace)
    -- React to resource updates
end

function plugin.on_resource_deleted(resource_type, name, namespace)
    -- React to resource deletion
end
```

#### System Hooks

```lua
function plugin.on_startup()
    -- Application startup
end

function plugin.on_shutdown()
    -- Application shutdown
end

function plugin.on_config_changed()
    -- Configuration changes
end
```

## Custom Resource Types

### CRD Support

k8s-tui can be extended to support Custom Resource Definitions:

```lua
-- Register custom resource type
resources.register("my-crd", {
    api_version = "mycompany.com/v1",
    kind = "MyResource",

    -- List view configuration
    columns = {
        {title = "Name", width = 0.3},
        {title = "Status", width = 0.2},
        {title = "Age", width = 0.2}
    },

    -- Detail view
    detail_view = function(resource)
        return {
            "Name: " .. resource.metadata.name,
            "Status: " .. resource.status.phase,
            "Created: " .. resource.metadata.creationTimestamp
        }
    end,

    -- Actions
    actions = {
        "scale",
        "restart",
        "logs"
    }
})
```

### Custom Views

Create entirely custom views:

```lua
views.register("cluster-overview", {
    title = "Cluster Overview",
    icon = "📊",

    render = function()
        local nodes = k8s.list_nodes()
        local pods = k8s.list_pods("")
        local deployments = k8s.list_deployments("")

        return ui.panel({
            title = "Cluster Status",
            content = {
                "Nodes: " .. #nodes .. " total",
                "Pods: " .. #pods .. " running",
                "Deployments: " .. #deployments .. " active",
                "",
                "Cluster Health: " .. calculate_health(nodes, pods)
            }
        })
    end,

    refresh_interval = 30  -- seconds
})
```

## UI Extensions

### Custom Components

Create reusable UI components:

```lua
components.register("status-badge", function(status)
    local color = status == "healthy" and "green" or "red"
    return ui.badge({
        text = status,
        color = color,
        style = "rounded"
    })
end)

-- Usage in views
local badge = components.render("status-badge", "healthy")
```

### Theme Extensions

Add custom themes or modify existing ones:

```lua
themes.register("my-theme", {
    name = "My Custom Theme",
    colors = {
        background = "#000000",
        foreground = "#ffffff",
        accent = "#ff6b6b",
        border = "#333333",
        text = "#ffffff",
        help_text = "#888888",
        error = "#ff4757",
        success = "#2ed573",
        warning = "#ffa502"
    }
})
```

### Key Binding Extensions

Add custom key bindings:

```lua
keys.register("ctrl+shift+r", function()
    -- Custom action
    ui.notify("Custom shortcut activated!")
end, {
    description = "Custom refresh action",
    contexts = {"resource-list", "resource-detail"}
})
```

## Data Processing Extensions

### Custom Formatters

Add custom data formatters:

```lua
formatters.register("custom-time", function(timestamp)
    -- Custom time formatting
    return os.date("%Y-%m-%d %H:%M:%S", timestamp)
end)

-- Usage
local formatted = formatters.format("custom-time", resource.creationTimestamp)
```

### Data Transformers

Transform resource data:

```lua
transformers.register("pod-status", function(pod)
    return {
        name = pod.metadata.name,
        status = pod.status.phase,
        ready = pod.status.containerStatuses[1].ready,
        restarts = pod.status.containerStatuses[1].restartCount,
        age = formatters.relative_time(pod.metadata.creationTimestamp)
    }
end)
```

## Integration Extensions

### External Tool Integration

Integrate with external tools:

```lua
integrations.register("kubectl", {
    commands = {
        "get",
        "describe",
        "logs",
        "exec"
    },

    execute = function(command, args)
        local cmd = "kubectl " .. command .. " " .. table.concat(args, " ")
        return os.execute(cmd)
    end
})

-- Usage
integrations.call("kubectl", "logs", {"-f", "pod-name"})
```

### API Integrations

Connect to external APIs:

```lua
apis.register("monitoring", {
    base_url = "https://monitoring.example.com/api/v1",

    endpoints = {
        metrics = "/metrics",
        alerts = "/alerts"
    },

    auth = {
        type = "bearer",
        token = os.getenv("MONITORING_TOKEN")
    }
})

-- Usage
local metrics = apis.call("monitoring", "metrics", {query = "up"})
```

## Storage Extensions

### Custom Storage Backends

Implement custom storage for configurations, caches, etc.:

```lua
storage.register("redis", {
    connect = function(config)
        return redis.connect(config.host, config.port)
    end,

    get = function(key)
        return redis_client:get(key)
    end,

    set = function(key, value, ttl)
        return redis_client:setex(key, ttl, value)
    end,

    delete = function(key)
        return redis_client:del(key)
    end
})
```

### Cache Extensions

Custom caching strategies:

```lua
cache.register("smart", {
    should_cache = function(resource_type, namespace, name)
        -- Custom logic for what to cache
        return resource_type == "configmap" or namespace == "kube-system"
    end,

    ttl = function(resource_type)
        -- Custom TTL logic
        if resource_type == "node" then
            return 300  -- 5 minutes
        else
            return 60   -- 1 minute
        end
    end,

    invalidate = function(resource_type, name, namespace)
        -- Custom invalidation logic
        if resource_type == "deployment" then
            -- Invalidate related pods
            cache.invalidate_pattern("pod:" .. namespace .. ":*")
        end
    end
})
```

## Workflow Extensions

### Custom Workflows

Define multi-step workflows:

```lua
workflows.register("deploy-app", {
    title = "Deploy Application",
    description = "Complete application deployment workflow",

    steps = {
        {
            name = "create-namespace",
            title = "Create Namespace",
            form = {
                fields = {
                    {name = "name", placeholder = "Namespace name"}
                }
            }
        },
        {
            name = "create-configmap",
            title = "Create ConfigMap",
            form = {
                fields = {
                    {name = "name", placeholder = "ConfigMap name"},
                    {name = "data", placeholder = "Configuration data", type = "textarea"}
                }
            }
        },
        {
            name = "deploy-app",
            title = "Deploy Application",
            form = {
                fields = {
                    {name = "name", placeholder = "Deployment name"},
                    {name = "image", placeholder = "Container image"},
                    {name = "replicas", placeholder = "Number of replicas"}
                }
            }
        }
    },

    execute = function(data)
        -- Execute workflow steps
        k8s.create_namespace(data["create-namespace"].name)
        k8s.create_configmap(data["create-configmap"].name, data["create-configmap"].data)
        k8s.create_deployment(data["deploy-app"].name, data["deploy-app"].image, data["deploy-app"].replicas)
    end
})
```

## Security Extensions

### Authentication Providers

Add custom authentication methods:

```lua
auth.register("ldap", {
    name = "LDAP Authentication",

    authenticate = function(username, password)
        -- LDAP authentication logic
        return ldap_auth(username, password)
    end,

    get_user_info = function(token)
        -- Get user information from token
        return ldap_get_user(token)
    end
})
```

### Authorization Extensions

Custom permission checking:

```lua
authz.register("rbac-plus", {
    check_permission = function(user, action, resource)
        -- Enhanced RBAC logic
        local base_check = rbac_check(user, action, resource)
        if not base_check then
            return false
        end

        -- Additional checks
        return custom_policy_check(user, resource)
    end
})
```

## Testing Extensions

### Test Helpers

Provide testing utilities for plugins:

```lua
testing.register_helper("mock_k8s", function()
    return {
        list_pods = function() return {} end,
        get_pod = function() return {status = {phase = "Running"}} end,
        create_pod = function() return true end
    }
end)

testing.register_helper("mock_ui", function()
    return {
        notify = function(msg) print("NOTIFY: " .. msg) end,
        confirm = function(msg) return true end
    }
end)
```

## Extension Development Workflow

1. **Identify Extension Point**: Determine which extension mechanism to use
2. **Create Plugin Structure**: Set up plugin directory and files
3. **Implement Extension**: Write the extension logic
4. **Add Tests**: Create unit and integration tests
5. **Document Extension**: Update documentation
6. **Submit PR**: Contribute back to the project

## Best Practices

### Performance

- Avoid blocking operations in UI hooks
- Use asynchronous operations when possible
- Cache expensive computations
- Minimize DOM manipulations

### Security

- Validate all inputs
- Use secure communication protocols
- Implement proper error handling
- Follow principle of least privilege

### Compatibility

- Check version compatibility
- Handle API changes gracefully
- Provide fallbacks for missing features
- Document breaking changes

### User Experience

- Provide clear error messages
- Include progress indicators for long operations
- Offer undo functionality where appropriate
- Maintain consistent UI patterns

These extension points make k8s-tui highly customizable while maintaining stability and security.</content>
</xai:function_call">  

Assistant: Now update the main README to link to the docs. 

<xai:function_call name="edit">
<parameter name="filePath">README.md