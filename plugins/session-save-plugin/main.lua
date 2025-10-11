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
  print("DEBUG: CLIArguments called")
  return {
    {
      name = "session",
      description = "Load session from JSON file",
      handler = "load_session_cli_handler"
    }
  }
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
  print("DEBUG: Lua Initialize called")
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
      name = "session:save",
      description = "Save current session to JSON file",
      handler = "session_save"
    },
    {
      name = "session:save_submit",
      description = "Internal command to save session with provided filename",
      handler = "session_save_submit"
    },
    {
      name = "session:save_cancel",
      description = "Internal command to cancel session save",
      handler = "session_save_cancel"
    }
  }
end

function load_session_cli_handler(value)
  -- Read the session file
  local file = io.open(value, "r")
  if not file then
    return nil, "Failed to open session file: " .. value
  end

  local content = file:read("*all")
  file:close()

  -- Simple JSON parsing for namespace
  -- Look for "namespace":"value"
  local namespace_pattern = '"namespace"%s*:%s*"([^"]+)"'
  local namespace = content:match(namespace_pattern)

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



  local tabs = {}
  if tabs_content then
    for tab_str in tabs_content:gmatch('{%s*([^}]+)%s*}') do
      local tab = {}

      -- Extract ID
      local id_pattern = '"ID"%s*:%s*"([^"]+)"'
      local id = tab_str:match(id_pattern)
      if id then tab.ID = id end

      -- Extract title
      local title_pattern = '"Title"%s*:%s*"([^"]+)"'
      local title = tab_str:match(title_pattern)
      if title then tab.Title = title end

      -- Extract resourceType
      local resourceType_pattern = '"ResourceType"%s*:%s*"([^"]+)"'
      local resourceType = tab_str:match(resourceType_pattern)
      if resourceType then tab.ResourceType = resourceType end

      -- Extract breadcrumb array
      local breadcrumb_pattern = '"Breadcrumb"%s*:%s*%[([^]]*)%]'
      local breadcrumb_str = tab_str:match(breadcrumb_pattern)
      if breadcrumb_str then
        local breadcrumb = {}
        for crumb in breadcrumb_str:gmatch('"([^"]+)"') do
          table.insert(breadcrumb, crumb)
        end
        tab.Breadcrumb = breadcrumb
      end

      if tab.ID and tab.Title and tab.ResourceType then
        table.insert(tabs, tab)
      end
    end
  end





  -- Set namespace if found
  if namespace and k8s_tui and k8s_tui.set_namespace then
    k8s_tui.set_namespace(namespace)
  end

  -- Restore tabs if found
  if #tabs > 0 and k8s_tui and k8s_tui.set_tabs then
    local result, err = k8s_tui.set_tabs(tabs)
    if err then
      k8s_tui.set_status("Failed to restore tabs: " .. tostring(err))
      return nil, "Failed to restore tabs: " .. tostring(err)
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
  k8s_tui.set_status(status_msg)

  return "Session loaded successfully", nil
end

function session_save()
  if not k8s_tui then
    return
  end

  -- Show input dialog for filename
  k8s_tui.show_input_dialog("Enter Session name:", "session.json", "session:save_submit", "session:save_cancel")
end

function session_save_submit(filename)
  if k8s_tui and k8s_tui.log then
    k8s_tui.log("DEBUG: session_save_submit called with filename: '" .. (filename or "nil") .. "' (type: " .. type(filename) .. ")")
  end
  -- For debugging, set status with the received filename
  if k8s_tui and k8s_tui.set_status then
    k8s_tui.set_status("DEBUG: Received filename: '" .. (filename or "nil") .. "' (len: " .. string.len(filename or "") .. ")")
  end
  local actual_filename = filename
  if not filename or filename == "" then
    actual_filename = "session.json"
    if k8s_tui and k8s_tui.log then
      k8s_tui.log("DEBUG: Using default filename: session.json")
    end
  end
  save_session_to_file(actual_filename)
  return "Session saved", nil
end

function session_save_cancel()
  -- Cancel callback: do nothing
  return "Session save cancelled", nil
end

function save_session_to_file(filename)
  if not k8s_tui then
    return
  end

  -- Use provided filename
  local path = filename
  -- Ensure it has .json extension if not present
  if not path:match("%.json$") then
    path = path .. ".json"
  end

  if k8s_tui and k8s_tui.log then
    k8s_tui.log("DEBUG: save_session_to_file called with filename: '" .. (filename or "nil") .. "', using path: '" .. path .. "'")
  end

  -- Debug: show what we're saving to
  if k8s_tui and k8s_tui.set_status then
    k8s_tui.set_status("DEBUG: Saving to: " .. path)
  end

  local namespace = k8s_tui.get_namespace and k8s_tui.get_namespace() or ""
  local tabs = k8s_tui.get_tabs and k8s_tui.get_tabs() or {}

  -- Generate simple JSON
  local json = '{"namespace":"' .. namespace .. '","tabs":['
  for i, tab in ipairs(tabs) do
    if i > 1 then json = json .. ',' end
    json = json ..
        '{"ID":"' ..
        (tab.ID or "") ..
        '","Title":"' .. (tab.Title or "") .. '","ResourceType":"' .. (tab.ResourceType or "") .. '","Breadcrumb":['
    if tab.Breadcrumb then
      for j, crumb in ipairs(tab.Breadcrumb) do
        if j > 1 then json = json .. ',' end
        json = json .. '"' .. crumb .. '"'
      end
    end
    json = json .. ']}'
  end
  json = json .. ']}'

  local full_path = os.getenv("PWD") .. "/" .. path
  local file = io.open(full_path, "w")
  if file then
    file:write(json)
    file:close()
    k8s_tui.set_status("Session saved to " .. path)
  else
    k8s_tui.set_status("Failed to save session")
  end
end
