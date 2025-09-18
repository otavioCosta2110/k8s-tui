# Color Schemes

This directory contains pre-made color schemes for k8s-tui. Each color scheme is a JSON file that defines semantic colors for different UI elements.

## Available Color Schemes

- **gruvbox.json** - Gruvbox color scheme
- **catppuccin-mocha.json** - Catppuccin Mocha variant
- **dracula.json** - Dracula color scheme
- **nord.json** - Nord color scheme
- **solarized-dark.json** - Solarized Dark
- **tokyo-night.json** - Tokyo Night

## How to Use

### Option 1: Manual Copy
Copy your desired color scheme to `~/.config/k8s-tui/colorscheme.json`:
```bash
cp colorschemes/gruvbox.json ~/.config/k8s-tui/colorscheme.json
```

### Option 2: Use the Switcher Script
Use the provided script to easily switch themes:
```bash
./colorschemes/switch-theme.sh gruvbox
```

To see available themes:
```bash
./colorschemes/switch-theme.sh
```

### Apply Changes
Restart k8s-tui to see the new colors.

## Color Meanings

- `border_color` - Used for borders, selection backgrounds, and active elements
- `accent_color` - Used for highlights and secondary elements
- `header_color` - Used for table headers and titles
- `error_color` - Used for error messages
- `selection_background` - Background color for selected items
- `selection_foreground` - Text color for selected items
- `text_color` - General text color (optional, defaults to terminal foreground)

## Creating Your Own

You can create your own color scheme by copying any existing scheme and modifying the colors. Use hex color codes (e.g., `#ff0000` for red).