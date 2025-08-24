# Model Implementation Template

This document provides a comprehensive template for implementing new Kubernetes resource models in the k8s-tui project. Follow this pattern to ensure consistency and completeness.

## Overview

Each Kubernetes resource should have the following components:

1. **K8s Package** (`internal/k8s/`):
   - Resource type constant
   - Info struct (e.g., `ResourceInfo`)
   - Constructor function (`NewResource`)
   - Data fetching functions
   - Delete function
   - **Describe method** for detailed view

2. **Models Package** (`internal/ui/models/`):
   - Main model struct (e.g., `resourceModel`)
   - Constructor function (`NewResource`)
   - Data conversion methods
   - Component initialization
   - Details model (e.g., `resourceDetailsModel`)
   - Comprehensive tests

3. **Resource Integration**:
   - Add to `resource.go` functions
   - Add to `resource_data.go`
   - Update test files

## Step-by-Step Implementation

### 1. Add Resource Type Constant

**File**: `internal/k8s/client.go`

```go
const (
    // ... existing constants ...
    ResourceTypeNewResource    ResourceType = "newresource"
)
```

**File**: `internal/k8s/client_test.go`

```go
tests := []struct {
    name     string
    constant ResourceType
    expected string
}{
    // ... existing tests ...
    {"NewResource", ResourceTypeNewResource, "newresource"},
}
```

### 2. Create Resource K8s File

**File**: `internal/k8s/newresource.go`

```go
package k8s

import (
    "context"
    "fmt"
    "otaviocosta2110/k8s-tui/utils"
    "time"

    "gopkg.in/yaml.v3"
    corev1 "k8s.io/api/core/v1"
    // Import appropriate API group
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NewResourceInfo struct {
    Namespace string
    Name      string
    // Add relevant fields for your resource
    Age       string
    Raw       *ResourceType // Replace with actual Kubernetes type
    Client    Client
}

func NewNewResource(name, namespace string, k Client) *NewResourceInfo {
    return &NewResourceInfo{
        Name:      name,
        Namespace: namespace,
        Client:    k,
    }
}

func FetchNewResourceList(client Client, namespace string) ([]string, error) {
    resources, err := client.Clientset.API_GROUP().NewResources(namespace).List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to fetch newresources: %v", err)
    }

    names := make([]string, 0, len(resources.Items))
    for _, resource := range resources.Items {
        names = append(names, resource.Name)
    }

    return names, nil
}

func GetNewResourcesTableData(client Client, namespace string) ([]NewResourceInfo, error) {
    resources, err := client.Clientset.API_GROUP().NewResources(namespace).List(
        context.Background(),
        metav1.ListOptions{},
    )
    if err != nil {
        return nil, fmt.Errorf("failed to list newresources: %v", err)
    }

    var resourceInfos []NewResourceInfo
    for _, resource := range resources.Items {
        // Extract relevant data for table display
        resourceInfos = append(resourceInfos, NewResourceInfo{
            Namespace: resource.Namespace,
            Name:      resource.Name,
            // Set other fields...
            Age:       utils.FormatAge(resource.CreationTimestamp.Time),
            Raw:       resource.DeepCopy(),
            Client:    client,
        })
    }

    return resourceInfos, nil
}

func (n *NewResourceInfo) Fetch() error {
    resource, err := n.Client.Clientset.API_GROUP().NewResources(n.Namespace).Get(context.Background(), n.Name, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("failed to get newresource: %v", err)
    }
    n.Raw = resource
    return nil
}

func (n *NewResourceInfo) Describe() (string, error) {
    if n.Raw == nil {
        if err := n.Fetch(); err != nil {
            return "", fmt.Errorf("failed to fetch newresource: %v", err)
        }
    }

    events, err := n.Client.Clientset.CoreV1().Events(n.Namespace).List(context.Background(), metav1.ListOptions{
        FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=NewResource", n.Name, n.Namespace),
    })
    if err != nil {
        return "", fmt.Errorf("failed to get newresource events: %v", err)
    }

    data, err := n.DescribeNewResource(events)
    if err != nil {
        return "", fmt.Errorf("failed to describe newresource: %v", err)
    }

    yamlData, err := yaml.Marshal(data)
    if err != nil {
        return "", fmt.Errorf("failed to marshal newresource to YAML: %v", err)
    }

    return string(yamlData), nil
}

func (n *NewResourceInfo) DescribeNewResource(events *corev1.EventList) (map[string]any, error) {
    type Event struct {
        Type    string `yaml:"type"`
        Reason  string `yaml:"reason"`
        Age     string `yaml:"age"`
        From    string `yaml:"from"`
        Message string `yaml:"message"`
    }

    desc := map[string]any{
        "name":        n.Name,
        "namespace":   n.Namespace,
        "labels":      n.Raw.Labels,
        "annotations": n.Raw.Annotations,
        "created":     formatTime(n.Raw.CreationTimestamp),
        // Add resource-specific fields
    }

    // Add events
    if len(events.Items) > 0 {
        eventList := make([]Event, 0)
        for _, event := range events.Items {
            age := time.Since(event.LastTimestamp.Time).Round(time.Second)
            eventList = append(eventList, Event{
                Type:    event.Type,
                Reason:  event.Reason,
                Age:     age.String(),
                From:    event.Source.Component,
                Message: event.Message,
            })
        }
        desc["events"] = eventList
    }

    return desc, nil
}

func DeleteNewResource(client Client, namespace string, resourceName string) error {
    err := client.Clientset.API_GROUP().NewResources(namespace).Delete(context.Background(), resourceName, metav1.DeleteOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete newresource %s: %v", resourceName, err)
    }
    return nil
}
```

