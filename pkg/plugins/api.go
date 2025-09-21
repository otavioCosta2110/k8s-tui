package plugins

import (
	"fmt"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

// UIManager handles UI-related operations
type UIManager struct {
	headerComponents []UIInjectionPoint
	footerComponents []UIInjectionPoint
}

func NewUIManager() *UIManager {
	return &UIManager{
		headerComponents: make([]UIInjectionPoint, 0),
		footerComponents: make([]UIInjectionPoint, 0),
	}
}

func (ui *UIManager) AddHeaderComponent(component UIInjectionPoint) {
	ui.headerComponents = append(ui.headerComponents, component)
	logger.PluginDebug("ui", fmt.Sprintf("Added header component: %s", component.Component.Config["content"]))
}

func (ui *UIManager) AddFooterComponent(component UIInjectionPoint) {
	ui.footerComponents = append(ui.footerComponents, component)
	logger.PluginDebug("ui", fmt.Sprintf("Added footer component: %s", component.Component.Config["content"]))
}

func (ui *UIManager) GetHeaderComponents() []UIInjectionPoint {
	return ui.headerComponents
}

func (ui *UIManager) GetFooterComponents() []UIInjectionPoint {
	return ui.footerComponents
}

// CommandManager handles command registration and execution
type CommandManager struct {
	commands map[string]PluginCommand
}

func NewCommandManager() *CommandManager {
	return &CommandManager{
		commands: make(map[string]PluginCommand),
	}
}

func (cm *CommandManager) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	cm.commands[name] = PluginCommand{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	logger.PluginDebug("command", fmt.Sprintf("Registered command: %s - %s", name, description))
}

func (cm *CommandManager) ExecuteCommand(name string, args []string) (string, error) {
	if cmd, exists := cm.commands[name]; exists {
		return cmd.Handler(args)
	}
	return "", fmt.Errorf("command not found: %s", name)
}

func (cm *CommandManager) GetCommands() map[string]PluginCommand {
	return cm.commands
}

// EventManager handles event registration and triggering
type EventManager struct {
	eventHandlers map[PluginEvent][]func(data interface{}) error
}

func NewEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[PluginEvent][]func(data interface{}) error),
	}
}

func (em *EventManager) RegisterEventHandler(event PluginEvent, handler func(data interface{}) error) {
	em.eventHandlers[event] = append(em.eventHandlers[event], handler)
}

func (em *EventManager) TriggerEvent(event PluginEvent, data interface{}) {
	if handlers, exists := em.eventHandlers[event]; exists {
		for _, handler := range handlers {
			if err := handler(data); err != nil {
				logger.PluginError("event", fmt.Sprintf("Error in event handler for %s: %v", event, err))
			}
		}
	}
}

// ConfigManager handles configuration operations
type ConfigManager struct {
	config map[string]interface{}
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: make(map[string]interface{}),
	}
}

func (cm *ConfigManager) GetConfig(key string) interface{} {
	return cm.config[key]
}

func (cm *ConfigManager) SetConfig(key string, value interface{}) {
	cm.config[key] = value
	logger.PluginDebug("config", fmt.Sprintf("Set config %s = %v", key, value))
}

type PluginAPIImpl struct {
	currentNamespace string
	uiManager        *UIManager
	commandManager   *CommandManager
	eventManager     *EventManager
	configManager    *ConfigManager
	resourceRegistry *ResourceRegistry
	client           k8s.Client
}

func NewPluginAPI() *PluginAPIImpl {
	return &PluginAPIImpl{
		currentNamespace: "default",
		uiManager:        NewUIManager(),
		commandManager:   NewCommandManager(),
		eventManager:     NewEventManager(),
		configManager:    NewConfigManager(),
		resourceRegistry: NewResourceRegistry(),
	}
}

func (api *PluginAPIImpl) GetCurrentNamespace() string {
	return api.currentNamespace
}

func (api *PluginAPIImpl) SetCurrentNamespace(namespace string) {
	api.currentNamespace = namespace
	api.eventManager.TriggerEvent(EventNamespaceChanged, namespace)
}

func (api *PluginAPIImpl) SetStatusMessage(message string) {
	logger.Info(fmt.Sprintf("📢 Plugin Status: %s", message))
	// In a real implementation, this would update the UI status bar
}

