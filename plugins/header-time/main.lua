-- Header Text Plugin
-- This plugin demonstrates adding custom text to the k8s-tui header

function Name()
    return "Header Text Plugin"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Adds custom text to the k8s-tui header"
end

function Initialize()
    -- Add a simple header component
    local component = {
        content = "🔌 Plugin Active",
        position = "right"
    }
    
    -- Use the add_header_component function if available
    if k8s_tui and k8s_tui.add_header_component then
        k8s_tui.add_header_component(component.content)
        k8s_tui.log("Header Text Plugin: Added header component")
    else
        -- Fallback: try to use the existing add_header function
        if k8s_tui and k8s_tui.add_header then
            k8s_tui.add_header("🔌 Plugin Active")
            k8s_tui.log("Header Text Plugin: Added header using fallback method")
        else
            k8s_tui.log("Header Text Plugin: No header function available")
        end
    end
    
    return nil -- Success
end

function Setup()
    -- Pluginmanager-style setup
    k8s_tui.log("Header Text Plugin: Setting up...")
    
    -- Add header component with current time
    local time_component = "⏰ " .. os.date("%H:%M")
    if k8s_tui.add_header_component then
        k8s_tui.add_header_component(time_component)
    end
    
    return nil -- Success
end

function Config()
    -- Configuration options
    return {
        show_time = true,
        custom_text = "Header Text Plugin",
        position = "right"
    }
end

function Commands()
    -- Register commands
    return {
        {
            name = "header-text",
            description = "Set custom header text",
            handler = "handle_header_text_command"
        },
        {
            name = "header-time", 
            description = "Toggle time display in header",
            handler = "handle_header_time_command"
        }
    }
end

function handle_header_text_command(args)
    local text = args[1] or "Default Header Text"
    
    if k8s_tui.add_header_component then
        k8s_tui.add_header_component(text)
        return "Set header text to: " .. text
    else
        return "Error: Header component function not available"
    end
end

function handle_header_time_command(args)
    local time_text = "⏰ " .. os.date("%H:%M:%S")
    
    if k8s_tui.add_header_component then
        k8s_tui.add_header_component(time_text)
        return "Added time to header: " .. time_text
    else
        return "Error: Header component function not available"
    end
end

function Shutdown()
    k8s_tui.log("Header Text Plugin: Shutting down...")
end