# 🧪 Testing

This document covers the testing strategy, guidelines, and best practices for k8s-tui.

## 🎯 Testing Strategy

k8s-tui uses a multi-layered testing approach to ensure reliability and maintainability:

```
Testing Pyramid
    ┌─────────────────┐
    │   E2E Tests     │  ← Few, slow, high value
    ├─────────────────┤
    │ Integration     │  ← Moderate number, medium speed
    ├─────────────────┤
    │  Unit Tests     │  ← Many, fast, foundational
    └─────────────────┘
```

## 🧪 Unit Tests

Unit tests focus on individual functions and methods in isolation.

### Guidelines

1. **Test Public APIs**: Focus on testing exported functions and methods
2. **Mock Dependencies**: Use interfaces to mock external dependencies
3. **Table-Driven Tests**: Use for testing multiple scenarios
4. **Edge Cases**: Test boundary conditions and error cases

### Example Structure

```go
func TestPodModel_UpdateStatus(t *testing.T) {
    tests := []struct {
        name           string
        initialPods    []v1.Pod
        expectedStatus  string
        expectedError  bool
    }{
        {
            name: "running pod",
            initialPods: []v1.Pod{
                {
                    ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
                    Status: v1.PodStatus{Phase: v1.PodRunning},
                },
            },
            expectedStatus: "Running",
            expectedError:  false,
        },
        {
            name:          "empty pod list",
            initialPods:    []v1.Pod{},
            expectedStatus: "",
            expectedError:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockClient := &MockK8sClient{pods: tt.initialPods}
            model := NewPodModel(mockClient)

            // Execute
            err := model.UpdateStatus(context.Background())

            // Assert
            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedStatus, model.GetStatus())
            }
        })
    }
}
```

### Mock Objects

Create mock implementations for external dependencies:

```go
// MockK8sClient implements the k8s.Client interface for testing
type MockK8sClient struct {
    pods   []v1.Pod
    error  error
    called bool
}

func (m *MockK8sClient) ListPods(ctx context.Context) ([]v1.Pod, error) {
    m.called = true
    return m.pods, m.error
}

func (m *MockK8sClient) GetPod(ctx context.Context, name string) (*v1.Pod, error) {
    for _, pod := range m.pods {
        if pod.Name == name {
            return &pod, nil
        }
    }
    return nil, fmt.Errorf("pod not found: %s", name)
}
```

## 🔗 Integration Tests

Integration tests verify that multiple components work together correctly.

### Kubernetes Integration

```go
func TestKubernetesIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup test cluster (using envtest or kind)
    testEnv := &envtest.Environment{
        CRDDirectoryPaths: []string{filepath.Join("..", "..", "testdata", "crds")},
    }
    
    cfg, err := testEnv.Start()
    require.NoError(t, err)
    defer testEnv.Stop()

    // Create real client
    client, err := k8s.NewClient(cfg)
    require.NoError(t, err)

    // Test real operations
    pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
        },
        Spec: v1.PodSpec{
            Containers: []v1.Container{
                {Name: "test", Image: "nginx:alpine"},
            },
        },
    }

    created, err := client.CreatePod(context.Background(), pod)
    assert.NoError(t, err)
    assert.Equal(t, "test-pod", created.Name)
}
```

### UI Component Integration

```go
func TestTableComponentIntegration(t *testing.T) {
    // Setup
    model := NewTableModel([]string{"Name", "Status", "Age"})
    model.SetRows([]TableRow{
        {"pod-1", "Running", "1d"},
        {"pod-2", "Pending", "5m"},
    })

    // Test tea.Model interface
    _, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
    assert.NotNil(t, cmd)

    // Test view rendering
    view := model.View()
    assert.Contains(t, view, "pod-1")
    assert.Contains(t, view, "Running")
}
```

## 🌐 End-to-End Tests

E2E tests simulate real user workflows using the actual application.

### Test Scenarios

```go
func TestE2E_PodManagement(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Start application in test mode
    app := NewApp(WithTestConfig())
    go func() {
        if err := app.Run(); err != nil {
            t.Errorf("App failed: %v", err)
        }
    }()

    // Simulate user interactions
    app.SendKey(tea.KeyMsg{Type: tea.KeyEnter})
    app.SendKey(tea.KeyMsg{Type: tea.KeyDown})
    app.SendKey(tea.KeyMsg{Type: tea.KeyEnter})

    // Verify results
    time.Sleep(100 * time.Millisecond)
    assert.True(t, app.IsInDetailView())
}
```

## 📊 Test Coverage

### Coverage Goals

- **Overall Coverage**: >80%
- **Core Logic**: >90%
- **UI Components**: >85%
- **Kubernetes Integration**: >75%

### Coverage Commands

