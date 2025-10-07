# Session Save Plugin

A k8s-tui plugin that allows saving the current session (opened tabs) to a JSON file by pressing Ctrl+S.

## Features

- Saves current namespace
- Saves all opened tabs with their details (ID, title, resource type, breadcrumb)
- Saves timestamp
- Outputs to a configurable JSON file

## Installation

1. Place the `session-save-plugin` directory in your k8s-tui plugins directory (usually `~/.local/share/k8s-tui/plugins/` or as configured)
2. Restart k8s-tui

## Configuration

The plugin can be configured in your k8s-tui config file:

```json
{
  "plugins": {
    "session-save-plugin": {
      "enabled": true,
      "session_file": "session.json",
      "save_namespace": true,
      "save_timestamp": true
    }
  }
}
```

## Usage

1. Open some tabs in k8s-tui
2. Press Ctrl+S to save the current session
3. The session will be saved to `session.json` (or configured file)

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

The saved JSON file contains:

```json
{
  "timestamp": "2023-10-06 14:30:00",
  "namespace": "default",
  "tabs": [
    {
      "id": "tab-1",
      "title": "Pods",
      "resourceType": "pods",
      "breadcrumb": ["Resource List", "Pods"]
    }
  ]
}
```