### 3. Update Resource Functions

**File**: `internal/k8s/resource.go`

Add cases for the new resource in:
- `DeleteResource()`
- `ListResources()`
- `GetResourceInfo()`

### 4. Create Main Model

**File**: `internal/ui/models/newresources.go`

```go
package models

import (
    "fmt"
    "otaviocosta2110/k8s-tui/internal/k8s"
    "otaviocosta2110/k8s-tui/internal/ui/components"
    ui "otaviocosta2110/k8s-tui/internal/ui/components"
    "time"

    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
)

type newResourcesModel struct {
    *GenericResourceModel
    newResourcesInfo []k8s.NewResourceInfo
}

func NewNewResources(k k8s.Client, namespace string) (*newResourcesModel, error) {
    config := ResourceConfig{
        ResourceType:    k8s.ResourceTypeNewResource,
        Title:           "NewResources in " + namespace,
        ColumnWidths:    []float64{0.15, 0.25, 0.15, 0.15, 0.15, 0.15}, // IMPORTANT: Must match number of columns exactly
        RefreshInterval: 5 * time.Second,
        Columns: []table.Column{
            components.NewColumn("NAMESPACE", 0),
            components.NewColumn("NAME", 0),
            // Add columns for your resource fields
            components.NewColumn("AGE", 0),
        },
    }

    genericModel := NewGenericResourceModel(k, namespace, config)

    model := &newResourcesModel{
        GenericResourceModel: genericModel,
    }

    return model, nil
}

func (n *newResourcesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
    n.k8sClient = k

    if err := n.fetchData(); err != nil {
        return nil, err
    }

    onSelect := func(selected string) tea.Msg {
        resourceDetails, err := NewNewResourceDetails(*k, n.namespace, selected).InitComponent(k)
        if err != nil {
            return components.NavigateMsg{
                Error:   err,
                Cluster: *k,
            }
        }
        return components.NavigateMsg{
            NewScreen: resourceDetails,
        }
    }

    fetchFunc := func() ([]table.Row, error) {
        if err := n.fetchData(); err != nil {
            return nil, err
        }
        return n.dataToRows(), nil
    }

    tableModel := ui.NewTable(n.config.Columns, n.config.ColumnWidths, n.dataToRows(), n.config.Title, onSelect, 1, fetchFunc, nil)

    actions := map[string]func() tea.Cmd{
        "d": n.createDeleteAction(tableModel),
    }
    tableModel.SetUpdateActions(actions)

    return &autoRefreshModel{
        inner:           tableModel,
        refreshInterval: n.refreshInterval,
        k8sClient:       n.k8sClient,
    }, nil
}

func (n *newResourcesModel) fetchData() error {
    resourceInfo, err := k8s.GetNewResourcesTableData(*n.k8sClient, n.namespace)
    if err != nil {
        return fmt.Errorf("failed to fetch newresources: %v", err)
    }
    n.newResourcesInfo = resourceInfo

    n.resourceData = make([]ResourceData, len(resourceInfo))
    for idx, resource := range resourceInfo {
        n.resourceData[idx] = NewResourceData{&resource}
    }

    return nil
}

func (n *newResourcesModel) dataToRows() []table.Row {
    rows := make([]table.Row, len(n.newResourcesInfo))
    for idx, resource := range n.newResourcesInfo {
        rows[idx] = table.Row{
            resource.Namespace,
            resource.Name,
            // Add other fields...
            resource.Age,
        }
    }
    return rows
}
```

### 5. Create Details Model

**File**: `internal/ui/models/newresource_details.go`

```go
package models

import (
    "otaviocosta2110/k8s-tui/internal/k8s"
    "otaviocosta2110/k8s-tui/internal/ui/components"

    tea "github.com/charmbracelet/bubbletea"
)

type newResourceDetailsModel struct {
    resource   *k8s.NewResourceInfo
    k8sClient  *k8s.Client
    loading    bool
    err        error
}

func NewNewResourceDetails(k k8s.Client, namespace, resourceName string) *newResourceDetailsModel {
    return &newResourceDetailsModel{
        resource:   k8s.NewNewResource(resourceName, namespace, k),
        k8sClient:  &k,
        loading:    false,
        err:        nil,
    }
}

func (n *newResourceDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
    n.k8sClient = k

    desc, err := n.resource.Describe()
    if err != nil {
        return nil, err
    }

    return components.NewYAMLViewer("NewResource: "+n.resource.Name, desc), nil
}
```