func (api *PluginAPIImpl) AddHeaderComponent(component UIInjectionPoint) {
	api.uiManager.AddHeaderComponent(component)
}

func (api *PluginAPIImpl) AddFooterComponent(component UIInjectionPoint) {
	api.uiManager.AddFooterComponent(component)
}

func (api *PluginAPIImpl) GetHeaderComponents() []UIInjectionPoint {
	return api.uiManager.GetHeaderComponents()
}

func (api *PluginAPIImpl) GetFooterComponents() []UIInjectionPoint {
	return api.uiManager.GetFooterComponents()
}

func (api *PluginAPIImpl) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	api.commandManager.RegisterCommand(name, description, handler)
}

func (api *PluginAPIImpl) ExecuteCommand(name string, args []string) (string, error) {
	return api.commandManager.ExecuteCommand(name, args)
}

func (api *PluginAPIImpl) GetConfig(key string) any {
	return api.configManager.GetConfig(key)
}

func (api *PluginAPIImpl) SetConfig(key string, value any) {
	api.configManager.SetConfig(key, value)
}

func (api *PluginAPIImpl) RegisterEventHandler(event PluginEvent, handler func(data any) error) {
	api.eventManager.RegisterEventHandler(event, handler)
}

func (api *PluginAPIImpl) TriggerEvent(event PluginEvent, data any) {
	api.eventManager.TriggerEvent(event, data)
}

func (api *PluginAPIImpl) GetCommands() map[string]PluginCommand {
	return api.commandManager.GetCommands()
}

func (api *PluginAPIImpl) GetClient() k8s.Client {
	return api.client
}

func (api *PluginAPIImpl) SetClient(client k8s.Client) {
	api.client = client
}

// Kubernetes resource API methods using the resource registry

func (api *PluginAPIImpl) GetPods(namespace string, selector ...string) ([]k8s.PodInfo, error) {
	selectorStr := ""
	if len(selector) > 0 && selector[0] != "" {
		selectorStr = selector[0]
	}

	logger.Debug(fmt.Sprintf("PluginAPI GetPods called with namespace=%s, selector=%s", namespace, selectorStr))

	// Check if a custom handler is registered for pods
	handler, exists := api.resourceRegistry.GetHandler(k8s.ResourceTypePod)
	logger.Debug(fmt.Sprintf("Pod handler exists: %v, handler: %v", exists, handler))

	// If no custom handler is registered, or if a selector is provided (custom handlers don't support selectors),
	// use the direct k8s client call with selector
	if !exists || handler == nil || selectorStr != "" {
		logger.Debug(fmt.Sprintf("Using direct k8s client with selector: %s", selectorStr))
		return k8s.FetchPods(api.client, namespace, selectorStr)
	}

	// Use custom handler if registered and no selector is needed
	logger.Debug("Using custom handler (no selector provided)")
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypePod, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.PodInfo), nil
}

func (api *PluginAPIImpl) GetServices(namespace string) ([]k8s.ServiceInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeService, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.ServiceInfo), nil
}

func (api *PluginAPIImpl) GetDeployments(namespace string) ([]k8s.DeploymentInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeDeployment, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.DeploymentInfo), nil
}

func (api *PluginAPIImpl) GetConfigMaps(namespace string) ([]k8s.Configmap, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeConfigMap, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.Configmap), nil
}

func (api *PluginAPIImpl) GetSecrets(namespace string) ([]k8s.SecretInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeSecret, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.SecretInfo), nil
}

func (api *PluginAPIImpl) GetIngresses(namespace string) ([]k8s.IngressInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeIngress, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.IngressInfo), nil
}

func (api *PluginAPIImpl) GetJobs(namespace string) ([]k8s.JobInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeJob, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.JobInfo), nil
}

func (api *PluginAPIImpl) GetCronJobs(namespace string) ([]k8s.CronJobInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeCronJob, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.CronJobInfo), nil
}

func (api *PluginAPIImpl) GetDaemonSets(namespace string) ([]k8s.DaemonSetInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeDaemonSet, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.DaemonSetInfo), nil
}

func (api *PluginAPIImpl) GetStatefulSets(namespace string) ([]k8s.StatefulSetInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeStatefulSet, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.StatefulSetInfo), nil
}

