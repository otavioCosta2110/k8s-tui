# Resource Management

This guide covers how to view, create, edit, and delete Kubernetes resources using k8s-tui.

## Viewing Resources

### Resource Lists

Each resource type has a dedicated list view showing key information:

- **Pods**: Name, Namespace, Ready status, Status, Restarts, Age
- **Deployments**: Namespace, Name, Ready, Up-to-date, Available, Age
- **Services**: Namespace, Name, Type, Cluster-IP, External-IP, Ports, Age
- **ConfigMaps**: Namespace, Name, Data keys, Age
- **Secrets**: Namespace, Name, Type, Data keys, Age

### Detailed Views

Press `Enter` on any resource to see detailed information:

- **YAML Configuration**: Full resource definition
- **Status Information**: Current state and conditions
- **Related Resources**: Links to dependent resources
- **Events**: Recent events affecting the resource

### Sorting and Filtering

- Use arrow keys to navigate
- `/` to search by name
- Results update in real-time

## Creating Resources

### Using Creation Forms

1. Navigate to the desired resource type tab
2. Press `n` to open the creation form
3. Fill in the required fields
4. Use `Tab`/`Shift+Tab` to navigate between fields
5. Press `Enter` to submit

### Form Fields by Resource Type

#### Pods
- **Name**: Pod name (required)
- **Image**: Container image (required)
- **Namespace**: Target namespace (defaults to current)

#### Deployments
- **Name**: Deployment name (required)
- **Image**: Container image (required)
- **Replicas**: Number of replicas (default: 1)
- **Namespace**: Target namespace (defaults to current)

#### Services
- **Name**: Service name (required)
- **Type**: Service type - ClusterIP, NodePort, LoadBalancer (default: ClusterIP)
- **Port**: Service port (required)
- **Target Port**: Container port (required)
- **Namespace**: Target namespace (defaults to current)

#### ConfigMaps
- **Name**: ConfigMap name (required)
- **Namespace**: Target namespace (defaults to current)

#### Secrets
- **Name**: Secret name (required)
- **Type**: Secret type - Opaque, TLS, etc. (default: Opaque)
- **Namespace**: Target namespace (defaults to current)

#### Ingresses
- **Name**: Ingress name (required)
- **Host**: Domain name (required)
- **Path**: URL path (default: /)
- **Service Name**: Backend service name (required)
- **Service Port**: Backend service port (required)
- **Namespace**: Target namespace (defaults to current)

#### Jobs
- **Name**: Job name (required)
- **Image**: Container image (required)
- **Backoff Limit**: Retry attempts on failure (default: 6)
- **Namespace**: Target namespace (defaults to current)

#### CronJobs
- **Name**: CronJob name (required)
- **Image**: Container image (required)
- **Schedule**: Cron schedule expression (default: */5 * * * *)
- **Suspend**: Whether to suspend the job (default: false)
- **Namespace**: Target namespace (defaults to current)

## Editing Resources

### YAML Editing

1. Select a resource and press `Enter` to view details
2. Press `e` to open the YAML editor
3. Modify the configuration in your preferred editor
4. Save and exit to apply changes

### Supported Editors

k8s-tui respects the `$EDITOR` environment variable:

```bash
export EDITOR=vim
# or
export EDITOR=nano
# or
export EDITOR=code
```

## Deleting Resources

### Single Resource Deletion

1. Navigate to the resource list
2. Select the resource with arrow keys
3. Press `d` to delete
4. Confirm deletion

### Bulk Operations

Currently, k8s-tui supports single resource operations. For bulk operations, use `kubectl`.

## Resource Relationships

### Viewing Related Resources

From resource details, you can navigate to related resources:

- **Deployment → Pods**: See pods managed by the deployment
- **Service → Endpoints**: View pods backing the service
- **Pod → Logs**: Access container logs
- **ConfigMap/Secret → Consumers**: See which pods use the resource

### Cross-Resource Navigation

Use the breadcrumb navigation to move between related resources efficiently.

## Best Practices

### Resource Naming

- Use descriptive, consistent naming conventions
- Include environment identifiers (dev, staging, prod)
- Follow Kubernetes naming restrictions

### Namespace Organization

- Use namespaces to separate environments
- Apply resource quotas and RBAC policies
- Group related applications together

### Monitoring and Debugging

- Regularly check resource status and events
- Monitor resource usage and performance
- Use logs and events for troubleshooting

## Troubleshooting

### Common Issues

- **Permission Denied**: Check RBAC policies and service account permissions
- **Image Pull Errors**: Verify image registry access and credentials
- **Resource Quotas**: Check namespace resource limits
- **Network Policies**: Ensure proper network connectivity

See the [Troubleshooting](./troubleshooting.md) guide for detailed solutions.</content>