```bash
# Run tests with coverage
go test -v -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Check coverage by package
go test -coverprofile=coverage.out ./internal/k8s/resources
go tool cover -func=coverage.out
```

### Coverage Exclusions

```go
//go:generate go run github.com/cheekybits/gendoc -doc doc.go
//go:generate mockgen -source=client.go -destination=mock_client.go

// Exclude from coverage
//go:exclude coverage
func (m *PodModel) init() {
    // Initialization code not worth testing
}
```

## 🏃‍♂️ Running Tests

### Local Development

```bash
# Run all tests
go test -v ./...

# Run specific package
go test -v ./internal/k8s/resources

# Run with race detection
go test -race -v ./...

# Run benchmarks
go test -bench=. -v ./...

# Run tests with coverage
go test -cover -v ./...
```

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Run tests
        run: go test -v -race -cover ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## 🛠️ Testing Tools

### Testing Frameworks

- **Standard Library**: `testing` package
- **Assertions**: `testify/assert` for readable assertions
- **Mocking**: `testify/mock` or `gomock`
- **HTTP Testing**: `net/http/httptest`

### Kubernetes Testing

- **envtest**: Integration testing with real Kubernetes API
- **kind**: Local Kubernetes clusters for E2E testing
- **fake client**: `k8s.io/client-go/kubernetes/fake`

### UI Testing

- **Bubble Tea Testing**: Direct model testing
- **Terminal Emulation**: `github.com/maaslalani/typer` for terminal testing

## 📋 Test Data Management

### Fixtures

```go
// testdata/fixtures/pod.json
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "test-pod",
    "namespace": "default"
  },
  "spec": {
    "containers": [{
      "name": "test",
      "image": "nginx:alpine"
    }]
  }
}
```

### Test Utilities

```go
// testutils/utils.go
package testutils

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func CreateTestPod(t *testing.T, name string) *v1.Pod {
    return &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: "default",
            Labels: map[string]string{
                "test": "true",
            },
        },
        Spec: v1.PodSpec{
            Containers: []v1.Container{
                {
                    Name:  "test",
                    Image: "nginx:alpine",
                },
            },
        },
    }
}

func AssertNoError(t *testing.T, err error) {
    require.NoError(t, err, "Unexpected error: %v", err)
}
```

## 🔍 Performance Testing

### Benchmark Tests

```go
func BenchmarkPodList_Large(b *testing.B) {
    // Setup large dataset
    pods := make([]v1.Pod, 1000)
    for i := range pods {
        pods[i] = *CreateTestPod(fmt.Sprintf("pod-%d", i))
    }

    model := NewPodModel(&MockK8sClient{pods: pods})

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        model.UpdatePods(context.Background())
    }
}
```

### Memory Testing

```go
func TestMemoryUsage_PodModel(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)

    // Create many pod models
    models := make([]*PodModel, 1000)
    for i := range models {
        models[i] = NewPodModel(&MockK8sClient{})
    }

    runtime.GC()
    runtime.ReadMemStats(&m2)

    memUsed := m2.Alloc - m1.Alloc
    assert.Less(t, memUsed, uint64(10*1024*1024)) // Less than 10MB
}
```

## 🐛 Debugging Tests

### Verbose Output

```bash
# Verbose test output
go test -v ./...

# Test with specific flags
go test -v -run TestPodModel ./internal/k8s/resources

# Show test coverage in detail
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "total|PodModel"
```

### Test Debugging

```go
func TestDebuggingExample(t *testing.T) {
    if testing.Verbose() {
        t.Log("Debug information")
        fmt.Printf("Debug: %+v\n", someStruct)
    }

    // Use testify's require for immediate failure
    require.NotNil(t, someValue, "Value should not be nil")
    
    // Use assert for continued testing
    assert.Equal(t, "expected", actual, "Values should match")
}
```

## 📝 Best Practices

### DO ✅

1. **Write tests first** (TDD when possible)
2. **Test error paths** as well as success paths
3. **Use descriptive test names**
4. **Keep tests simple and focused**
5. **Mock external dependencies**
6. **Use table-driven tests** for multiple scenarios
7. **Test edge cases and boundaries**
8. **Maintain high test coverage**

### DON'T ❌

1. **Don't test implementation details**
2. **Don't ignore test failures**
3. **Don't write complex test logic**
4. **Don't rely on external services** in unit tests
5. **Don't use sleep()** for synchronization
6. **Don't test private functions** directly
7. **Don't skip error handling** in tests
8. **Don't write tests without assertions**

## 🔄 Continuous Testing

### Watch Mode

```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Run tests on file changes
air -c .air.toml
```

### Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests before commit
go test ./... || exit 1

# Run linter
golangci-lint run || exit 1

# Format code
gofmt -s -w . || exit 1
```

This comprehensive testing approach ensures k8s-tui remains reliable, maintainable, and bug-free across releases.