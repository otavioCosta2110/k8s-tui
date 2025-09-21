-- Neovim-Style Header Plugin for k8s-tui
-- Demonstrates Neovim-style plugin architecture with setup, config, commands, and hooks

-- Plugin metadata
function Name()
    return "neovim-header"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Neovim-style plugin that adds dynamic content to the header"
end

-- Default configuration
function Config()
    return {
        show_namespace = true,
        show_time = true,
        custom_message = "🚀 k8s-tui",
        update_interval = 30
    }
end

-- Setup function (called with user configuration)
function Setup(opts)
    print("Setting up Neovim Header Plugin with options:")
    for k, v in pairs(opts) do
        print("  " .. k .. " = " .. v)
    end

    -- Store configuration
    config = opts

    -- Add header component
    if config.custom_message then
        k8s_tui.add_header(config.custom_message)
    end

    return nil
end

-- Initialize the plugin
function Initialize()
    print("Neovim Header Plugin initialized")

    -- Add a status message
    k8s_tui.set_status("Neovim-style plugin loaded!")

    return nil
end

-- Shutdown the plugin
function Shutdown()
    print("Neovim Header Plugin shutting down")
    return nil
end

-- Commands provided by this plugin
function Commands()
    return {
        {
            name = "header:status",
            description = "Show header plugin status"
        },
        {
            name = "header:config",
            description = "Show current header configuration"
        },
        {
            name = "header:update",
            description = "Update header with current time"
        }
    }
end

-- Hooks that this plugin registers for
function Hooks()
    return {
        {
            event = "app_started",
            handler = "on_app_started"
        },
        {
            event = "namespace_changed",
            handler = "on_namespace_changed"
        }
    }
end

-- Hook handlers
function on_app_started(data)
    print("Header plugin: App started event received")
    k8s_tui.set_status("Header plugin ready!")
end

function on_namespace_changed(data)
    print("Header plugin: Namespace changed to " .. data)
    local current_ns = k8s_tui.get_namespace()
    k8s_tui.set_status("Switched to namespace: " .. current_ns)
end

-- Command handlers (these would be called when commands are executed)
function handle_header_status(args)
    return "Header plugin is active with " .. #args .. " arguments", nil
end

function handle_header_config(args)
    local config_str = ""
    for k, v in pairs(config) do
        config_str = config_str .. k .. "=" .. tostring(v) .. " "
    end
    return "Header config: " .. config_str, nil
end

function handle_header_update(args)
    local time_str = os.date("%H:%M:%S")
    k8s_tui.add_header("🕐 " .. time_str)
    return "Header updated with current time: " .. time_str, nil
end
