# 🤝 Contributing

We welcome contributions to k8s-tui! This guide will help you get started with contributing to the project.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- A Kubernetes cluster (for testing)
- Docker (optional, for containerized testing)

### Development Setup

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/your-username/k8s-tui.git
   cd k8s-tui
   ```

2. **Add Upstream Remote**
   ```bash
   git remote add upstream https://github.com/otavioCosta2110/k8s-tui.git
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   ```

4. **Build and Test**
   ```bash
   go build -v ./...
   go test -v ./...
   ```

## 📋 Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Changes

- Follow the existing code style and conventions
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./internal/k8s/resources

# Run with coverage
go test -v -cover ./...
```

### 4. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with a clear message
git commit -m "feat: add new feature description"

# Use conventional commit messages:
# feat: new feature
# fix: bug fix
# docs: documentation changes
# style: formatting changes
# refactor: code refactoring
# test: adding or updating tests
# chore: maintenance tasks
```

### 5. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create a pull request on GitHub
```

## 🏗️ Project Structure

Understanding the project structure will help you make effective contributions:

- `cmd/` - Application entry points
- `internal/` - Private application code
  - `app/` - Core application logic and UI
  - `k8s/` - Kubernetes integration
  - `plugins/` - Plugin system
- `pkg/` - Public library code
- `assets/` - Static assets (themes, etc.)
- `plugins/` - Example and built-in plugins
- `docs/` - Documentation

## 🧪 Testing Guidelines

### Writing Tests

1. **Unit Tests**: Test individual functions and methods
   ```go
   func TestPodModel_UpdateStatus(t *testing.T) {
       // Test implementation
   }
   ```

2. **Table-Driven Tests**: For testing multiple scenarios
   ```go
   func TestFormatBytes(t *testing.T) {
       tests := []struct {
           name     string
           input    int64
           expected string
       }{
           {"bytes", 1024, "1.0 KB"},
           {"kilobytes", 1048576, "1.0 MB"},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result := FormatBytes(tt.input)
               assert.Equal(t, tt.expected, result)
           })
       }
   }
   ```

3. **Integration Tests**: Test component interactions
   ```go
   func TestKubernetesClient_ListPods(t *testing.T) {
       // Integration test with mock Kubernetes
   }
   ```

### Test Coverage

- Aim for >80% code coverage on new code
- Write tests for both success and error paths
- Test edge cases and boundary conditions

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -race -v ./...

# Run benchmarks
go test -bench=. -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Code Style

### Go Conventions

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle errors properly

### Example Code Style

```go
// PodModel represents the model for pod management
type PodModel struct {
    client k8s.Client
    pods   []v1.Pod
    selected int
    loading bool
    err     error
}

// NewPodModel creates a new pod model instance
func NewPodModel(client k8s.Client) *PodModel {
    return &PodModel{
        client: client,
        pods:   make([]v1.Pod, 0),
    }
}

// UpdatePods fetches the latest pod list from Kubernetes
func (m *PodModel) UpdatePods(ctx context.Context) error {
    pods, err := m.client.ListPods(ctx)
    if err != nil {
        return fmt.Errorf("failed to list pods: %w", err)
    }
    
    m.pods = pods
    return nil
}
```

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment Information**
   - k8s-tui version
   - Go version
   - OS and architecture
   - Kubernetes version

2. **Steps to Reproduce**
   - Clear, step-by-step instructions
   - Expected vs actual behavior

3. **Additional Context**
   - Screenshots if applicable
   - Error messages or logs
   - Configuration files (redacted if necessary)

### Bug Report Template

```markdown
## Bug Description
Brief description of the bug

## Environment
- k8s-tui version: 
- Go version:
- OS:
- Kubernetes version:

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Additional Context
Any additional information
```

## ✨ Feature Requests

When requesting features:

1. **Use a clear title** describing the feature
2. **Describe the problem** you're trying to solve
3. **Propose a solution** if you have one
4. **Consider alternatives** and why your solution is preferred
5. **Additional context** like use cases or examples

### Feature Request Template

```markdown
## Feature Description
Clear description of the feature

## Problem Statement
What problem does this solve?

## Proposed Solution
How should this work?

## Alternatives Considered
Other approaches and why they weren't chosen

## Additional Context
Use cases, examples, or additional information
```

## 🔧 Plugin Development

If you're contributing plugins:

1. **Follow Plugin Guidelines** in the [[API/Plugins|Plugin Development Guide]]
2. **Test Thoroughly** with different Kubernetes versions
3. **Document Usage** with clear examples
4. **Handle Errors** gracefully
5. **Version Compatibility** - specify minimum k8s-tui version

## 📖 Documentation

When contributing documentation:

1. **Update User Docs** for user-facing features
2. **Update Developer Docs** for API changes
3. **Add Examples** for complex features
4. **Test Links** to ensure they work
5. **Follow Style Guide** for consistency

## 🔄 Pull Request Process

### Before Submitting

1. **Test thoroughly** - ensure all tests pass
2. **Update documentation** if needed
3. **Check formatting** with `gofmt`
4. **Run linter** if available (`golangci-lint run`)

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests
2. **Code Review**: Maintainers review for quality and consistency
3. **Approval**: At least one maintainer approval required
4. **Merge**: Squash and merge to maintain clean history

## 🏷️ Release Process

Releases follow semantic versioning:

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check the [[Home|documentation]] first

## 🙏 Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- GitHub contributor statistics

Thank you for contributing to k8s-tui! 🎉