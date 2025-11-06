# 🏗️ Architecture

This document describes the high-level architecture of k8s-tui and its key components.

## System Overview

k8s-tui is built using a modular architecture that separates concerns between UI, business logic, and Kubernetes interactions.

```
┌─────────────────────────────────────────────────────────────┐
│                        k8s-tui                              │
├─────────────────────────────────────────────────────────────┤
│  Terminal UI (Bubble Tea)                                   │
│  ├── Components (Tables, Forms, Lists)                      │
│  ├── Styles & Themes                                        │
│  └── Event Handling                                         │
├─────────────────────────────────────────────────────────────┤
│  Application Layer                                           │
│  ├── Models & State Management                              │
│  ├── Resource Handlers                                      │
│  └── Plugin Manager                                         │
├─────────────────────────────────────────────────────────────┤
│  Kubernetes Layer                                           │
│  ├── Client Wrapper                                         │
│  ├── Resource Abstractions                                 │
│  └── Multi-Cluster Support                                 │
├─────────────────────────────────────────────────────────────┤
│  Plugin System                                              │
│  ├── Lua Runtime                                           │
│  ├── API Hooks                                             │
│  └── Extension Points                                       │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### Terminal UI Layer
- **Framework**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Components**: Reusable UI components for tables, forms, and navigation
- **Styling**: Theme system with multiple color schemes
- **Events**: Message-based event handling

### Application Layer
- **Models**: State management for different views and resources
- **Handlers**: Business logic for resource operations
- **Navigation**: Tab-based navigation system
- **Auto-refresh**: Background updates for resource lists

### Kubernetes Layer
- **Client**: Wrapper around client-go for Kubernetes API access
- **Resources**: Abstractions for different Kubernetes resource types
- **Multi-cluster**: Support for managing multiple clusters simultaneously

### Plugin System
- **Runtime**: Embedded Lua interpreter for plugin execution
- **API**: Hooks for extending functionality
- **Isolation**: Sandboxed plugin execution

## Data Flow

```
User Input → UI Events → Application Logic → Kubernetes API → Response → UI Update
```

1. User interacts with terminal interface
2. UI generates events (keyboard, mouse)
3. Application processes events and updates state
4. Kubernetes client makes API calls
5. Response updates application state
6. UI re-renders with new data

## Key Design Principles

### Separation of Concerns
- UI components are decoupled from business logic
- Kubernetes interactions are abstracted
- Plugin system is isolated from core functionality

### Extensibility
- Plugin system allows custom functionality
- Theme system supports custom styling
- Resource handlers can be extended

### Performance
- Efficient terminal rendering
- Minimal API calls through caching
- Background updates without blocking UI

### Multi-Cluster Support
- Isolated client connections
- Independent state per cluster
- Unified interface for cluster management

## Module Structure

```
internal/
├── app/                    # Application layer
│   ├── cli/               # Command-line interface
│   ├── config/            # Configuration management
│   └── ui/                # UI components and logic
│       ├── components/     # Reusable UI components
│       ├── models/         # View models and state
│       └── styles/         # Styling and themes
├── k8s/                    # Kubernetes layer
│   ├── client/            # Kubernetes client wrapper
│   ├── resources/         # Resource-specific logic
│   └── types/             # Type definitions
└── plugins/                # Plugin system
```

## Concurrency Model

- **Main Thread**: UI rendering and event handling
- **Background Goroutines**: Auto-refresh, plugin execution
- **Channels**: Communication between components
- **Mutexes**: Thread-safe state access

## State Management

- **Centralized State**: Main model holds application state
- **Immutable Updates**: State changes create new state objects
- **Event-Driven**: State updates triggered by events
- **Persistence**: Plugin state can be persisted

## Error Handling

- **Graceful Degradation**: Errors don't crash the application
- **User Feedback**: Clear error messages in the UI
- **Recovery**: Automatic recovery from transient failures
- **Logging**: Comprehensive error logging

## Testing Strategy

- **Unit Tests**: Individual component testing
- **Integration Tests**: Component interaction testing
- **End-to-End Tests**: Full workflow testing
- **Mock Kubernetes**: Isolated testing without real clusters