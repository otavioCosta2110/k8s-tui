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
