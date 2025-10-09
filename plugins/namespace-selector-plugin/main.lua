-- Namespace Selector Plugin for k8s-tui
-- Allows switching namespaces by typing the name

function Name()
    return "namespace-selector-plugin"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Plugin to switch namespaces by typing the namespace name"
end

function Config()
    return {
        enabled = true
    }
end

function Setup(opts)
    return nil
end

function Initialize()
    return nil
end

function Shutdown()
    return nil
end

function Hooks()
    return {}
end

function Commands()
    return {
        {
            name = "namespace:switch",
            description = "Switch to a different namespace by typing the name",
            handler = "namespace_switch"
        }
    }
end

function namespace_switch()
    if not k8s_tui then
        return
    end

    k8s_tui.log("namespace_switch called")

    -- Get current namespace for default
    local current_ns = k8s_tui.get_namespace and k8s_tui.get_namespace() or "default"

    -- Prompt for namespace name
    k8s_tui.show_text_input("Switch Namespace", "Enter namespace name (current: " .. current_ns .. ")", function(namespace)
        if namespace == "" then
            k8s_tui.set_status("Namespace switch cancelled - no name provided")
            return
        end

        -- Basic validation
        if namespace:match("%s") then
            k8s_tui.set_status("Invalid namespace name: cannot contain spaces")
            return
        end

        -- Trim whitespace
        namespace = namespace:gsub("^%s*(.-)%s*$", "%1")

        if namespace == "" then
            k8s_tui.set_status("Invalid namespace name: cannot be empty")
            return
        end

        -- Switch namespace
        k8s_tui.set_namespace(namespace)
        k8s_tui.set_status("Switched to namespace: " .. namespace)
    end, function()
        k8s_tui.set_status("Namespace switch cancelled")
    end)
end