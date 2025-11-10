package plugins

import (
	"fmt"
	"sync"

	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

// ClusterContext holds cluster-specific information for plugins
type ClusterContext struct {
	ID        string
	Name      string
	Client    k8s.Client
	Settings  map[string]interface{}
	Namespace string
}

// GlobalPluginManager manages a single plugin instance across multiple clusters
type GlobalPluginManager struct {
	*PluginManager // Embed the existing plugin manager
	mu             sync.RWMutex
	clusters       map[string]*ClusterContext // cluster ID -> context
	currentCluster string                     // current active cluster ID
}

// NewGlobalPluginManager creates a new global plugin manager
func NewGlobalPluginManager(pluginDir string) *GlobalPluginManager {
	pm := NewPluginManager(pluginDir)

	gpm := &GlobalPluginManager{
		PluginManager: pm,
		clusters:      make(map[string]*ClusterContext),
	}

	// Set the global manager reference in the API
	pm.SetGlobalManagerReference(gpm)

	return gpm
}

// AddCluster adds a new cluster context to the global plugin manager
func (gpm *GlobalPluginManager) AddCluster(id, name string, client k8s.Client) {
	gpm.AddClusterWithNamespace(id, name, client, "default")
}

// AddClusterWithNamespace adds a new cluster context with a specific namespace
func (gpm *GlobalPluginManager) AddClusterWithNamespace(id, name string, client k8s.Client, namespace string) {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	gpm.clusters[id] = &ClusterContext{
		ID:        id,
		Name:      name,
		Client:    client,
		Settings:  make(map[string]interface{}),
		Namespace: namespace,
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Added cluster %s (%s) with namespace %s", name, id, namespace))
}

// RemoveCluster removes a cluster context from the global plugin manager
func (gpm *GlobalPluginManager) RemoveCluster(id string) {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	if cluster, exists := gpm.clusters[id]; exists {
		delete(gpm.clusters, id)

		// If this was the current cluster, switch to first available
		if gpm.currentCluster == id {
			if len(gpm.clusters) > 0 {
				for newID := range gpm.clusters {
					gpm.SwitchToCluster(newID)
					break
				}
			} else {
				gpm.currentCluster = ""
				gpm.api.SetClient(k8s.Client{}) // Clear client
			}
		}

		logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Removed cluster %s", cluster.Name))
	}
}

// SwitchToCluster switches the active cluster context for plugins
func (gpm *GlobalPluginManager) SwitchToCluster(id string) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[id]
	if !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	gpm.currentCluster = id
	gpm.api.SetClient(cluster.Client)
	gpm.api.SetCurrentNamespace(cluster.Namespace)

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Switched to cluster %s (%s) with namespace %s", cluster.Name, id, cluster.Namespace))
	return nil
}

// GetCurrentCluster returns the current active cluster context
func (gpm *GlobalPluginManager) GetCurrentCluster() *ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	if gpm.currentCluster == "" {
		return nil
	}

	return gpm.clusters[gpm.currentCluster]
}

// GetCluster returns a specific cluster context by ID
func (gpm *GlobalPluginManager) GetCluster(id string) *ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	return gpm.clusters[id]
}

// GetAllClusters returns all cluster contexts
func (gpm *GlobalPluginManager) GetAllClusters() map[string]*ClusterContext {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	result := make(map[string]*ClusterContext)
	for id, cluster := range gpm.clusters {
		result[id] = cluster
	}
	return result
}

// GetClusterNames returns all cluster names
func (gpm *GlobalPluginManager) GetClusterNames() []string {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	var names []string
	for _, cluster := range gpm.clusters {
		names = append(names, cluster.Name)
	}
	return names
}

// SetClusterNamespace sets the namespace for a specific cluster
func (gpm *GlobalPluginManager) SetClusterNamespace(clusterID, namespace string) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Namespace = namespace

	// If this is the current cluster, also update the API's namespace
	if gpm.currentCluster == clusterID {
		gpm.api.SetCurrentNamespace(namespace)
	}

	logger.Info(fmt.Sprintf("🔌 Global Plugin Manager: Set namespace %s for cluster %s (%s)", namespace, cluster.Name, clusterID))
	return nil
}

