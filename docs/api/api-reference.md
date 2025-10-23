# API Reference

This document provides a comprehensive reference for k8s-tui's internal APIs and interfaces.

## Go API Reference

### Core Interfaces

#### ResourceData Interface

```go
type ResourceData interface {
    GetName() string
    GetNamespace() string
    GetColumns() table.Row
}
```

Implemented by all resource types to provide consistent data access.

#### Client Interface

```go
type Client struct {
    Clientset      kubernetes.Interface
    Config         *rest.Config
    Namespace      string
    KubeconfigPath string
}
```

Main Kubernetes client interface.

**Methods:**
- `NewClient(kubeconfigPath, namespace string) (*Client, error)`
- `GetPods(namespace string) ([]PodInfo, error)`
- `GetDeployments(namespace string) ([]DeploymentInfo, error)`
- `GetServices(namespace string) ([]ServiceInfo, error)`

### Resource Types

#### Pod Operations

```go
type Pod struct {
    Name      string
    Namespace string
    Client    Client
    Raw       *corev1.Pod
}

func NewPod(name, namespace string, client Client) *Pod
func (p *Pod) Create(name, image, namespace string) error
func (p *Pod) Delete() error
func (p *Pod) Fetch() error
func (p *Pod) GetLogs(container string, follow bool) (string, error)
func (p *Pod) Exec(command []string) error
```

#### Deployment Operations

```go
type DeploymentInfo struct {
    Name      string
    Namespace string
    Ready     string
    UpToDate  string
    Available string
    Age       string
    Raw       *appsv1.Deployment
    Client    Client
}

func NewDeployment(name, namespace string, client Client) *DeploymentInfo
func (d *DeploymentInfo) Create(name, image, namespace string) error
func (d *DeploymentInfo) Scale(replicas int32) error
func (d *DeploymentInfo) GetPods() ([]PodInfo, error)
func (d *DeploymentInfo) Describe() (string, error)
```

#### Service Operations

```go
type ServiceInfo struct {
    Name       string
    Namespace  string
    Type       string
    ClusterIP  string
    ExternalIP string
    Ports      string
    Age        string
    Raw        *corev1.Service
    Client     Client
}

func NewService(name, namespace string, client Client) *ServiceInfo
func (s *ServiceInfo) Create(name, svcType, port, targetPort, namespace string) error
func (s *ServiceInfo) GetEndpoints() ([]EndpointInfo, error)
```

### UI Components

#### Table Component

```go
type TableModel struct {
    table     table.Model
    columns   []table.Column
    rows      []table.Row
    title     string
    onSelect  func(string) tea.Msg
    selection int
}

func NewTable(columns []table.Column, columnWidths []float64, rows []table.Row, title string, onSelect func(string) tea.Msg, selection int, fetchFunc func() ([]table.Row, error), updateFunc func() tea.Cmd) *TableModel
func (t *TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (t *TableModel) View() string
func (t *TableModel) SetUpdateActions(actions map[string]func() tea.Cmd)
```

#### Form Components

```go
type CreateForm struct {
    inputs   []textinput.Model
    fields   []Field
    current  int
    title    string
    onSubmit func(map[string]string) tea.Msg
    onCancel func() tea.Msg
}

type Field struct {
    Name         string
    Placeholder  string
    DefaultValue string
    Row          int
}

func NewCreateForm(title string, resourceType string, fields []Field) *CreateForm
func (f *CreateForm) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (f *CreateForm) View() string
```

### Model Interfaces

#### GenericResourceModel

```go
type GenericResourceModel struct {
    k8sClient     *k8s.Client
    namespace     string
    config        ResourceConfig
    data          []types.ResourceData
    table         *ui.TableModel
    refreshInterval time.Duration
    lastRefresh   time.Time
}

type ResourceConfig struct {
    ResourceType    types.ResourceType
    Title           string
    ColumnWidths    []float64
    RefreshInterval time.Duration
    Columns         []table.Column
}

func NewGenericResourceModel(client *k8s.Client, namespace string, config ResourceConfig) *GenericResourceModel
```

## Lua Plugin API

### Core Modules

#### k8s Module

```lua
-- Get current context
local context = k8s.get_context()

-- List resources
local pods = k8s.list_pods(namespace)
local deployments = k8s.list_deployments(namespace)
local services = k8s.list_services(namespace)

-- Get specific resource
local pod = k8s.get_pod(namespace, name)
local deployment = k8s.get_deployment(namespace, name)

-- Execute kubectl commands
local result = k8s.exec("get pods -o json")

-- Watch resources
k8s.watch_pods(namespace, function(event, pod)
    -- Handle pod events
end)
```

#### ui Module

