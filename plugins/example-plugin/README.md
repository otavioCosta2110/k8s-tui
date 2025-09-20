# Example Plugin for k8s-tui

This is an example plugin that demonstrates how to extend k8s-tui with custom resource types.

## Building the Plugin

To build this plugin as a shared library:

```bash
go build -buildmode=plugin -o example-plugin.so main.go
```

## Installing the Plugin

Copy the `example-plugin.so` file to the plugins directory:

```bash
mkdir -p ./plugins
cp example-plugin.so ./plugins/
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

To create your own plugin:

1. Implement the `plugins.Plugin` interface
2. For resource plugins, implement `plugins.ResourcePlugin`
3. For UI extensions, implement `plugins.UIPlugin`
4. Export your plugin instance as `var Plugin YourPluginType`
5. Build with `-buildmode=plugin`

## Plugin Interface

```go
type ResourcePlugin interface {
    Plugin
    GetResourceTypes() []CustomResourceType
    GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error)
    DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error
    GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error)
}
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

This plugin system allows you to extend k8s-tui with custom Kubernetes resources, third-party CRDs, or any other data sources you want to monitor and manage through the TUI interface.