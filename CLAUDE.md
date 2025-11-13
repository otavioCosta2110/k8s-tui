# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

k8s-tui is a powerful terminal-based user interface for managing Kubernetes resources built with Bubble Tea (Go TUI framework). It provides full CRUD operations, multi-cluster support, and a Lua-based plugin system inspired by Neovim.

## Build and Test Commands

```bash
# Build the application
go build -o k8s-tui ./cmd

# Build all packages
go build -v ./...

# Run tests
go test -v ./...

# Run with custom kubeconfig
./k8s-tui --kubeconfig ~/.kube/config

# Run with multiple clusters
./k8s-tui --kubeconfig ~/.kube/cluster1 --kubeconfig ~/.kube/cluster2

# Run with plugins
./k8s-tui --plugin-dir ./plugins

# Test specific package
go test -v ./internal/app/ui/models
go test -v ./pkg/plugins
```

## Architecture

### Multi-Cluster Architecture

The application uses a two-level model hierarchy:

1. **MultiClusterModel** (`internal/app/ui/main.go:77`)
   - Top-level model that manages multiple clusters
   - Contains array of `*AppModel` instances (one per cluster)
   - Manages cluster switching via `clusterTabComponent`
   - Handles `currentCluster` index to track active cluster
   - Each cluster has its own isolated plugin manager

2. **AppModel** (`internal/app/ui/main.go:28`)
   - Represents a single cluster connection
   - Contains:
     - `tabManager`: manages resource tabs (Pods, Deployments, etc.)
     - `header`: displays cluster info and tabs
     - `kube`: Kubernetes client (`internal/k8s/resources/client.go`)
     - `pluginManager`: cluster-specific plugin instance
     - `breadcrumbTrail`: navigation history
   - Coordinates between UI components and K8s resources

### Key Components

- **TabManager** (`internal/app/ui/models/main_model.go`)
  - Manages multiple resource tabs per cluster
  - Each tab contains a resource-specific model (Pods, Deployments, etc.)
  - Handles tab creation, switching, and closing

- **Kubernetes Client** (`internal/k8s/resources/client.go`)
  - Wraps k8s.io/client-go
  - Provides resource CRUD operations
  - Handles namespace management
  - `GetClusterName()` extracts display name from cluster server URL

- **Plugin System** (`pkg/plugins/`)
  - Lua-based plugins with Neovim-inspired API
  - `GlobalPluginManager`: manages plugins across clusters
  - `ClusterContext`: holds cluster-specific state for plugins
  - Plugins can hook into events: `app_started`, `namespace_changed`, etc.

### Plugin Architecture

Plugins are Lua scripts with structured lifecycle:

```lua
function Name() -- Plugin identifier
function Version() -- Semver version
function Description() -- User-facing description
function Config() -- Default configuration
function Setup(opts) -- Called with user config at startup
function Initialize() -- Called after setup
function Commands() -- Register custom commands
function Hooks() -- Register event handlers
function CLIArguments() -- Register CLI arguments
```

Plugins access the `k8s_tui` global API:
- Session management: `get_tabs()`, `set_tabs()`, `get_namespace()`, `set_namespace()`
- UI: `set_status()`, `add_header()`, `show_input_dialog()`
- Navigation: `get_breadcrumb_trail()`, `set_breadcrumb_trail()`

### Multi-Cluster Plugin Management

Important: Each cluster maintains its own plugin manager instance. When implementing multi-cluster features in plugins:

1. The `GlobalPluginManager` tracks all clusters via `ClusterContext` objects
2. Each cluster has: `ID`, `Name`, `Client`, `Namespace`, `Settings`
3. Plugins can access cluster info via the API:
   - `k8s_tui.get_current_cluster()` - returns current cluster context
   - `k8s_tui.get_all_clusters()` - returns all clusters
   - Plugin state persists per cluster in `ClusterContext.Settings`

### Session Save Plugin

The `session-save-plugin` demonstrates:
- Saving session state to JSON (namespace, tabs, breadcrumb)
- Loading session state via CLI argument (`--session <file>`)
- Custom input dialog for user interaction

**To make it multi-cluster aware**, you need to:
1. Use `k8s_tui.get_all_clusters()` to get all cluster contexts
2. Save each cluster's state (Index, Name, Namespace, Kubeconfig, IsActive)
3. Save `current_cluster` to identify active cluster
4. When loading, restore all clusters and set the active one

## Code Structure

```
internal/
├── app/
│   ├── cli/           # CLI argument parsing
│   ├── config/        # App configuration (themes, kubeconfig)
│   └── ui/            # Bubble Tea UI implementation
│       ├── components/    # Reusable UI components (tables, forms, tabs)
│       ├── models/        # Resource-specific models (pods, deployments)
│       ├── main.go        # MultiClusterModel and AppModel
│       └── update_handlers.go # Message handlers
├── k8s/
│   ├── resources/     # K8s resource operations (CRUD)
│   └── types/         # K8s type interfaces
pkg/
├── format/            # Data formatting utilities (time, bytes)
├── logger/            # Application logging
└── plugins/           # Plugin system
    ├── manager.go         # Plugin loading and lifecycle
    ├── global_manager.go  # Multi-cluster plugin coordination
    ├── lua_plugin.go      # Lua integration
    └── resource_handlers.go # Plugin resource type extensions
```

## Important Patterns

### Bubble Tea Message Flow

1. User input generates `tea.KeyMsg`
2. `MultiClusterModel.Update()` routes to active cluster's `AppModel`
3. `AppModel.Update()` handles global keys or forwards to `TabManager`
4. `TabManager` forwards to active tab's model
5. Models return commands (e.g., API calls) that generate new messages
6. Messages bubble back up through the model hierarchy

### Adding New Resource Types

1. Create model in `internal/app/ui/models/<resource>.go`
2. Implement `InitComponent()`, `Init()`, `Update()`, `View()`
3. Implement `GetTable()` to return the underlying table component
4. Add to `model_factory.go` to enable tab creation
5. Add resource operations in `internal/k8s/resources/<resource>.go`

### Plugin Development

Plugins live in `./plugins/<plugin-name>/main.lua`. Key considerations:

- **Error Handling**: Always return `nil` for success, error string for failures
- **Logging**: Use `k8s_tui.log(message)` (logs to `~/.local/state/k8s-tui/logs/plugins/`)
- **State Management**: Store per-cluster state in the plugin's cluster context
- **API Calls**: All K8s operations go through the cluster's client
- **Testing**: Run with `go run cmd/main.go --plugin-dir ./plugins`

## Gotchas

1. **Plugin Manager Isolation**: Each cluster has its own `GlobalPluginManager`. When switching clusters, the global plugin manager reference is updated via `plugins.SetGlobalPluginManager()`.

2. **Namespace vs Cluster Switching**:
   - Namespace switching happens within a cluster (updates `AppModel.kube.Namespace`)
   - Cluster switching changes the entire `AppModel` instance

3. **Tab State**: When implementing session restore, tabs are cluster-specific. Each cluster maintains its own `TabManager` with separate tabs.

4. **Kubeconfig Paths**: Empty string `""` means use default kubeconfig (`~/.kube/config` or `KUBECONFIG` env var).

5. **Bubble Tea Alt Screen**: Application runs in alt screen mode (`tea.WithAltScreen()`), so it clears terminal on exit.

6. **Color Schemes**: Theme configuration in `~/.config/k8s-tui/config.json` affects global styles in `internal/app/ui/styles/global.go`.