### 6. Add Resource Data

**File**: `internal/ui/models/resource_data.go`

```go
type NewResourceData struct {
    *k8s.NewResourceInfo
}

func (n NewResourceData) GetName() string {
    return n.Name
}

func (n NewResourceData) GetNamespace() string {
    return n.Namespace
}

func (n NewResourceData) GetColumns() table.Row {
    return table.Row{
        n.Namespace,
        n.Name,
        // Add other fields...
        n.Age,
    }
}
```

### 7. Create Comprehensive Tests

**File**: `internal/ui/models/newresources_test.go`

```go
package models

import (
    "otaviocosta2110/k8s-tui/internal/k8s"
    "testing"
    "time"
)

func TestNewNewResources(t *testing.T) {
    client := k8s.Client{Namespace: "default"}
    model, err := NewNewResources(client, "default")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if model == nil {
        t.Error("Expected model to be non-nil")
    }
    if model.namespace != "default" {
        t.Error("Expected namespace to be 'default'")
    }
    if model.config.ResourceType != k8s.ResourceTypeNewResource {
        t.Error("Expected ResourceType to be ResourceTypeNewResource")
    }
}

func TestNewResourcesModelDataToRows(t *testing.T) {
    client := k8s.Client{Namespace: "default"}
    model, err := NewNewResources(client, "default")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    // Set mock data
    model.newResourcesInfo = []k8s.NewResourceInfo{
        {
            Name:      "test-resource",
            Namespace: "default",
            // Set other fields...
            Age:       "1h",
        },
    }

    rows := model.dataToRows()
    if len(rows) != 1 {
        t.Error("Expected 1 row")
    }
    if rows[0][1] != "test-resource" {
        t.Error("Resource name mismatch in row")
    }
}

func TestNewResourcesModelWithEmptyData(t *testing.T) {
    client := k8s.Client{Namespace: "default"}
    model, err := NewNewResources(client, "default")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    rows := model.dataToRows()
    if len(rows) != 0 {
        t.Error("Expected 0 rows for empty data")
    }
}

func TestNewResourcesModelConfig(t *testing.T) {
    client := k8s.Client{Namespace: "test-namespace"}
    model, err := NewNewResources(client, "test-namespace")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if model.config.ResourceType != k8s.ResourceTypeNewResource {
        t.Error("Config ResourceType not set correctly")
    }
    if model.config.Title != "NewResources in test-namespace" {
        t.Error("Config Title not set correctly")
    }
    if model.config.RefreshInterval != 5*time.Second {
        t.Error("Config RefreshInterval not set correctly")
    }
}

func TestNewResourceDetails(t *testing.T) {
    client := k8s.Client{Namespace: "default"}
    model := NewNewResourceDetails(client, "default", "test-resource")

    if model == nil {
        t.Error("Expected model to be non-nil")
    }
    if model.resource.Name != "test-resource" {
        t.Error("Expected resource name to be 'test-resource'")
    }
    if model.resource.Namespace != "default" {
        t.Error("Expected resource namespace to be 'default'")
    }
}
```

## Checklist

Before completing implementation, verify:

- [ ] Resource type constant added to `client.go`
- [ ] Resource type test added to `client_test.go`
- [ ] K8s file created with all necessary functions
- [ ] Describe method implemented with proper YAML output
- [ ] Resource functions updated in `resource.go`
- [ ] Main model created with proper table structure
- [ ] **ColumnWidths array length matches number of Columns exactly** (critical - causes panic if mismatched)
- [ ] Details model created for navigation
- [ ] Resource data struct added to `resource_data.go`
- [ ] Comprehensive tests created (including column width tests)
- [ ] All tests pass
- [ ] Code builds successfully

## Notes

1. **API Groups**: Replace `API_GROUP()` with the appropriate Kubernetes API group (e.g., `AppsV1()`, `CoreV1()`, `NetworkingV1()`)

2. **Resource Types**: Replace `*ResourceType` with the actual Kubernetes resource type (e.g., `*appsv1.Deployment`, `*corev1.Pod`)

3. **Column Configuration**: Adjust column widths and names based on the specific resource fields

4. **Error Handling**: Follow the existing error handling patterns in the codebase

5. **Naming Conventions**: Use consistent naming (e.g., `NewResourceInfo`, `NewResourceData`, `newResourcesModel`)

6. **Testing**: Include tests for various scenarios including empty data, multiple items, and different states