-- Example Lua Plugin for k8s-tui
-- Neovim-style plugin demonstrating Lua-based custom resource functionality

-- Plugin metadata
function Name()
    return "example-lua-plugin"
end

function Version()
    return "2.0.0"
end

function Description()
    return "Example plugin demonstrating Lua-based custom resource functionality"
end

-- Default configuration
function Config()
    return {
        enabled = true,
        show_status = true,
        status_message = "Example Plugin Active",
        refresh_interval = 10,
        demo_resources = true
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    print("Setting up Example Plugin with options:")
    for k, v in pairs(opts) do
        print("  " .. k .. " = " .. tostring(v))
    end

    -- Store configuration
    config = opts

    -- Register resources if enabled
    if config.demo_resources then
        registerResources()
    end

    -- Setup UI if enabled
    if config.show_status then
        setupUIComponents()
    end

    return nil
end

-- Initialize the plugin
function Initialize()
    print("Example Lua plugin initialized")

    -- Set initial status
    k8s_tui.set_status("Example plugin ready")

    return nil
end

-- Shutdown the plugin
function Shutdown()
    print("Example Lua plugin shutting down")
    return nil
end

-- Register custom resources with the application
function registerResources()
    -- This would register resources with the k8s_tui API
    -- For now, we'll keep the legacy functions for backward compatibility
    print("Registered example resources")
end

-- Setup UI components using the new API
function setupUIComponents()
    if config.status_message then
        k8s_tui.add_header(config.status_message)
    end
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "example:status",
            description = "Show example plugin status"
        },
        {
            name = "example:resources",
            description = "List example resources"
        },
        {
            name = "example:config",
            description = "Show current configuration"
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
    print("Example Plugin: App started event received")
    k8s_tui.set_status("Example plugin initialized")
end

function on_namespace_changed(data)
    print("Example Plugin: Namespace changed to " .. data)
    k8s_tui.set_status("Example plugin active in namespace: " .. data)
end

-- Legacy function for backward compatibility (will be removed)
function GetResourceTypes()
    if not config.demo_resources then
        return {}
    end

    return {
        {
            Name = "ExampleResources",
            Type = "exampleresource",
            Icon = "󰐧",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.30, 0.30, 0.19, 0.15},
                },
            },
            RefreshIntervalSeconds = config.refresh_interval or 10,
            Namespaced = true,
            Category = "Examples",
            Description = "Demonstrates custom resource functionality"
        }
    }
end

-- Fetch resource data
function GetResourceData(resourceType, namespace)
    if resourceType ~= "exampleresource" then
        return nil, "unsupported resource type: " .. resourceType
    end

    -- Return example data
    return {
        {
            Name = "example-resource-1",
            Namespace = namespace,
            Status = "Running",
            Age = "8m",
        },
        {
            Name = "example-resource-2",
            Namespace = namespace,
            Status = "Pending",
            Age = "2m",
        },
    }, nil
end

-- Delete a resource
function DeleteResource(resourceType, namespace, name)
    if resourceType ~= "exampleresource" then
        return "unsupported resource type: " .. resourceType
    end

    return nil
end

-- Get resource information
function GetResourceInfo(resourceType, namespace, name)
    if resourceType ~= "exampleresource" then
        return nil, "unsupported resource type: " .. resourceType
    end

    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "5m",
    }, nil
end

-- Legacy function for backward compatibility (will be removed)
function GetUIExtensions()
    if not config.show_status then
        return {}
    end

    return {
        {
            Name = "example-status",
            Type = "ui_injection",
            InjectionPoints = {
                {
                    Location = "header",
                    Position = "right",
                    Priority = 10,
                    Component = {
                        Type = "text",
                        Config = {
                            content = config.status_message or "Example Plugin Active",
                            style = "success"
                        },
                        Style = {
                            ForegroundColor = "#00FF00",
                            BackgroundColor = "#000000"
                        }
                    },
                    DataSource = "static",
                    UpdateInterval = 0
                }
            }
        }
    }
end
