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

The saved JSON file contains:

```json
{
  "namespace": "default",
  "tabs": [
    {
      "ID": "tab-1",
      "Title": "Pods",
      "ResourceType": "pods",
      "Breadcrumb": ["Resource List", "Pods"]
    }
  ]
}
```