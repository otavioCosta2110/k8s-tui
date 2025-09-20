# EndpointSlice Plugin

A k8s-tui plugin that adds support for viewing Kubernetes EndpointSlice resources.

## Overview

EndpointSlices in Kubernetes track the IP addresses and ports of pods that back Services. This plugin provides a dedicated view for examining endpoint slice information, which is essential for understanding service-to-pod networking and troubleshooting connectivity issues. EndpointSlice is the modern replacement for the deprecated Endpoints API.

## Features

- **EndpointSlice Resource View**: Dedicated table view for Kubernetes EndpointSlices
- **Service Mapping**: Shows which service each endpoint slice belongs to
- **Address & Port Information**: Displays IP addresses and port configurations
- **Readiness Status**: Shows ready and not-ready endpoints
- **Namespace Awareness**: Works with current namespace context

## Resource Added

### EndpointSlices (`endpoints`)
Displays endpoint slice information including:
- Endpoint slice name
- IP addresses of backing pods (with readiness status)
- Port configurations
- Associated service name
- Age of the endpoint slice

## Configuration

```lua
{
    enabled = true,           -- Enable/disable the plugin
    refresh_interval = 30,    -- Refresh interval in seconds
    show_in_header = false    -- Show status in header
}
```

## Usage

1. Navigate to "EndpointSlices" in the resource list
2. View endpoint slice information in a structured table format
3. Each row shows endpoint slice details including addresses and ports
4. Use the provided commands for additional functionality

## Commands

- `endpoints:list` - List all endpoint slices in current namespace
- `endpoints:services` - Show endpoint slices grouped by services

## Real Data

The plugin queries the actual Kubernetes API to display real endpoint information from your cluster, including:
- Endpoint names and associated services
- IP addresses of backing pods
- Port configurations and protocols
- Endpoint age and creation timestamps
- Ready and not-ready pod addresses

## Understanding EndpointSlices

EndpointSlices are automatically created and managed by Kubernetes when you create a Service. They are the modern replacement for the deprecated Endpoints API and provide:
- Scalable endpoint tracking for large numbers of pods
- Better performance and resource usage
- Support for multiple IP families (IPv4/IPv6)
- Topology-aware endpoint selection
- Readiness status tracking

This plugin helps you:
- Debug service connectivity issues
- Understand load balancing distribution
- Monitor endpoint health and readiness
- Verify service-to-pod mappings
- Troubleshoot networking problems

## Plugin Architecture

This plugin demonstrates:
- Pluginmanager-style plugin structure
- Custom resource type registration for networking resources
- Integration with the modern Kubernetes discovery API
- Proper categorization in the UI
- Future-proof design using EndpointSlice instead of deprecated Endpoints</content>
</xai:function_call">Now let me test that the plugin loads correctly by running the test program. I should see the endpoints plugin in the list of loaded plugins. 

<xai:function_call name="bash">
<parameter name="command">cd /home/otavio/dev/k8s-tui && go run test_pluginmanager_plugins.go 2>&1 | grep -A5 -B5 "endpoints"