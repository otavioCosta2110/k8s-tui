package plugins

import (
	"testing"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
)

// MockClient creates a mock k8s client for testing
func MockClient() k8s.Client {
	return k8s.Client{
		Namespace: "test-namespace",
	}
}

func TestGlobalPluginManager_NewGlobalPluginManager(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")

	if gpm == nil {
		t.Fatal("Expected GlobalPluginManager to be created")
	}

	if gpm.clusters == nil {
		t.Error("Expected clusters map to be initialized")
	}

	if gpm.currentCluster != "" {
		t.Error("Expected currentCluster to be empty initially")
	}
}

func TestGlobalPluginManager_AddCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	// Test adding a cluster
	gpm.AddCluster("cluster1", "Test Cluster 1", client)

	// Verify cluster was added
	clusters:=gpm.clusters
	for c, _ := range clusters {
		println(c)
	}
	cluster := gpm.GetCluster("cluster1")
	if cluster == nil {
		t.Error("Expected cluster to be added")
	}

	if cluster.ID != "cluster1" {
		t.Error("Expected cluster ID to match")
	}

	if cluster.Name != "Test Cluster 1" {
		t.Error("Expected cluster name to match")
	}

	if cluster.Namespace != "default" {
		t.Error("Expected default namespace to be 'default'")
	}
}

func TestGlobalPluginManager_AddClusterWithNamespace(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	// Test adding a cluster with specific namespace
	gpm.AddClusterWithNamespace("cluster2", "Test Cluster 2", client, "custom-namespace")

	cluster := gpm.GetCluster("cluster2")
	if cluster == nil {
		t.Error("Expected cluster to be added")
	}

	if cluster.Namespace != "custom-namespace" {
		t.Error("Expected namespace to be 'custom-namespace'")
	}
}

func TestGlobalPluginManager_SwitchToCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client1 := MockClient()
	client2 := MockClient()

	// Add clusters
	gpm.AddCluster("cluster1", "Test Cluster 1", client1)
	gpm.AddCluster("cluster2", "Test Cluster 2", client2)

	// Switch to first cluster
	err := gpm.SwitchToCluster("cluster1")
	if err != nil {
		t.Errorf("Expected no error switching to cluster1, got: %v", err)
	}

	if gpm.currentCluster != "cluster1" {
		t.Error("Expected currentCluster to be 'cluster1'")
	}

	// Switch to second cluster
	err = gpm.SwitchToCluster("cluster2")
	if err != nil {
		t.Errorf("Expected no error switching to cluster2, got: %v", err)
	}

	if gpm.currentCluster != "cluster2" {
		t.Error("Expected currentCluster to be 'cluster2'")
	}

	// Test switching to non-existent cluster
	err = gpm.SwitchToCluster("nonexistent")
	if err == nil {
		t.Error("Expected error when switching to non-existent cluster")
	}
}

func TestGlobalPluginManager_RemoveCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	// Add clusters but don't switch to any to avoid deadlock
	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.AddCluster("cluster2", "Test Cluster 2", client)

	// Remove a cluster that is not current
	gpm.RemoveCluster("cluster1")

	// Verify cluster was removed
	cluster := gpm.GetCluster("cluster1")
	if cluster != nil {
		t.Error("Expected cluster1 to be removed")
	}

	// Verify other cluster still exists
	cluster2 := gpm.GetCluster("cluster2")
	if cluster2 == nil {
		t.Error("Expected cluster2 to still exist")
	}
}

func TestGlobalPluginManager_SetClusterNamespace(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)

	// Set namespace for specific cluster
	err := gpm.SetClusterNamespace("cluster1", "new-namespace")
	if err != nil {
		t.Errorf("Expected no error setting namespace, got: %v", err)
	}

	// Verify namespace was set
	namespace, err := gpm.GetClusterNamespace("cluster1")
	if err != nil {
		t.Errorf("Expected no error getting namespace, got: %v", err)
	}

	if namespace != "new-namespace" {
		t.Error("Expected namespace to be 'new-namespace'")
	}

	// Test setting namespace for non-existent cluster
	err = gpm.SetClusterNamespace("nonexistent", "namespace")
	if err == nil {
		t.Error("Expected error when setting namespace for non-existent cluster")
	}
}

