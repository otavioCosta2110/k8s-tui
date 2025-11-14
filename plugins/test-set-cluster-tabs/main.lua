-- Test set_cluster_tabs API function
function Name()
  return "test-set-cluster-tabs"
end

function Version()
  return "1.0.0"
end

function Description()
  return "Test set_cluster_tabs API function"
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
      name = "test_set_clusters",
      description = "Test set_cluster_tabs function",
      handler = "test_set_clusters_handler"
    }
  }
end

function test_set_clusters_handler(value)
  -- Create test cluster configurations
  local clusters = {
    {
      KubeconfigPath = "",
      ClusterName = "Test Cluster 1",
      Namespace = "default"
    },
    {
      KubeconfigPath = "/home/otavio/.kube/config",
      ClusterName = "Test Cluster 2", 
      Namespace = "kube-system"
    }
  }
  
  -- Call set_cluster_tabs API
  local result = k8s_tui.set_cluster_tabs(clusters)
  
  if result then
    k8s_tui.set_status("ERROR: Failed to set cluster tabs: " .. tostring(result))
    return "Failed to set cluster tabs: " .. tostring(result), nil
  else
    k8s_tui.set_status("SUCCESS: Set cluster tabs successfully")
    return "Set cluster tabs successfully", nil
  end
end