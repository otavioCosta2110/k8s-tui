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

## Quick Start

1. Ensure you have access to a Kubernetes cluster via `kubectl` or `kubeconfig`
2. Run k8s-tui:

```bash
./k8s-tui
```

3. Use Tab to navigate between resource types
4. Use arrow keys to browse resources
5. Press Enter to view details
6. Press 'n' to create new resources
7. Press 'd' to delete resources

## Key Bindings

### Global
- `Tab` / `Shift+Tab`: Switch between resource tabs
- `Ctrl+C` / `q`: Quit
- `?`: Show help

### Resource Lists
- `↑` / `↓` / `j` / `k`: Navigate resources
- `Enter`: View resource details
- `n`: Create new resource
- `d`: Delete selected resource
- `r`: Refresh
- `/`: Search
- `Esc`: Go back

### Forms
- `Tab` / `Shift+Tab`: Navigate between fields
- `Enter`: Next field / Submit
- `Esc`: Cancel

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

Available themes:
- `catppuccin-mocha`
- `dracula`
- `gruvbox`
- `nord`
- `one-dark`
- `solarized-dark`
- `tokyo-night`
- `transparent`

The configuration file is automatically created on first run. You can also customize individual colors in the `colors` section of the config file.

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
