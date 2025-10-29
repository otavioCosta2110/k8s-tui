# Navigation Guide

k8s-tui uses a tab-based interface for navigating between different Kubernetes resource types and managing your clusters.

## Multi-Cluster Navigation

When managing multiple clusters, a cluster tab bar appears at the top:

- **Cluster 1, Cluster 2, etc.**: Switch between configured clusters
- `F1-F3`: Quick cluster switching (up to 3 clusters)
- `Ctrl + N`: Add a new cluster dynamically
- `Tab`/`Shift + Tab`: Navigate cluster tabs

Each cluster maintains its own:
- Namespace context
- Resource tabs
- Plugin state

## Resource Tab Navigation

Within each cluster, navigate resource types using tabs:

- **Pods**: Manage pod lifecycle
- **Deployments**: Handle deployment scaling and updates
- **Services**: Configure service networking
- **ConfigMaps**: Manage configuration data
- **Secrets**: Handle sensitive configuration
- **Ingresses**: Control external access
- **Jobs**: Run one-time tasks
- **CronJobs**: Schedule recurring tasks
- **DaemonSets**: Manage node-level agents
- **StatefulSets**: Handle stateful applications
- **Namespaces**: Organize resources
- **Nodes**: Monitor cluster nodes

### Switching Resource Tabs

- `Tab`: Move to next resource tab
- `Shift + Tab`: Move to previous resource tab
- `1-9`: Jump directly to tab number (if available)

## Screen Transitions

### Resource List → Details

From any resource list:
- `Enter`: Open detailed view of selected resource
- Shows YAML configuration, status, and related resources

### Details → Edit Mode

From resource details:
- `e`: Open YAML editor for modifications
- Uses your configured `$EDITOR`

### Create New Resources

From any resource list:
- `n`: Open creation form for that resource type
- Guided form with validation

### Back Navigation

- `Esc`: Return to previous screen
- `q` or `Ctrl+C`: Quit application

## Keyboard Shortcuts

### Global Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `Ctrl+C` | Quit |
| `q` | Quit |
| `?` | Show help |

### Resource Lists

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate items |
| `j/k` | Navigate items (vim-style) |
| `Enter` | View details |
| `n` | Create new |
| `d` | Delete selected |
| `r` | Refresh |
| `/` | Search |
| `Esc` | Back |

### Forms

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Enter` | Submit/Next |
| `Esc` | Cancel |

## Search and Filtering

- `/`: Enter search mode
- Type to filter resources by name
- `Enter`: Select first match
- `Esc`: Exit search

## Auto-Refresh

Most screens auto-refresh every 5 seconds by default. Use `r` to manual refresh.

## Context Awareness

The interface remembers your current namespace and cluster context across screens.</content>
</xai:function_call">  

Assistant: Now resource management. 

<xai:function_call name="write">
<parameter name="filePath">docs/user-guide/resource-management.md