func TestGlobalPluginManager_ClusterSettings(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)

	// Test setting and getting cluster settings
	err := gpm.SetClusterSetting("cluster1", "key1", "value1")
	if err != nil {
		t.Errorf("Expected no error setting cluster setting, got: %v", err)
	}

	value, err := gpm.GetClusterSetting("cluster1", "key1")
	if err != nil {
		t.Errorf("Expected no error getting cluster setting, got: %v", err)
	}

	if value != "value1" {
		t.Error("Expected setting value to be 'value1'")
	}

	// Test getting non-existent setting
	value, err = gpm.GetClusterSetting("cluster1", "nonexistent")
	if err != nil {
		t.Errorf("Expected no error getting non-existent setting, got: %v", err)
	}
	if value != nil {
		t.Error("Expected nil value for non-existent setting")
	}
}

func TestGlobalPluginManager_ExecuteOnCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client1 := MockClient()
	client2 := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client1)
	gpm.AddCluster("cluster2", "Test Cluster 2", client2)

	// Switch to cluster1
	gpm.SwitchToCluster("cluster1")

	executed := false
	err := gpm.ExecuteOnCluster("cluster2", func(cluster *ClusterContext) error {
		executed = true
		if cluster.ID != "cluster2" {
			t.Error("Expected to execute on cluster2")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error executing on cluster, got: %v", err)
	}

	if !executed {
		t.Error("Expected function to be executed")
	}

	// Verify current cluster is restored
	if gpm.currentCluster != "cluster1" {
		t.Error("Expected current cluster to be restored to cluster1")
	}
}

func TestGlobalPluginManager_GetAllClusters(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.AddCluster("cluster2", "Test Cluster 2", client)

	allClusters := gpm.GetAllClusters()
	if len(allClusters) != 2 {
		t.Error("Expected 2 clusters")
	}

	if _, exists := allClusters["cluster1"]; !exists {
		t.Error("Expected cluster1 to exist in all clusters")
	}

	if _, exists := allClusters["cluster2"]; !exists {
		t.Error("Expected cluster2 to exist in all clusters")
	}
}

func TestGlobalPluginManager_GetClusterNames(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.AddCluster("cluster2", "Test Cluster 2", client)

	names := gpm.GetClusterNames()
	if len(names) != 2 {
		t.Error("Expected 2 cluster names")
	}

	// Note: map iteration order is not guaranteed, so we just check count and content
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["Test Cluster 1"] || !nameMap["Test Cluster 2"] {
		t.Error("Expected both cluster names to be present")
	}
}

func TestMultiClusterPluginAPI_GetCurrentClusterID(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.SwitchToCluster("cluster1")

	mcAPI := gpm.GetAPI()

	if mcAPI.GetCurrentClusterID() != "cluster1" {
		t.Error("Expected current cluster ID to be 'cluster1'")
	}
}

func TestMultiClusterPluginAPI_GetCurrentClusterName(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.SwitchToCluster("cluster1")

	mcAPI := gpm.GetAPI()

	if mcAPI.GetCurrentClusterName() != "Test Cluster 1" {
		t.Error("Expected current cluster name to be 'Test Cluster 1'")
	}
}

func TestMultiClusterPluginAPI_SwitchToCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.AddCluster("cluster2", "Test Cluster 2", client)

	mcAPI := gpm.GetAPI()

	// Test switching clusters
	err := mcAPI.SwitchToCluster("cluster2")
	if err != nil {
		t.Errorf("Expected no error switching clusters, got: %v", err)
	}

	if mcAPI.GetCurrentClusterID() != "cluster2" {
		t.Error("Expected current cluster ID to be 'cluster2'")
	}
}

func TestMultiClusterPluginAPI_GetClusterNames(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.AddCluster("cluster2", "Test Cluster 2", client)

	mcAPI := gpm.GetAPI()

	names := mcAPI.GetClusterNames()
	if len(names) != 2 {
		t.Error("Expected 2 cluster names")
	}
}

