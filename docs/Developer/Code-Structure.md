# 📁 Code Structure

This document explains the organization of the k8s-tui codebase and the purpose of each directory and module.

## Directory Overview

```
k8s-tui/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── app/               # Core application logic
│   ├── k8s/               # Kubernetes integration
│   └── plugins/           # Plugin system
├── pkg/                   # Public library code
├── assets/                # Static assets (themes, etc.)
├── plugins/               # Example and built-in plugins
├── scripts/               # Build and utility scripts
├── testdata/              # Test fixtures and data
└── docs/                  # Documentation
```

## Core Directories

### `cmd/`
Contains the main application entry points.

```
cmd/
├── main.go               # Application entry point
└── main_test.go          # Integration tests
```

**Purpose**: Bootstrap the application, parse command-line arguments, and start the UI.

### `internal/app/`
Core application logic and UI components.

```
internal/app/
├── cli/                  # Command-line interface
│   └── cli.go           # CLI argument parsing
├── config/               # Configuration management
│   ├── colorscheme.go   # Theme configuration
│   └── kubeconfig_location.go # Kubeconfig handling
└── ui/                   # User interface
    ├── components/       # Reusable UI components
    ├── models/          # View models and state management
    ├── styles/          # Styling and themes
    └── *.go             # Main UI logic and initialization
```

**Key Components**:
- **Components**: Reusable UI elements (tables, forms, lists)
- **Models**: State management for different views (pods, deployments, etc.)
- **Styles**: Theme system and styling definitions

### `internal/k8s/`
Kubernetes client and resource management.

```
internal/k8s/
├── client/               # Kubernetes client wrapper
│   └── editor.go        # YAML editor integration
├── resources/            # Resource-specific implementations
│   ├── pod.go          # Pod resource handling
│   ├── deployment.go   # Deployment resource handling
│   └── ...             # Other resource types
└── types/               # Type definitions and interfaces
    └── interfaces.go    # Core interfaces
```

**Key Components**:
- **Client**: Wrapper around client-go with multi-cluster support
- **Resources**: Resource-specific logic for CRUD operations
- **Types**: Shared interfaces and type definitions

### `internal/plugins/`
Plugin system implementation.

```
internal/plugins/
├── api.go               # Plugin API definitions
├── global.go            # Global plugin state
├── interfaces.go        # Plugin interfaces
├── lua_plugin.go        # Lua plugin runtime
├── lua_utils.go         # Lua utility functions
├── manager.go           # Plugin manager
├── pluginmanager_plugin.go # Built-in plugin manager
└── resource_handlers.go # Resource handling hooks
```

## Public Libraries (`pkg/`)

### `pkg/format/`
Formatting utilities.

```
pkg/format/
├── format_bytes.go      # Byte formatting (KB, MB, GB)
└── format_time.go       # Time formatting utilities
```

### `pkg/logger/`
Logging utilities.

```
pkg/logger/
└── logger.go           # Structured logging
```

### `pkg/plugins/`
Public plugin API.

```
pkg/plugins/
├── api.go              # Public plugin API
├── global.go           # Global plugin state
├── interfaces.go       # Plugin interfaces
├── lua_plugin.go       # Lua plugin implementation
├── lua_utils.go        # Lua utilities
├── manager.go          # Plugin manager
├── pluginmanager_plugin.go # Plugin manager plugin
└── resource_handlers.go # Resource handlers
```

## Static Assets (`assets/`)

```
assets/
└── colorschemes/        # Color theme definitions
    ├── catppuccin-mocha.json
    ├── dracula.json
    ├── gruvbox.json
    ├── nord.json
    ├── one-dark.json
    ├── solarized-dark.json
    ├── tokyo-night.json
    ├── transparent.json
    └── switch-theme.sh  # Theme switching script
```

## Plugin Examples (`plugins/`)

```
plugins/
├── example-plugin/      # Basic plugin example
├── pluginmanager-header/ # Header customization plugin
└── session-save-plugin/ # Session persistence plugin
```

## Build and Scripts (`scripts/`)

```
scripts/
└── test_pluginmanager_plugins.go # Plugin testing utilities
```

## Test Data (`testdata/`)

```
testdata/
└── fixtures/            # Test fixtures
    └── sample-pod.json  # Sample pod definition
```

## Code Organization Principles

### 1. Separation of Concerns
- **UI Logic**: `internal/app/ui/`
- **Business Logic**: `internal/app/`
- **Kubernetes Logic**: `internal/k8s/`
- **Plugin Logic**: `internal/plugins/`

### 2. Dependency Direction
```
cmd/ → internal/ → pkg/
```
- Higher-level packages depend on lower-level packages
- `internal/` can import from `pkg/`
- `pkg/` should not import from `internal/`

### 3. Interface-Based Design
- Core interfaces defined in `internal/*/types/`
- Implementations in respective packages
- Easy testing and mocking

### 4. Plugin Architecture
- Core functionality in `internal/`
- Plugin API in `pkg/plugins/`
- Example plugins in `plugins/`

## Naming Conventions

### Files
- **PascalCase** for exported types: `pod_model.go`
- **snake_case** for utilities: `kubeconfig_location.go`
- **Test files**: `*_test.go`

### Packages
- **Lowercase**, single words when possible
- **Descriptive**: `resources`, `components`, `styles`
- **No plural forms** unless necessary: `plugin` not `plugins`

### Types and Functions
- **PascalCase** for exported: `PodModel`, `NewPodModel()`
- **camelCase** for unexported: `podModel`, `newPodModel()`

## Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party libraries
    "github.com/charmbracelet/bubbletea"
    
    // Internal packages
    "github.com/otavioCosta2110/k8s-tui/internal/app/ui"
    "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
    
    // Local packages
    "github.com/otavioCosta2110/k8s-tui/pkg/format"
)
```

## Testing Structure

### Unit Tests
- Co-located with source files: `model.go` → `model_test.go`
- Table-driven tests for complex logic
- Mock interfaces for external dependencies

### Integration Tests
- End-to-end workflows in `cmd/main_test.go`
- Resource handling tests in `internal/k8s/resources/`
- UI component tests in `internal/app/ui/components/`

### Test Fixtures
- Sample Kubernetes objects in `testdata/fixtures/`
- Mock configurations for testing
- Test utilities in `scripts/`

## Build Process

The application follows standard Go build practices:
- `go build ./cmd` builds the main application
- `go test ./...` runs all tests
- `go fmt ./...` formats code
- `golangci-lint run` for linting (if available)