```lua
-- Notifications
ui.notify(message, level)  -- level: "info", "warn", "error"

-- Dialogs
local confirmed = ui.confirm(message)
local input = ui.prompt(message, default)

-- Navigation
ui.switch_tab(tab_name)
ui.open_resource(resource_type, name, namespace)

-- External actions
ui.open_url(url)
ui.open_editor(file_path)
```

#### log Module

```lua
log.info(message)
log.warn(message)
log.error(message)
log.debug(message)
log.set_level(level)  -- "debug", "info", "warn", "error"
```

#### http Module

```lua
-- GET request
local response = http.get(url, options)

-- POST request
local response = http.post(url, {
    headers = {["Content-Type"] = "application/json"},
    body = json.encode(data)
})

-- Response structure
response = {
    status = 200,
    headers = {...},
    body = "..."
}
```

### Plugin Hooks

```lua
function plugin.on_startup()
    -- Called when k8s-tui starts
end

function plugin.on_shutdown()
    -- Called when k8s-tui shuts down
end

function plugin.on_tab_changed(tab_name)
    -- Called when user switches tabs
end

function plugin.on_resource_selected(resource_type, name, namespace)
    -- Called when user selects a resource
end

function plugin.on_resource_created(resource_type, name, namespace)
    -- Called when a resource is created
end

function plugin.on_resource_deleted(resource_type, name, namespace)
    -- Called when a resource is deleted
end

function plugin.on_key_pressed(key)
    -- Called for every key press
    -- Return true to consume the key
end
```

### Custom Views

```lua
views.register("my-view", {
    title = "My Custom View",
    icon = "🔧",

    list = function(namespace)
        -- Return array of items to display
        return {
            {name = "item1", status = "active"},
            {name = "item2", status = "inactive"}
        }
    end,

    detail = function(namespace, name)
        -- Return detailed information
        return {
            name = name,
            status = "active",
            created = "2024-01-01T00:00:00Z"
        }
    end,

    actions = {
        "start",
        "stop",
        "restart"
    },

    on_action = function(namespace, name, action)
        -- Handle custom actions
        if action == "start" then
            -- Start logic
        end
    end
})
```

### Commands

```lua
commands.register("my-command", function(args)
    -- Handle custom command
    log.info("Executing my-command with args: " .. table.concat(args, " "))
end)

-- Usage in k8s-tui:
-- :my-command arg1 arg2 arg3
```

### Configuration

```lua
-- Access configuration
local config = plugin.get_config()

-- Plugin-specific config
local api_url = config.api_url
local timeout = config.timeout or 30
```

## REST API (Future)

When k8s-tui exposes a REST API, it will provide endpoints for:

### Resources

```
GET    /api/v1/resources/{type}           # List resources
GET    /api/v1/resources/{type}/{name}    # Get resource
POST   /api/v1/resources/{type}           # Create resource
PUT    /api/v1/resources/{type}/{name}    # Update resource
DELETE /api/v1/resources/{type}/{name}    # Delete resource
```

### Plugins

```
GET    /api/v1/plugins                     # List plugins
POST   /api/v1/plugins/{name}/install      # Install plugin
DELETE /api/v1/plugins/{name}              # Uninstall plugin
```

### System

```
GET    /api/v1/health                      # Health check
GET    /api/v1/metrics                     # Metrics
GET    /api/v1/logs                        # Application logs
```

## Error Codes

### Kubernetes Errors

- `ECONNREFUSED`: Cannot connect to Kubernetes API
- `EUNAUTHORIZED`: Authentication failed
- `EFORBIDDEN`: Insufficient permissions
- `ENOTFOUND`: Resource not found
- `ECONFLICT`: Resource version conflict

### Application Errors

- `EINVALIDINPUT`: Invalid user input
- `EPLUGINERROR`: Plugin execution failed
- `ECONFIGERROR`: Configuration error
- `ENETWORKERROR`: Network communication failed

## Data Types

### Resource Objects

```typescript
interface Resource {
    apiVersion: string;
    kind: string;
    metadata: {
        name: string;
        namespace: string;
        uid: string;
        creationTimestamp: string;
        labels: Record<string, string>;
        annotations: Record<string, string>;
    };
    spec: any;
    status: any;
}
```

### Event Objects

```typescript
interface Event {
    type: "ADDED" | "MODIFIED" | "DELETED";
    resource: Resource;
    oldResource?: Resource;
}
```

### Plugin Objects

```typescript
interface Plugin {
    name: string;
    version: string;
    description: string;
    author: string;
    permissions: string[];
    config: Record<string, any>;
}
```

This API reference provides the foundation for extending and integrating with k8s-tui programmatically.</content>
</xai:function_call">  

Assistant: Now extension points. 

<xai:function_call name="write">
<parameter name="filePath">docs/api/extension-points.md