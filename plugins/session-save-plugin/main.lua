 -- print("DEBUG: session-save-plugin Lua file loaded")

function Name()
    return "session-save-plugin"
end

function Version()
    return "1.0.0"
end

function Description()
    return "Plugin to save current session to JSON file with Ctrl+S"
end

function CLIArguments()
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: CLIArguments function called")
    end
    local args = {
        {
            name = "session",
            description = "Load session from JSON file",
            handler = "load_session_cli_handler"
        }
    }
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Returning CLI arguments table with " .. #args .. " arguments")
    end
    return args
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

function load_session_cli_handler(value)
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: load_session_cli_handler called with value: " .. tostring(value))
    end

    -- Read the session file
    local file = io.open(value, "r")
    if not file then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Failed to open session file: " .. value)
        end
        return nil, "Failed to open session file: " .. value
    end

    local content = file:read("*all")
    file:close()

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Session file content: " .. content)
    end

    -- Simple JSON parsing for namespace
    -- Look for "namespace":"value"
    local namespace_pattern = '"namespace"%s*:%s*"([^"]+)"'
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Looking for namespace with pattern: " .. namespace_pattern)
    end
    local namespace = content:match(namespace_pattern)
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Namespace pattern match result: " .. (namespace or "nil"))
    end
    if namespace then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Found namespace: '" .. namespace .. "'")
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No namespace found in content")
        end
    end

    -- Parse tabs array
    -- Look for "tabs": [ ... ] and extract the array content
    -- Find the position after "tabs":[
    local tabs_start = content:find('"tabs"%s*:%s*%[')
    local tabs_content = nil
    if tabs_start then
        -- Find the matching closing bracket
        local bracket_count = 0
        local in_string = false
        local escape_next = false
        local end_pos = tabs_start - 1

        for i = tabs_start, #content do
            local char = content:sub(i, i)
            if escape_next then
                escape_next = false
            elseif char == '\\' then
                escape_next = true
            elseif char == '"' then
                in_string = not in_string
            elseif not in_string then
                if char == '[' then
                    bracket_count = bracket_count + 1
                elseif char == ']' then
                    bracket_count = bracket_count - 1
                    if bracket_count == 0 then
                        end_pos = i
                        break
                    end
                end
            end
        end

        if end_pos > tabs_start then
            -- Extract content between brackets
            local start_bracket = content:find('%[', tabs_start)
            if start_bracket then
                tabs_content = content:sub(start_bracket + 1, end_pos - 1)
            end
        end
    end

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Tabs content extraction result: " .. (tabs_content or "nil"))
    end
    if tabs_content then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Found tabs_content: '" .. tabs_content .. "'")
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No tabs_content found in content")
        end
    end

    local tabs = {}
    if tabs_content then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Parsing tabs from content")
            k8s_tui.log("DEBUG: Starting to parse individual tab objects from tabs_content")
        end
        for tab_str in tabs_content:gmatch('{%s*([^}]+)%s*}') do
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Processing tab_str: '" .. tab_str .. "'")
            end
            local tab = {}

            -- Extract ID
            local id_pattern = '"id"%s*:%s*"([^"]+)"'
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Looking for ID with pattern: " .. id_pattern)
            end
            local id = tab_str:match(id_pattern)
            if id then tab.ID = id end
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: ID match result: " .. (id or "nil"))
            end

            -- Extract title
            local title_pattern = '"title"%s*:%s*"([^"]+)"'
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Looking for title with pattern: " .. title_pattern)
            end
            local title = tab_str:match(title_pattern)
            if title then tab.Title = title end
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Title match result: " .. (title or "nil"))
            end

            -- Extract resourceType
            local resourceType_pattern = '"resourceType"%s*:%s*"([^"]+)"'
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Looking for resourceType with pattern: " .. resourceType_pattern)
            end
            local resourceType = tab_str:match(resourceType_pattern)
            if resourceType then tab.ResourceType = resourceType end
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: ResourceType match result: " .. (resourceType or "nil"))
            end

            -- Extract breadcrumb array
            local breadcrumb_pattern = '"breadcrumb"%s*:%s*%[([^]]*)%]'
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Looking for breadcrumb with pattern: " .. breadcrumb_pattern)
            end
            local breadcrumb_str = tab_str:match(breadcrumb_pattern)
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Breadcrumb match result: " .. (breadcrumb_str or "nil"))
            end
            if breadcrumb_str then
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Found breadcrumb_str: '" .. breadcrumb_str .. "'")
                    k8s_tui.log("DEBUG: Parsing breadcrumb crumbs")
                end
                local breadcrumb = {}
                for crumb in breadcrumb_str:gmatch('"([^"]+)"') do
                    table.insert(breadcrumb, crumb)
                    if k8s_tui and k8s_tui.log then
                        k8s_tui.log("DEBUG: Added breadcrumb crumb: '" .. crumb .. "'")
                    end
                end
                tab.Breadcrumb = breadcrumb
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Breadcrumb array has " .. #breadcrumb .. " elements")
                end
            else
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: No breadcrumb found")
                end
            end

            if tab.ID and tab.Title and tab.ResourceType then
                table.insert(tabs, tab)
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Added tab to tabs table")
                end
            else
                if k8s_tui and k8s_tui.log then
                    k8s_tui.log("DEBUG: Tab missing required fields, skipping")
                end
            end
        end
    end

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Total tabs parsed: " .. #tabs)
    end

    -- Check if k8s_tui API is available
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Checking if k8s_tui is available: " .. tostring(k8s_tui ~= nil))
    end
    if k8s_tui then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: k8s_tui type: " .. type(k8s_tui))
        end
        if k8s_tui.set_namespace then
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.set_namespace exists")
            end
        else
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.set_namespace does not exist")
            end
        end
        if k8s_tui.restore_tabs then
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs exists")
            end
        else
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs does not exist")
            end
        end
    end

    -- Set namespace if found
    if namespace then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Calling k8s_tui.set_namespace with: " .. namespace)
        end
        if k8s_tui and k8s_tui.set_namespace then
            k8s_tui.set_namespace(namespace)
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.set_namespace called successfully")
            end
        else
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.set_namespace not available")
            end
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No namespace found, skipping set_namespace")
        end
    end

    -- Restore tabs if found
    if #tabs > 0 then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Calling k8s_tui.restore_tabs with " .. #tabs .. " tabs")
        end
        for i, tab in ipairs(tabs) do
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Tab " .. i .. ": ID=" .. (tab.ID or "nil") .. ", Title=" .. (tab.Title or "nil") .. ", ResourceType=" .. (tab.ResourceType or "nil"))
                if tab.Breadcrumb then
                    k8s_tui.log("DEBUG: Tab " .. i .. " Breadcrumb: " .. table.concat(tab.Breadcrumb, " > "))
                end
            end
        end
        if k8s_tui and k8s_tui.restore_tabs then
            k8s_tui.restore_tabs(tabs)
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs called")
            end
        else
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs not available")
            end
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No tabs found, skipping restore_tabs")
        end
    end

    -- Restore tabs if found
    if #tabs > 0 then
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: Calling k8s_tui.restore_tabs with " .. #tabs .. " tabs")
        end
        for i, tab in ipairs(tabs) do
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: Tab " .. i .. ": ID=" .. (tab.ID or "nil") .. ", Title=" .. (tab.Title or "nil") .. ", ResourceType=" .. (tab.ResourceType or "nil"))
                if tab.Breadcrumb then
                    k8s_tui.log("DEBUG: Tab " .. i .. " Breadcrumb: " .. table.concat(tab.Breadcrumb, " > "))
                end
            end
        end
        local result, err = k8s_tui.restore_tabs(tabs)
        if err then
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs failed with error: " .. tostring(err))
            end
            k8s_tui.set_status("Failed to restore tabs: " .. tostring(err))
            return nil, "Failed to restore tabs: " .. tostring(err)
        else
            if k8s_tui and k8s_tui.log then
                k8s_tui.log("DEBUG: k8s_tui.restore_tabs succeeded with result: " .. tostring(result))
            end
        end
    else
        if k8s_tui and k8s_tui.log then
            k8s_tui.log("DEBUG: No tabs found, skipping restore_tabs")
        end
    end

    -- Set status
    local status_msg = "Session loaded from " .. value
    if namespace then
        status_msg = status_msg .. " (namespace: " .. namespace .. ")"
    end
    if #tabs > 0 then
        status_msg = status_msg .. " (tabs: " .. #tabs .. ")"
    end
    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: Setting final status: " .. status_msg)
    end
    k8s_tui.set_status(status_msg)

    if k8s_tui and k8s_tui.log then
        k8s_tui.log("DEBUG: load_session_cli_handler completed successfully")
    end
    return "Session loaded successfully", nil
end
