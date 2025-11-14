# Cluster Tab Test Plugin

This plugin demonstrates how to add cluster tabs programmatically using the k8s-tui plugin API.

## Features

- **Add Test Cluster**: Adds a new cluster tab using the default kubeconfig
- **List Clusters**: Lists all available clusters with their IDs and namespaces
- **Switch to Cluster**: Switches to a specific cluster by ID

## CLI Arguments

### `--add-test-cluster`
Adds a new cluster tab with default kubeconfig and "default" namespace.

### `--list-clusters`
Lists all available clusters showing:
- Cluster name
- Cluster ID  
- Current namespace

### `--switch-to-cluster <cluster_id>`
Switches to the specified cluster by its ID.

## Usage

1. Load the plugin by placing it in the plugins directory
2. Use CLI arguments when starting k8s-tui
3. The plugin will demonstrate programmatic cluster management

## Example

```bash
# Add a new cluster tab
./k8s-tui --plugin-dir ./plugins --add-test-cluster

# Show all clusters
./k8s-tui --plugin-dir ./plugins --list-clusters

# Switch to cluster with ID "1"
./k8s-tui --plugin-dir ./plugins --switch-to-cluster 1
```
/add_test_cluster        # Add a new cluster tab
/list_clusters          # Show all clusters
/switch_to_cluster 1    # Switch to cluster with ID "1"
```

## Implementation Details

This plugin uses the following k8s-tui API functions:
- `k8s_tui.add_cluster_tab(kubeconfig_path, cluster_name, namespace)`
- `k8s_tui.get_clusters()`
- `k8s_tui.switch_to_cluster(cluster_id)`

These functions allow plugins to dynamically manage cluster tabs in the UI.