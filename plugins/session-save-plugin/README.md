# Session Save Plugin

A k8s-tui plugin that allows saving the current session (opened tabs) to a JSON file by pressing Ctrl+S.

## Features

- Saves current namespace
- Saves all opened tabs with their details (ID, title, resource type, breadcrumb)
- Prompts user for custom filename
- Outputs to a JSON file in the current directory

## Installation

1. Place the `session-save-plugin` directory in your k8s-tui plugins directory (usually `~/.local/share/k8s-tui/plugins/` or as configured)
2. Restart k8s-tui

## Usage

1. Open some tabs in k8s-tui
2. Press Ctrl+S to save the current session
3. Enter a filename for the session (e.g., "my-session" or "my-session.json")
4. Press Enter to save, or Esc to cancel
5. The session will be saved to the specified file in the current directory

## Key Binding

To bind the save command to Ctrl+S, add to your k8s-tui config:

```json
{
  "key_bindings": {
    "ctrl+s": "session:save"
  }
}
```

## Output Format

The saved JSON file contains multi-cluster information:

```json
{
  "clusters": [
    {
      "Index": 0,
      "ID": "0",
      "Name": "cluster-name",
      "Namespace": "default",
      "Kubeconfig": "https://api.cluster.example.com",
      "IsActive": false
    },
    {
      "Index": 1,
      "ID": "1",
      "Name": "another-cluster",
      "Namespace": "production",
      "Kubeconfig": "https://api.another.example.com",
      "IsActive": true
    }
  ],
  "current_cluster": {
    "ID": "1",
    "Name": "another-cluster",
    "Namespace": "production",
    "Kubeconfig": "https://api.another.example.com",
    "IsActive": true
  },
  "session": {
    "namespace": "production",
    "breadcrumb": ["Resource List", "Deployments"],
    "tabs": [
      {
        "ID": "tab-1",
        "Title": "Deployments",
        "ResourceType": "Deployments",
        "CurrentIndex": 0,
        "Breadcrumb": ["Resource List", "Deployments"],
        "Metadata": {}
      }
    ]
  }
}
```

### Format Details

- **clusters**: Array of all connected clusters with their configuration
  - `Index`: Zero-based index of the cluster
  - `ID`: Unique identifier for the cluster
  - `Name`: Display name of the cluster
  - `Namespace`: Current namespace for this cluster
  - `Kubeconfig`: Kubeconfig path or cluster API endpoint
  - `IsActive`: Whether this is the currently active cluster

- **current_cluster**: The currently active cluster (same structure as cluster array items)

- **session**: Current session state for the active cluster
  - `namespace`: Active namespace
  - `breadcrumb`: Navigation breadcrumb trail
  - `tabs`: Array of open resource tabs with their state

### Backwards Compatibility

The plugin can load both the old single-cluster format and the new multi-cluster format.