# Code Architecture & Development

This document provides an overview of the k8s-tui codebase architecture, development patterns, and recent refactoring improvements.

## Project Structure

```
k8s-tui/
├── cmd/                    # Application entry point
├── internal/
│   ├── app/
│   │   ├── cli/           # Command-line interface
│   │   ├── config/        # Configuration management
│   │   └── ui/           # User interface components
│   │       ├── components/ # Reusable UI components
│   │       ├── models/     # UI model implementations
│   │       └── styles/    # Styling and themes
│   └── k8s/
│       ├── client/         # Kubernetes client wrapper
│       ├── resources/      # Kubernetes resource handlers
│       └── types/         # Type definitions and interfaces
├── pkg/                   # Public packages
│   ├── format/           # Utility formatting functions
│   ├── logger/           # Logging utilities
│   └── plugins/         # Plugin system
├── plugins/              # Example plugins
├── wiki/                 # Documentation
└── assets/              # Static assets (themes, etc.)
```

## Recent Refactoring Improvements

### Naming Convention Standardization

The codebase has been refactored to follow consistent naming conventions:

**Constructor Functions:**
- `NewDeployment` → `NewDeploymentInfo`
- `NewService` → `NewServiceInfo`
- Pattern: `NewXxxInfo` for resource info constructors

**Parameter Names:**
- `client` → `kubernetesClient` for clarity
- `k` → `kubernetesClient` for descriptiveness
- Use descriptive names that indicate purpose

### Error Handling Improvements

Enhanced error handling throughout the codebase:

```go
// Before
return fmt.Errorf("failed to get deployment: %v", err)

// After  
return fmt.Errorf("failed to fetch deployment %s/%s: %w", d.Namespace, d.Name, err)
```

- Added context to error messages (resource names, namespaces)
- Used error wrapping with `%w` for proper error chains
- Consistent error message formatting

### Code Organization

**Resource Handlers:**
- Each Kubernetes resource has dedicated handler in `/internal/k8s/resources/`
- Consistent patterns across all resource types
- Proper separation of concerns

**UI Components:**
- Reusable components in `/internal/app/ui/components/`
- Model-specific logic in `/internal/app/ui/models/`
- Consistent styling system

## Development Guidelines

### Adding New Resources

1. Create resource handler in `/internal/k8s/resources/`
2. Follow naming convention: `NewXxxInfo()`
3. Implement standard interface methods:
   - `Fetch()` - Get single resource
   - `FetchList()` - Get resource names
   - `GetTableData()` - Get table display data
4. Add to resource factory in `/internal/app/ui/models/resource_model.go`

### Adding UI Components

1. Create component in `/internal/app/ui/components/`
2. Follow existing patterns for Update/Init/View methods
3. Use consistent styling from `/internal/app/ui/styles/`
4. Add proper error handling and validation

### Plugin Development

1. Create plugin in `/plugins/` directory
2. Follow plugin interface in `/pkg/plugins/interfaces.go`
3. Use Lua API for extensibility
4. Register plugin in plugin manager

## Testing Strategy

### Test Organization
- Unit tests alongside source files (`*_test.go`)
- Integration tests in separate test packages
- Table-driven tests for multiple scenarios

### Test Patterns
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        // test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Performance Considerations

### Resource Loading
- Lazy loading of resource details
- Efficient table data generation
- Proper error handling to prevent cascading failures

### UI Updates
- Efficient re-rendering with Bubble Tea
- Minimal state updates
- Proper command batching

## Future Improvements

### Identified Refactoring Opportunities

1. **Plugin System**: Break down monolithic plugin files
2. **Generic Resource Handler**: Eliminate code duplication
3. **Update Handlers**: Split into focused handler types
4. **Error Types**: Create custom error types for better error handling
5. **UI Components**: Better separation of concerns

### Code Quality Goals
- Reduce code duplication by 30%
- Improve test coverage to 90%+
- Implement comprehensive error types
- Add performance benchmarks
- Enhance plugin system capabilities

## Contributing

When contributing to k8s-tui:

1. Follow the established naming conventions
2. Add comprehensive tests for new features
3. Update relevant documentation in `/wiki/`
4. Ensure all tests pass: `go test -v ./...`
5. Format code: `gofmt -w .`
6. Build successfully: `go build -v ./...`

## Architecture Decisions

### Bubble Tea Framework
Chosen for terminal UI development due to:
- Excellent state management
- Composable architecture
- Strong community support
- Built-in cross-platform support

### Plugin System
Lua-based plugin system provides:
- Easy extensibility
- Safe sandboxed execution
- Rich API for UI integration
- Hot-reloading capabilities

### Kubernetes Client
Uses official Go client for:
- Full API coverage
- Proper authentication handling
- Efficient resource management
- Community-maintained reliability