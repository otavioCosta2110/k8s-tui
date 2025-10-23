# Build Process

This guide covers how to build, test, and release k8s-tui across different platforms and environments.

## Prerequisites

### System Requirements

- **Go 1.21+**: [Installation guide](https://golang.org/doc/install)
- **Git**: Version control system
- **Make** (optional): Build automation

### Development Setup

```bash
# Clone repository
git clone https://github.com/otavioCosta2110/k8s-tui.git
cd k8s-tui

# Initialize Go modules
go mod download

# Verify setup
go version
go env GOPATH
```

## Local Development Build

### Basic Build

```bash
# Build for current platform
go build -o k8s-tui ./cmd

# Build with verbose output
go build -v -o k8s-tui ./cmd

# Build with race detection (for debugging)
go build -race -o k8s-tui ./cmd
```

### Development Flags

```bash
# Include debug information
go build -ldflags "-X main.version=dev" -o k8s-tui ./cmd

# Strip debug information for smaller binary
go build -ldflags "-s -w" -o k8s-tui ./cmd

# Include build time and version
go build -ldflags "-X main.version=$(git describe --tags) -X main.buildTime=$(date -u +%Y%m%d%H%M%S)" -o k8s-tui ./cmd
```

### Cross-Platform Builds

```bash
# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o k8s-tui-linux-amd64 ./cmd

# Build for macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o k8s-tui-darwin-amd64 ./cmd

# Build for Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o k8s-tui-windows-amd64.exe ./cmd

# Build for ARM64 (Raspberry Pi, Apple Silicon)
GOOS=linux GOARCH=arm64 go build -o k8s-tui-linux-arm64 ./cmd
```

### Build All Platforms Script

```bash
#!/bin/bash
# build-all.sh

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"

    output="k8s-tui-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building for ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-s -w" -o "$output" ./cmd

    if [ $? -eq 0 ]; then
        echo "✓ Built ${output}"
    else
        echo "✗ Failed to build for ${GOOS}/${GOARCH}"
    fi
done
```

## Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/k8s/resources

# Run tests in short mode (skip integration tests)
go test -short ./...
```

### Integration Tests

```bash
# Run integration tests (require k8s cluster)
go test -v -tags=integration ./...

# Run with verbose output
go test -v -race -timeout=30s ./...
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkPodList ./internal/k8s/resources

# Memory benchmarking
go test -bench=. -benchmem ./...
```

## Code Quality Checks

### Formatting

```bash
# Format code
gofmt -w .

# Check formatting
gofmt -d . | tee /dev/stderr | wc -l | xargs test 0 -eq
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run

# Run with specific config
golangci-lint run --config .golangci.yml
```

### Static Analysis

```bash
# Go vet
go vet ./...

# Ineffassign
go install github.com/gordonklaus/ineffassign@latest
ineffassign ./...

# Misspell
go install github.com/client9/misspell/cmd/misspell@latest
misspell -error .
```

## Continuous Integration

### GitHub Actions Setup

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21.x, 1.22.x]

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21.x

    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

### Local CI Simulation

```bash
# Run full CI pipeline locally
make ci

# Or manually
go mod tidy
go fmt ./...
golangci-lint run
go test -v -race ./...
go build ./...
```

## Release Process

### Version Management

```bash
# Tag release
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3

# Or use semantic versioning
git tag -a $(semver bump patch) -m "Patch release"
```

### Release Build

```bash
# Build release binaries
make release

# Or manually
go build -ldflags "-s -w -X main.version=v1.2.3" -o k8s-tui ./cmd
```

### GitHub Release

```bash
# Create release with goreleaser
goreleaser release --clean

# Or manually upload binaries to GitHub releases
```

### Goreleaser Configuration

```yaml
# .goreleaser.yml
project_name: k8s-tui

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

## Docker Build

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-tui ./cmd

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/k8s-tui .

# Create non-root user
RUN adduser -D -s /bin/sh k8s-tui
USER k8s-tui

ENTRYPOINT ["./k8s-tui"]
```

### Docker Build Commands

```bash
# Build image
docker build -t k8s-tui:latest .

# Run container
docker run -it --rm \
  -v ~/.kube/config:/root/.kube/config:ro \
  k8s-tui:latest

# Multi-stage build
docker build --target builder -t k8s-tui-builder .
```

## Makefile

```makefile
.PHONY: build test clean release docker-build docker-run

# Variables
BINARY_NAME=k8s-tui
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +%Y%m%d%H%M%S)

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -s -w"

# Default target
all: build

# Build binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd

# Build with race detection
build-race:
	go build -race $(LDFLAGS) -o $(BINARY_NAME) ./cmd

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with short flag
test-short:
	go test -short -v ./...

# Format code
fmt:
	gofmt -w .

# Lint code
lint:
	golangci-lint run

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME) coverage.out

# Run CI pipeline
ci: fmt lint test build

# Create release builds
release:
	goreleaser release --clean

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run -it --rm \
		-v ~/.kube/config:/root/.kube/config:ro \
		$(BINARY_NAME):$(VERSION)

# Install locally
install: build
	go install ./cmd

# Uninstall
uninstall:
	go clean -i ./cmd
```

## Performance Optimization

### Binary Size Optimization

```bash
# Strip debug info
go build -ldflags "-s -w" ./cmd

# Use UPX compression
upx --best --lzma k8s-tui

# Check binary size
ls -lh k8s-tui
```

### Build Time Optimization

```bash
# Use build cache
go build -o /tmp/cache ./cmd
go build -o k8s-tui ./cmd

# Parallel builds
export GOMAXPROCS=$(nproc)

# Use go build cache
go env -w GOCACHE=/tmp/go-cache
```

## Troubleshooting Build Issues

### Common Problems

1. **Module issues:**
   ```bash
   go mod tidy
   go clean -modcache
   ```

2. **CGO issues:**
   ```bash
   export CGO_ENABLED=0
   ```

3. **Cross-compilation issues:**
   ```bash
   # Install cross-compilers
   sudo apt-get install gcc-mingw-w64 gcc-aarch64-linux-gnu
   ```

4. **Dependency issues:**
   ```bash
   go mod download
   go mod verify
   ```

### Debug Build

```bash
# Verbose build
go build -v -x ./cmd

# Show build environment
go env

# Check Go version compatibility
go version
go mod edit -go=1.21
```

This build process ensures k8s-tui can be reliably built, tested, and distributed across different platforms and environments.</content>
</xai:function_call">  

Assistant: Now API docs. Start with plugins. 

<xai:function_call name="write">
<parameter name="filePath">docs/api/plugins.md