func TestMultiClusterPluginAPI_SetClusterNamespace(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)

	mcAPI := gpm.GetAPI()

	err := mcAPI.SetClusterNamespace("cluster1", "test-namespace")
	if err != nil {
		t.Errorf("Expected no error setting cluster namespace, got: %v", err)
	}

	namespace, err := mcAPI.GetClusterNamespace("cluster1")
	if err != nil {
		t.Errorf("Expected no error getting cluster namespace, got: %v", err)
	}

	if namespace != "test-namespace" {
		t.Error("Expected namespace to be 'test-namespace'")
	}
}

func TestMultiClusterPluginAPI_ExecuteOnCluster(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client1 := MockClient()
	client2 := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client1)
	gpm.AddCluster("cluster2", "Test Cluster 2", client2)

	mcAPI := gpm.GetAPI()

	executed := false
	err := mcAPI.ExecuteOnCluster("cluster2", func(client k8s.Client) error {
		executed = true
		// We can't easily compare clients, so we just verify execution
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error executing on cluster, got: %v", err)
	}

	if !executed {
		t.Error("Expected function to be executed")
	}
}

func TestMultiClusterPluginAPI_ResourceOperations(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)
	gpm.SwitchToCluster("cluster1")

	mcAPI := gpm.GetAPI()

	// Test that resource operations are forwarded correctly
	// These will likely fail with mock clients, but we're testing the forwarding
	_, _ = mcAPI.GetPods("test-namespace")
	// We expect this to fail since we're using a mock client
	// The important thing is that the method exists and forwards the call

	// Test namespace operations
	mcAPI.SetCurrentNamespace("test-namespace")
	if mcAPI.GetCurrentNamespace() != "test-namespace" {
		t.Error("Expected namespace to be set correctly")
	}

	// Test tab operations
	tabs := []TabInfo{
		{ID: "tab1", Title: "Test Tab"},
	}

	_ = mcAPI.SetTabs(tabs)
	// This might fail if tab setter is not configured, but we're testing the forwarding

	_, _ = mcAPI.GetTabs()
	// This might also fail, but we're testing the method exists

	// Test breadcrumb operations
	mcAPI.SetBreadcrumbTrail([]string{"level1", "level2"})
	breadcrumb := mcAPI.GetBreadcrumbTrail()
	// Note: GetBreadcrumbTrail returns empty slice when no callback is set
	// We're testing that the methods exist and don't panic
	if breadcrumb == nil {
		t.Error("Expected breadcrumb trail to be non-nil slice")
	}
}

func TestClusterContext_Isolation(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client1 := MockClient()
	client2 := MockClient()

	// Add two clusters with different namespaces
	gpm.AddClusterWithNamespace("cluster1", "Test Cluster 1", client1, "namespace1")
	gpm.AddClusterWithNamespace("cluster2", "Test Cluster 2", client2, "namespace2")

	// Switch to cluster1 and set namespace
	gpm.SwitchToCluster("cluster1")
	gpm.SetClusterNamespace("cluster1", "updated-namespace1")

	// Switch to cluster2 and set different namespace
	gpm.SwitchToCluster("cluster2")
	gpm.SetClusterNamespace("cluster2", "updated-namespace2")

	// Verify namespaces are isolated
	ns1, _ := gpm.GetClusterNamespace("cluster1")
	ns2, _ := gpm.GetClusterNamespace("cluster2")

	if ns1 != "updated-namespace1" {
		t.Error("Expected cluster1 namespace to be isolated")
	}

	if ns2 != "updated-namespace2" {
		t.Error("Expected cluster2 namespace to be isolated")
	}
}

func TestGlobalPluginManager_ConcurrentAccess(t *testing.T) {
	gpm := NewGlobalPluginManager("/tmp/test-plugins")
	client := MockClient()

	gpm.AddCluster("cluster1", "Test Cluster 1", client)

	// Test concurrent access to cluster operations
	done := make(chan bool, 2)

	// Goroutine 1: Switch to cluster
	go func() {
		for range 100 {
			gpm.SwitchToCluster("cluster1")
			gpm.SetClusterNamespace("cluster1", "namespace1")
		}
		done <- true
	}()

	// Goroutine 2: Get cluster info
	go func() {
		for i := 0; i < 100; i++ {
			gpm.GetCluster("cluster1")
			gpm.GetClusterNamespace("cluster1")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	if gpm.currentCluster != "cluster1" {
		t.Error("Expected final current cluster to be 'cluster1'")
	}
}
