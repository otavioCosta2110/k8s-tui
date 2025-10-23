# Architecture Overview

k8s-tui is built with a modular, event-driven architecture using the Bubble Tea framework for terminal UI development.

## High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Terminal UI   │◄──►│  Application    │◄──►│  Kubernetes     │
│   (Bubble Tea)  │    │   Logic         │    │   API Client    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Components    │    │   Models        │    │   Resources     │
│   (UI Widgets)  │    │   (Screens)     │    │   (API Calls)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. Terminal UI Layer (Bubble Tea)

**Framework:** [Charm's Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Event-driven:** Handles keyboard input, updates, and rendering
- **Composable:** UI components can be nested and combined
- **Functional:** Pure functions for state updates

**Key Concepts:**
- **Model:** State representation
- **Update:** State transition function
- **View:** Rendering function
- **Cmd:** Side effects (async operations)

### 2. Application Logic Layer

**Main Components:**
- **AppModel:** Central state management
- **Tab Manager:** Handles multiple resource screens
- **Update Handlers:** Process user actions and API responses
- **UI Injector:** Manages component lifecycle

**State Management:**
- Centralized application state
- Event-driven updates
- Immutable state transitions

### 3. Kubernetes API Layer

**Client:** Official Kubernetes Go client
- **REST API:** Direct HTTP calls to Kubernetes API server
- **Authentication:** Supports all k8s auth methods
- **Caching:** Intelligent resource caching
- **Watching:** Real-time resource updates

## Component Architecture

### UI Components

Located in `internal/app/ui/components/`

```
components/
├── create_form.go      # Generic resource creation forms
├── editor.go           # YAML editor integration
├── help.go             # Context-sensitive help
├── list.go             # Generic list views
├── navigatemsg.go      # Navigation messages
├── spinner.go          # Loading indicators
├── table.go            # Tabular data display
├── textinput.go        # Text input widgets
├── yaml.go             # YAML viewer/editor
└── ...
```

**Design Patterns:**
- **Composition over inheritance**
- **Functional options**
- **Interface-based design**

### Resource Models

Located in `internal/app/ui/models/`

```
models/
├── pods.go             # Pod management screen
├── deployments.go      # Deployment management
├── services.go         # Service management
├── configmaps.go       # ConfigMap management
├── secrets.go          # Secret management
├── ingresses.go        # Ingress management
├── jobs.go             # Job management
├── cronjobs.go         # CronJob management
├── daemonsets.go       # DaemonSet management
├── statefulsets.go     # StatefulSet management
├── namespaces.go       # Namespace management
├── nodes.go            # Node management
└── ...
```

**Model Structure:**
- **GenericResourceModel:** Base functionality
- **Resource-specific models:** Specialized behavior
- **Auto-refresh models:** Background updates

### Kubernetes Resources

Located in `internal/k8s/resources/`

```
resources/
├── client.go           # Kubernetes client setup
├── pod.go              # Pod operations
├── deployment.go       # Deployment operations
├── service.go          # Service operations
├── configmap.go        # ConfigMap operations
├── secret.go           # Secret operations
├── ingress.go          # Ingress operations
├── job.go              # Job operations
├── cronjob.go          # CronJob operations
├── daemonset.go        # DaemonSet operations
├── statefulset.go      # StatefulSet operations
├── namespace.go        # Namespace operations
├── node.go             # Node operations
└── ...
```

**Resource Operations:**
- **CRUD operations:** Create, Read, Update, Delete
- **Listing:** Paginated resource lists
- **Watching:** Real-time updates
- **Relationships:** Cross-resource navigation

## Data Flow

### User Input Flow

```
User Input → Bubble Tea → Update Handler → State Update → View Render
```

### API Request Flow

```
User Action → Update Handler → Resource Method → K8s API → Response → State Update
```

### Plugin Flow

```
Plugin Event → Plugin Manager → Lua Runtime → UI Update → State Change
```

## Key Design Decisions

### 1. Event-Driven Architecture

- **Reactive:** UI responds to state changes
- **Decoupled:** Components communicate via messages
- **Testable:** Pure functions for state transitions

### 2. Separation of Concerns

- **UI Layer:** Pure presentation logic
- **Business Logic:** State management and validation
- **Data Layer:** API interactions and caching

### 3. Plugin Architecture

- **Extensible:** Lua-based plugin system
- **Isolated:** Plugins run in separate runtime
- **Secure:** Sandboxed execution environment

### 4. Resource Abstraction

- **Generic Models:** Common functionality shared
- **Specialized Views:** Resource-specific displays
- **Unified API:** Consistent interface across resources

## Performance Considerations

### 1. Caching Strategy

- **Resource caching:** Avoid redundant API calls
- **Smart invalidation:** Cache based on resource version
- **Background updates:** Non-blocking refresh operations

### 2. Memory Management

- **Lazy loading:** Load data on demand
- **Pagination:** Limit displayed items
- **Garbage collection:** Clean up unused resources

### 3. Network Optimization

- **Batching:** Group related API calls
- **Compression:** Efficient data transfer
- **Connection pooling:** Reuse HTTP connections

## Security Architecture

### 1. Authentication

- **Kubernetes auth:** Leverages k8s authentication
- **Token management:** Secure credential handling
- **RBAC integration:** Respects cluster permissions

### 2. Plugin Security

- **Sandboxing:** Isolated Lua execution
- **Permission model:** Granular access controls
- **Code review:** Plugin validation

### 3. Data Protection

- **No data persistence:** Sensitive data not stored
- **Secure communication:** HTTPS to API server
- **Audit logging:** Operation tracking

## Extensibility Points

### 1. Plugin API

- **Custom views:** New resource displays
- **Enhanced functionality:** Additional operations
- **Integration:** External service connections

### 2. Component Library

- **Reusable widgets:** Custom UI components
- **Theming:** Color scheme extensions
- **Layout:** Alternative screen layouts

### 3. Resource Types

- **CRD support:** Custom resource definitions
- **Provider extensions:** Cloud-specific resources
- **Custom operations:** Specialized workflows

## Deployment Architecture

### 1. Standalone Application

- **Single binary:** Easy distribution
- **No external dependencies:** Self-contained
- **Cross-platform:** Linux, macOS, Windows

### 2. Container Deployment

- **Docker images:** Official container images
- **Kubernetes deployment:** Run in-cluster
- **Service account:** RBAC permissions

### 3. CI/CD Integration

- **Automated builds:** GitHub Actions
- **Release automation:** Version tagging
- **Binary distribution:** GitHub Releases

## Monitoring and Observability

### 1. Logging

- **Structured logging:** JSON format logs
- **Log levels:** Debug, info, warn, error
- **Configurable output:** File and console

### 2. Metrics

- **Performance metrics:** Response times, error rates
- **Resource usage:** Memory, CPU, network
- **User analytics:** Feature usage tracking

### 3. Error Handling

- **Graceful degradation:** Continue on partial failures
- **User feedback:** Clear error messages
- **Recovery mechanisms:** Automatic retry logic

## Future Architecture Considerations

### 1. Microservices

- **API server:** Separate backend service
- **Web UI:** Browser-based interface
- **Mobile clients:** Native mobile apps

### 2. Cloud Integration

- **Multi-cloud:** AWS, GCP, Azure support
- **Managed services:** EKS, GKE, AKS integration
- **Hybrid deployments:** On-prem + cloud

### 3. Advanced Features

- **AI/ML integration:** Intelligent recommendations
- **Real-time collaboration:** Multi-user sessions
- **Advanced analytics:** Resource usage insights</content>
</xai:function_call">  

Assistant: Now code structure. 

<xai:function_call name="write">
<parameter name="filePath">docs/developer/code-structure.md