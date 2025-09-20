-- Cluster Monitor Plugin for k8s-tui
-- Neovim-style plugin demonstrating advanced UI components and data visualization

print("Cluster Monitor: Lua script loaded")

-- Plugin metadata
function Name()
    return "cluster-monitor"
end

function Version()
    return "2.0.0"
end

function Description()
    return "Monitor Kubernetes cluster status with advanced visualizations"
end

-- Default configuration
function Config()
    print("Cluster Monitor: Config() function called")
    return {
        show_cluster_health = true,
        show_node_health = true,
        refresh_interval = 30,
        enable_ui_injections = true,
        cluster_health_icon = "🟢",
        monitoring_message = "Monitoring Active"
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    print("Cluster Monitor: Setup() function called with options:")
    for k, v in pairs(opts) do
        print("  " .. k .. " = " .. tostring(v))
    end

    -- Store configuration
    config = opts

    -- Register custom resources
    registerResources()

    -- Add UI components if enabled
    if config.enable_ui_injections then
        setupUIComponents()
    end

    return nil
end

-- Initialize the plugin
function Initialize()
    print("Cluster Monitor plugin initialized")

    -- Set initial status
    k8s_tui.set_status("Cluster monitoring active")

    return nil
end

-- Shutdown the plugin
function Shutdown()
    print("Cluster Monitor plugin shutting down")
    return nil
end

-- Register custom resources with the application
function registerResources()
    -- This would register resources with the k8s_tui API
    -- For now, we'll keep the legacy functions for backward compatibility
    print("Registered cluster monitoring resources")
end

-- Legacy function for backward compatibility (will be removed)
function GetResourceTypes()
    return {
        {
            Name = "Cluster Status",
            Type = "clusterstatus",
            Icon = "󰒋",
            DisplayComponent = {
                Type = "yaml",
                Config = {
                    -- YAML configuration options can be added here
                },
                Style = {
                    Width = 80,
                    Height = 25,
                    Border = "rounded",
                    ForegroundColor = "#A1EFD3",
                    BackgroundColor = "#1E1E2E",
                    BorderColor = "#F28FAD"
                }
            },
            RefreshIntervalSeconds = config.refresh_interval or 30,
            Namespaced = false,
            Category = "Monitoring",
            Description = "Real-time cluster health and status information"
        },
        {
            Name = "Node Health",
            Type = "nodehealth",
            Icon = "󰇄",
            DisplayComponent = {
                Type = "chart",
                Config = {
                    -- Chart configuration options can be added here
                },
                Style = {
                    Width = 60,
                    Height = 15,
                    Border = "rounded",
                    ForegroundColor = "#FFFFFF",
                    BackgroundColor = "#1E1E2E",
                    BorderColor = "#CBA6F7"
                }
            },
            RefreshIntervalSeconds = 15,
            Namespaced = false,
            Category = "Monitoring",
            Description = "Node health metrics and status visualization"
        }
    }
end

-- Fetch resource data
function GetResourceData(resourceType, namespace)
    if resourceType == "clusterstatus" then
        return getClusterStatusData()
    elseif resourceType == "nodehealth" then
        return getNodeHealthData()
    else
        return nil, "unsupported resource type: " .. resourceType
    end
end

function getClusterStatusData()
    return {
        {
            Name = "cluster-info",
            Status = "Healthy",
            Age = "2h",
            api_version = "v1.28.0",
            nodes_ready = "3/3",
            pods_running = "24/25",
            services_active = "8",
            last_updated = os.date("%Y-%m-%d %H:%M:%S")
        }
    }, nil
end

function getNodeHealthData()
    return {
        {
            Name = "node-1",
            Status = "Ready",
            Age = "45d",
            cpu_usage = "65%",
            memory_usage = "72%",
            disk_usage = "45%"
        },
        {
            Name = "node-2",
            Status = "Ready",
            Age = "30d",
            cpu_usage = "58%",
            memory_usage = "68%",
            disk_usage = "52%"
        },
        {
            Name = "node-3",
            Status = "Ready",
            Age = "15d",
            cpu_usage = "71%",
            memory_usage = "75%",
            disk_usage = "38%"
        }
    }, nil
end

-- Delete a resource (not applicable for monitoring data)
function DeleteResource(resourceType, namespace, name)
    return "Monitoring resources cannot be deleted"
end

-- Get resource information
function GetResourceInfo(resourceType, namespace, name)
    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "N/A",
    }, nil
end

-- Setup UI components using the new API
function setupUIComponents()
    if config.show_cluster_health then
        local health_icon = config.cluster_health_icon or "🟢"
        k8s_tui.add_header(health_icon .. " Cluster Healthy")
    end

    if config.monitoring_message then
        k8s_tui.set_status(config.monitoring_message)
    end
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "cluster:status",
            description = "Show cluster status information"
        },
        {
            name = "cluster:health",
            description = "Check cluster health"
        },
        {
            name = "nodes:health",
            description = "Show node health information"
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
    print("Cluster Monitor: App started event received")
    k8s_tui.set_status("Cluster monitoring initialized")
end

function on_namespace_changed(data)
    print("Cluster Monitor: Namespace changed to " .. data)
    -- Could update monitoring based on namespace change
end

-- Legacy function for backward compatibility (will be removed)
function GetUIExtensions()
    return {
        {
            Name = "cluster-health-indicator",
            Type = "ui_injection",
            InjectionPoints = {
                {
                    Location = "header",
                    Position = "center",
                    Priority = 20,
                    Component = {
                        Type = "text",
                        Config = {
                            content = config.cluster_health_icon .. " Cluster Healthy",
                            style = "success"
                        },
                        Style = {
                            ForegroundColor = "#A1EFD3",
                            BackgroundColor = "#000000"
                        }
                    },
                    DataSource = "cluster_status",
                    UpdateInterval = config.refresh_interval or 30
                }
            }
        }
    }
end