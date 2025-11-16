-- Example Lua Plugin for k8s-tui
-- Demonstrates custom resource type functionality with Global Plugin Manager

-- Global configuration variable
local config = {}

-- Plugin metadata
function Name()
    return "example-custom-resource-plugin"
end

function Version()
    return "2.1.0"
end

function Description()
    return "Example plugin that adds a custom resource type (ExampleResources) to the resource list"
end

-- Default configuration
function Config()
    return {
        enabled = true,
        show_status = true,
        status_message = "Custom Resource Plugin Active",
        refresh_interval = 10,
        demo_resources = true,
        resource_icon = "📦"
    }
end

function CLIArguments()
    return {}
end

-- Setup function (called with user configuration)
function Setup(opts)
    -- Initialize default config if opts is nil
    if not opts then
        opts = Config()
    end
    
    k8s_tui.log("Setting up Example Custom Resource Plugin with options:")
    for k, v in pairs(opts) do
        k8s_tui.log("  " .. k .. " = " .. tostring(v))
    end

    -- Store configuration
    config = opts

    -- Register custom resource type if enabled
    if config.demo_resources then
        k8s_tui.log("Registering custom resource type: ExampleResources")
        registerCustomResource()
    end

    -- Setup UI if enabled
    if config.show_status then
        setupUIComponents()
    end

    return nil
end

-- Initialize the plugin
function Initialize()
    k8s_tui.log("Example Custom Resource Plugin initialized")

    -- Set initial status
    k8s_tui.set_status("Custom resource plugin ready")

    return nil
end

-- Shutdown the plugin
function Shutdown()
    k8s_tui.log("Example Custom Resource Plugin shutting down")
    return nil
end

-- Register custom resource type
function registerCustomResource()
    -- This function sets up the custom resource type
    -- The actual registration happens through GetResourceTypes()
    k8s_tui.log("Custom resource type registration prepared")
end

-- Setup UI components using the API
function setupUIComponents()
    if config.status_message then
        k8s_tui.log("Setting up UI components: " .. config.status_message)
    end
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "custom:status",
            description = "Show custom resource plugin status",
            handler = "handle_custom_status"
        },
        {
            name = "custom:list",
            description = "List all custom resources",
            handler = "handle_custom_list"
        },
        {
            name = "custom:config",
            description = "Show current plugin configuration",
            handler = "handle_custom_config"
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
    k8s_tui.log("Custom Resource Plugin: App started event received")
    k8s_tui.set_status("Custom resource plugin initialized - ExampleResources available")
end

function on_namespace_changed(data)
    k8s_tui.log("Custom Resource Plugin: Namespace changed to " .. data)
    k8s_tui.set_status("Custom resource plugin active in namespace: " .. data)
end

-- Command handlers
function handle_custom_status(args)
    return "Custom Resource Plugin Status: Active (ExampleResources registered)", nil
end

function handle_custom_list(args)
    return "Custom Resources: Use the resource list to view ExampleResources", nil
end

function handle_custom_config(args)
    local config_str = "Plugin Configuration:\n"
    for k, v in pairs(config) do
        config_str = config_str .. "  " .. k .. " = " .. tostring(v) .. "\n"
    end
    return config_str, nil
end

-- Resource type definition function
function GetResourceTypes()
    if not config.demo_resources then
        return {}
    end

    return {
        {
            Name = "ExampleResources",
            Type = "exampleresource",
            Icon = config.resource_icon or "📦",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.25, 0.25, 0.20, 0.15, 0.15},
                },
            },
            RefreshIntervalSeconds = config.refresh_interval or 10,
            Namespaced = true,
            Category = "Custom",
            Description = "Example custom resources demonstrating plugin functionality"
        }
    }
end

-- Fetch resource data
function GetResourceData(resourceType, namespace)
    if resourceType ~= "exampleresource" then
        return nil, "unsupported resource type: " .. resourceType
    end

    k8s_tui.log("Fetching ExampleResources data for namespace: " .. namespace)

    -- Return example data with proper table structure
    return {
        {
            Name = "example-app-v1",
            Namespace = namespace,
            Status = "Running",
            Version = "v1.0.0",
            Replicas = "3",
            Age = "2h15m",
        },
        {
            Name = "example-db-v2",
            Namespace = namespace,
            Status = "Pending",
            Version = "v2.1.0",
            Replicas = "1",
            Age = "45m",
        },
        {
            Name = "example-cache",
            Namespace = namespace,
            Status = "Failed",
            Version = "v1.5.2",
            Replicas = "2",
            Age = "1h30m",
        },
        {
            Name = "example-worker",
            Namespace = namespace,
            Status = "Running",
            Version = "v3.0.1",
            Replicas = "5",
            Age = "30m",
        }
    }, nil
end

-- Delete a resource
function DeleteResource(resourceType, namespace, name)
    if resourceType ~= "exampleresource" then
        return "unsupported resource type: " .. resourceType
    end

    k8s_tui.log("Deleting ExampleResource: " .. name .. " in namespace: " .. namespace)
    
    -- Simulate deletion logic
    -- In a real plugin, you would call the Kubernetes API here
    
    return nil
end

-- Get resource information
function GetResourceInfo(resourceType, namespace, name)
    if resourceType ~= "exampleresource" then
        return nil, "unsupported resource type: " .. resourceType
    end

    k8s_tui.log("Getting info for ExampleResource: " .. name .. " in namespace: " .. namespace)

    -- Return detailed resource information
    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "1h45m",
        Version = "v2.1.0",
        Replicas = 3,
        Status = "Running",
        Created = "2025-01-15T10:30:00Z",
        Labels = {
            app = "example",
            version = "v2.1.0",
            environment = "production"
        },
        Annotations = {
            description = "Example custom resource instance",
            ["maintained-by"] = "example-team"
        }
    }, nil
end

-- UI extensions function
function GetUIExtensions()
    if not config.show_status then
        return {}
    end

    return {
        {
            Name = "custom-resource-status",
            Type = "ui_injection",
            InjectionPoints = {
                {
                    Location = "header",
                    Position = "right",
                    Priority = 10,
                    Component = {
                        Type = "text",
                        Config = {
                            content = config.status_message or "Custom Resources Active",
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