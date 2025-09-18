#!/bin/bash

# k8s-tui Color Scheme Switcher
# Usage: ./switch-theme.sh <theme-name>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$HOME/.config/k8s-tui"
THEME="$1"

if [ -z "$THEME" ]; then
    echo "Usage: $0 <theme-name>"
    echo ""
    echo "Available themes:"
    ls -1 "$SCRIPT_DIR"/*.json | grep -v README | sed 's/.*\///' | sed 's/\.json$//'
    exit 1
fi

THEME_FILE="$SCRIPT_DIR/$THEME.json"

if [ ! -f "$THEME_FILE" ]; then
    echo "Error: Theme '$THEME' not found."
    echo ""
    echo "Available themes:"
    ls -1 "$SCRIPT_DIR"/*.json | grep -v README | sed 's/.*\///' | sed 's/\.json$//'
    exit 1
fi

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Copy the theme
cp "$THEME_FILE" "$CONFIG_DIR/colorscheme.json"

echo "Switched to $THEME theme!"
echo "Restart k8s-tui to see the changes."