// GetClusterNamespace gets the namespace for a specific cluster
func (gpm *GlobalPluginManager) GetClusterNamespace(clusterID string) (string, error) {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	logger.Info(fmt.Sprintf("🔍 DEBUG: GetClusterNamespace called for clusterID %s", clusterID))
	logger.Info(fmt.Sprintf("🔍 DEBUG: Available clusters: %v", gpm.getClusterIDs()))

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		logger.Error(fmt.Sprintf("🔍 DEBUG: Cluster %s not found in global manager", clusterID))
		return "", fmt.Errorf("cluster %s not found", clusterID)
	}

	logger.Info(fmt.Sprintf("🔍 DEBUG: Found cluster %s with namespace %s", clusterID, cluster.Namespace))
	return cluster.Namespace, nil
}

// Helper method to get all cluster IDs for debugging
func (gpm *GlobalPluginManager) getClusterIDs() []string {
	var ids []string
	for id := range gpm.clusters {
		ids = append(ids, id)
	}
	return ids
}

// SetClusterSetting sets a cluster-specific setting
func (gpm *GlobalPluginManager) SetClusterSetting(clusterID, key string, value interface{}) error {
	gpm.mu.Lock()
	defer gpm.mu.Unlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.Settings[key] = value
	return nil
}

// GetClusterSetting gets a cluster-specific setting
func (gpm *GlobalPluginManager) GetClusterSetting(clusterID, key string) (interface{}, error) {
	gpm.mu.RLock()
	defer gpm.mu.RUnlock()

	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	return cluster.Settings[key], nil
}

// ExecuteOnCluster executes a function with a specific cluster's context
func (gpm *GlobalPluginManager) ExecuteOnCluster(clusterID string, fn func(*ClusterContext) error) error {
	gpm.mu.Lock()
	cluster, exists := gpm.clusters[clusterID]
	if !exists {
		gpm.mu.Unlock()
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	// Save current cluster
	oldCluster := gpm.currentCluster
	oldClient := gpm.api.GetClient()

	// Switch to target cluster
	gpm.currentCluster = clusterID
	gpm.api.SetClient(cluster.Client)
	gpm.mu.Unlock()

	// Execute function
	err := fn(cluster)

	// Restore previous cluster
	gpm.mu.Lock()
	gpm.currentCluster = oldCluster
	gpm.api.SetClient(oldClient)
	gpm.mu.Unlock()

	return err
}

// GetAPI returns the plugin API with multi-cluster support
func (gpm *GlobalPluginManager) GetAPI() *MultiClusterPluginAPI {
	return &MultiClusterPluginAPI{
		api:           gpm.PluginManager.GetAPI(),
		globalManager: gpm,
	}
}

// MultiClusterPluginAPI extends PluginAPI with multi-cluster capabilities
type MultiClusterPluginAPI struct {
	api           *PluginAPIImpl
	globalManager *GlobalPluginManager
}

// GetCurrentClusterID returns the current active cluster ID
func (mc *MultiClusterPluginAPI) GetCurrentClusterID() string {
	return mc.globalManager.currentCluster
}

// GetCurrentClusterName returns the current active cluster name
func (mc *MultiClusterPluginAPI) GetCurrentClusterName() string {
	cluster := mc.globalManager.GetCurrentCluster()
	if cluster == nil {
		return ""
	}
	return cluster.Name
}

// SwitchToCluster switches to a different cluster
func (mc *MultiClusterPluginAPI) SwitchToCluster(clusterID string) error {
	return mc.globalManager.SwitchToCluster(clusterID)
}

// GetClusterNames returns all available cluster names
func (mc *MultiClusterPluginAPI) GetClusterNames() []string {
	return mc.globalManager.GetClusterNames()
}

// SetClusterNamespace sets the namespace for a specific cluster
func (mc *MultiClusterPluginAPI) SetClusterNamespace(clusterID, namespace string) error {
	return mc.globalManager.SetClusterNamespace(clusterID, namespace)
}

// GetClusterNamespace gets the namespace for a specific cluster
func (mc *MultiClusterPluginAPI) GetClusterNamespace(clusterID string) (string, error) {
	return mc.globalManager.GetClusterNamespace(clusterID)
}

// ExecuteOnCluster executes a function in the context of a specific cluster
func (mc *MultiClusterPluginAPI) ExecuteOnCluster(clusterID string, fn func(client k8s.Client) error) error {
	return mc.globalManager.ExecuteOnCluster(clusterID, func(cluster *ClusterContext) error {
		return fn(cluster.Client)
	})
}

// Forward all PluginAPI methods to the embedded api
func (mc *MultiClusterPluginAPI) GetCurrentNamespace() string {
	return mc.api.GetCurrentNamespace()
}

func (mc *MultiClusterPluginAPI) SetCurrentNamespace(namespace string) {
	mc.api.SetCurrentNamespace(namespace)
}

func (mc *MultiClusterPluginAPI) SetStatusMessage(message string) {
	mc.api.SetStatusMessage(message)
}

func (mc *MultiClusterPluginAPI) AddHeaderComponent(component UIInjectionPoint) {
	mc.api.AddHeaderComponent(component)
}

func (mc *MultiClusterPluginAPI) AddFooterComponent(component UIInjectionPoint) {
	mc.api.AddFooterComponent(component)
}

func (mc *MultiClusterPluginAPI) GetHeaderComponents() []UIInjectionPoint {
	return mc.api.GetHeaderComponents()
}

func (mc *MultiClusterPluginAPI) GetFooterComponents() []UIInjectionPoint {
	return mc.api.GetFooterComponents()
}

func (mc *MultiClusterPluginAPI) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	mc.api.RegisterCommand(name, description, handler)
}

