# Agent Instructions for k8s-tui

## Build/Lint/Test Commands

### Build
```bash
go build -v ./...
```

### Test All
```bash
go test -v ./...
```

### Test Single Package
```bash
go test -v ./internal/k8s
go test -v ./internal/ui/models
```

### Test Single Function
```bash
go test -v -run TestResourceTypeConstants ./internal/k8s
```

### Format Code
```bash
gofmt -w .
```

### Lint (if golangci-lint is available)
```bash
golangci-lint run
```

## Code Style Guidelines

### Imports
- Standard library imports first
- Third-party imports second
- Local project imports last
- Use blank lines to separate groups

### Naming Conventions
- **Variables/Functions**: camelCase for unexported, PascalCase for exported
- **Types/Structs**: PascalCase
- **Constants**: PascalCase for exported, camelCase for unexported
- **Interfaces**: PascalCase, often suffixed with "er" (e.g., ResourceManager)

### Error Handling
- Functions return `(result, error)` pattern
- Always check and handle errors
- Use `fmt.Errorf` for error wrapping
- Return `nil` explicitly for successful operations

### Testing
- Use table-driven tests for multiple test cases
- Test files named `*_test.go`
- Use `t.Run()` for subtests
- Test both success and error paths

### Code Organization
- Use interfaces for abstraction
- Group related functionality in packages
- Keep functions focused and single-purpose
- Use meaningful variable names

### Go Specific
- Use `gofmt` for formatting
- Follow Go idioms and conventions
- Use struct embedding appropriately
- Prefer composition over inheritance