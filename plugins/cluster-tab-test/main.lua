-- Cluster Tab Test Plugin
-- This plugin demonstrates adding cluster tabs programmatically

function Name()
    k8s_tui.log("DEBUG: Name() function called")
    return "cluster-tab-test"
end

function Version()
    k8s_tui.log("DEBUG: Version() function called")
    return "1.0.0"
end

function Description()
    k8s_tui.log("DEBUG: Description() function called")
    return "Test plugin for adding cluster tabs programmatically"
end

function Initialize()
    k8s_tui.log("DEBUG: Initialize() function called - Cluster Tab Test Plugin initialized")
    return nil
end

function Shutdown()
    k8s_tui.log("DEBUG: Shutdown() function called - Cluster Tab Test Plugin shutdown")
    return nil
end

function Setup(opts)
    k8s_tui.log("DEBUG: Setup() function called with opts: " .. tostring(opts))
    return nil
end

function Config()
    k8s_tui.log("DEBUG: Config() function called")
    return {}
end

function Commands()
    k8s_tui.log("DEBUG: Commands() function called - returning empty commands")
    return {}
end

function CLIArguments()
    k8s_tui.log("DEBUG: CLIArguments() function called - registering CLI arguments")
    
    local args = {
        {
            name = "add-test-cluster",
            description = "Add a test cluster tab",
            handler = "handle_add_test_cluster_arg"
        },
        {
            name = "list-clusters", 
            description = "List all available clusters",
            handler = "handle_list_clusters_arg"
        },
        {
            name = "switch-to-cluster",
            description = "Switch to a specific cluster by ID", 
            handler = "handle_switch_to_cluster_arg"
        }
    }
    
    k8s_tui.log("DEBUG: Registering " .. #args .. " CLI arguments")
    for i, arg in ipairs(args) do
        k8s_tui.log("DEBUG:   - " .. arg.name .. " (" .. arg.description .. ")")
    end
    
    return args
end

function Hooks()
    k8s_tui.log("DEBUG: Hooks() function called - returning empty hooks")
    return {}
end

function handle_add_test_cluster_arg(value)
    k8s_tui.log("DEBUG: handle_add_test_cluster_arg called with value: '" .. tostring(value) .. "'")
    
    -- Try to add a cluster with default kubeconfig path
    -- This will use the user's default kubeconfig
    k8s_tui.log("DEBUG: Calling k8s_tui.add_cluster_tab('', 'Test Cluster', 'default')")
    local err = k8s_tui.add_cluster_tab("", "Test Cluster", "default")
    
    if err then
        k8s_tui.log("DEBUG: add_cluster_tab returned error: " .. tostring(err))
        print("Error adding cluster: " .. err)
        return err
    else
        k8s_tui.log("DEBUG: add_cluster_tab succeeded")
        print("Successfully added test cluster tab")
        return nil
    end
end

function handle_list_clusters_arg(value)
    k8s_tui.log("DEBUG: === handle_list_clusters_arg called with value: '" .. tostring(value) .. "' ===")
    k8s_tui.log("DEBUG: Calling k8s_tui.get_clusters()")
    
    local clusters = k8s_tui.get_clusters()
    k8s_tui.log("DEBUG: get_clusters returned type: " .. type(clusters))
    
    if clusters == nil then
        k8s_tui.log("DEBUG: clusters is nil!")
        print("Available clusters: (nil)")
        return nil
    end
    
    k8s_tui.log("DEBUG: clusters length: " .. #clusters)
    
    -- Check if it's a table but empty
    local count = 0
    for k, v in pairs(clusters) do
        count = count + 1
        k8s_tui.log("DEBUG: Found cluster at key " .. tostring(k) .. ": " .. tostring(v))
    end
    k8s_tui.log("DEBUG: Actual cluster count using pairs: " .. count)
    
    local result = "Available clusters:\n"
    
    if #clusters == 0 then
        k8s_tui.log("DEBUG: No clusters found - this is expected when no kubeconfigs are provided")
        result = result .. "  No clusters available. Try running with a kubeconfig or use --add-test-cluster first."
    else
        for i, cluster in ipairs(clusters) do
            k8s_tui.log("DEBUG: Cluster " .. i .. ": " .. cluster.Name .. " (ID: " .. cluster.ID .. ", NS: " .. cluster.Namespace .. ")")
            result = result .. string.format("  %d. %s (ID: %s, Namespace: %s)\n", 
                i, cluster.Name, cluster.ID, cluster.Namespace)
        end
    end
    
    print("=== PLUGIN OUTPUT ===")
    print(result)
    print("=== END PLUGIN OUTPUT ===")
    return nil
end

function handle_switch_to_cluster_arg(value)
    k8s_tui.log("DEBUG: handle_switch_to_cluster_arg called with value: '" .. tostring(value) .. "'")
    
    if not value or value == "" then
        k8s_tui.log("DEBUG: No cluster ID provided")
        print("Error: --switch-to-cluster requires a cluster ID")
        return "cluster ID required"
    end
    
    k8s_tui.log("DEBUG: Calling k8s_tui.switch_to_cluster('" .. value .. "')")
    local err = k8s_tui.switch_to_cluster(value)
    
    if err then
        k8s_tui.log("DEBUG: switch_to_cluster returned error: " .. tostring(err))
        print("Error switching to cluster: " .. err)
        return err
    else
        k8s_tui.log("DEBUG: switch_to_cluster succeeded")
        print("Successfully switched to cluster " .. value)
        return nil
    end
end