func (mc *MultiClusterPluginAPI) ExecuteCommand(name string, args []string) (string, error) {
	return mc.api.ExecuteCommand(name, args)
}

func (mc *MultiClusterPluginAPI) RegisterCLIArgument(name, description string, handler func(value string) error) {
	mc.api.RegisterCLIArgument(name, description, handler)
}

func (mc *MultiClusterPluginAPI) GetCLIArguments() map[string]CLIArgument {
	return mc.api.GetCLIArguments()
}

func (mc *MultiClusterPluginAPI) HasCLIArgument(name string) bool {
	return mc.api.HasCLIArgument(name)
}

func (mc *MultiClusterPluginAPI) ExecuteCLIArgument(name string, value string) error {
	return mc.api.ExecuteCLIArgument(name, value)
}

func (mc *MultiClusterPluginAPI) GetConfig(key string) any {
	return mc.api.GetConfig(key)
}

func (mc *MultiClusterPluginAPI) SetConfig(key string, value any) {
	mc.api.SetConfig(key, value)
}

func (mc *MultiClusterPluginAPI) GetClient() k8s.Client {
	return mc.api.GetClient()
}

func (mc *MultiClusterPluginAPI) SetClient(client k8s.Client) {
	mc.api.SetClient(client)
}

func (mc *MultiClusterPluginAPI) GetPods(namespace string, selector ...string) ([]k8s.PodInfo, error) {
	return mc.api.GetPods(namespace, selector...)
}

func (mc *MultiClusterPluginAPI) GetServices(namespace string) ([]k8s.ServiceInfo, error) {
	return mc.api.GetServices(namespace)
}

func (mc *MultiClusterPluginAPI) GetDeployments(namespace string) ([]k8s.DeploymentInfo, error) {
	return mc.api.GetDeployments(namespace)
}

func (mc *MultiClusterPluginAPI) GetConfigMaps(namespace string) ([]k8s.Configmap, error) {
	return mc.api.GetConfigMaps(namespace)
}

