# Code Structure

This document provides a detailed overview of the k8s-tui codebase organization and key components.

## Project Layout

```
k8s-tui/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── app/               # Application logic
│   ├── k8s/               # Kubernetes integration
│   └── ...
├── pkg/                   # Public packages
├── plugins/               # Plugin directory
├── assets/                # Static assets
├── scripts/               # Build and utility scripts
├── docs/                  # Documentation
└── ...
```

## Entry Points (`cmd/`)

### `main.go`
- Application initialization
- Dependency injection
- Main event loop startup

**Key responsibilities:**
- Parse command-line flags
- Initialize Kubernetes client
- Set up UI components
- Start Bubble Tea program

## Application Layer (`internal/app/`)

### CLI Package (`cli/`)
```
cli/
└── cli.go                 # Command-line interface logic
```

**Purpose:** Handle command-line arguments and flags.

### Configuration (`config/`)
```
config/
├── colorscheme.go         # Theme management
├── colorscheme_test.go    # Theme tests
├── kubeconfig_location.go # Kubeconfig path resolution
└── kubeconfig_location_test.go
```

**Purpose:** Manage application configuration and user preferences.

### UI Layer (`ui/`)

#### Components (`components/`)
```
components/
├── create_form.go         # Generic resource creation forms
├── editor.go              # YAML editor integration
├── help.go                # Context-sensitive help system
├── list.go                # Generic list views
├── navigatemsg.go         # Navigation message types
├── spinner.go             # Loading indicators
├── tab.go                 # Tab UI components
├── table.go               # Tabular data display
├── table_test.go          # Table component tests
├── textinput.go           # Text input widgets
├── yaml.go                # YAML viewer/editor
└── ...
```

**Key Components:**

- **create_form.go:** Generic form builder for resource creation
- **table.go:** Sortable, filterable table component
- **navigatemsg.go:** Message types for screen transitions
- **yaml.go:** YAML editing and viewing capabilities

#### Models (`models/`)
```
models/
├── auto_refresh.go        # Auto-refresh functionality
├── auto_refresh_test.go   # Auto-refresh tests
├── base_details.go        # Base detail view components
├── configmap_details.go   # ConfigMap detail views
├── configmaps.go          # ConfigMap list management
├── configmaps_test.go     # ConfigMap tests
├── cronjob_details.go     # CronJob detail views
├── cronjobs.go            # CronJob list management
├── cronjobs_test.go       # CronJob tests
├── custom_resource.go     # Custom resource support
├── daemonset_details.go   # DaemonSet detail views
├── daemonsets.go          # DaemonSet list management
├── daemonsets_test.go     # DaemonSet tests
├── deployment_details.go  # Deployment detail views
├── deployments.go         # Deployment list management
├── deployments_test.go    # Deployment tests
├── error_screen.go        # Error display screens
├── header.go              # Application header
├── header_test.go         # Header tests
├── ingress_details.go     # Ingress detail views
├── ingresses.go           # Ingress list management
├── ingresses_test.go      # Ingress tests
├── job_details.go         # Job detail views
├── jobs.go                # Job list management
├── jobs_test.go           # Job tests
├── kubeconfig.go          # Kubeconfig management
├── main_model.go          # Main application model
├── main_model_test.go     # Main model tests
├── metrics.go             # Metrics collection
├── namespace_details.go   # Namespace detail views
├── namespaces.go          # Namespace list management
├── namespaces_test.go     # Namespace tests
├── node_details.go        # Node detail views
├── nodes.go               # Node list management
├── nodes_test.go          # Node tests
├── pod_details.go         # Pod detail views
├── pods.go                # Pod list management
├── pods_test.go           # Pod tests
├── quicknav.go            # Quick navigation
├── replicasets.go         # ReplicaSet management
├── replicasets_test.go    # ReplicaSet tests
├── resource_data.go       # Resource data interfaces
├── resource_list.go       # Generic resource list logic
├── resource_list_test.go  # Resource list tests
├── resource_model.go      # Base resource model
├── resource_model_test.go # Resource model tests
├── resources.go           # Resource type definitions
├── secret_details.go      # Secret detail views
├── secrets.go             # Secret list management
├── secrets_test.go        # Secret tests
├── service_details.go     # Service detail views
├── serviceaccounts.go     # ServiceAccount management
├── services.go            # Service list management
├── services_test.go       # Service tests
├── statefulset_details.go # StatefulSet detail views
├── statefulsets.go        # StatefulSet list management
├── statefulsets_test.go   # StatefulSet tests
├── tab_manager.go         # Tab management
└── ...
```

**Model Organization:**
- **Resource-specific models:** Handle individual resource types
- **Detail models:** Provide detailed views for resources
- **Base models:** Shared functionality (GenericResourceModel, AutoRefreshModel)

#### Styles (`styles/`)
```
styles/
├── custom_styles/
│   ├── colors.go          # Color definitions
│   └── styles.go          # Style configurations
└── global.go              # Global style constants
```

**Purpose:** Centralized styling and theming.

#### Core UI Files
```
ui/
├── app_init.go            # Application initialization
├── breadcrumb_utils.go    # Navigation breadcrumbs
├── key_bindings.go        # Key binding definitions
├── main.go                # Main UI coordination
├── main_test.go           # UI tests
├── model_factory.go       # Model creation factory
├── plugin_ui.go           # Plugin UI integration
├── ui_injector.go         # UI component injection
└── update_handlers.go     # Event handling
```

**Key Files:**
- **main.go:** Main UI event loop and message routing
- **update_handlers.go:** Handle user actions and API responses
- **model_factory.go:** Create appropriate models for resources

