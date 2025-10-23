# Contributing Guide

Welcome! We're glad you're interested in contributing to k8s-tui. This guide will help you get started with development, coding standards, and the contribution process.

## Development Setup

### Prerequisites

- **Go 1.21+**: [Download here](https://golang.org/dl/)
- **Kubernetes cluster**: Local (kind, minikube) or remote
- **kubectl**: Configured for your cluster
- **Git**: Version control

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/otavioCosta2110/k8s-tui.git
cd k8s-tui

# Install dependencies
go mod download

# Build the project
go build -v ./...

# Run tests
go test -v ./...
```

### Development Environment

```bash
# Set up pre-commit hooks (if available)
# Enable Go modules
export GO111MODULE=on

# Configure your editor for Go development
# Recommended: Go extension for VS Code, vim-go, etc.
```

## Development Workflow

### 1. Choose an Issue

- Check [GitHub Issues](https://github.com/otavioCosta2110/k8s-tui/issues) for open tasks
- Look for "good first issue" or "help wanted" labels
- Comment on the issue to indicate you're working on it

### 2. Create a Branch

```bash
# Create and switch to a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-number-description
```

### 3. Make Changes

- Follow the [code structure](code-structure.md) guidelines
- Write tests for new functionality
- Update documentation as needed
- Ensure code compiles and tests pass

### 4. Test Your Changes

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/k8s/resources

# Run with race detection
go test -race -v ./...

# Run benchmarks
go test -bench=. -v ./...
```

### 5. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add support for custom resource definitions

- Add CRD discovery functionality
- Implement generic CRD views
- Add CRD creation forms

Closes #123"
```

#### Commit Message Format

Follow conventional commits:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Testing
- `chore`: Maintenance

**Examples:**
```
feat: add dark mode support
fix(ui): resolve table rendering issue
docs: update installation guide
```

### 6. Push and Create PR

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a Pull Request on GitHub
# Fill out the PR template with details
```

## Code Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

```go
// Good
type KubernetesClient struct {
    config *rest.Config
    clientset *kubernetes.Clientset
}

func (kc *KubernetesClient) GetPods(namespace string) ([]PodInfo, error)

// Bad
type k8sClient struct {
    conf *rest.Config
    cs *kubernetes.Clientset
}

func (k *k8sClient) getPods(ns string) ([]PodInfo, error)
```

### Error Handling

```go
// Good
func (r *Resource) Get(name string) (*Resource, error) {
    if name == "" {
        return nil, fmt.Errorf("name cannot be empty")
    }

    resource, err := r.client.Get(name)
    if err != nil {
        return nil, fmt.Errorf("failed to get resource %s: %w", name, err)
    }

    return resource, nil
}

// Bad - don't ignore errors
func (r *Resource) Get(name string) *Resource {
    resource, _ := r.client.Get(name)
    return resource
}
```

### Documentation

```go
// Good
// GetPod retrieves a pod by name from the specified namespace.
// Returns an error if the pod is not found or if the request fails.
func (c *Client) GetPod(namespace, name string) (*PodInfo, error) {
    // implementation
}

// Bad - no documentation
func (c *Client) GetPod(namespace, name string) (*PodInfo, error) {
    // implementation
}
```

## Testing Guidelines

### Unit Tests

```go
func TestPodCreation(t *testing.T) {
    // Arrange
    client := &mockClient{}
    pod := NewPod("test-pod", "default", client)

    // Act
    err := pod.Create("nginx", "latest", "default")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test-pod", pod.Name)
}
```

### Table-Driven Tests

```go
func TestValidateResourceName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool
    }{
        {"valid name", "my-pod", true},
        {"empty name", "", false},
        {"invalid chars", "my_pod!", false},
        {"too long", strings.Repeat("a", 254), false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateResourceName(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Tests

```go
func TestPodLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Set up test cluster
    client := setupTestCluster(t)

    // Test pod creation
    pod := NewPod("test-pod", "default", client)
    err := pod.Create("nginx", "latest", "default")
    require.NoError(t, err)

    // Test pod retrieval
    retrieved, err := client.GetPod("default", "test-pod")
    require.NoError(t, err)
    assert.Equal(t, "nginx", retrieved.Image)

    // Clean up
    err = pod.Delete()
    assert.NoError(t, err)
}
```

### Test Coverage

Aim for >80% code coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Pull Request Process

### PR Checklist

- [ ] Tests pass: `go test -v ./...`
- [ ] Code formatted: `gofmt -w .`
- [ ] Linting passes: `golangci-lint run`
- [ ] Documentation updated
- [ ] Commit messages follow conventional format
- [ ] PR description includes:
  - What changes were made
  - Why they were needed
  - How to test the changes
  - Screenshots/videos if UI changes

### PR Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How the changes were tested

## Screenshots
If applicable, add screenshots

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Code style checked
```

### Review Process

1. **Automated Checks**: CI runs tests, linting, and builds
2. **Code Review**: At least one maintainer reviews the code
3. **Discussion**: Address review comments and iterate
4. **Approval**: Maintainers approve the PR
5. **Merge**: Squash and merge with conventional commit message

## Adding New Features

### 1. Resource Support

To add support for a new Kubernetes resource:

1. **Create resource operations** in `internal/k8s/resources/`
2. **Add UI model** in `internal/app/ui/models/`
3. **Update tab manager** to include the new resource
4. **Add creation form** fields
5. **Update help text**

### 2. UI Components

When adding new UI components:

1. **Follow Bubble Tea patterns**
2. **Make components reusable**
3. **Add proper styling support**
4. **Include accessibility features**

### 3. Plugins

For new plugin functionality:

1. **Define plugin interface**
2. **Implement Lua bindings**
3. **Add security checks**
4. **Document plugin API**

## Reporting Security Issues

If you discover a security vulnerability:

1. **Don't create a public issue**
2. **Email maintainers directly**
3. **Provide detailed reproduction steps**
4. **Allow time for fix before disclosure**

## Getting Help

- **GitHub Discussions**: Ask questions and get community help
- **Discord/Slack**: Join our community chat (if available)
- **Documentation**: Check the docs first
- **Issues**: Search existing issues before creating new ones

## Recognition

Contributors are recognized in:
- GitHub contributor stats
- CHANGELOG.md entries
- Release notes
- Project README

Thank you for contributing to k8s-tui! 🎉</content>
</xai:function_call">  

Assistant: Now testing. 

<xai:function_call name="write">
<parameter name="filePath">docs/developer/testing.md