func (mc *MultiClusterPluginAPI) GetSecrets(namespace string) ([]k8s.SecretInfo, error) {
	return mc.api.GetSecrets(namespace)
}

func (mc *MultiClusterPluginAPI) GetIngresses(namespace string) ([]k8s.IngressInfo, error) {
	return mc.api.GetIngresses(namespace)
}

func (mc *MultiClusterPluginAPI) GetJobs(namespace string) ([]k8s.JobInfo, error) {
	return mc.api.GetJobs(namespace)
}

func (mc *MultiClusterPluginAPI) GetCronJobs(namespace string) ([]k8s.CronJobInfo, error) {
	return mc.api.GetCronJobs(namespace)
}

func (mc *MultiClusterPluginAPI) GetDaemonSets(namespace string) ([]k8s.DaemonSetInfo, error) {
	return mc.api.GetDaemonSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetStatefulSets(namespace string) ([]k8s.StatefulSetInfo, error) {
	return mc.api.GetStatefulSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetReplicaSets(namespace string) ([]k8s.ReplicaSetInfo, error) {
	return mc.api.GetReplicaSets(namespace)
}

func (mc *MultiClusterPluginAPI) GetNodes() ([]k8s.NodeInfo, error) {
	return mc.api.GetNodes()
}

func (mc *MultiClusterPluginAPI) GetNamespaces() ([]string, error) {
	return mc.api.GetNamespaces()
}

func (mc *MultiClusterPluginAPI) GetServiceAccounts(namespace string) ([]k8s.ServiceAccountInfo, error) {
	return mc.api.GetServiceAccounts(namespace)
}

func (mc *MultiClusterPluginAPI) DeletePod(namespace, name string) error {
	return mc.api.DeletePod(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteService(namespace, name string) error {
	return mc.api.DeleteService(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteDeployment(namespace, name string) error {
	return mc.api.DeleteDeployment(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteConfigMap(namespace, name string) error {
	return mc.api.DeleteConfigMap(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteSecret(namespace, name string) error {
	return mc.api.DeleteSecret(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteIngress(namespace, name string) error {
	return mc.api.DeleteIngress(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteJob(namespace, name string) error {
	return mc.api.DeleteJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteCronJob(namespace, name string) error {
	return mc.api.DeleteCronJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteDaemonSet(namespace, name string) error {
	return mc.api.DeleteDaemonSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteStatefulSet(namespace, name string) error {
	return mc.api.DeleteStatefulSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteReplicaSet(namespace, name string) error {
	return mc.api.DeleteReplicaSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DeleteServiceAccount(namespace, name string) error {
	return mc.api.DeleteServiceAccount(namespace, name)
}

func (mc *MultiClusterPluginAPI) GetTabs() ([]TabInfo, error) {
	return mc.api.GetTabs()
}

func (mc *MultiClusterPluginAPI) SetTabs(tabs []TabInfo) error {
	return mc.api.SetTabs(tabs)
}

func (mc *MultiClusterPluginAPI) SetTabSetterCallback(callback func()) {
	mc.api.SetTabSetterCallback(callback)
}

func (mc *MultiClusterPluginAPI) GetBreadcrumbTrail() []string {
	return mc.api.GetBreadcrumbTrail()
}

func (mc *MultiClusterPluginAPI) SetBreadcrumbTrail(breadcrumb []string) {
	mc.api.SetBreadcrumbTrail(breadcrumb)
}

func (mc *MultiClusterPluginAPI) ShowInputDialog(title, placeholder, submitCommand, cancelCommand string) {
	mc.api.ShowInputDialog(title, placeholder, submitCommand, cancelCommand)
}

func (mc *MultiClusterPluginAPI) GetCurrentResourceType() string {
	return mc.api.GetCurrentResourceType()
}

func (mc *MultiClusterPluginAPI) GetHelp(resourceType string) (title, content string) {
	return mc.api.GetHelp(resourceType)
}

func (mc *MultiClusterPluginAPI) RegisterHelp(resourceType string, title, content string) {
	mc.api.RegisterHelp(resourceType, title, content)
}

func (mc *MultiClusterPluginAPI) RegisterEventHandler(event PluginEvent, handler func(data any) error) {
	mc.api.RegisterEventHandler(event, handler)
}

func (mc *MultiClusterPluginAPI) TriggerEvent(event PluginEvent, data any) {
	mc.api.TriggerEvent(event, data)
}

func (mc *MultiClusterPluginAPI) GetCommands() map[string]PluginCommand {
	return mc.api.GetCommands()
}

func (mc *MultiClusterPluginAPI) SetNamespaceCallback(callback func(namespace string)) {
	mc.api.SetNamespaceCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetStatusCallback(callback func(message string)) {
	mc.api.SetStatusCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetBreadcrumbCallback(callback func(breadcrumb []string)) {
	mc.api.SetBreadcrumbCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetShowInputDialogCallback(callback func(title, placeholder, submitCommand, cancelCommand string)) {
	mc.api.SetShowInputDialogCallback(callback)
}

func (mc *MultiClusterPluginAPI) SetTabGetter(getter func() ([]TabInfo, error)) {
	mc.api.SetTabGetter(getter)
}

func (mc *MultiClusterPluginAPI) SetTabSetter(setter func(tabs []TabInfo) error) {
	mc.api.SetTabSetter(setter)
}

func (mc *MultiClusterPluginAPI) GetTabSetter() func(tabs []TabInfo) error {
	return mc.api.GetTabSetter()
}

func (mc *MultiClusterPluginAPI) RegisterResourceHandler(resourceType k8s.ResourceType, handler ResourceHandler) {
	mc.api.RegisterResourceHandler(resourceType, handler)
}

func (mc *MultiClusterPluginAPI) GetSupportedResourceTypes() []k8s.ResourceType {
	return mc.api.GetSupportedResourceTypes()
}

func (mc *MultiClusterPluginAPI) GetResourceHandler(resourceType k8s.ResourceType) (ResourceHandler, bool) {
	return mc.api.GetResourceHandler(resourceType)
}

func (mc *MultiClusterPluginAPI) SetCurrentResourceType(resourceType string) {
	mc.api.SetCurrentResourceType(resourceType)
}

func (mc *MultiClusterPluginAPI) GetBreadcrumbCallback(callback func() []string) {
	mc.api.GetBreadcrumbCallback(callback)
}

func (mc *MultiClusterPluginAPI) DescribePod(namespace, name string) (string, error) {
	return mc.api.DescribePod(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeService(namespace, name string) (string, error) {
	return mc.api.DescribeService(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeDeployment(namespace, name string) (string, error) {
	return mc.api.DescribeDeployment(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeConfigMap(namespace, name string) (string, error) {
	return mc.api.DescribeConfigMap(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeSecret(namespace, name string) (string, error) {
	return mc.api.DescribeSecret(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeIngress(namespace, name string) (string, error) {
	return mc.api.DescribeIngress(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeJob(namespace, name string) (string, error) {
	return mc.api.DescribeJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeCronJob(namespace, name string) (string, error) {
	return mc.api.DescribeCronJob(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeDaemonSet(namespace, name string) (string, error) {
	return mc.api.DescribeDaemonSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeStatefulSet(namespace, name string) (string, error) {
	return mc.api.DescribeStatefulSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeReplicaSet(namespace, name string) (string, error) {
	return mc.api.DescribeReplicaSet(namespace, name)
}

func (mc *MultiClusterPluginAPI) DescribeNode(name string) (string, error) {
	return mc.api.DescribeNode(name)
}

func (mc *MultiClusterPluginAPI) DescribeServiceAccount(namespace, name string) (string, error) {
	return mc.api.DescribeServiceAccount(namespace, name)
}
