# Agent Instructions for k8s-tui

## Build/Lint/Test Commands
- **Build**: `go build -v ./...`
- **Test All**: `go test -v ./...`
- **Test Package**: `go test -v ./internal/k8s`
- **Test Function**: `go test -v -run TestResourceTypeConstants ./internal/k8s`
- **Format**: `gofmt -w .`
- **Lint**: `golangci-lint run` (if available)

## Code Style Guidelines
- **Imports**: Standard → Third-party → Local (blank lines between groups)
- **Naming**: PascalCase for exported types/functions, camelCase for unexported
- **Error Handling**: Return `(result, error)`, check/handle all errors, use `fmt.Errorf`
- **Testing**: Table-driven tests with `t.Run()`, test success/error paths
- **Organization**: Interfaces for abstraction, single-purpose functions, meaningful names
- **Go Idioms**: Use `gofmt`, struct embedding, composition over inheritance

## Documentation Structure
- **Wiki Directory**: Located at `/wiki/` containing user documentation
- **User Guide**: Comprehensive documentation in `/wiki/User-Guide/`
  - `Configuration.md` - Configuration options and setup
  - `Key-Bindings.md` - Keyboard shortcuts and navigation
  - `Navigation.md` - How to navigate the interface
  - `Resource-Management.md` - Managing Kubernetes resources
  - `Troubleshooting.md` - Common issues and solutions
- **Home Page**: `/wiki/Home.md` - Main documentation entry point
- **Documentation Updates**: When adding features, update relevant wiki pages. Always check if documentation changes are needed when modifying key bindings, configuration options, or user-facing features.

## Cluster Tab Creation Flow (Ctrl+N)
When user presses `Ctrl+N`:
1. **Key Binding**: `internal/app/ui/main.go:332` - detects "ctrl+n" key press
2. **Kubeconfig Selector**: Creates `KubeconfigSelectorModel` (`internal/app/ui/models/kubeconfig_selector.go:23`) to show file selection
3. **File Selection**: User selects kubeconfig file, triggers `KubeconfigSelectedMsg`
4. **Namespace Selector**: Creates `NamespaceSelectorModel` for namespace selection (`internal/app/ui/main.go:357`)
5. **Cluster Creation**: After namespace selection, creates new `AppModel` and adds to clusters slice (`internal/app/ui/main.go:374-375`)
6. **Tab Addition**: Adds new tab to `clusterTabComponent` with title "Cluster X" (`internal/app/ui/main.go:378`)
7. **Switch**: Sets new cluster as active (`internal/app/ui/main.go:379-380`)

## Plugin System (`pkg/plugins/`)

### Architecture Overview
The plugin system enables extensibility through Lua-based plugins with two main styles:
- **Legacy plugins**: Basic Lua scripts with simple interfaces
- **Pluginmanager-style plugins**: Advanced plugins with setup, config, commands, and hooks

### Key Components

#### Plugin Manager (`manager.go`)
- **`NewPluginManager(pluginDir string)`**: Creates new plugin manager instance
- **`LoadPlugins()`**: Scans plugin directory for `.lua` files and loads them
- **`setupBasicLuaAPI(L *lua.LState)`**: Sets up basic k8s_tui API for Lua plugins
- **`loadLuaPlugin(path string)`**: Loads and initializes individual Lua plugin

#### Plugin API (`api.go`)
- **`NewPluginAPI()`**: Creates new plugin API instance with managers for UI, commands, events, etc.
- **UI Management**: `AddHeaderComponent()`, `AddFooterComponent()` for UI extensions
- **Resource Access**: `GetPods()`, `GetServices()`, `GetDeployments()`, etc. for K8s resources
- **Tab Management**: `GetTabs()`, `SetTabs()` for tab manipulation
- **Event System**: `RegisterEventHandler()`, `TriggerEvent()` for plugin communication

#### Plugin Interfaces (`interfaces.go`)
- **`Plugin`**: Base interface with `Name()`, `Version()`, `Description()`, `Initialize()`, `Shutdown()`
- **`ResourcePlugin`**: For custom K8s resources with `GetResourceTypes()`, `GetResourceData()`
- **`UIPlugin`**: For UI extensions with `GetUIExtensions()`
- **`PluginAPI`**: Complete interface for plugin functionality access

### Plugin Loading Process
1. **Discovery**: Scan plugin directory for `.lua` files
2. **Validation**: Check for required functions (`Name()`, `Initialize()`)
3. **Style Detection**: Determine if plugin is legacy or pluginmanager-style
4. **API Setup**: Install appropriate Lua API functions
5. **Initialization**: Call plugin's `Initialize()` method
6. **Registration**: Register plugin in appropriate registry (resource/UI)

### Lua API Endpoints
- **Namespace**: `get_namespace()`, `set_namespace()`
- **Tabs**: `get_tabs()`, `set_tabs()`
- **Status**: `set_status()`
- **Resources**: `get_pods()`, `get_services()`, `get_deployments()`, etc.
- **Actions**: `delete_pod()`, `delete_service()`, etc.
- **UI**: `add_header()`, `show_input_dialog()`
- **Commands**: `register_command()`, `register_cli_argument()`

### Plugin Development Guidelines
- Use pluginmanager-style for new plugins (Setup, Config, Commands, Hooks)
- Implement proper error handling in Lua functions
- Use the provided logging functions for debugging
- Follow the existing naming conventions for API calls
- Test plugins with different K8s resource types