func (api *PluginAPIImpl) GetReplicaSets(namespace string) ([]k8s.ReplicaSetInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeReplicaSet, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.ReplicaSetInfo), nil
}

func (api *PluginAPIImpl) GetNodes() ([]k8s.NodeInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeNode, "")
	if err != nil {
		return nil, err
	}
	return result.([]k8s.NodeInfo), nil
}

func (api *PluginAPIImpl) GetNamespaces() ([]string, error) {
	return k8s.FetchNamespaces(api.client)
}

func (api *PluginAPIImpl) GetServiceAccounts(namespace string) ([]k8s.ServiceAccountInfo, error) {
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeServiceAccount, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.ServiceAccountInfo), nil
}

// Delete methods using resource registry

func (api *PluginAPIImpl) DeletePod(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypePod, namespace, name)
}

func (api *PluginAPIImpl) DeleteService(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeService, namespace, name)
}

func (api *PluginAPIImpl) DeleteDeployment(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeDeployment, namespace, name)
}

func (api *PluginAPIImpl) DeleteConfigMap(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeConfigMap, namespace, name)
}

func (api *PluginAPIImpl) DeleteSecret(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeSecret, namespace, name)
}

func (api *PluginAPIImpl) DeleteIngress(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeIngress, namespace, name)
}

func (api *PluginAPIImpl) DeleteJob(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeJob, namespace, name)
}

func (api *PluginAPIImpl) DeleteCronJob(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeCronJob, namespace, name)
}

func (api *PluginAPIImpl) DeleteDaemonSet(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeDaemonSet, namespace, name)
}

func (api *PluginAPIImpl) DeleteStatefulSet(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeStatefulSet, namespace, name)
}

func (api *PluginAPIImpl) DeleteReplicaSet(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeReplicaSet, namespace, name)
}

func (api *PluginAPIImpl) DeleteServiceAccount(namespace, name string) error {
	return api.resourceRegistry.DeleteResource(api.client, k8s.ResourceTypeServiceAccount, namespace, name)
}

// Describe methods for individual resources using resource registry

func (api *PluginAPIImpl) DescribePod(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypePod, namespace, name)
}

func (api *PluginAPIImpl) DescribeService(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeService, namespace, name)
}

func (api *PluginAPIImpl) DescribeDeployment(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeDeployment, namespace, name)
}

func (api *PluginAPIImpl) DescribeConfigMap(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeConfigMap, namespace, name)
}

func (api *PluginAPIImpl) DescribeSecret(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeSecret, namespace, name)
}

func (api *PluginAPIImpl) DescribeIngress(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeIngress, namespace, name)
}

func (api *PluginAPIImpl) DescribeJob(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeJob, namespace, name)
}

func (api *PluginAPIImpl) DescribeCronJob(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeCronJob, namespace, name)
}

func (api *PluginAPIImpl) DescribeDaemonSet(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeDaemonSet, namespace, name)
}

func (api *PluginAPIImpl) DescribeStatefulSet(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeStatefulSet, namespace, name)
}

func (api *PluginAPIImpl) DescribeReplicaSet(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeReplicaSet, namespace, name)
}

func (api *PluginAPIImpl) DescribeNode(name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeNode, "", name)
}

func (api *PluginAPIImpl) DescribeServiceAccount(namespace, name string) (string, error) {
	return api.resourceRegistry.DescribeResource(api.client, k8s.ResourceTypeServiceAccount, namespace, name)
}

// Plugin extensibility methods

// RegisterResourceHandler allows plugins to register custom resource handlers
func (api *PluginAPIImpl) RegisterResourceHandler(resourceType k8s.ResourceType, handler ResourceHandler) {
	api.resourceRegistry.RegisterHandler(resourceType, handler)
	logger.PluginDebug("api", fmt.Sprintf("Registered custom handler for resource type: %s", resourceType))
}

// GetSupportedResourceTypes returns all currently supported resource types
func (api *PluginAPIImpl) GetSupportedResourceTypes() []k8s.ResourceType {
	return api.resourceRegistry.GetSupportedTypes()
}

// GetResourceHandler returns the handler for a specific resource type
func (api *PluginAPIImpl) GetResourceHandler(resourceType k8s.ResourceType) (ResourceHandler, bool) {
	return api.resourceRegistry.GetHandler(resourceType)
}
