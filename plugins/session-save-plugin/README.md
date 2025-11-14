# Session Management Plugin

A k8s-tui plugin that provides complete session management - saving and loading sessions (opened tabs, namespaces, clusters) to/from JSON files.

## Features

- **Save Sessions**: Save current namespace, all opened tabs with their details (ID, title, resource type, breadcrumb)
- **Load Sessions**: Restore complete multi-cluster sessions with all tabs and navigation state
- **Multi-Cluster Support**: Save and load sessions with multiple clusters and their configurations
- **CLI Integration**: Load sessions via command line argument
- **Interactive Prompts**: User-friendly file selection dialogs

## Installation

1. Place the `session-save-plugin` directory in your k8s-tui plugins directory (usually `~/.local/share/k8s-tui/plugins/` or as configured)
2. Restart k8s-tui

## Usage

### Saving Sessions

1. Open some tabs and navigate in k8s-tui
2. Press Ctrl+S to save the current session
3. Enter a filename for the session (e.g., "my-session" or "my-session.json")
4. Press Enter to save, or Esc to cancel
5. The session will be saved to the specified file in the current directory

### Loading Sessions

#### Interactive Loading
1. Press Ctrl+L (if bound) or use the command palette
2. Execute `session:load` command
3. Enter the session file path (e.g., "my-session.json")
4. Press Enter to load, or Esc to cancel
5. The plugin will recreate all clusters and switch to the active one

#### CLI Loading
```bash
k8s-tui --session my-session.json
```

### Key Bindings

To bind the session commands to keyboard shortcuts, add to your k8s-tui config:

```json
{
  "key_bindings": {
    "ctrl+s": "session:save",
    "ctrl+l": "session:load"
  }
}
```

### Commands

- `session:save` - Save current session to JSON file
- `session:load` - Load session from JSON file
- `--session <filename>` - CLI argument to load session on startup

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

### Session Loading Process

When loading a session, the plugin:

1. **Parses the JSON file** and extracts cluster configurations
2. **Creates cluster tabs** using `k8s_tui.add_cluster_tab()` for each cluster
3. **Restores namespaces** and cluster configurations
4. **Switches to active cluster** using `k8s_tui.switch_to_cluster()`
5. **Updates status** with loading results

### Error Handling

The plugin provides comprehensive error handling:
- File not found or read errors
- Invalid JSON format
- Missing cluster configurations
- API function failures
- Graceful degradation with informative status messages

### Backwards Compatibility

The plugin can load both the old single-cluster format and the new multi-cluster format.

### Session File Format

The plugin uses the same JSON format as before, ensuring compatibility with existing session files. See the format examples below for details.