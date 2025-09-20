-- Custom Metrics Plugin for k8s-tui
-- Demonstrates adding custom resource endpoints via plugins

-- Plugin metadata
function Name()
    return "custom-metrics"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Custom metrics and monitoring resource plugin"
end

-- Default configuration
function Config()
    return {
        enabled = true,
        refresh_interval = 30,
        show_in_header = true,
        metrics_types = {"cpu", "memory", "network", "storage"}
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    print("Setting up Custom Metrics Plugin")
    config = opts or Config()

    -- Register resources
    registerResources()

    -- Setup UI components
    if config.show_in_header then
        setupUIComponents()
    end

    return nil
end

-- Initialize the plugin
function Initialize()
    print("Custom Metrics plugin initialized")
    k8s_tui.set_status("Custom Metrics plugin ready")
    return nil
end

-- Shutdown the plugin
function Shutdown()
    print("Custom Metrics plugin shutting down")
    return nil
end

-- Register custom resources
function registerResources()
    print("Registered custom metrics resources")
end

-- Setup UI components
function setupUIComponents()
    k8s_tui.add_header("📊 Custom Metrics Active")
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "metrics:list",
            description = "List available custom metrics"
        },
        {
            name = "metrics:top",
            description = "Show top resource consumers"
        },
        {
            name = "metrics:alerts",
            description = "Show active metric alerts"
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
    print("Custom Metrics Plugin: App started")
    k8s_tui.set_status("Custom Metrics monitoring active")
end

function on_namespace_changed(data)
    print("Custom Metrics Plugin: Namespace changed to " .. data)
end

-- Define custom resource types
function GetResourceTypes()
    if not config.enabled then
        return {}
    end

    return {
        {
            Name = "Custom Metrics",
            Type = "custommetrics",
            Icon = "📊",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.25, 0.20, 0.15, 0.15, 0.25},
                },
            },
            Columns = {
                {Title = "Resource", Width = 20},
                {Title = "Type", Width = 15},
                {Title = "Value", Width = 12},
                {Title = "Status", Width = 10},
                {Title = "Last Updated", Width = 18}
            },
            RefreshIntervalSeconds = config.refresh_interval,
            Namespaced = true,
            Category = "Monitoring",
            Description = "Custom application and system metrics"
        },
        {
            Name = "Metric Alerts",
            Type = "metricalerts",
            Icon = "🚨",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.30, 0.20, 0.15, 0.20, 0.15},
                },
            },
            Columns = {
                {Title = "Alert Name", Width = 25},
                {Title = "Severity", Width = 12},
                {Title = "Threshold", Width = 15},
                {Title = "Current Value", Width = 15},
                {Title = "Status", Width = 10}
            },
            RefreshIntervalSeconds = 15,
            Namespaced = true,
            Category = "Monitoring",
            Description = "Active metric alerts and thresholds"
        }
    }
end

-- Fetch resource data for Custom Metrics
function GetResourceData(resourceType, namespace)
    if resourceType == "custommetrics" then
        return getCustomMetricsData(namespace)
    elseif resourceType == "metricalerts" then
        return getMetricAlertsData(namespace)
    else
        return nil, "unsupported resource type: " .. resourceType
    end
end

-- Get custom metrics data
function getCustomMetricsData(namespace)
    local current_time = os.date("%H:%M:%S")

    return {
        {
            Name = "web-server-" .. namespace,
            Namespace = namespace,
            Type = "cpu",
            Value = "45.2%",
            Status = "normal",
            Last_Updated = current_time
        },
        {
            Name = "web-server-" .. namespace,
            Namespace = namespace,
            Type = "memory",
            Value = "234MB",
            Status = "warning",
            Last_Updated = current_time
        },
        {
            Name = "database-" .. namespace,
            Namespace = namespace,
            Type = "cpu",
            Value = "12.8%",
            Status = "normal",
            Last_Updated = current_time
        },
        {
            Name = "database-" .. namespace,
            Namespace = namespace,
            Type = "memory",
            Value = "1.2GB",
            Status = "normal",
            Last_Updated = current_time
        },
        {
            Name = "api-gateway-" .. namespace,
            Namespace = namespace,
            Type = "network",
            Value = "2.1MB/s",
            Status = "normal",
            Last_Updated = current_time
        },
        {
            Name = "cache-" .. namespace,
            Namespace = namespace,
            Type = "storage",
            Value = "45GB",
            Status = "normal",
            Last_Updated = current_time
        }
    }, nil
end

-- Get metric alerts data
function getMetricAlertsData(namespace)
    return {
        {
            Name = "High Memory Usage",
            Namespace = namespace,
            Severity = "warning",
            Threshold = ">80%",
            Current_Value = "85.3%",
            Status = "active"
        },
        {
            Name = "Low Disk Space",
            Namespace = namespace,
            Severity = "critical",
            Threshold = "<10GB",
            Current_Value = "8.2GB",
            Status = "active"
        },
        {
            Name = "High CPU Usage",
            Namespace = namespace,
            Severity = "info",
            Threshold = ">70%",
            Current_Value = "45.2%",
            Status = "resolved"
        }
    }, nil
end

-- Delete a resource
function DeleteResource(resourceType, namespace, name)
    if resourceType ~= "custommetrics" and resourceType ~= "metricalerts" then
        return "unsupported resource type: " .. resourceType
    end

    -- In a real implementation, this would delete the metric/alert
    print("Deleted " .. resourceType .. " resource: " .. name .. " in namespace: " .. namespace)
    return nil
end

-- Get resource information
function GetResourceInfo(resourceType, namespace, name)
    if resourceType ~= "custommetrics" and resourceType ~= "metricalerts" then
        return nil, "unsupported resource type: " .. resourceType
    end

    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "2m",
        Status = "Active"
    }, nil
end