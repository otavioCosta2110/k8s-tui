# GEMINI.md

## Project Overview

This project is a terminal user interface (TUI) for interacting with Kubernetes clusters. It is written in Go and uses the `bubbletea` and `bubbles` libraries for the TUI, and the official Kubernetes Go client for interacting with the Kubernetes API.

The application provides a user-friendly way to view and manage various Kubernetes resources, such as Pods, Deployments, Services, and more. It features a tabbed interface, allowing users to switch between different resource views and details.

## Building and Running

The project can be built, tested, and run using standard Go commands.

### Build

To build the application, run the following command:

```sh
go build -v ./...
```

### Test

To run the tests, use the following command:

```sh
go test -v ./...
```

### Run

To run the application, execute the following command:

```sh
go run ./cmd/main.go
```

## Development Conventions

The codebase is organized into several packages, each with a specific responsibility:

-   `cmd`: Contains the main entry point of the application.
-   `internal`: Holds the core application logic, including the UI, Kubernetes client, and configuration.
-   `pkg`: Contains reusable packages that can be shared across different projects.
-   `utils`: Provides utility functions used throughout the application.

The UI is built using the Model-View-Update (MVU) architecture, which is a common pattern in `bubbletea` applications. The `internal/ui/models` package defines the models for the various UI components, such as tabs, resource lists, and headers.

The `internal/k8s` package provides a client for interacting with the Kubernetes API. It includes functions for getting, listing, and describing various Kubernetes resources.

The `internal/ui/models/resource_list.go` file defines a `ResourceFactory` that is responsible for creating different resource models based on the resource type. This design makes it easy to extend the application to support new Kubernetes resources in the future.
