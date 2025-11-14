function Name()
  return "session-save-plugin"
end

function Version()
  return "1.0.0"
end

function Description()
  return "Plugin to save and load sessions from JSON files"
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
    },
    {
      name = "session:load",
      description = "Load session from JSON file",
      handler = "session_load"
    },
    {
      name = "session:load_submit",
      description = "Internal command to load session with provided filename",
      handler = "session_load_submit"
    },
    {
      name = "session:load_cancel",
      description = "Internal command to cancel session load",
      handler = "session_load_cancel"
    }
  }
end

function load_session_cli_handler(value)
  k8s_tui.log("DEBUG: load_session_cli_handler called with value: " .. (value or "nil"))
  
  if not value or value == "" then
    k8s_tui.set_status("Error: Session file path required")
    return "Error: Session file path required", nil
  end
  
  return load_session_from_file(value)
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

function session_load()
  if not k8s_tui then
    return
  end

  -- Show input dialog for filename
  k8s_tui.show_input_dialog("Enter session file path:", "session.json", "session:load_submit", "session:load_cancel")
end

function session_load_submit(filename)
  if k8s_tui and k8s_tui.log then
    k8s_tui.log("DEBUG: session_load_submit called with filename: '" .. (filename or "nil") .. "' (type: " .. type(filename) .. ")")
  end
  
  local actual_filename = filename
  if not filename or filename == "" then
    actual_filename = "session.json"
    if k8s_tui and k8s_tui.log then
      k8s_tui.log("DEBUG: Using default filename: session.json")
    end
  end
  
  return load_session_from_file(actual_filename)
end

function session_load_cancel()
  -- Cancel callback: do nothing
  return "Session load cancelled", nil
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

