-- EndpointSlice Plugin for k8s-tui
-- Adds support for viewing Kubernetes EndpointSlice resources

-- Plugin metadata
function Name()
    return "endpoints"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Kubernetes EndpointSlice resource viewer plugin"
end

-- Default configuration
function Config()
    return {
        enabled = true,
        refresh_interval = 30,
        show_in_header = false
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    -- print("Setting up Endpoints Plugin")
    config = opts or Config()
    return nil
end

-- Initialize the plugin
function Initialize()
    -- print("Endpoints plugin initialized")
    if config and config.show_in_header then
        k8s_tui.set_status("Endpoints plugin ready")
    end
    return nil
end

-- Shutdown the plugin
function Shutdown()
    -- print("Endpoints plugin shutting down")
    return nil
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "endpoints:list",
            description = "List all endpoint slices in current namespace"
        },
        {
            name = "endpoints:services",
            description = "Show endpoint slices grouped by services"
        }
    }
end

-- Hooks that this plugin registers for
function Hooks()
    return {
        {
            event = "namespace_changed",
            handler = "on_namespace_changed"
        }
    }
end

-- Hook handlers
function on_namespace_changed(data)
    -- print("Endpoints Plugin: Namespace changed to " .. data)
end

-- Define custom resource types
function GetResourceTypes()
    if not config.enabled then
        return {}
    end

    return {
        {
            Name = "EndpointSlices",
            Type = "endpoints",
            Icon = "🔗",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.25, 0.20, 0.15, 0.20, 0.20},
                },
            },
            Columns = {
                {Title = "Name", Width = 30},
                {Title = "Addresses", Width = 25},
                {Title = "Ports", Width = 15},
                {Title = "Service", Width = 20},
                {Title = "Age", Width = 10}
            },
            RefreshIntervalSeconds = config.refresh_interval,
            Namespaced = true,
            Category = "Networking",
            Description = "Kubernetes endpoint slices that track pod IP addresses for services"
        }
    }
end

-- Fetch resource data for EndpointSlices
function GetResourceData(resourceType, namespace)
    if resourceType == "endpoints" then
        return getEndpointSlicesData(namespace)
    else
        return nil, "unsupported resource type: " .. resourceType
    end
end

-- Get endpoint slices data
function getEndpointSlicesData(namespace)
    -- Query the Kubernetes API for real endpoints data
    local result = k8s_tui.get_endpoints(namespace)

    -- Check if result is a string (error message)
    if type(result) == "string" then
        -- print("Error fetching endpoints: " .. result)
        return {}, result
    end

    -- Result should be a table of endpoint objects
    local endpoints = {}
    for i, endpoint in ipairs(result) do
        table.insert(endpoints, {
            Name = endpoint.Name,
            Namespace = endpoint.Namespace,
            Addresses = endpoint.Addresses,
            Ports = endpoint.Ports,
            Service = endpoint.Service,
            Age = endpoint.Age
        })
    end

    return endpoints, nil
end

-- Delete a resource
function DeleteResource(resourceType, namespace, name)
    if resourceType ~= "endpoints" then
        return "unsupported resource type: " .. resourceType
    end

    -- In a real implementation, this would delete the endpoint slice
    -- Note: EndpointSlices are typically managed by Services and cannot be deleted directly
    -- print("Cannot delete endpoint slices directly - they are managed by Services")
    return "endpoint slices are managed by services and cannot be deleted directly"
end

-- Get resource information
function GetResourceInfo(resourceType, namespace, name)
    if resourceType ~= "endpoints" then
        return nil, "unsupported resource type: " .. resourceType
    end

    return {
        Name = name,
        Namespace = namespace,
        Kind = "EndpointSlice",
        Age = "2h",
        Status = "Active"
    }, nil
end