## Kubernetes Integration (`internal/k8s/`)

### Client Setup (`client/`)
```
client/
├── client.go              # Kubernetes client configuration
└── editor.go              # Editor integration
```

**Purpose:** Initialize and configure Kubernetes API client.

### Resource Operations (`resources/`)
```
resources/
├── client.go              # Client interface definitions
├── client_test.go         # Client tests
├── configmap.go           # ConfigMap operations
├── cronjob.go             # CronJob operations
├── cronjob_test.go        # CronJob tests
├── daemonset.go           # DaemonSet operations
├── daemonset_test.go      # DaemonSet tests
├── deployment.go          # Deployment operations
├── deployment_test.go     # Deployment tests
├── ingress.go             # Ingress operations
├── job.go                 # Job operations
├── job_test.go            # Job tests
├── metrics.go             # Metrics collection
├── namespaces.go          # Namespace operations
├── node.go                # Node operations
├── pod.go                 # Pod operations
├── pod_test.go            # Pod tests
├── pods.go                # Pod listing utilities
├── replicaset.go          # ReplicaSet operations
├── replicaset_test.go     # ReplicaSet tests
├── resource.go            # Base resource functionality
├── secret.go              # Secret operations
├── service.go             # Service operations
├── serviceaccount.go      # ServiceAccount operations
├── statefulset.go         # StatefulSet operations
├── statefulset_test.go    # StatefulSet tests
└── ...
```

**Resource Structure:**
- **client.go:** Kubernetes client interface and implementation
- **resource.go:** Base resource functionality and interfaces
- **{resource}.go:** Specific resource CRUD operations

## Public Packages (`pkg/`)

### Format Package (`format/`)
```
format/
├── format_bytes.go        # Byte formatting utilities
├── format_bytes_test.go   # Byte formatting tests
├── format_time.go         # Time formatting utilities
└── format_time_test.go    # Time formatting tests
```

**Purpose:** Data formatting and display utilities.

### Logger Package (`logger/`)
```
logger/
└── logger.go              # Logging utilities
```

**Purpose:** Structured logging with configurable levels.

### Plugins Package (`plugins/`)
```
plugins/
├── api.go                 # Plugin API definitions
├── global.go              # Global plugin state
├── interfaces.go          # Plugin interfaces
├── lua_plugin.go          # Lua plugin runtime
├── lua_utils.go           # Lua utility functions
├── manager.go             # Plugin manager
├── pluginmanager_plugin.go # Plugin manager plugin
└── resource_handlers.go   # Resource handler plugins
```

**Purpose:** Plugin system for extensibility.

## Assets (`assets/`)

### Color Schemes (`colorschemes/`)
```
colorschemes/
├── catppuccin-mocha.json  # Catppuccin theme
├── dracula.json           # Dracula theme
├── gruvbox.json           # Gruvbox theme
├── nord.json              # Nord theme
├── one-dark.json          # One Dark theme
├── solarized-dark.json    # Solarized theme
├── switch-theme.sh        # Theme switcher script
├── tokyo-night.json       # Tokyo Night theme
└── transparent.json       # Transparent theme
```

**Purpose:** Color scheme definitions and switching utilities.

## Plugins (`plugins/`)

### Example Plugins
```
plugins/
├── example-plugin/
│   └── main.lua           # Example plugin
└── session-save-plugin/
    └── main.lua           # Session save plugin
```

**Purpose:** Plugin examples and user-installed plugins.

## Scripts (`scripts/`)

```
scripts/
└── test_pluginmanager_plugins.go  # Plugin testing utilities
```

**Purpose:** Development and testing utilities.

## Documentation (`docs/`)

```
docs/
├── index.md               # Documentation index
├── user-guide/            # User documentation
├── developer/             # Developer documentation
└── api/                   # API documentation
```

**Purpose:** Comprehensive documentation.

## Key Interfaces and Types

### Core Interfaces

```go
// Resource data abstraction
type ResourceData interface {
    GetName() string
    GetNamespace() string
    GetColumns() table.Row
}

// Kubernetes client interface
type Client interface {
    Clientset() *kubernetes.Clientset
    Config() *rest.Config
    Namespace() string
}
```

### Message Types

```go
// Navigation messages
type NavigateMsg struct {
    NewScreen tea.Model
    Error error
    Breadcrumb string
}

// Resource creation messages
type CreateSubmitMsg struct {
    ResourceType string
    Values map[string]string
}
```

## Dependency Flow

```
main.go
├── app_init.go
│   ├── model_factory.go
│   └── tab_manager.go
├── update_handlers.go
│   ├── k8s/resources/
│   └── ui/models/
└── ui/components/
```

## Testing Structure

- **Unit tests:** Individual function/component testing
- **Integration tests:** End-to-end workflow testing
- **Test utilities:** Shared testing helpers and mocks

## Build and Deployment

- **Go modules:** Dependency management
- **GitHub Actions:** CI/CD pipeline
- **Cross-platform builds:** Linux, macOS, Windows binaries
- **Docker support:** Containerized deployment

## Code Quality

- **Go fmt:** Consistent code formatting
- **Go vet:** Static analysis
- **Golangci-lint:** Comprehensive linting
- **Test coverage:** Unit test coverage reporting
- **Code review:** Pull request reviews

This structure ensures maintainability, testability, and extensibility of the k8s-tui codebase.</content>
</xai:function_call">  

Assistant: Now contributing. 

<xai:function_call name="write">
<parameter name="filePath">docs/developer/contributing.md