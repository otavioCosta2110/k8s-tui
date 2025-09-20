# Example Plugin for k8s-tui

This is an example plugin that demonstrates how to extend k8s-tui with custom resource types.

## Installing the Plugin

Copy the `main.lua` file to the plugins directory:

```bash
mkdir -p ./plugins
cp main.lua ./plugins/
```

## Running k8s-tui with the Plugin

Start k8s-tui with the plugin directory:

```bash
./k8s-tui --plugin-dir ./plugins
```

## Plugin Features

This example plugin provides:

- **ExampleResources**: A custom resource type that displays example data
- Basic CRUD operations (Create, Read, Update, Delete)
- Custom table columns and formatting
- Integration with k8s-tui's UI system

## Developing Custom Plugins

To create a Lua plugin:

1. Define required functions: `Name()`, `Version()`, `Description()`, `Initialize()`
2. For resource plugins, define: `GetResourceTypes()`, `GetResourceData()`, etc.
3. For UI extensions, define: `GetUIExtensions()`
4. Save as `.lua` file

## Plugin Functions

```lua
-- Required functions
function Name() return "my-plugin" end
function Version() return "1.0.0" end
function Description() return "My custom plugin" end
function Initialize() return nil end  -- Return nil for success, string for error
function Shutdown() return nil end    -- Optional cleanup function

-- Resource plugin functions
function GetResourceTypes()
    return {
        {
            Name = "MyResources",
            Type = "myresource",
            Icon = "🔌",
            Columns = {
                {Title = "Name", Width = 20},
                {Title = "Status", Width = 10},
            },
            RefreshIntervalSeconds = 30,
            Namespaced = true,
        }
    }
end

function GetResourceData(resourceType, namespace)
    return {
        {
            Name = "example-resource",
            Namespace = namespace,
            Status = "Running",
            Age = "5m",
        }
    }, nil  -- Return data table and error (nil for success)
end

function DeleteResource(resourceType, namespace, name)
    -- Implement delete logic
    return nil  -- Return nil for success, string for error
end

function GetResourceInfo(resourceType, namespace, name)
    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "5m",
    }, nil
end

-- UI extension functions (optional)
function GetUIExtensions()
    return {}  -- Return empty table if no extensions
end
```

## Custom Resource Type Definition

```go
type CustomResourceType struct {
    Name           string
    Type           string
    Icon           string
    Columns        []table.Column
    RefreshInterval time.Duration
    Namespaced     bool
}
```

This plugin system allows you to extend k8s-tui with custom Kubernetes resources, third-party CRDs, or any other data sources you want to monitor and manage through the TUI interface. Lua plugins provide maximum flexibility with easy development and runtime modification capabilities.