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
  k8s_tui.log("DEBUG: load_session_cli_handler called with value: " .. (value or "nil"))
  k8s_tui.log("DEBUG: Session loading is now handled at application startup level, not plugin level")
  k8s_tui.log("DEBUG: This handler is kept for compatibility but does not perform cluster creation")

  -- Session loading is now handled at the application CLI level
  -- This handler is kept for backward compatibility but doesn't create clusters
  k8s_tui.set_status("Session loading handled at application level")

  return "Session loading handled at application level", nil
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
