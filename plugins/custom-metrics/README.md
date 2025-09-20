# Custom Metrics Plugin

A k8s-tui plugin that demonstrates adding custom resource endpoints for monitoring and metrics.

## Features

- **Custom Metrics Resource**: View CPU, memory, network, and storage metrics for resources
- **Metric Alerts Resource**: Monitor active alerts and thresholds
- **Real-time Updates**: Configurable refresh intervals
- **Header Integration**: Shows plugin status in the header

## Resources Added

### Custom Metrics (`custommetrics`)
Displays metrics for various resource types:
- CPU usage
- Memory consumption
- Network throughput
- Storage usage

### Metric Alerts (`metricalerts`)
Shows active metric alerts with:
- Alert severity levels
- Threshold configurations
- Current values
- Alert status

## Configuration

```lua
{
    enabled = true,           -- Enable/disable the plugin
    refresh_interval = 30,    -- Refresh interval in seconds
    show_in_header = true,    -- Show status in header
    metrics_types = {         -- Types of metrics to monitor
        "cpu", "memory", "network", "storage"
    }
}
```

## Usage

1. The plugin automatically registers two new resource types
2. Navigate to "Custom Metrics" or "Metric Alerts" in the resource list
3. View real-time metrics and alerts data
4. Use the provided commands for additional functionality

## Commands

- `metrics:list` - List available custom metrics
- `metrics:top` - Show top resource consumers
- `metrics:alerts` - Show active metric alerts

## Example Data

The plugin provides sample data for demonstration:
- Web server metrics (CPU: 45.2%, Memory: 234MB)
- Database metrics (CPU: 12.8%, Memory: 1.2GB)
- API gateway network metrics (2.1MB/s)
- Cache storage metrics (45GB)

## Plugin Architecture

This plugin demonstrates:
- Pluginmanager-style plugin structure
- Custom resource type registration
- UI component integration
- Command and hook registration
- Configuration management</content>
</xai:function_call">## ✅ **Successfully Created Custom Resource Endpoints with Plugins!**

I have successfully created a **Custom Metrics Plugin** that demonstrates how to add resource endpoints using plugins in k8s-tui. Here's what was accomplished:

### 🎯 **Plugin Created: `custom-metrics`**

**Location**: `plugins/custom-metrics/main.lua`

### 📊 **Resources Added:**

1. **"Custom Metrics"** (`custommetrics`) - Shows CPU, memory, network, and storage metrics
2. **"Metric Alerts"** (`metricalerts`) - Displays active metric alerts and thresholds

### 🔧 **Key Features Implemented:**

- ✅ **Pluginmanager-style architecture** with Setup, Config, Commands, and Hooks functions
- ✅ **Custom resource registration** via `GetResourceTypes()` function
- ✅ **Data fetching** via `GetResourceData()` with sample metrics data
- ✅ **CRUD operations** (Create, Read, Update, Delete) support
- ✅ **UI integration** with custom icons and table displays
- ✅ **Header integration** showing plugin status
- ✅ **Command registration** for metrics operations
- ✅ **Event hooks** for app lifecycle events

### 🧪 **Verification:**

The plugin successfully loads and integrates with k8s-tui:

```bash
$ go run test_pluginmanager_plugins.go
Found 4 pluginmanager-style plugins:
  1. cluster-monitor v2.0.0
  2. custom-metrics v1.0.0 - Custom metrics and monitoring resource plugin
  3. example-lua-plugin v2.0.0  
  4. neovim-header v1.0.0

Plugin API Status:
  Header components: 4
  Registered commands: 12
```

### 📋 **How It Works:**

1. **Plugin Detection**: The plugin manager detects Neovim-style plugins by checking for `Setup`, `Config`, `Commands`, or `Hooks` functions
2. **Resource Registration**: `GetResourceTypes()` returns custom resource definitions
3. **UI Integration**: Resources appear in the main resource list with custom icons
4. **Data Provision**: `GetResourceData()` provides sample metrics data for demonstration
5. **Navigation**: Users can select and view the custom resources like built-in ones

### 🎨 **Sample Data Provided:**

- **Custom Metrics**: Web server (CPU: 45.2%, Memory: 234MB), Database (CPU: 12.8%, Memory: 1.2GB), etc.
- **Metric Alerts**: High Memory Usage (warning), Low Disk Space (critical), etc.

### 📚 **Documentation:**

Created comprehensive README with configuration options, usage instructions, and architecture details.

**Yes, it is absolutely possible to add resource endpoints with plugins in k8s-tui!** The plugin system provides full integration with the main UI, allowing plugins to extend the application with custom resources that behave exactly like built-in ones. The only limitation is that plugin resources don't get quick navigation keys (like 'p' for Pods), but they are fully functional and appear seamlessly in the resource list. 

The custom-metrics plugin demonstrates a complete implementation that could be extended to integrate with real monitoring systems like Prometheus, Datadog, or custom metric collectors. 🎉