# k8s-tui Plugin System

k8s-tui supports a plugin system that allows you to extend the application with custom resource types and UI functionality.

## Overview

The plugin system uses Go's native plugin mechanism to load shared libraries (.so files) at runtime. Plugins can:

- Add custom resource types to the resource list
- Provide custom data fetching and display logic
- Extend the UI with new components and functionality
- Integrate with external systems and APIs

## Plugin Architecture

### Core Interfaces

Plugins implement one or more of the following interfaces:

- `Plugin`: Base interface for all plugins
- `ResourcePlugin`: For plugins that provide custom resources
- `UIPlugin`: For plugins that extend the UI

### Plugin Loading

Plugins are loaded from a configurable directory (default: `./plugins`) at application startup. Each plugin is a `.so` file that exports a `Plugin` variable.

## Creating a Plugin

### 1. Set up your plugin project

```bash
mkdir my-k8s-plugin
cd my-k8s-plugin
go mod init my-k8s-plugin
```

### 2. Implement the plugin

```go
package main

import (
    "otaviocosta2110/k8s-tui/internal/k8s"
    "otaviocosta2110/k8s-tui/internal/plugins"
    "otaviocosta2110/k8s-tui/internal/types"
    "time"
    "github.com/charmbracelet/bubbles/table"
)

type MyPlugin struct {
    name    string
    version string
}

var Plugin MyPlugin

func init() {
    Plugin = MyPlugin{
        name:    "my-plugin",
        version: "1.0.0",
    }
}

func (p MyPlugin) Name() string { return p.name }
func (p MyPlugin) Version() string { return p.version }
func (p MyPlugin) Description() string { return "My custom plugin" }

func (p MyPlugin) Initialize() error {
    // Plugin initialization logic
    return nil
}

func (p MyPlugin) Shutdown() error {
    // Plugin cleanup logic
    return nil
}

func (p MyPlugin) GetResourceTypes() []plugins.CustomResourceType {
    return []plugins.CustomResourceType{
        {
            Name:  "MyResources",
            Type:  "myresource",
            Icon:  "🔧",
            Columns: []table.Column{
                {Title: "Name", Width: 20},
                {Title: "Status", Width: 10},
                {Title: "Age", Width: 10},
            },
            RefreshInterval: 30 * time.Second,
            Namespaced:      true,
        },
    }
}

func (p MyPlugin) GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
    // Fetch and return your custom resource data
    return []types.ResourceData{
        &MyResourceData{
            name:   "example-resource",
            status: "Running",
            age:    "5m",
        },
    }, nil
}

func (p MyPlugin) DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error {
    // Implement delete logic
    return nil
}

func (p MyPlugin) GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error) {
    // Return resource information
    return &k8s.ResourceInfo{
        Name:      name,
        Namespace: namespace,
        Kind:      k8s.ResourceType(resourceType),
        Age:       "5m",
    }, nil
}

func (p MyPlugin) GetUIExtensions() []plugins.UIExtension {
    return []plugins.UIExtension{}
}

// Implement ResourceData interface
type MyResourceData struct {
    name   string
    status string
    age    string
}

func (m *MyResourceData) GetName() string { return m.name }
func (m *MyResourceData) GetNamespace() string { return "" }
func (m *MyResourceData) GetColumns() table.Row {
    return table.Row{m.name, m.status, m.age}
}
```

### 3. Build the plugin

```bash
go build -buildmode=plugin -o my-plugin.so main.go
```

### 4. Install and use

```bash
# Copy to plugins directory
mkdir -p ./plugins
cp my-plugin.so ./plugins/

# Run k8s-tui with plugins
./k8s-tui --plugin-dir ./plugins
```

## Configuration

### Command Line Options

- `--plugin-dir`: Directory containing plugin files (default: `./plugins`)

### Environment Variables

- `K8S_TUI_PLUGIN_DIR`: Alternative way to specify plugin directory

## Plugin Development Tips

### Best Practices

1. **Error Handling**: Always handle errors appropriately and return meaningful error messages
2. **Resource Naming**: Use consistent naming conventions for your custom resources
3. **Performance**: Be mindful of refresh intervals and data fetching performance
4. **Thread Safety**: Ensure your plugin code is thread-safe if it maintains state

### Testing

Test your plugins by:
1. Building with `go build -buildmode=plugin`
2. Loading in k8s-tui and verifying functionality
3. Testing error conditions and edge cases

### Debugging

Enable debug logging to troubleshoot plugin issues:

```bash
./k8s-tui --plugin-dir ./plugins 2>&1 | grep -i plugin
```

## Example Plugins

See the `example-plugin/` directory for a complete working example that demonstrates:

- Custom resource type definition
- Data fetching and display
- Basic CRUD operations
- Plugin lifecycle management

## Security Considerations

- Plugins run with the same permissions as the main application
- Be cautious when loading plugins from untrusted sources
- Validate all inputs and outputs in your plugin code
- Avoid executing system commands or accessing sensitive resources

## Troubleshooting

### Common Issues

1. **Plugin not loading**: Check file permissions and ensure it's a valid `.so` file
2. **Import errors**: Verify all required packages are available
3. **Runtime errors**: Check plugin logs and error messages

### Getting Help

- Check the example plugin for reference implementations
- Review the plugin interfaces in `internal/plugins/`
- File issues on the GitHub repository

## Future Enhancements

Planned improvements to the plugin system:

- Plugin configuration files
- Hot-reloading of plugins
- Plugin marketplace/registry
- Enhanced UI extension APIs
- Plugin dependency management