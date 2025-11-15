-- Test session save with multiple clusters
function Name()
  return "test-session-save"
end

function Version()
  return "1.0.0"
end

function Description()
  return "Test session save functionality"
end

function Initialize()
  return nil
end

function Shutdown()
  return nil
end

function CLIArguments()
  return {
    {
      name = "test_save_session",
      description = "Test saving session with multiple clusters",
      handler = "test_save_session_handler"
    }
  }
end

function test_save_session_handler(value)
  -- Test sync all clusters sessions
  if k8s_tui.sync_all_clusters_sessions then
    k8s_tui.log("DEBUG: Testing sync_all_clusters_sessions")
    local sync_result = k8s_tui.sync_all_clusters_sessions()
    k8s_tui.set_status("Sync result: " .. tostring(sync_result))
  else
    k8s_tui.set_status("ERROR: sync_all_clusters_sessions not available")
    return "sync_all_clusters_sessions not available", nil
  end
  
  -- Test get all clusters
  local all_clusters = k8s_tui.get_all_clusters()
  if all_clusters then
    for i, cluster in ipairs(all_clusters) do
      local tab_count = cluster.tabs and #cluster.tabs or 0
      k8s_tui.log("Cluster " .. i .. ": " .. tostring(cluster.Name) .. " has " .. tab_count .. " tabs")
    end
    k8s_tui.set_status("Found " .. #all_clusters .. " clusters")
    return "Found " .. #all_clusters .. " clusters", nil
  else
    k8s_tui.set_status("ERROR: Could not get clusters")
    return "Could not get clusters", nil
  end
end