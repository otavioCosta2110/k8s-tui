# ⚙️ Configuration Guide

k8s-tui is designed to work out-of-the-box with standard Kubernetes configurations, but offers several customization options.

## Kubernetes Configuration

### Kubeconfig Discovery

k8s-tui uses a specific approach for kubeconfig discovery and selection:

#### Configuration Sources

1. **Command Line Arguments**
   - Specify kubeconfig files directly: `k8s-tui --kubeconfig /path/to/config`
   - Multiple kubeconfigs: `k8s-tui --kubeconfig config1 --kubeconfig config2`

2. **Interactive Kubeconfig Selection**
   - If no kubeconfig is specified, k8s-tui scans `~/.kube/` directory
   - Presents a list of available kubeconfig files
   - User selects the desired kubeconfig from the interface

3. **Environment Variable**
   - Sets `KUBECONFIG` environment variable when a kubeconfig is selected
   - Also sets `KUBERNETES_MASTER` for compatibility

#### Configuration Loading Process

```bash
# k8s-tui follows this process:
# 1. Check for --kubeconfig command line arguments
# 2. If none provided, scan ~/.kube/ directory for config files
# 3. Show interactive selection menu
# 4. Set selected config as KUBECONFIG environment variable
# 5. Create Kubernetes client using client-go with the selected path
```

#### Kubeconfig File Discovery

- **Scanned Directory**: `~/.kube/`
- **File Types**: All non-directory files in the kubeconfig directory
- **Selection**: Interactive UI-based selection from available files

#### Client Configuration

- **Namespace**: Can be set via `--namespace` flag or selected interactively

### Kubeconfig Setup

k8s-tui uses the standard Kubernetes configuration files and environment variables:

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

Create custom themes by adding JSON files to `~/.local/share/k8s-tui/themes/`:

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

Plugins are loaded from the directory specified by `plugin_dir` specified at ~/.config/k8s-tui or by the `--plugin-dir` argument (default: `./plugins`).

### Installing Plugins

1. Download or create plugin files
2. Place them in your plugin directory
3. Restart k8s-tui with the correct plugin directory

### Plugin Configuration

Plugins can be configured through:
- **Command line arguments**: `k8s-tui --my-plugin-setting=value`
- **Configuration file**: Plugin-specific settings in `~/.config/k8s-tui/config.json`

### Available Plugins

Example plugins included:
- **example-plugin**: Demonstrates custom resource types
- **pluginmanager-header**: Adds custom header content
- **session-save-plugin**: Saves and restores session state

See [PLUGINS.md](../../PLUGINS.md) for development details and examples.

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

### Command Line Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| `--kubeconfig` | Specify kubeconfig file path | `--kubeconfig ~/.kube/config` |
| `--namespace` | Set default namespace | `--namespace default` |
| `--plugin-dir` | Override plugin directory | `--plugin-dir ./my-plugins` |

### Plugin Arguments

Custom plugin arguments can be passed using `--<plugin-arg>=<value>` format:

```bash
# Example plugin arguments
k8s-tui --my-plugin-setting=value --another-flag=true
```

## Troubleshooting Configuration

### Kubeconfig Selection Issues

```bash
# Check if ~/.kube directory exists
ls -la ~/.kube/

# Verify kubeconfig files are present
ls -la ~/.kube/*

# Test specific kubeconfig file
kubectl --kubeconfig ~/.kube/config cluster-info

# Check current context in specific config
kubectl --kubeconfig ~/.kube/config config current-context

# Test cluster connectivity
kubectl cluster-info

# Verify authentication
kubectl auth can-i list pods
```

### Common Selection Problems

1. **No kubeconfig files found**
   ```bash
   # Create default kubeconfig directory
   mkdir -p ~/.kube
   # Copy your config file to ~/.kube/
   cp /path/to/your/config ~/.kube/my-cluster
   ```

2. **Invalid kubeconfig format**
   ```bash
   # Validate kubeconfig syntax
   kubectl --kubeconfig ~/.kube/config config view --minify
   
   # Check for common issues
   kubectl --kubeconfig ~/.kube/config config get-contexts
   ```

3. **Permission issues**
   ```bash
   # Check file permissions
   ls -la ~/.kube/
   
   # Fix permissions if needed
   chmod 600 ~/.kube/config
   ```

4. **Command line kubeconfig not working**
   ```bash
   # Test with explicit kubeconfig path
   k8s-tui --kubeconfig /absolute/path/to/config
   
   # Verify file exists and is readable
   ls -la /absolute/path/to/config
   ```

## Advanced Configuration

### Custom Resource Definitions

k8s-tui automatically discovers CRDs. For custom views, create plugins.

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
