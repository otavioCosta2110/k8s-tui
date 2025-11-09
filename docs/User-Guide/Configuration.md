# ⚙️ Configuration Guide

k8s-tui is designed to work out-of-the-box with standard Kubernetes configurations, but offers several customization options.

## Kubernetes Configuration

### Kubeconfig Setup

k8s-tui uses the standard Kubernetes configuration files and environment variables:

#### Configuration Sources (in order of precedence)

1. **KUBECONFIG environment variable**
   ```bash
   export KUBECONFIG=/path/to/config:/another/config
   ```

2. **Default kubeconfig location**
   ```bash
   # User-specific config
   ~/.kube/config

   # System-wide configs
   /etc/kubernetes/admin.conf
   ```

3. **In-cluster configuration** (when running inside a Kubernetes pod)

#### Multiple Clusters

k8s-tui supports managing multiple Kubernetes clusters simultaneously with isolated views and plugin managers.

##### Command Line Configuration

Specify multiple kubeconfig files:

```bash
# Multiple kubeconfig files
k8s-tui --kubeconfig ~/.kube/cluster1 --kubeconfig ~/.kube/cluster2

# Or using environment variable
export KUBECONFIG=~/.kube/cluster1:~/.kube/cluster2
k8s-tui
```

##### Dynamic Cluster Addition

While running k8s-tui:

1. Press `Ctrl + N` to add a new cluster
2. Select a kubeconfig file from `~/.kube/`
3. Choose a namespace for the cluster
4. The cluster will be added with a new tab

##### Cluster Isolation

Each cluster maintains:
- Separate Kubernetes client connection
- Independent plugin manager instance
- Isolated resource views
- Separate namespace context

##### Plugin Directory

All clusters share the same plugin directory:

```bash
# Set plugin directory (shared across clusters)
k8s-tui --plugin-dir ./my-plugins --kubeconfig cluster1 --kubeconfig cluster2
```

#### Context Switching

Switch between contexts using kubectl:

```bash
# List available contexts
kubectl config get-contexts

# Switch context
kubectl config use-context my-cluster

# Verify current context
kubectl config current-context
```

### Authentication

k8s-tui supports all standard Kubernetes authentication methods:

- **X.509 certificates**
- **Bearer tokens**
- **OIDC authentication**
- **AWS IAM**
- **Azure AD**
- **GCP service accounts**

Ensure your kubeconfig contains the appropriate authentication credentials.

## Theme Configuration

k8s-tui supports multiple color schemes for different environments and preferences.

### Available Themes

- **Catppuccin Mocha**: Warm, pastel colors
- **Dracula**: Dark theme with vibrant accents
- **Gruvbox**: Retro, earthy colors
- **Nord**: Arctic-inspired color palette
- **One Dark**: VS Code inspired theme
- **Solarized Dark**: Low-contrast, eye-friendly
- **Tokyo Night**: Modern dark theme
- **Transparent**: Minimal styling

### Switching Themes

Use the provided theme switcher script:

```bash
cd assets/colorschemes
./switch-theme.sh
```

Follow the interactive prompts to select your preferred theme.

### Custom Themes

Create custom themes by adding JSON files to `assets/colorschemes/`:

```json
{
  "name": "My Custom Theme",
  "background": "#000000",
  "foreground": "#ffffff",
  "accent": "#ff0000",
  "border": "#333333",
  "text": "#ffffff",
  "helpText": "#888888",
  "error": "#ff0000",
  "success": "#00ff00",
  "warning": "#ffff00"
}
```

## Plugin Configuration

### Plugin Directory

Plugins are loaded from the `plugins/` directory in the project root.

### Installing Plugins

1. Download or create plugin files (`.lua`)
2. Place them in the `plugins/` directory
3. Restart k8s-tui

### Plugin Manager

Use the built-in plugin manager:

1. Navigate to the Plugin Manager tab
2. Browse available plugins
3. Install/uninstall as needed

See [Plugins](../api/plugins.md) for development details.

## Application Settings

### Auto-Refresh Interval

Configure how often resource lists refresh:

```bash
# Currently not configurable via config file
# Default: 5 seconds
```

### Editor Configuration

Set your preferred YAML editor:

```bash
export EDITOR=vim
# or
export EDITOR=nano
# or
export EDITOR=code
```

### Terminal Settings

For optimal experience:

- **Terminal size**: Minimum 120x30 characters
- **Color support**: 256-color or truecolor terminal
- **Unicode support**: For proper icon display

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file(s) | `~/.kube/config` |
| `EDITOR` | YAML editor command | System default |
| `KUBERNETES_SERVICE_HOST` | API server host (in-cluster) | - |
| `KUBERNETES_SERVICE_PORT` | API server port (in-cluster) | - |

## Troubleshooting Configuration

### Connection Issues

```bash
# Test cluster connectivity
kubectl cluster-info

# Check current context
kubectl config current-context

# Verify authentication
kubectl auth can-i list pods
```

### Theme Issues

```bash
# Reset to default theme
cd assets/colorschemes
cp one-dark.json current-theme.json
```

### Plugin Problems

```bash
# Check plugin syntax
lua -l plugin.lua

# View plugin logs
tail -f ~/.k8s-tui/logs/plugin.log
```

## Advanced Configuration

### Custom Resource Definitions

k8s-tui automatically discovers CRDs. For custom views, create plugins.

### Network Policies

Ensure your cluster allows connections from k8s-tui to the API server.

### RBAC Permissions

Required permissions for full functionality:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-tui-user
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets", "namespaces", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "statefulsets", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup kubeconfig
cp ~/.kube/config ~/.kube/config.backup

# Backup themes
cp -r assets/colorschemes ~/k8s-tui-themes-backup
```

### Reset Configuration

```bash
# Reset themes
cd assets/colorschemes
git checkout .

# Clear plugin cache
rm -rf ~/.k8s-tui/cache
```</content>
