# k8s-tui

A powerful terminal-based user interface for managing Kubernetes resources. Browse, create, edit, and delete Kubernetes resources with an intuitive TUI built with Bubble Tea.

![Build Status](https://github.com/otavioCosta2110/k8s-tui/workflows/Go/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## Features

- **Multi-Cluster Support**: Manage multiple Kubernetes clusters from a single interface
- **Resource Management**: Full CRUD operations for all major Kubernetes resources
- **Interactive Forms**: Create new resources with guided forms
- **Real-time Updates**: Auto-refreshing views with configurable intervals
- **Plugin System**: Extend functionality with Lua plugins
- **Theme Support**: Multiple color schemes including Catppuccin, Dracula, Nord, and more
- **YAML Editing**: Edit resources directly in your preferred editor
- **Search & Filter**: Quickly find resources across namespaces

![main dashboard](./assets/screenshots/main_dashboard.png)

## Supported Resources

- Pods
- Deployments
- Services
- ConfigMaps
- Secrets
- Ingresses
- Jobs
- CronJobs
- DaemonSets
- StatefulSets
- Namespaces
- Nodes

## Installation

### From Source

```bash
git clone https://github.com/otavioCosta2110/k8s-tui.git
cd k8s-tui
go build -o k8s-tui ./cmd
```

### Binary Releases

Download pre-built binaries from the [releases page](https://github.com/otavioCosta2110/k8s-tui/releases).

## Usage

### Command Line Arguments

```bash
# Basic usage
./k8s-tui

# Specify kubeconfig file
./k8s-tui --kubeconfig ~/.kube/cluster-config

# Use multiple kubeconfig files
./k8s-tui --kubeconfig ~/.kube/cluster1 --kubeconfig ~/.kube/cluster2

# Set default namespace
./k8s-tui --namespace production

# Use custom plugin directory
./k8s-tui --plugin-dir ./my-plugins

# Pass plugin-specific arguments
./k8s-tui --my-plugin-setting=value --another-flag=true

# Help
./k8s-tui --help
```

### Available Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| `--kubeconfig` | Path to kubeconfig file(s) | `--kubeconfig ~/.kube/config` |
| `--namespace` | Default namespace to use | `--namespace default` |
| `--plugin-dir` | Plugin directory path | `--plugin-dir ./plugins` |
| `--help`, `-h` | Show help message | `--help` |
| `--<plugin-arg>` | Custom plugin arguments | `--my-setting=value` |

## Quick Start

1. Ensure you have access to a Kubernetes cluster via `kubectl` or `kubeconfig`
2. Run k8s-tui:

```bash
./k8s-tui
```
3. Use up and down arrows to navigate between resource types
4. Use arrow keys to browse resources
![main dashboard](./assets/screenshots/pod_list.png)
5. Press Enter to view details
![main dashboard](./assets/screenshots/pod_detail.png)

6. Press 'n' to create new resources
![main dashboard](./assets/screenshots/pod_new.png)

7. Press 'd' to delete resources

## Key Bindings

### Global Shortcuts
- `left/right`: Switch between resource tabs
- `Ctrl+C` / `q`: Quit application
- `?`: Show context-sensitive help

### Multi-Cluster Navigation
- `Ctrl+Left`: Previous cluster
- `Ctrl+Right`: Next cluster  
- `Ctrl+N`: Add new cluster

### Resource Lists
- `↑` / `↓` / `j` / `k`: Navigate resources
- `Page Up` / `Page Down`: Navigate by page
- `Home` / `End`: Jump to top/bottom
- `Enter`: View resource details
- `n`: Create new resource
- `d`: Delete selected resource
- `e`: Edit resource (from details view)
- `r`: Refresh resource list
- `Esc`: Go back

### Resource Details
- `e`: Edit YAML in external editor
- `r`: Refresh resource data
- `Esc/q`: Return to resource list

### Creation Forms
- `Tab` / `Shift+Tab`: Navigate between fields
- `↑` / `↓`: Alternative field navigation
- `Enter`: Submit form or move to next field
- `Esc`: Cancel creation

### Tab-Specific Shortcuts
- **Pods**: `l` (view logs), `v` (view manifest)
- **Deployments**: `s` (scale), `v` (view pods)
- **Services**: `v` (view endpoints)

For a complete reference, see [Key Bindings Guide](docs/User-Guide/Key-Bindings.md).

## Configuration

### Kubeconfig

k8s-tui uses the standard Kubernetes configuration:

- `KUBECONFIG` environment variable
- `~/.kube/config` file
- In-cluster configuration (when running in a pod)

### Themes

Configure themes by editing the configuration file at `~/.config/k8s-tui/config.json`:

```json
{
  "theme": "catppuccin-mocha"
}
```

## Theme Showcase

<details>
<summary><strong>catppuccin-mocha</strong></summary>

![Catppuccin Mocha Theme Preview](./assets/screenshots/catppuccin-mocha.png)
</details>

<details>
<summary><strong>dracula</strong></summary>

![Dracula Theme Preview](./assets/screenshots/dracula.png)
</details>

<details>
<summary><strong>gruvbox</strong></summary>

![Gruvbox Theme Preview](./assets/screenshots/gruvbox.png)
</details>

<details>
<summary><strong>nord</strong></summary>

![Nord Theme Preview](./assets/screenshots/nord.png)
</details>

<details>
<summary><strong>one-dark</strong></summary>

![One Dark Theme Preview](./assets/screenshots/one-dark.png)
</details>

<details>
<summary><strong>solarized-dark</strong></summary>

![Solarized Dark Theme Preview](./assets/screenshots/solarized-dark.png)
</details>

<details>
<summary><strong>tokyo-night</strong></summary>

![Tokyo Night Theme Preview](./assets/screenshots/tokyo-night.png)
</details>

<details>
<summary><strong>transparent</strong></summary>

![Transparent Theme Preview](./assets/screenshots/transparent.png)
</details>

## Plugins

Extend k8s-tui functionality with Lua plugins. See [PLUGINS.md](PLUGINS.md) for details.

## Development

### Prerequisites

- Go 1.21+
- Access to a Kubernetes cluster

### Building

```bash
go build -v ./...
```

### Testing

```bash
go test -v ./...
```

### Code Structure

```
internal/
├── app/           # UI and application logic
│   ├── cli/       # Command line interface
│   ├── config/    # Configuration handling
│   └── ui/        # User interface components
│       ├── components/  # Reusable UI components
│       └── models/      # Resource-specific models
└── k8s/           # Kubernetes API interactions
    ├── client/    # Kubernetes client setup
    └── resources/ # Resource-specific operations
pkg/               # Shared packages
├── format/        # Data formatting utilities
├── logger/        # Logging utilities
└── plugins/       # Plugin system
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal app framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Kubernetes Go Client](https://github.com/kubernetes/client-go) - Kubernetes API client

## Support

- [Issues](https://github.com/otavioCosta2110/k8s-tui/issues)
- [Discussions](https://github.com/otavioCosta2110/k8s-tui/discussions)

---

**Note**: This project is not affiliated with the Kubernetes project or the Cloud Native Computing Foundation.</content>
