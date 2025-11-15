-- Test cluster-specific tabs and breadcrumb functions
function Name()
  return "test-cluster-specific-tabs"
end

function Version()
  return "1.0.0"
end

function Description()
  return "Test cluster-specific tabs and breadcrumb functions"
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
      name = "test_cluster_tabs",
      description = "Test getting tabs for specific clusters",
      handler = "test_cluster_tabs_handler"
    }
  }
end

function test_cluster_tabs_handler(value)
  -- Test get_all_clusters first
  local all_clusters = k8s_tui.get_all_clusters()
  if not all_clusters then
    k8s_tui.set_status("ERROR: Could not get clusters")
    return "Could not get clusters", nil
  end
  
  k8s_tui.set_status("Found " .. #all_clusters .. " clusters")
  
  -- Test getting tabs for each cluster
  for i, cluster in ipairs(all_clusters) do
    if cluster.ID then
      -- Test new get_cluster_tabs function
      local tabs = k8s_tui.get_cluster_tabs(cluster.ID)
      local tab_count = tabs and #tabs or 0
      
      -- Test new get_cluster_breadcrumb function
      local breadcrumb = k8s_tui.get_cluster_breadcrumb(cluster.ID)
      local breadcrumb_count = breadcrumb and #breadcrumb or 0
      
      k8s_tui.log("Cluster " .. i .. " (" .. tostring(cluster.ID) .. "): " .. tab_count .. " tabs, " .. breadcrumb_count .. " breadcrumb items")
      
      -- Verify tabs are different between clusters
      if i > 1 then
        local prev_cluster = all_clusters[i-1]
        if prev_cluster and prev_cluster.ID then
          local prev_tabs = k8s_tui.get_cluster_tabs(prev_cluster.ID)
          local prev_count = prev_tabs and #prev_tabs or 0
          
          if tab_count ~= prev_count then
            k8s_tui.log("SUCCESS: Cluster " .. i .. " has " .. tab_count .. " tabs, different from previous cluster's " .. prev_count .. " tabs")
          else
            k8s_tui.log("INFO: Cluster " .. i .. " has same number of tabs as previous cluster: " .. tab_count)
          end
        end
      end
    end
  end
  
  k8s_tui.set_status("Tested cluster-specific tabs for " .. #all_clusters .. " clusters")
  return "Tested cluster-specific tabs", nil
end