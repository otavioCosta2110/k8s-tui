# Testing Guide

This guide covers the testing strategy, tools, and best practices for k8s-tui development.

## Testing Strategy

k8s-tui employs a multi-layered testing approach:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user workflows
- **Performance Tests**: Ensure responsive UI and efficient API calls

## Testing Tools

### Go Testing Framework

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests in short mode (skip integration tests)
go test -short ./...

# Run race detection
go test -race ./...
```

### Testify

Enhanced assertions and test utilities:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPodCreation(t *testing.T) {
    client := &mockClient{}
    pod := NewPod("test", "default", client)

    err := pod.Create("nginx", "latest", "default")

    assert.NoError(t, err)
    assert.Equal(t, "test", pod.Name)
    require.NotNil(t, pod.Raw)
}
```

### Mocking

Use interfaces for testable code:

```go
type Client interface {
    GetPod(namespace, name string) (*PodInfo, error)
    CreatePod(pod *corev1.Pod) error
}

// Mock implementation
type mockClient struct {
    pods map[string]*PodInfo
}

func (m *mockClient) GetPod(namespace, name string) (*PodInfo, error) {
    key := namespace + "/" + name
    if pod, exists := m.pods[key]; exists {
        return pod, nil
    }
    return nil, fmt.Errorf("pod not found")
}
```

## Unit Testing

### Testing Functions

```go
func ValidateResourceName(name string) error {
    if len(name) > 253 {
        return fmt.Errorf("name too long")
    }
    if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(name) {
        return fmt.Errorf("invalid name format")
    }
    return nil
}

func TestValidateResourceName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected error
    }{
        {"valid name", "my-pod", nil},
        {"empty name", "", fmt.Errorf("invalid name format")},
        {"too long", strings.Repeat("a", 254), fmt.Errorf("name too long")},
        {"invalid chars", "my_pod!", fmt.Errorf("invalid name format")},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateResourceName(tt.input)
            if tt.expected == nil {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expected.Error())
            }
        })
    }
}
```

### Testing Methods

```go
func (p *Pod) GetStatus() string {
    if p.Raw == nil {
        return "Unknown"
    }
    return string(p.Raw.Status.Phase)
}

func TestPod_GetStatus(t *testing.T) {
    pod := &Pod{
        Raw: &corev1.Pod{
            Status: corev1.PodStatus{
                Phase: corev1.PodRunning,
            },
        },
    }

    status := pod.GetStatus()
    assert.Equal(t, "Running", status)
}
```

### Testing UI Components

```go
func TestTextInput_Update(t *testing.T) {
    model := NewTextInput("Test", "placeholder", nil, nil)

    // Test initial state
    assert.Equal(t, "placeholder", model.Placeholder)

    // Test key input
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
    updated, cmd := model.Update(msg)

    updatedModel, ok := updated.(*TextInputModel)
    require.True(t, ok)
    assert.Equal(t, "h", updatedModel.textInput.Value())
    assert.NotNil(t, cmd)
}
```

## Integration Testing

### Component Integration

```go
func TestTableWithData(t *testing.T) {
    columns := []table.Column{
        {Title: "Name", Width: 20},
        {Title: "Status", Width: 10},
    }

    rows := []table.Row{
        {"pod-1", "Running"},
        {"pod-2", "Pending"},
    }

    table := NewTable(columns, rows, "Test Table")

    // Test rendering
    view := table.View()
    assert.Contains(t, view, "Test Table")
    assert.Contains(t, view, "pod-1")

    // Test selection
    msg := tea.KeyMsg{Type: tea.KeyDown}
    updated, _ := table.Update(msg)
    updatedTable := updated.(*TableModel)
    assert.Equal(t, 1, updatedTable.table.Cursor())
}
```

### API Integration

```go
func TestPodFetch_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    client := setupTestClient(t)
    pod := NewPod("nginx-pod", "default", client)

    // This would require a real k8s cluster
    err := pod.Fetch()
    if err != nil {
        t.Logf("Pod not found (expected in test env): %v", err)
    }
}
```

## End-to-End Testing

### Manual Testing Checklist

- [ ] Application starts without errors
- [ ] Tab navigation works
- [ ] Resource lists load
- [ ] Search functionality works
- [ ] Create forms submit correctly
- [ ] Edit functionality works
- [ ] Delete operations work
- [ ] Theme switching works
- [ ] Plugin loading works

### Automated E2E Tests

