# Namespace Selector Plugin

A simple plugin for k8s-tui that allows switching namespaces by typing the namespace name.

## Features

- `namespace:switch` command - Prompts for a namespace name and switches to it
- Input validation to prevent invalid namespace names
- Shows current namespace as default in the prompt

## Usage

Run the command `namespace:switch` from within k8s-tui to switch namespaces interactively.

## Example

```
:namespace:switch
```

This will show a text input dialog where you can type the namespace name.