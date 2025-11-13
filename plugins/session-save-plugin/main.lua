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
  k8s_tui.log("DEBUG: CLIArguments called")
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

  -- Only support multi-cluster format
  local is_multicluster = content:match('"clusters"%s*:%s*%[')
  if not is_multicluster then
    return nil, "Only multi-cluster session format is supported"
  end
  
  k8s_tui.log("DEBUG: Loading multi-cluster session format")
  
  local namespace = nil
  local breadcrumb = {}
  local tabs = {}
  
  -- Extract current cluster index
  local current_cluster_pattern = '"current_cluster"%s*:%s*{[^}]*"Index"%s*:%s*(%d+)'
  local current_cluster_index = tonumber(content:match(current_cluster_pattern))
  
  if not current_cluster_index then
    k8s_tui.log("ERROR: Could not find current cluster index")
    return nil, "Could not find current cluster index in session file"
  end
  
  k8s_tui.log("DEBUG: Current cluster index: " .. current_cluster_index)
  
  -- Find the Nth cluster in the clusters array and extract its session
  local cluster_count = 0
  local pos = content:find('"clusters"%s*:%s*%[')
  if not pos then
    return nil, "Could not find clusters array in session file"
  end
  
  pos = pos + 1  -- Move past the opening bracket
  
  -- Find each cluster object
  while cluster_count <= current_cluster_index do
    local cluster_start = content:find('{', pos)
    if not cluster_start then
      break
    end
    
    -- Find the end of this cluster object
    local bracket_count = 0
    local in_string = false
    local escape_next = false
    local cluster_end = cluster_start
    
    for i = cluster_start, #content do
      local char = content:sub(i, i)
      if escape_next then
        escape_next = false
      elseif char == '\\' then
        escape_next = true
      elseif char == '"' then
        in_string = not in_string
      elseif not in_string then
        if char == '{' then
          bracket_count = bracket_count + 1
        elseif char == '}' then
          bracket_count = bracket_count - 1
          if bracket_count == 0 then
            cluster_end = i
            break
          end
        end
      end
    end
    
    if cluster_count == current_cluster_index then
      -- This is the current cluster, extract its session data
      local cluster_content = content:sub(cluster_start, cluster_end)
      k8s_tui.log("DEBUG: Found current cluster, extracting session data")
      
      -- Extract session object from this cluster using a more robust approach
      local session_start = cluster_content:find('"session"%s*:%s*{')
      if session_start then
        -- Find the opening brace of session object
        local brace_pos = cluster_content:find('{', session_start)
        if brace_pos then
          -- Find the matching closing brace for session object
          local bracket_count = 0
          local in_string = false
          local escape_next = false
          local session_end = brace_pos
          
          for i = brace_pos, #cluster_content do
            local char = cluster_content:sub(i, i)
            if escape_next then
              escape_next = false
            elseif char == '\\' then
              escape_next = true
            elseif char == '"' then
              in_string = not in_string
            elseif not in_string then
              if char == '{' then
                bracket_count = bracket_count + 1
              elseif char == '}' then
                bracket_count = bracket_count - 1
                if bracket_count == 0 then
                  session_end = i
                  break
                end
              end
            end
          end
          
          local session_content = cluster_content:sub(brace_pos, session_end)
          k8s_tui.log("DEBUG: Found session object: " .. session_content)
          
          -- Extract namespace from session
          namespace = session_content:match('"namespace"%s*:%s*"([^"]+)"')
          k8s_tui.log("DEBUG: Extracted namespace: " .. tostring(namespace))
          
          -- Extract breadcrumb from session
          local breadcrumb_match = session_content:match('"breadcrumb"%s*:%s*%[([^]]*)%]')
          if breadcrumb_match then
            for crumb in breadcrumb_match:gmatch('"([^"]+)"') do
              table.insert(breadcrumb, crumb)
            end
            k8s_tui.log("DEBUG: Extracted " .. #breadcrumb .. " breadcrumb items")
          end
          
          -- Extract tabs from session
          local tabs_match = session_content:match('"tabs"%s*:%s*(%b[])')
          if tabs_match then
            k8s_tui.log("DEBUG: Found tabs array: " .. tabs_match)
            for tab_str in tabs_match:gmatch('{%s*([^}]+)%s*}') do
              local tab = {}
              
              -- Extract ID
              local id = tab_str:match('"ID"%s*:%s*"([^"]+)"')
              if id then tab.ID = id end

              -- Extract title
              local title = tab_str:match('"Title"%s*:%s*"([^"]+)"')
              if title then tab.Title = title end

              -- Extract resourceType
              local resourceType = tab_str:match('"ResourceType"%s*:%s*"([^"]+)"')
              if resourceType then tab.ResourceType = resourceType end

              -- Extract currentIndex
              local currentIndex = tab_str:match('"CurrentIndex"%s*:%s*(%d+)')
              if currentIndex then tab.CurrentIndex = tonumber(currentIndex) end

              -- Extract breadcrumb array
              local tab_breadcrumb_match = tab_str:match('"Breadcrumb"%s*:%s*%[([^]]*)%]')
              if tab_breadcrumb_match then
                local tab_breadcrumb = {}
                for crumb in tab_breadcrumb_match:gmatch('"([^"]+)"') do
                  table.insert(tab_breadcrumb, crumb)
                end
                tab.Breadcrumb = tab_breadcrumb
              end

              -- Extract metadata object
              local metadata_match = tab_str:match('"Metadata"%s*:%s*{([^}]*)}')
              if metadata_match then
                local metadata = {}
                for k, v in metadata_match:gmatch('"([^"]+)"%s*:%s*"([^"]+)"') do
                  metadata[k] = v
                end
                tab.Metadata = metadata
              end

              if tab.ID and tab.Title and tab.ResourceType then
                table.insert(tabs, tab)
                k8s_tui.log("DEBUG: Added tab: " .. tab.Title .. " (ID: " .. tab.ID .. ")")
              end
            end
            k8s_tui.log("DEBUG: Total tabs extracted: " .. #tabs)
          end
        end
      else
        k8s_tui.log("ERROR: Could not find session object in current cluster")
      end
      break
    end
    
    cluster_count = cluster_count + 1
    pos = cluster_end + 1
  end
  
  if #tabs == 0 then
    k8s_tui.log("ERROR: Could not find session data for current cluster")
    return nil, "Could not find session data for current cluster"
  end

  -- Set namespace if found
  if namespace and k8s_tui and k8s_tui.set_namespace then
    k8s_tui.set_namespace(namespace)
    k8s_tui.log("DEBUG: Set namespace to: " .. namespace)
  end

  -- Restore tabs if found
  if #tabs > 0 and k8s_tui and k8s_tui.set_tabs then
    local result, err = k8s_tui.set_tabs(tabs)
    if err then
      k8s_tui.set_status("Failed to restore tabs: " .. tostring(err))
      return nil, "Failed to restore tabs: " .. tostring(err)
    end
    k8s_tui.log("DEBUG: Restored " .. #tabs .. " tabs")
  end

  -- Restore breadcrumb if found
  if #breadcrumb > 0 and k8s_tui and k8s_tui.set_breadcrumb_trail then
    k8s_tui.set_breadcrumb_trail(breadcrumb)
    k8s_tui.log("DEBUG: Restored breadcrumb with " .. #breadcrumb .. " items")
  end

  -- Set status
  local status_msg = "Session loaded from " .. value
  if namespace then
    status_msg = status_msg .. " (namespace: " .. namespace .. ")"
  end
  if #tabs > 0 then
    status_msg = status_msg .. " (tabs: " .. #tabs .. ")"
  end
  if #breadcrumb > 0 then
    status_msg = status_msg .. " (breadcrumb: " .. table.concat(breadcrumb, " > ") .. ")"
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

  -- Sync current cluster session data before getting all clusters
  if k8s_tui.sync_current_cluster_session then
    k8s_tui.log("DEBUG: Calling sync_current_cluster_session")
    local sync_result = k8s_tui.sync_current_cluster_session()
    k8s_tui.log("DEBUG: sync_current_cluster_session returned: " .. tostring(sync_result))
  else
    k8s_tui.log("ERROR: sync_current_cluster_session function not available!")
  end

  -- Get all clusters (now includes session data)
  k8s_tui.log("DEBUG: About to call get_all_clusters")
  local all_clusters = k8s_tui.get_all_clusters and k8s_tui.get_all_clusters() or {}
  k8s_tui.log("DEBUG: get_all_clusters returned, type: " .. type(all_clusters))

  if all_clusters then
    k8s_tui.log("DEBUG: Number of clusters: " .. #all_clusters)
    for i, cluster in ipairs(all_clusters) do
      k8s_tui.log("DEBUG: Cluster " .. i .. ": " .. tostring(cluster.Name) .. " (ID: " .. tostring(cluster.ID) .. ")")
    end
  else
    k8s_tui.log("ERROR: all_clusters is nil!")
  end

  local current_cluster = k8s_tui.get_current_cluster and k8s_tui.get_current_cluster() or nil
  if current_cluster then
    k8s_tui.log("DEBUG: Current cluster: " .. tostring(current_cluster.Name))
  else
    k8s_tui.log("ERROR: current_cluster is nil!")
  end

  -- Generate JSON with multi-cluster support (new format)
  local json = '{"clusters":['

  -- Save all clusters with their session data
  for i, cluster in ipairs(all_clusters) do
    if i > 1 then json = json .. ',' end
    json = json .. '{'
    json = json .. '"Index":' .. (cluster.Index or 0) .. ','
    json = json .. '"Name":"' .. (cluster.Name or "") .. '",'
    json = json .. '"Namespace":"' .. (cluster.Namespace or "default") .. '",'
    json = json .. '"Kubeconfig":"' .. (cluster.Kubeconfig or "") .. '",'
    json = json .. '"IsActive":' .. (cluster.IsActive and "true" or "false") .. ','

    -- Add session data for this cluster
    json = json .. '"session":{'
    json = json .. '"namespace":"' .. (cluster.Namespace or "default") .. '",'
    json = json .. '"breadcrumb":['
    if cluster.Breadcrumb then
      for j, crumb in ipairs(cluster.Breadcrumb) do
        if j > 1 then json = json .. ',' end
        json = json .. '"' .. crumb .. '"'
      end
    end
    json = json .. '],"tabs":['
    if cluster.Tabs then
      for j, tab in ipairs(cluster.Tabs) do
        if j > 1 then json = json .. ',' end
        json = json .. '{'
        json = json .. '"ID":"' .. (tab.ID or "") .. '",'
        json = json .. '"Title":"' .. (tab.Title or "") .. '",'
        json = json .. '"ResourceType":"' .. (tab.ResourceType or "") .. '",'
        json = json .. '"CurrentIndex":' .. (tab.CurrentIndex or 0) .. ','
        json = json .. '"Breadcrumb":['
        if tab.Breadcrumb then
          for k, crumb in ipairs(tab.Breadcrumb) do
            if k > 1 then json = json .. ',' end
            json = json .. '"' .. crumb .. '"'
          end
        end
        json = json .. '],"Metadata":{'
        if tab.Metadata then
          local first = true
          for k, v in pairs(tab.Metadata) do
            if not first then json = json .. ',' end
            json = json .. '"' .. k .. '":"' .. tostring(v) .. '"'
            first = false
          end
        end
        json = json .. '}'
        json = json .. '}'
      end
    end
    json = json .. ']}'
    json = json .. '}'
  end

  json = json .. '],"current_cluster":'

  -- Save current cluster info (without session, as it's in the clusters array)
  if current_cluster then
    -- Find the matching cluster in all_clusters to get the Index
    local current_index = 0
    for i, cluster in ipairs(all_clusters) do
      if cluster.ID == current_cluster.ID then
        current_index = cluster.Index
        break
      end
    end
    json = json .. '{'
    json = json .. '"Index":' .. current_index .. ','
    json = json .. '"Name":"' .. (current_cluster.Name or "") .. '",'
    json = json .. '"Namespace":"' .. (current_cluster.Namespace or "default") .. '",'
    json = json .. '"Kubeconfig":"' .. (current_cluster.Kubeconfig or "") .. '",'
    json = json .. '"IsActive":true'
    json = json .. '}'
  else
    json = json .. 'null'
  end

  json = json .. '}'

  local full_path = os.getenv("PWD") .. "/" .. path
  local file = io.open(full_path, "w")
  if file then
    file:write(json)
    file:close()
    k8s_tui.set_status("Session saved to " .. path .. " with " .. #all_clusters .. " cluster(s)")
  else
    k8s_tui.set_status("Failed to save session")
  end
end