```bash
# Example using a test framework
func TestFullWorkflow(t *testing.T) {
    // Set up test environment
    app := setupTestApp(t)

    // Simulate user interactions
    app.SendKey("tab")      // Switch to pods tab
    app.SendKey("n")        // Create new pod
    app.TypeText("test-pod") // Enter pod name
    app.SendKey("tab")      // Next field
    app.TypeText("nginx")   // Enter image
    app.SendKey("enter")    // Submit

    // Verify pod was created
    pods := app.GetPods()
    assert.Contains(t, pods, "test-pod")
}
```

## Performance Testing

### Benchmarking

```go
func BenchmarkPodList(b *testing.B) {
    client := setupBenchmarkClient(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := GetPods(client, "default")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkTableRender(b *testing.B) {
    rows := generateTestRows(1000)
    table := NewTable(testColumns, rows, "Benchmark")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = table.View()
    }
}
```

### Memory Testing

```go
func TestMemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping memory test")
    }

    // Record initial memory
    var m runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m)
    initialAlloc := m.Alloc

    // Perform operations
    for i := 0; i < 1000; i++ {
        createTestResource(t)
    }

    // Check memory usage
    runtime.GC()
    runtime.ReadMemStats(&m)
    finalAlloc := m.Alloc

    // Allow some memory growth but not excessive
    growth := float64(finalAlloc-initialAlloc) / float64(initialAlloc)
    assert.Less(t, growth, 0.1, "Memory growth too high")
}
```

## Test Organization

### Directory Structure

```
internal/
├── app/
│   └── ui/
│       ├── components/
│       │   ├── table_test.go
│       │   └── textinput_test.go
│       └── models/
│           ├── pods_test.go
│           └── deployments_test.go
└── k8s/
    └── resources/
        ├── pod_test.go
        └── deployment_test.go
```

### Test Naming

```go
// Good
func TestPodCreate_ValidInput(t *testing.T)
func TestTableRender_EmptyState(t *testing.T)
func TestClient_GetPods_Error(t *testing.T)

// Bad
func TestPod(t *testing.T)
func TestTable(t *testing.T)
func TestError(t *testing.T)
```

### Test Helpers

```go
// test_helpers.go
func setupTestClient(t *testing.T) *Client {
    // Create mock or test client
}

func createTestPod(t *testing.T, name string) *Pod {
    client := setupTestClient(t)
    return NewPod(name, "default", client)
}

func assertPodEqual(t *testing.T, expected, actual *Pod) {
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Namespace, actual.Namespace)
}
```

## Continuous Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - run: go test -v -race -coverprofile=coverage.out ./...
    - uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Code Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage by function
go tool cover -func=coverage.out

# Minimum coverage check
go test -coverprofile=coverage.out ./...
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
if (( $(echo "$coverage < 80.0" | bc -l) )); then
    echo "Coverage $coverage% is below 80%"
    exit 1
fi
```

## Mocking and Stubs

### Kubernetes API Mocking

```go
import "k8s.io/client-go/kubernetes/fake"

func setupFakeClient() *Client {
    fakeClientset := fake.NewSimpleClientset()

    // Add test data
    fakeClientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: "test-pod",
            Namespace: "default",
        },
    }, metav1.CreateOptions{})

    return &Client{
        Clientset: fakeClientset,
        Namespace: "default",
    }
}
```

### HTTP Mocking

```go
import "net/http/httptest"

func setupMockServer() *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock API responses
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"apiVersion": "v1", "kind": "Pod"}`))
    }))
}
```

## Debugging Tests

### Verbose Output

```bash
go test -v -run TestPodCreate
```

### Debug Flags

```go
// Enable debug logging in tests
os.Setenv("K8S_TUI_LOG_LEVEL", "debug")

// Test with specific inputs
t.Logf("Testing with input: %v", testInput)
```

### Test Timeouts

```go
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test")
    }

    done := make(chan bool)
    go func() {
        slowOperation()
        done <- true
    }()

    select {
    case <-done:
        // Success
    case <-time.After(30 * time.Second):
        t.Fatal("Operation timed out")
    }
}
```

## Best Practices

### Test Isolation

- Each test should be independent
- Clean up resources after tests
- Use unique names to avoid conflicts

### Test Readability

- Use descriptive test names
- Include comments explaining complex tests
- Use table-driven tests for similar cases

### Performance

- Use `testing.Short()` for slow tests
- Parallelize tests when possible
- Mock external dependencies

### Maintenance

- Keep tests up to date with code changes
- Remove obsolete tests
- Refactor tests with code refactoring

This testing strategy ensures k8s-tui remains reliable, maintainable, and performant as it evolves.</content>
</xai:function_call">  

Assistant: Now build process. 

<xai:function_call name="write">
<parameter name="filePath">docs/developer/build-process.md