function load_session_from_file(filename)
  if not k8s_tui then
    return "Error: k8s_tui API not available", nil
  end

  if k8s_tui.log then
    k8s_tui.log("DEBUG: load_session_from_file called with filename: '" .. filename .. "'")
  end

  -- Ensure .json extension
  local path = filename
  if not path:match("%.json$") then
    path = path .. ".json"
  end

  local full_path = os.getenv("PWD") .. "/" .. path
  local file = io.open(full_path, "r")
  if not file then
    local error_msg = "Failed to open session file: " .. path
    k8s_tui.set_status(error_msg)
    return error_msg, nil
  end

  local content = file:read("*all")
  file:close()

  if not content or content == "" then
    local error_msg = "Session file is empty: " .. path
    k8s_tui.set_status(error_msg)
    return error_msg, nil
  end

  -- Parse JSON (simple JSON parsing for our specific format)
  local session_data = parse_session_json(content)
  if not session_data then
    local error_msg = "Failed to parse session file: " .. path
    k8s_tui.set_status(error_msg)
    return error_msg, nil
  end

  if k8s_tui.log then
    k8s_tui.log("DEBUG: Parsed session data with " .. #(session_data.clusters or {}) .. " clusters")
  end

  -- Load clusters from session
  local clusters = session_data.clusters or {}
  local loaded_clusters = {}
  local active_cluster_id = nil

  for i, cluster_data in ipairs(clusters) do
    if k8s_tui.log then
      k8s_tui.log("DEBUG: Loading cluster " .. i .. ": " .. (cluster_data.Name or "unknown"))
    end

    -- Add cluster tab using the API
    local kubeconfig = cluster_data.Kubeconfig or ""
    local name = cluster_data.Name or "Cluster " .. i
    local namespace = cluster_data.Namespace or "default"
    
    k8s_tui.log("DEBUG: Adding cluster - Name: '" .. name .. "', Namespace: '" .. namespace .. "', Kubeconfig: '" .. kubeconfig .. "'")
    
    local result = k8s_tui.add_cluster_tab(kubeconfig, name, namespace)

    if result and result.cluster_id then
      loaded_clusters[i] = result.cluster_id
      k8s_tui.log("DEBUG: Added cluster with ID: " .. result.cluster_id)

      -- Store active cluster info
      if cluster_data.IsActive or (session_data.current_cluster and session_data.current_cluster.Index == (cluster_data.Index or i-1)) then
        active_cluster_id = result.cluster_id
      end
    else
      k8s_tui.log("ERROR: Failed to add cluster: " .. (cluster_data.Name or "unknown"))
    end
  end

  -- Switch to active cluster
  if active_cluster_id then
    k8s_tui.log("DEBUG: Switching to active cluster: " .. active_cluster_id)
    local switch_result = k8s_tui.switch_to_cluster(active_cluster_id)
    if switch_result then
      k8s_tui.set_status("Session loaded: " .. #loaded_clusters .. " cluster(s), active: " .. active_cluster_id)
      return "Session loaded successfully", nil
    else
      k8s_tui.set_status("Session loaded but failed to switch to active cluster")
      return "Session loaded but failed to switch to active cluster", nil
    end
  else
    k8s_tui.set_status("Session loaded: " .. #loaded_clusters .. " cluster(s) (no active cluster)")
    return "Session loaded but no active cluster found", nil
  end
end

function parse_session_json(content)
  -- Simple JSON parser for our specific session format
  -- This is a basic implementation that handles the session file structure
  
  local session_data = {}
  
  k8s_tui.log("DEBUG: Parsing session JSON, content length: " .. #content)
  
  -- Extract clusters array
  local clusters_start = content:find('"clusters"%s*:%s*%[')
  if not clusters_start then
    k8s_tui.log("ERROR: No clusters array found in session file")
    return nil
  end
  
  k8s_tui.log("DEBUG: Found clusters at position: " .. clusters_start)
  
  local bracket_start = content:find('%[', clusters_start)
  k8s_tui.log("DEBUG: Found bracket at position: " .. (bracket_start or "nil"))
  
  if not bracket_start then
    k8s_tui.log("ERROR: Could not find opening bracket")
    return nil
  end
  
  local clusters_end = find_matching_bracket(content, bracket_start)
  if not clusters_end then
    k8s_tui.log("ERROR: Invalid clusters array format")
    return nil
  end
  
  k8s_tui.log("DEBUG: Clusters end at position: " .. clusters_end)
  
  local clusters_end = find_matching_bracket(content, content:find('%[', clusters_start))
  if not clusters_end then
    k8s_tui.log("ERROR: Invalid clusters array format")
    return nil
  end
  
  -- Extract content inside the brackets (excluding the brackets themselves)
  local clusters_content = content:sub(bracket_start + 1, clusters_end - 1)
  session_data.clusters = parse_clusters_array(clusters_content)
  
  -- Extract current_cluster
  local current_start = content:find('"current_cluster"%s*:%s*{', clusters_end)
  if current_start then
    local bracket_start = content:find('{', current_start)
    local current_end = find_matching_bracket(content, bracket_start)
    if current_end then
      local current_str = content:sub(current_start, current_end)
      session_data.current_cluster = parse_current_cluster(current_str)
    end
  end
  
  return session_data
end

function find_matching_bracket(str, start_pos)
  local bracket = str:sub(start_pos, start_pos)
  local open_char, close_char
  
  if bracket == '[' then
    open_char, close_char = '[', ']'
  elseif bracket == '{' then
    open_char, close_char = '{', '}'
  else
    return nil
  end
  
  local depth = 1
  local pos = start_pos + 1
  
  while pos <= #str and depth > 0 do
    local char = str:sub(pos, pos)
    if char == open_char then
      depth = depth + 1
    elseif char == close_char then
      depth = depth - 1
    end
    pos = pos + 1
  end
  
  if depth == 0 then
    return pos - 1
  else
    return nil
  end
end

function parse_clusters_array(clusters_str)
  local clusters = {}
  
  k8s_tui.log("DEBUG: Parsing clusters array, string length: " .. #clusters_str)
  
  -- Extract individual cluster objects - clusters_str should start with { and end with }
  local pos = 1  -- Start after opening {
  
  while pos < #clusters_str do
    local cluster_start = clusters_str:find('{', pos)
    if not cluster_start then 
      k8s_tui.log("DEBUG: No more cluster objects found")
      break 
    end
    
    k8s_tui.log("DEBUG: Found cluster start at position: " .. cluster_start)
    
    local cluster_end = find_matching_bracket(clusters_str, cluster_start)
    if not cluster_end then 
      k8s_tui.log("ERROR: Failed to find cluster end bracket")
      break 
    end
    
    k8s_tui.log("DEBUG: Found cluster end at position: " .. cluster_end)
    
    local cluster_str = clusters_str:sub(cluster_start, cluster_end)
    local cluster = parse_cluster_object(cluster_str)
    
    if cluster then
      table.insert(clusters, cluster)
      k8s_tui.log("DEBUG: Successfully parsed cluster " .. #clusters)
    else
      k8s_tui.log("ERROR: Failed to parse cluster object")
    end
    
    pos = cluster_end + 1
  end
  
  while pos < (content_end or #clusters_str) do
    local cluster_start = clusters_str:find('{', pos)
    if not cluster_start then 
      k8s_tui.log("DEBUG: No more cluster objects found")
      break 
    end
    
    k8s_tui.log("DEBUG: Found cluster start at position: " .. cluster_start)
    
    local cluster_end = find_matching_bracket(clusters_str, cluster_start)
    if not cluster_end then 
      k8s_tui.log("ERROR: Failed to find cluster end bracket")
      break 
    end
    
    k8s_tui.log("DEBUG: Found cluster end at position: " .. cluster_end)
    
    local cluster_str = clusters_str:sub(cluster_start, cluster_end)
    local cluster = parse_cluster_object(cluster_str)
    
    if cluster then
      table.insert(clusters, cluster)
      k8s_tui.log("DEBUG: Successfully parsed cluster " .. #clusters)
    else
      k8s_tui.log("ERROR: Failed to parse cluster object")
    end
    
    pos = cluster_end + 1
  end
  
  k8s_tui.log("DEBUG: Parsed " .. #clusters .. " clusters total")
  return clusters
end

function parse_cluster_object(cluster_str)
  local cluster = {}
  
  -- Extract Index
  cluster.Index = extract_number_field(cluster_str, "Index")
  
  -- Extract Name
  cluster.Name = extract_string_field(cluster_str, "Name")
  
  -- Extract Namespace
  cluster.Namespace = extract_string_field(cluster_str, "Namespace")
  
  -- Extract Kubeconfig
  cluster.Kubeconfig = extract_string_field(cluster_str, "Kubeconfig")
  
  -- Extract IsActive
  local is_active = extract_string_field(cluster_str, "IsActive")
  cluster.IsActive = (is_active == "true")
  
  -- Extract session data if present
  local session_start = cluster_str:find('"session"%s*:%s*{')
  if session_start then
    local bracket_start = cluster_str:find('{', session_start)
    local session_end = find_matching_bracket(cluster_str, bracket_start)
    if session_end then
      cluster.session = parse_session_data(cluster_str:sub(session_start, session_end))
    end
  end
  
  return cluster
end

function parse_current_cluster(current_str)
  local current = {}
  
  current.Index = extract_number_field(current_str, "Index")
  current.Name = extract_string_field(current_str, "Name")
  current.Namespace = extract_string_field(current_str, "Namespace")
  current.Kubeconfig = extract_string_field(current_str, "Kubeconfig")
  
  local is_active = extract_string_field(current_str, "IsActive")
  current.IsActive = (is_active == "true")
  
  return current
end

function parse_session_data(session_str)
  local session = {}
  
  session.namespace = extract_string_field(session_str, "namespace") or "default"
  
  -- Parse breadcrumb array
  local breadcrumb_start = session_str:find('"breadcrumb"%s*:%s*%[')
  if breadcrumb_start then
    local bracket_start = session_str:find('%[', breadcrumb_start)
    local breadcrumb_end = find_matching_bracket(session_str, bracket_start)
    if breadcrumb_end then
      session.breadcrumb = parse_string_array(session_str:sub(breadcrumb_start, breadcrumb_end))
    end
  end
  
  -- Parse tabs array
  local tabs_start = session_str:find('"tabs"%s*:%s*%[')
  if tabs_start then
    local bracket_start = session_str:find('%[', tabs_start)
    local tabs_end = find_matching_bracket(session_str, bracket_start)
    if tabs_end then
      session.tabs = parse_tabs_array(session_str:sub(tabs_start, tabs_end))
    end
  end
  
  return session
end

function parse_string_array(array_str)
  local items = {}
  local pos = array_str:find('%[') + 1
  local end_pos = array_str:find(']', pos) - 1
  
  while pos < end_pos do
    local item_start, item_end = array_str:find('"([^"]*)"', pos)
    if item_start and item_start <= end_pos then
      table.insert(items, array_str:sub(item_start + 1, item_end - 1))
      pos = item_end + 1
    else
      break
    end
  end
  
  return items
end

function parse_tabs_array(tabs_str)
  local tabs = {}
  local pos = tabs_str:find('%[') + 1
  local end_pos = tabs_str:find(']', pos) - 1
  
  while pos < end_pos do
    local tab_start = tabs_str:find('{', pos)
    if not tab_start then break end
    
    local tab_end = find_matching_bracket(tabs_str, tab_start)
    if not tab_end then break end
    
    local tab_str = tabs_str:sub(tab_start, tab_end)
    local tab = parse_tab_object(tab_str)
    
    if tab then
      table.insert(tabs, tab)
    end
    
    pos = tab_end + 1
  end
  
  return tabs
end

function parse_tab_object(tab_str)
  local tab = {}
  
  tab.ID = extract_string_field(tab_str, "ID")
  tab.Title = extract_string_field(tab_str, "Title")
  tab.ResourceType = extract_string_field(tab_str, "ResourceType")
  tab.CurrentIndex = extract_number_field(tab_str, "CurrentIndex") or 0
  
  -- Parse breadcrumb array
  local breadcrumb_start = tab_str:find('"Breadcrumb"%s*:%s*%[')
  if breadcrumb_start then
    local bracket_start = tab_str:find('%[', breadcrumb_start)
    local breadcrumb_end = find_matching_bracket(tab_str, bracket_start)
    if breadcrumb_end then
      tab.Breadcrumb = parse_string_array(tab_str:sub(breadcrumb_start, breadcrumb_end))
    end
  end
  
  -- Parse Metadata object
  local metadata_start = tab_str:find('"Metadata"%s*:%s*{')
  if metadata_start then
    local bracket_start = tab_str:find('{', metadata_start)
    local metadata_end = find_matching_bracket(tab_str, bracket_start)
    if metadata_end then
      tab.Metadata = parse_metadata_object(tab_str:sub(metadata_start, metadata_end))
    end
  end
  
  return tab
end

function parse_metadata_object(metadata_str)
  local metadata = {}
  
  -- Simple key-value parsing for metadata
  for key, value in metadata_str:gmatch('"(.-)"%s*:%s*"([^"]*)"') do
    metadata[key] = value
  end
  
  return metadata
end

function extract_string_field(str, field_name)
  local pattern = '"' .. field_name .. '"%s*:%s*"([^"]*)"'
  local match = str:match(pattern)
  return match
end

function extract_number_field(str, field_name)
  local pattern = '"' .. field_name .. '"%s*:%s*(%d+)'
  local match = str:match(pattern)
  if match then
    return tonumber(match)
  end
  return nil
end
