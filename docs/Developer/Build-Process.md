# 🔨 Build Process

This document covers the build process, packaging, and distribution of k8s-tui.

## 🏗️ Build Overview

k8s-tui is a Go application that can be built for multiple platforms and architectures. The build process includes compilation, testing, packaging, and optional release creation.

## 📋 Prerequisites

### Development Requirements

- **Go 1.21+** - Primary build requirement
- **Git** - Version control and build metadata
- **Make** - Build automation (optional but recommended)

### Build Tools

- **gofmt** - Code formatting (included with Go)
- **golangci-lint** - Linting (optional)
- **upx** - Binary compression (optional)

## 🚀 Basic Build

### Simple Build

```bash
# Build for current platform
go build -o k8s-tui ./cmd

# Build with verbose output
go build -v -o k8s-tui ./cmd

# Build with build info
go build -ldflags "-X main.version=1.0.0" -o k8s-tui ./cmd
```

### Development Build

```bash
# Build with development flags
go build \
  -ldflags "-X main.version=dev -X main.commit=$(git rev-parse HEAD)" \
  -o k8s-tui ./cmd
```

## 🎯 Cross-Platform Builds

### Manual Cross-Compilation

```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o k8s-tui-linux-amd64 ./cmd
GOOS=darwin GOARCH=amd64 go build -o k8s-tui-darwin-amd64 ./cmd
GOOS=windows GOARCH=amd64 go build -o k8s-tui-windows-amd64.exe ./cmd

# Build for ARM architectures
GOOS=linux GOARCH=arm64 go build -o k8s-tui-linux-arm64 ./cmd
GOOS=darwin GOARCH=arm64 go build -o k8s-tui-darwin-arm64 ./cmd
```

### Using Makefile

```makefile
# Makefile
.PHONY: build clean test lint

# Variables
BINARY_NAME=k8s-tui
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Build targets
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd

clean:
	rm -f $(BINARY_NAME)*

test:
	go test -v ./...

lint:
	golangci-lint run

# Development targets
dev: build
	./$(BINARY_NAME)

install: build
	cp $(BINARY_NAME) /usr/local/bin/
```

## 📦 Build Configuration

### Build Variables

The application supports build-time variables:

```go
// cmd/main.go
package main

import (
    "fmt"
    "runtime"
)

var (
    version   = "dev"
    commit    = "unknown"
    buildTime = "unknown"
)

func main() {
    if version == "dev" {
        fmt.Printf("k8s-tui development build\n")
        fmt.Printf("Go version: %s\n", runtime.Version())
        fmt.Printf("Commit: %s\n", commit)
        fmt.Printf("Build time: %s\n", buildTime)
    }
}
```

### Build Tags

Use build tags for conditional compilation:

```go
// +build !release

// Development-only code
func enableDebugFeatures() {
    // Debug functionality
}
```

```go
// +build release

// Release-only code
func enableProductionFeatures() {
    // Production optimizations
}
```

## 🧪 Testing During Build

### Pre-build Tests

```bash
# Run tests before building
go test -v ./... && go build -o k8s-tui ./cmd

# Run tests with race detection
go test -race -v ./... && go build -o k8s-tui ./cmd

# Run tests with coverage
go test -cover -v ./... && go build -o k8s-tui ./cmd
```

### Integration in Makefile

```makefile
build: test
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd

test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-cover:
	go test -cover -v ./...
```

## 🔧 Build Optimization

### Binary Size Optimization

```bash
# Strip debug information
go build -ldflags "-s -w" -o k8s-tui ./cmd

# Optimize for size
go build -ldflags "-s -w" -gcflags="-l=4" -o k8s-tui ./cmd

# Compress with UPX (optional)
upx --best k8s-tui
```

### Performance Optimization

```bash
# Optimize for speed
go build -ldflags "-s -w" -gcflags="-l=2" -o k8s-tui ./cmd

# Disable inlining for debugging
go build -gcflags="-l=4" -o k8s-tui ./cmd
```

## 🐳 Container Builds

### Dockerfile

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-tui ./cmd

# Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/k8s-tui .
COPY --from=builder /app/assets ./assets

CMD ["./k8s-tui"]
```

### Build and Push

```bash
# Build Docker image
docker build -t k8s-tui:latest .

# Tag and push
docker tag k8s-tui:latest your-registry/k8s-tui:latest
docker push your-registry/k8s-tui:latest
```

## 🚀 Release Process

### Automated Releases with GitHub Actions

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-linux-amd64 ./cmd
          GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-darwin-amd64 ./cmd
          GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-windows-amd64.exe ./cmd
      
      - name: Create Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false
      
      - name: Upload Assets
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ steps.create_release.outputs.upload_url }}
          asset_path: ./k8s-tui-linux-amd64
          asset_name: k8s-tui-linux-amd64
          asset_content_type: application/octet-stream
```

### Manual Release Process

```bash
# 1. Update version
git tag -a v1.0.0 -m "Release version 1.0.0"

# 2. Build release binaries
VERSION=v1.0.0
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-linux-amd64 ./cmd
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-darwin-amd64 ./cmd
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o k8s-tui-windows-amd64.exe ./cmd

# 3. Create checksums
sha256sum k8s-tui-* > checksums.txt

# 4. Push tag
git push origin v1.0.0

# 5. Create GitHub release and upload binaries
```

## 📋 Build Verification

### Binary Verification

```bash
# Check binary information
file k8s-tui
ldd k8s-tui  # Linux dependencies
otool -L k8s-tui  # macOS dependencies

# Check version info
./k8s-tui --version
./k8s-tui --help
```

### Smoke Testing

```bash
# Basic functionality test
./k8s-tui --help
./k8s-tui --version

# Test with mock Kubernetes (if available)
./k8s-tui --dry-run
```

## 🔍 Debugging Builds

### Build Debugging

```bash
# Verbose build output
go build -x -o k8s-tui ./cmd

# Show build commands
go build -x -work ./cmd

# Keep temporary files
go build -work -x ./cmd
```

### Dependency Issues

```bash
# Check module dependencies
go mod why -m github.com/some/package
go mod graph

# Clean module cache
go clean -modcache
go mod download
```

## 📊 Build Performance

### Build Caching

```bash
# Enable build cache
export GOCACHE=/tmp/go-build

# Clear build cache
go clean -cache

# Use remote build cache (if configured)
go build -cache=remote://my-cache.example.com ./cmd
```

### Parallel Builds

```bash
# Build with parallelism
go build -p 4 ./cmd

# Test with parallelism
go test -p 4 ./...
```

## 🛠️ Development Workflow

### Local Development

```bash
# Install dependencies
go mod download

# Run tests
go test -v ./...

# Build for development
go build -o k8s-tui ./cmd

# Run application
./k8s-tui
```

### Pre-commit Checks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Format code
gofmt -s -w .

# Run linter
golangci-lint run

# Run tests
go test -v ./...

# Build to ensure it compiles
go build -o /tmp/k8s-tui ./cmd
```

## 📝 Build Configuration Files

### .golangci.yml

```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/otavioCosta2110/k8s-tui
```

### .gitignore

```gitignore
# Binaries
k8s-tui
k8s-tui-*

# Test cache
*.test

# Build artifacts
*.exe
*.dll
*.so
*.dylib

# Go build cache
.cache/

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db
```

This comprehensive build process ensures k8s-tui can be reliably built, tested, and distributed across multiple platforms while maintaining quality and performance standards.