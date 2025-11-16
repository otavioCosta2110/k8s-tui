# Example Custom Resource Plugin for k8s-tui

This is an enhanced example plugin that demonstrates how to extend k8s-tui with custom resource types using the Global Plugin Manager.

## Installing the Plugin

Copy `main.lua` file to plugins directory:

```bash
mkdir -p ./plugins
cp main.lua ./plugins/
```

## Running k8s-tui with Plugin

Start k8s-tui with plugin directory:

```bash
./k8s-tui --plugin-dir ./plugins
```

## Plugin Features

This example plugin provides:

- **ExampleResources**: A custom resource type that displays example data in the resource list
- **Global Plugin Manager Integration**: Works seamlessly with multi-cluster environments
- **Custom Resource Type**: Adds "exampleresource" type to the available resources
- **Table Display**: Shows resources with custom columns (Name, Namespace, Status, Version, Replicas, Age)
- **CRUD Operations**: Supports delete and info operations for custom resources
- **Commands**: Provides custom commands (`custom:status`, `custom:list`, `custom:config`)
- **UI Extensions**: Optional header component showing plugin status
- **Hooks**: Responds to app startup and namespace changes
- **Multi-Cluster Support**: Works across different cluster contexts

## Resource Type Details

The plugin registers a custom resource type called **ExampleResources** with the following properties:

- **Type**: `exampleresource`
- **Icon**: 📦 (configurable)
- **Category**: Custom
- **Namespaced**: Yes
- **Refresh Interval**: 10 seconds (configurable)
- **Table Columns**: Name, Namespace, Status, Version, Replicas, Age

## Sample Data

The plugin provides example data including:
- `example-app-v1` (Running, 3 replicas)
- `example-db-v2` (Pending, 1 replica)  
- `example-cache` (Failed, 2 replicas)
- `example-worker` (Running, 5 replicas)

## Developing Custom Plugins

To create a Lua plugin that works with Global Plugin Manager:

1. **Required Functions**: `Name()`, `Version()`, `Description()`, `Initialize()`
2. **Optional Functions**: `Setup()`, `Config()`, `Commands()`, `Hooks()`, `Shutdown()`
3. **Resource Plugin Functions**: `GetResourceTypes()`, `GetResourceData()`, `DeleteResource()`, `GetResourceInfo()`
4. **UI Extension Functions**: `GetUIExtensions()`
5. Save as `.lua` file in plugin directory

## Plugin Functions Reference

### Required Functions
```lua
function Name() return "my-plugin" end
function Version() return "1.0.0" end
function Description() return "My custom plugin" end
function Initialize() return nil end  -- Return nil for success, string for error
function Shutdown() return nil end    -- Optional cleanup function
```

### Configuration Functions
```lua
function Config()
    return {
        enabled = true,
        refresh_interval = 10,
        custom_option = "value"
    }
end

function Setup(opts)
    -- Called with user configuration
    config = opts
    return nil
end
```

### Resource Plugin Functions
```lua
function GetResourceTypes()
    return {
        {
            Name = "MyResources",
            Type = "myresource",
            Icon = "🔌",
            DisplayComponent = {
                Type = "table",
                Config = {
                    ColumnWidths = {0.25, 0.25, 0.20, 0.15, 0.15},
                },
            },
            RefreshIntervalSeconds = 30,
            Namespaced = true,
            Category = "Custom",
            Description = "My custom resource type"
        }
    }
end

function GetResourceData(resourceType, namespace)
    return {
        {
            Name = "example-resource",
            Namespace = namespace,
            Status = "Running",
            Version = "v1.0.0",
            Replicas = "3",
            Age = "5m",
        }
    }, nil  -- Return data table and error (nil for success)
end

function DeleteResource(resourceType, namespace, name)
    -- Implement delete logic
    k8s_tui.log("Deleting resource: " .. name)
    return nil  -- Return nil for success, string for error
end

function GetResourceInfo(resourceType, namespace, name)
    return {
        Name = name,
        Namespace = namespace,
        Kind = resourceType,
        Age = "5m",
        Version = "v1.0.0",
        Status = "Running",
        Labels = {app = "example"},
        Annotations = {description = "Example resource"}
    }, nil
end
```

### Commands and Hooks
```lua
function Commands()
    return {
        {
            name = "myplugin:status",
            description = "Show plugin status",
            handler = "handle_status"
        }
    }
end

function Hooks()
    return {
        {
            event = "app_started",
            handler = "on_app_started"
        }
    }
end

function handle_status(args)
    return "Plugin is running", nil
end

function on_app_started(data)
    k8s_tui.set_status("Plugin initialized")
end
```

### UI Extensions (Optional)
```lua
function GetUIExtensions()
    return {
        {
            Name = "my-ui-extension",
            Type = "ui_injection",
            InjectionPoints = {
                {
                    Location = "header",
                    Position = "right",
                    Priority = 10,
                    Component = {
                        Type = "text",
                        Config = {
                            content = "My Plugin Active",
                            style = "success"
                        }
                    }
                }
            }
        }
    }
end
```

## Global Plugin Manager API

The plugin provides access to the `k8s_tui` API with functions like:

- `k8s_tui.log(message)` - Log messages
- `k8s_tui.set_status(message)` - Set status message
- `k8s_tui.get_namespace()` - Get current namespace
- `k8s_tui.set_namespace(namespace)` - Set current namespace
- `k8s_tui.get_tabs()` - Get current tabs
- `k8s_tui.set_tabs(tabs)` - Set tabs
- Multi-cluster functions for cluster management

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