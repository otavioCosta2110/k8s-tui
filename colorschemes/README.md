# K8s TUI Configuration

K8s TUI supports comprehensive configuration through JSON files located in your user configuration directory.

## Directory Structure

```
~/.config/k8s-tui/
├── config.json          # Main application configuration
└── colorscheme.json     # Legacy colorscheme (auto-generated)

~/.local/share/k8s-tui/
└── themes/              # Theme files (auto-copied from installation)
    ├── catppuccin-mocha.json
    ├── dracula.json
    ├── gruvbox.json
    ├── nord.json
    ├── one-dark.json
    ├── solarized-dark.json
    ├── tokyo-night.json
    └── transparent.json
```

## Main Configuration (config.json)

The main configuration file allows you to customize various aspects of K8s TUI:

```json
{
  "theme": "catppuccin-mocha",
  "refresh_interval_seconds": 10,
  "auto_refresh": true,
  "default_namespace": "default",
  "key_bindings": {
    "quit": "q",
    "help": "?",
    "refresh": "r",
    "back": "[",
    "forward": "]",
    "new_tab": "ctrl+t",
    "close_tab": "ctrl+w",
    "quick_nav": "g"
  },
  "colors": {
    "border_color": "#89b4fa",
    "accent_color": "#f38ba8",
    "header_color": "#f9e2af",
    "error_color": "#f38ba8",
    "selection_background": "#89b4fa",
    "selection_foreground": "#1e1e2e",
    "text_color": "#cdd6f4",
    "background_color": "#1e1e2e",
    "yaml_key_color": "#89b4fa",
    "yaml_value_color": "#cdd6f4",
    "yaml_title_color": "#f9e2af",
    "help_text_color": "#a6adc8",
    "header_value_color": "#a6e3a1",
    "header_loading_color": "#fab387"
  }
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `theme` | Name of the theme to use (without .json extension) | `"default"` |
| `refresh_interval_seconds` | How often to refresh data in seconds | `10` |
| `auto_refresh` | Whether to automatically refresh data | `true` |
| `default_namespace` | Default namespace to use when connecting | `"default"` |
| `key_bindings` | Custom key bindings for various actions | See example above |
| `colors` | Color scheme when not using a theme | Default color scheme |

### Key Bindings

You can customize the following key bindings:

- `quit`: Exit the application
- `help`: Show help screen
- `refresh`: Manually refresh data
- `back`: Navigate back in breadcrumb history
- `forward`: Navigate forward in breadcrumb history
- `new_tab`: Create a new tab
- `close_tab`: Close current tab
- `quick_nav`: Open quick navigation

## Color Scheme

The color scheme defines the appearance of various UI elements:

| Color | Description |
|-------|-------------|
| `border_color` | Color of borders and dividers |
| `accent_color` | Color for accents and highlights |
| `header_color` | Color for header text and titles |
| `error_color` | Color for error messages |
| `selection_background` | Background color for selected items |
| `selection_foreground` | Text color for selected items |
| `text_color` | Default text color |
| `background_color` | Default background color |
| `yaml_key_color` | Color for YAML keys |
| `yaml_value_color` | Color for YAML values |
| `yaml_title_color` | Color for YAML section titles |
| `help_text_color` | Color for help text |
| `header_value_color` | Color for values in header (namespace, counts) |
| `header_loading_color` | Color for loading indicators in header |

## Using Themes

K8s TUI comes with several pre-configured themes. To use a theme, set the `theme` field in your `config.json`:

```json
{
  "theme": "dracula"
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

## Custom Themes

You can create custom themes by adding JSON files to `~/.local/share/k8s-tui/themes/`. Use the same structure as the built-in themes.

## Automatic Setup

When you first run K8s TUI, it will automatically:

1. Create the necessary directories (`~/.config/k8s-tui/` and `~/.local/share/k8s-tui/themes/`)
2. Copy all built-in themes to the themes directory
3. Create a default `config.json` file
4. Create a legacy `colorscheme.json` file for backward compatibility

## Migration from Legacy Configuration

If you have an existing `colorscheme.json` file, K8s TUI will automatically migrate it to the new configuration format while preserving your custom colors.

## Legacy Theme Switching

For backward compatibility, you can still use the theme switching script:

```bash
./colorschemes/switch-theme.sh dracula
```

This will update your `colorscheme.json` file with the selected theme.