# Key Bindings Reference

This comprehensive reference covers all keyboard shortcuts available in k8s-tui, organized by screen and context.

## Global Shortcuts

Available from any screen:

| Key Combination | Action | Description |
|-----------------|--------|-------------|
| `Left` | Previous Tab | Switch to previous resource tab |
| `Right` | Next Tab | Switch to next resource tab |
| `Ctrl + C` | Quit | Exit application |
| `q` | Quit | Exit application |
| `?` | Help | Show context-sensitive help |

## Multi-Cluster Navigation

Available when managing multiple clusters:

| Key Combination | Action | Description |
|-----------------|--------|-------------|
| `Ctrl + Left` | Previous Cluster | Switch to previous cluster |
| `Ctrl + Right` | Next Cluster | Switch to next cluster |
| `Ctrl + N` | Add New Cluster | Open cluster addition workflow |

## Resource List Screens

Available when viewing lists of resources (Pods, Deployments, etc.):

### Navigation
| Key | Action | Description |
|-----|--------|-------------|
| `↑` | Up | Move selection up |
| `↓` | Down | Move selection down |
| `j` | Down | Move selection down (vim-style) |
| `k` | Up | Move selection up (vim-style) |
| `Page Up` | Page Up | Move up by page |
| `Page Down` | Page Down | Move down by page |
| `Home` | Top | Jump to first item |
| `End` | Bottom | Jump to last item |

### Actions
| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | View Details | Open detailed view of selected resource |
| `n` | Create New | Open creation form for resource type |
| `d` | Delete | Delete selected resource |
| `e` | Edit | Edit resource YAML (from details) |
| `r` | Refresh | Manual refresh of resource list |
| `/` | Search | Enter search/filter mode |
| `Esc` | Back | Return to previous screen |

### Search Mode
| Key | Action | Description |
|-----|--------|-------------|
| `Any text` | Filter | Type to filter resources |
| `Enter` | Select | Select first matching item |
| `Esc` | Exit | Exit search mode |
| `Backspace` | Delete | Remove last character |

## Resource Details Screens

Available when viewing individual resource details:

| Key | Action | Description |
|-----|--------|-------------|
| `e` | Edit YAML | Open resource in YAML editor |
| `d` | Delete | Delete current resource |
| `r` | Refresh | Refresh resource data |
| `Tab` | Next Section | Move between detail sections |
| `Shift + Tab` | Previous Section | Move between detail sections |
| `Esc` | Back | Return to resource list |

## Creation Forms

Available when creating new resources:

### Navigation
| Key | Action | Description |
|-----|--------|-------------|
| `Tab` | Next Field | Move to next form field |
| `Shift + Tab` | Previous Field | Move to previous form field |
| `↑` | Previous Field | Alternative navigation |
| `↓` | Next Field | Alternative navigation |

### Actions
| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Submit/Next | Submit form or move to next field |
| `Esc` | Cancel | Cancel creation and return |

## YAML Editor

Available when editing resource YAML:

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl + S` | Save | Save changes and apply |
| `Ctrl + C` | Cancel | Discard changes |
| `Esc` | Cancel | Discard changes |

*Note: Editor behavior depends on your configured `$EDITOR`*

## Tab-Specific Shortcuts

### Pods Tab
| Key | Action | Description |
|-----|--------|-------------|
| `l` | View Logs | View logs for selected pod |
| `v` | View Manifest | View pod YAML manifest |

### Deployments Tab
| Key | Action | Description |
|-----|--------|-------------|
| `s` | Scale | Scale deployment replicas |
| `v` | View Pods | View pods managed by deployment |

### Services Tab
| Key | Action | Description |
|-----|--------|-------------|
| `v` | View Endpoints | View service endpoints |

## Context-Sensitive Help

Press `?` from any screen to see available actions for that context.

## Customization

Key bindings are currently not customizable. For custom behavior, consider:

- Using plugins for extended functionality
- Contributing new key bindings to the project
- Using external tools for complex operations

## Vim-Compatible Bindings

k8s-tui includes vim-style navigation:

- `j`/`k` for up/down (instead of arrow keys)
- Relative navigation patterns
- Modal-like interface behavior

## Accessibility

- All functions accessible via keyboard
- High contrast color schemes available
- Screen reader compatible (text-based interface)
- Consistent navigation patterns

## Tips

- Use `Tab` to quickly switch between resource types
- Combine `r` with auto-refresh for real-time monitoring
- Use search (`/`) to quickly find resources in large lists
- Press `?` frequently to discover new features</content>
