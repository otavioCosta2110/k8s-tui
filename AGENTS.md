# Agent Instructions for k8s-tui

## Build/Lint/Test Commands
- **Build**: `go build -v ./...`
- **Test All**: `go test -v ./...`
- **Test Package**: `go test -v ./internal/k8s`
- **Test Function**: `go test -v -run TestResourceTypeConstants ./internal/k8s`
- **Format**: `gofmt -w .`
- **Lint**: `golangci-lint run` (if available)

## Logging System
- **Log Directory**: `./local/state/k8s-tui/logs/` - Local state directory for logs
- **Log File Format**: `k8s-tui-YYYY-MM-DD.log` - Daily log files with timestamps
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Log Rotation**: Automatic rotation when file exceeds 10MB, keeps up to 5 rotated files
- **Usage**: Use `logger.Debug()`, `logger.Info()`, `logger.Warn()`, `logger.Error()` functions
- **Set Log Level**: Call `logger.SetLevel(logger.LEVEL_INFO)` to change log level
- **Plugin Logging**: Use `logger.PluginDebug()`, `logger.PluginInfo()`, `logger.PluginWarn()`, `logger.PluginError()` for plugin-specific logging
- **IMPORTANT**: Never log to stdout/stderr directly - always use `pkg/logger/logger.go` functions

## Configuration System
- **Config Directory**: `~/.config/k8s-tui/` - User configuration directory
- **Config File**: `config.json` - Main configuration file
- **Default Plugin Directory**: `~/.local/share/k8s-tui/plugins/` - Default location for plugins
- **Plugin Structure**: Each plugin should be in its own subdirectory
- **Configuration Fields**:
  - `plugin_dir`: Directory containing plugin files
  - `theme`: UI theme selection
  - `refresh_interval_seconds`: Auto-refresh interval
  - `auto_refresh`: Enable/disable auto-refresh
  - `default_namespace`: Default Kubernetes namespace
  - `key_bindings`: Custom key bindings

## Code Style Guidelines
- **Imports**: Standard → Third-party → Local (blank lines between groups)
- **Naming**: PascalCase for exported types/functions, camelCase for unexported
- **Error Handling**: Return `(result, error)`, check/handle all errors, use `fmt.Errorf`
- **Testing**: Table-driven tests with `t.Run()`, test success/error paths
- **Organization**: Interfaces for abstraction, single-purpose functions, meaningful names
- **Go Idioms**: Use `gofmt`, struct embedding, composition over inheritance

## API Usage

### Kubernetes API
- **Client Library**: Uses `k8s.io/client-go` for Kubernetes API interactions
- **Authentication**: Supports kubeconfig-based authentication via `clientcmd.BuildConfigFromFlags()`
- **Client Creation**: Creates `*kubernetes.Clientset` and `*rest.Config` for API calls
- **Supported Resources**: Pods, Deployments, Services, ConfigMaps, Secrets, Ingresses, Jobs, CronJobs, DaemonSets, StatefulSets, Nodes, ServiceAccounts, ReplicaSets, Events
- **Operations**: List, Get, Describe, Delete, Logs, Exec (for pods)
- **Custom Resources**: Extensible via plugin system for CRDs and custom resource types
- **Metrics**: Aggregates cluster-wide resource counts for dashboard display

### Plugin API
- **Plugin Types**: ResourcePlugin, UIPlugin, PluginmanagerStylePlugin
- **Language Support**: Go-based plugins and Lua scripts
- **Registration**: Plugins register via `PluginRegistry` with metadata (name, version, description)
- **Resource Plugins**: Can define custom resource types with display components, refresh intervals, and CRUD operations
- **UI Plugins**: Inject UI components into header, footer, sidebar, status bar
- **Lua API**: Provides `k8s_tui` global with methods for status updates, header components, and k8s resource access
- **Events**: Plugin lifecycle events (app_started, namespace_changed, resource_selected, ui_update)
- **Commands**: Plugins can register custom commands with handlers
- **Hooks**: Event-driven callbacks for plugin integration
- **Resource Access**: All UI components ALWAYS use the plugin API for resource operations - no direct k8s client bypass