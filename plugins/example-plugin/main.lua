-- Example Lua Plugin for k8s-tui
-- This demonstrates how to create a plugin using Lua scripting

-- Plugin metadata
function Name()
    return "example-lua-plugin"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Example plugin demonstrating Lua-based custom resource functionality"
end

-- Initialize the plugin
function Initialize()
    print("Example Lua plugin initialized")
    return nil  -- Return nil for success, string for error
end

-- Shutdown the plugin
function Shutdown()
    print("Example Lua plugin shutting down")
    return nil
end

-- Define custom resource types
function GetResourceTypes()
    return {
        {
            Name = "ExampleResources",
            Type = "exampleresource",
            Icon = "󰐧",
            Columns = {
                {Title = "Name", Width = 10},
                {Title = "Namespace", Width = 10},
                {Title = "Status", Width = 10},
                {Title = "Age", Width = 10},
            },
            RefreshIntervalSeconds = 10,
            Namespaced = true,
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

    print("Deleting example resource " .. namespace .. "/" .. name)
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

-- UI Extensions (optional)
function GetUIExtensions()
    return {}  -- Return empty table if no extensions
end
