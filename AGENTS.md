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