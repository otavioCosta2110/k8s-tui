package plugins

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

// Helper functions for styling help content to reduce verbosity
func getHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getSpaceStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getHelpTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HelpTextColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getSectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor))
}

func getBulletStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
}

// helpLine creates a formatted help line with key and value
func helpLine(key, value string) string {
	return getKeyStyle().Render(key) + getSpaceStyle().Render(" ") + getValueStyle().Render(value)
}

// helpSection creates a formatted help section header
func helpSection(title string) string {
	return getSectionStyle().Render(title) + "\n"
}

// helpText creates formatted help text
func helpText(text string) string {
	return getTextStyle().Render(text)
}

// helpItem creates a complete help item with bullet point
func helpItem(bullet, description string) string {
	return getBulletStyle().Render("• ") + getKeyStyle().Render(bullet) + getSpaceStyle().Render(" ") + getValueStyle().Render(description) + "\n"
}

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

type CLIArgumentManager struct {
	arguments map[string]CLIArgument
}

func NewCLIArgumentManager() *CLIArgumentManager {
	return &CLIArgumentManager{
		arguments: make(map[string]CLIArgument),
	}
}

func (cam *CLIArgumentManager) RegisterArgument(name, description string, handler func(value string) error) {
	cam.arguments[name] = CLIArgument{
		Name:        name,
		Description: description,
		Handler:     handler,
	}

}

func (cam *CLIArgumentManager) ExecuteArgument(name string, value string) error {
	if arg, exists := cam.arguments[name]; exists {
		return arg.Handler(value)
	}
	return fmt.Errorf("CLI argument not found: %s", name)
}

func (cam *CLIArgumentManager) GetArguments() map[string]CLIArgument {
	return cam.arguments
}

func (cam *CLIArgumentManager) HasArgument(name string) bool {
	_, exists := cam.arguments[name]
	return exists
}

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
	currentNamespace        string
	currentResourceType     string
	uiManager               *UIManager
	commandManager          *CommandManager
	cliArgumentManager      *CLIArgumentManager
	eventManager            *EventManager
	configManager           *ConfigManager
	resourceRegistry        *ResourceRegistry
	client                  k8s.Client
	tabGetter               func() ([]TabInfo, error)
	tabSetter               func(tabs []TabInfo) error
	setTabSetterCallback    func()
	setNamespaceCallback    func(namespace string)
	setStatusCallback       func(message string)
	setBreadcrumbCallback   func(breadcrumb []string)
	getBreadcrumbCallback   func() []string
	showInputDialogCallback func(title, placeholder, submitCommand, cancelCommand string)
	customHelp              map[string]struct {
		title   string
		content string
	}
	globalManager *GlobalPluginManager // Reference to global manager for per-cluster namespace support

	// Cluster management callbacks
	addClusterTabCallback         func(kubeconfigPath, clusterName, namespace string) error
	setClusterTabsCallback        func(clusters []ClusterTabConfig) error
	getClustersCallback           func() []ClusterInfo
	switchToClusterCallback       func(clusterID string) error
	getTabsForClusterCallback     func(clusterID string) ([]TabInfo, error)
	setTabsForClusterCallback     func(clusterID string, tabs []TabInfo) error
	restoreTabsForClusterCallback func(clusterID string) error
}

// NewPluginAPI creates a new plugin API instance with all managers initialized
// Provides the main interface for plugins to interact with k8s-tui
func NewPluginAPI() *PluginAPIImpl {
	return NewPluginAPIWithGlobalManager(nil)
}

// NewPluginAPIWithGlobalManager creates a new plugin API instance with optional global manager
func NewPluginAPIWithGlobalManager(globalManager *GlobalPluginManager) *PluginAPIImpl {
	return &PluginAPIImpl{
		currentNamespace:   "default",
		uiManager:          NewUIManager(),
		commandManager:     NewCommandManager(),
		cliArgumentManager: NewCLIArgumentManager(),
		eventManager:       NewEventManager(),
		configManager:      NewConfigManager(),
		resourceRegistry:   NewResourceRegistry(),
		globalManager:      globalManager,
		customHelp: make(map[string]struct {
			title   string
			content string
		}),
	}
}

// GetCurrentNamespace returns the currently active Kubernetes namespace
func (api *PluginAPIImpl) GetCurrentNamespace() string {
	return api.currentNamespace
}

// SetCurrentNamespace changes the active Kubernetes namespace and triggers events
func (api *PluginAPIImpl) SetCurrentNamespace(namespace string) {
	logger.Info(fmt.Sprintf("DEBUG: SetCurrentNamespace called with: %s", namespace))
	api.currentNamespace = namespace
	// Note: Cluster namespace is managed by GlobalPluginManager.SwitchToCluster()
	// to avoid deadlocks. Don't call back into globalManager here.

	if api.setNamespaceCallback != nil {
		logger.Info("DEBUG: Calling setNamespaceCallback")
		api.setNamespaceCallback(namespace)
	} else {
		logger.Info("DEBUG: setNamespaceCallback is nil")
	}
	api.eventManager.TriggerEvent(EventNamespaceChanged, namespace)
}

// SetStatusMessage displays a status message in the TUI interface
func (api *PluginAPIImpl) SetStatusMessage(message string) {
	logger.Info(fmt.Sprintf("📢 Plugin Status: %s", message))
	if api.setStatusCallback != nil {
		api.setStatusCallback(message)
	}
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

// RegisterCommand registers a new command that can be called from plugins
func (api *PluginAPIImpl) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	api.commandManager.RegisterCommand(name, description, handler)
}

func (api *PluginAPIImpl) ExecuteCommand(name string, args []string) (string, error) {
	return api.commandManager.ExecuteCommand(name, args)
}

func (api *PluginAPIImpl) RegisterCLIArgument(name, description string, handler func(value string) error) {
	api.cliArgumentManager.RegisterArgument(name, description, handler)
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

func (api *PluginAPIImpl) GetCLIArguments() map[string]CLIArgument {
	return api.cliArgumentManager.GetArguments()
}

func (api *PluginAPIImpl) HasCLIArgument(name string) bool {
	return api.cliArgumentManager.HasArgument(name)
}

func (api *PluginAPIImpl) ExecuteCLIArgument(name string, value string) error {
	return api.cliArgumentManager.ExecuteArgument(name, value)
}

func (api *PluginAPIImpl) GetClient() k8s.Client {
	return api.client
}

func (api *PluginAPIImpl) SetClient(client k8s.Client) {
	api.client = client
}

// GetPods retrieves pod information from Kubernetes cluster with optional label selector
func (api *PluginAPIImpl) GetPods(namespace string, selector ...string) ([]k8s.PodInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}

	selectorStr := ""
	if len(selector) > 0 && selector[0] != "" {
		selectorStr = selector[0]
	}

	logger.Debug(fmt.Sprintf("PluginAPI GetPods called with namespace=%s, selector=%s", namespace, selectorStr))

	handler, exists := api.resourceRegistry.GetHandler(k8s.ResourceTypePod)
	logger.Debug(fmt.Sprintf("Pod handler exists: %v, handler: %v", exists, handler))

	if !exists || handler == nil || selectorStr != "" {
		logger.Debug(fmt.Sprintf("Using direct k8s client with selector: %s", selectorStr))
		return k8s.FetchPods(api.client, namespace, selectorStr)
	}

	logger.Debug("Using custom handler (no selector provided)")
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypePod, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.PodInfo), nil
}

func (api *PluginAPIImpl) GetServices(namespace string) ([]k8s.ServiceInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeService, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.ServiceInfo), nil
}

func (api *PluginAPIImpl) GetDeployments(namespace string) ([]k8s.DeploymentInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeDeployment, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.DeploymentInfo), nil
}

func (api *PluginAPIImpl) GetConfigMaps(namespace string) ([]k8s.Configmap, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeConfigMap, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.Configmap), nil
}

func (api *PluginAPIImpl) GetSecrets(namespace string) ([]k8s.SecretInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeSecret, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.SecretInfo), nil
}

func (api *PluginAPIImpl) GetIngresses(namespace string) ([]k8s.IngressInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeIngress, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.IngressInfo), nil
}

func (api *PluginAPIImpl) GetJobs(namespace string) ([]k8s.JobInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeJob, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.JobInfo), nil
}

func (api *PluginAPIImpl) GetCronJobs(namespace string) ([]k8s.CronJobInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeCronJob, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.CronJobInfo), nil
}

func (api *PluginAPIImpl) GetDaemonSets(namespace string) ([]k8s.DaemonSetInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeDaemonSet, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.DaemonSetInfo), nil
}

func (api *PluginAPIImpl) GetStatefulSets(namespace string) ([]k8s.StatefulSetInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeStatefulSet, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.StatefulSetInfo), nil
}

func (api *PluginAPIImpl) GetReplicaSets(namespace string) ([]k8s.ReplicaSetInfo, error) {
	if namespace == "" {
		namespace = api.currentNamespace
	}
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
	if namespace == "" {
		namespace = api.currentNamespace
	}
	result, err := api.resourceRegistry.GetResource(api.client, k8s.ResourceTypeServiceAccount, namespace)
	if err != nil {
		return nil, err
	}
	return result.([]k8s.ServiceAccountInfo), nil
}

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

// Restart methods for supported resources

func (api *PluginAPIImpl) RestartPod(namespace, name string) error {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	pod := k8s.NewPodInfo(name, namespace, api.client)
	return pod.Restart()
}

func (api *PluginAPIImpl) RestartDeployment(namespace, name string) error {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	deployment := k8s.NewDeploymentInfo(name, namespace, api.client)
	return deployment.Restart()
}

func (api *PluginAPIImpl) RestartReplicaSet(namespace, name string) error {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	replicaset := k8s.NewReplicaSetInfo(name, namespace, api.client)
	return replicaset.Restart()
}

func (api *PluginAPIImpl) RestartStatefulSet(namespace, name string) error {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	statefulset := k8s.NewStatefulSetInfo(name, namespace, api.client)
	return statefulset.Restart()
}

func (api *PluginAPIImpl) RestartDaemonSet(namespace, name string) error {
	if namespace == "" {
		namespace = api.currentNamespace
	}
	daemonset := k8s.NewDaemonSetInfo(name, namespace, api.client)
	return daemonset.Restart()
}

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

func (api *PluginAPIImpl) RegisterResourceHandler(resourceType k8s.ResourceType, handler ResourceHandler) {
	api.resourceRegistry.RegisterHandler(resourceType, handler)
	logger.PluginDebug("api", fmt.Sprintf("Registered custom handler for resource type: %s", resourceType))
}

func (api *PluginAPIImpl) GetSupportedResourceTypes() []k8s.ResourceType {
	return api.resourceRegistry.GetSupportedTypes()
}

func (api *PluginAPIImpl) GetResourceHandler(resourceType k8s.ResourceType) (ResourceHandler, bool) {
	return api.resourceRegistry.GetHandler(resourceType)
}

func (api *PluginAPIImpl) GetTabs() ([]TabInfo, error) {
	if api.tabGetter != nil {
		return api.tabGetter()
	}
	return nil, fmt.Errorf("tab getter not set")
}

func (api *PluginAPIImpl) SetTabs(tabs []TabInfo) error {
	if api.tabSetter != nil {
		err := api.tabSetter(tabs)
		if err != nil {
			return err
		}
		if api.setTabSetterCallback != nil {
			api.setTabSetterCallback()
		}
		return nil
	}
	return fmt.Errorf("tab setter not set")
}

func (api *PluginAPIImpl) SetTabGetter(getter func() ([]TabInfo, error)) {
	api.tabGetter = getter
}

func (api *PluginAPIImpl) SetTabSetter(setter func(tabs []TabInfo) error) {
	api.tabSetter = setter
}

func (api *PluginAPIImpl) GetTabSetter() func(tabs []TabInfo) error {
	return api.tabSetter
}

func (api *PluginAPIImpl) SetNamespaceCallback(callback func(namespace string)) {
	api.setNamespaceCallback = callback
}

func (api *PluginAPIImpl) SetStatusCallback(callback func(message string)) {
	api.setStatusCallback = callback
}

func (api *PluginAPIImpl) SetTabSetterCallback(callback func()) {
	api.setTabSetterCallback = callback
}

func (api *PluginAPIImpl) SetBreadcrumbCallback(callback func(breadcrumb []string)) {
	api.setBreadcrumbCallback = callback
}

func (api *PluginAPIImpl) GetBreadcrumbCallback(callback func() []string) {
	api.getBreadcrumbCallback = callback
}

func (api *PluginAPIImpl) GetBreadcrumbTrail() []string {
	if api.getBreadcrumbCallback != nil {
		return api.getBreadcrumbCallback()
	}
	return []string{}
}

func (api *PluginAPIImpl) SetBreadcrumbTrail(breadcrumb []string) {
	if api.setBreadcrumbCallback != nil {
		api.setBreadcrumbCallback(breadcrumb)
	}
}

func (api *PluginAPIImpl) ShowInputDialog(title, placeholder, submitCommand, cancelCommand string) {
	if api.showInputDialogCallback != nil {
		api.showInputDialogCallback(title, placeholder, submitCommand, cancelCommand)
	}
}

func (api *PluginAPIImpl) SetShowInputDialogCallback(callback func(title, placeholder, submitCommand, cancelCommand string)) {
	api.showInputDialogCallback = callback
}

func (api *PluginAPIImpl) SetAddClusterTabCallback(callback func(kubeconfigPath, clusterName, namespace string) error) {
	api.addClusterTabCallback = callback
}

func (api *PluginAPIImpl) SetGetClustersCallback(callback func() []ClusterInfo) {
	api.getClustersCallback = callback
}

func (api *PluginAPIImpl) SetSwitchToClusterCallback(callback func(clusterID string) error) {
	api.switchToClusterCallback = callback
}

func (api *PluginAPIImpl) SetClusterTabsCallback(callback func(clusters []ClusterTabConfig) error) {
	api.setClusterTabsCallback = callback
}

func (api *PluginAPIImpl) SetGetTabsForClusterCallback(callback func(clusterID string) ([]TabInfo, error)) {
	api.getTabsForClusterCallback = callback
}

func (api *PluginAPIImpl) SetRestoreTabsForClusterCallback(callback func(clusterID string) error) {
	// This callback will be set on the global manager if available
	if api.globalManager != nil {
		api.globalManager.restoreTabsForClusterCallback = callback
	}
}

func (api *PluginAPIImpl) SetSetTabsForClusterCallback(callback func(clusterID string, tabs []TabInfo) error) {
	// This callback will be set on the global manager if available
	if api.globalManager != nil {
		api.globalManager.setTabsForClusterCallback = callback
	}
}

func (api *PluginAPIImpl) GetCurrentResourceType() string {
	return api.currentResourceType
}

func (api *PluginAPIImpl) SetCurrentResourceType(resourceType string) {
	api.currentResourceType = resourceType
}

func (api *PluginAPIImpl) GetHelp(resourceType string) (title, content string) {
	if customHelp, exists := api.customHelp[resourceType]; exists {
		return customHelp.title, customHelp.content
	}

	helpMap := map[string]struct {
		title   string
		content string
	}{
		"Pods": {
			title: "Pods Help",
			content: helpText("Pods are the smallest deployable units in Kubernetes.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate pods") +
				helpItem("enter:", "View pod logs") +
				helpItem("v:", "View pod details") +
				helpItem("e:", "Execute a command in the pod") +
				helpItem("E:", "View pod events") +
				helpItem("t:", "View resource usage") +
				helpItem("R:", "Restart pod") +
				helpItem("d:", "Delete selected pods") +
				helpItem("n:", "Create new pod") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search pods") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("Pod Status:") +
				helpItem("Running:", "Pod is running successfully") +
				helpItem("Pending:", "Pod is being scheduled") +
				helpItem("Failed:", "Pod has failed") +
				helpItem("Succeeded:", "Pod completed successfully") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("View logs:", "Enter on a pod to see logs") +
				helpItem("View details:", "Press 'v' to see pod details") +
				helpItem("View events:", "Press 'E' to see pod events") +
				helpItem("View resource usage:", "Press 't' to see CPU/memory usage") +
				helpItem("Restart pod:", "Press 'R' to restart the pod (only works for managed pods)") +
				helpItem("Delete pod:", "Select with space, then press 'd'") +
				helpItem("Create pod:", "Press 'n' to open create form") +
				helpItem("Refresh:", "Press 'r' to update the list") + "\n\n" +

				helpSection("Events View Key Bindings:") +
				helpItem("r:", "Refresh events") +
				helpItem("esc:", "Go back"),
		},
		"Deployments": {
			title: "Deployments Help",
			content: helpText("Deployments manage the deployment and scaling of applications.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate deployments") +
				helpItem("enter:", "View deployment pods") +
				helpItem("v:", "View deployment details") +
				helpItem("L:", "View deployment logs") +
				helpItem("E:", "View deployment events") +
				helpItem("R:", "Restart deployment") +
				helpItem("d:", "Delete selected deployments") +
				helpItem("n:", "Create new deployment") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search deployments") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("Deployment Status:") +
				helpItem("Ready:", "Shows ready/desired replicas") +
				helpItem("Updated:", "Shows updated replicas") +
				helpItem("Available:", "Shows available replicas") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("Scale deployment:", "Enter to view details and scale") +
				helpItem("Update image:", "Use deployment details view") +
				helpItem("View pods:", "See associated pods in details") +
				helpItem("View logs:", "Press 'L' to see deployment logs") +
				helpItem("View events:", "Press 'E' to see deployment events") +
				helpItem("Restart deployment:", "Press 'R' to trigger a rollout restart") +
				helpItem("Create deployment:", "Press 'n' to open create form") + "\n\n" +

				helpSection("Events View Key Bindings:") +
				helpItem("g:", "Go to top") +
				helpItem("G:", "Go to bottom") +
				helpItem("r:", "Refresh events") +
				helpItem("esc:", "Go back"),
		},
		"Services": {
			title: "Services Help",
			content: helpText("Services expose applications running on pods.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate services") +
				helpItem("enter:", "View service details") +
				helpItem("E:", "View service events") +
				helpItem("d:", "Delete selected services") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search services") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("Service Types:") +
				helpItem("ClusterIP:", "Internal cluster access") +
				helpItem("NodePort:", "External access via node port") +
				helpItem("LoadBalancer:", "Cloud load balancer") +
				helpItem("ExternalName:", "DNS alias") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("View endpoints:", "See pods backing the service") +
				helpItem("Check connectivity:", "Use service details") + "\n\n" +

				helpSection("Events View Key Bindings:") +
				helpItem("g:", "Go to top") +
				helpItem("G:", "Go to bottom") +
				helpItem("r:", "Refresh events") +
				helpItem("esc:", "Go back"),
		},
		"Ingresses": {
			title: "Ingresses Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Ingresses manage external access to services.

Key Bindings:
↑/↓/j/k: Navigate ingresses
enter: View ingress details
d: Delete selected ingresses
r: Refresh
/: Search ingresses
esc: Go back

Common Actions:
View rules: See routing rules
Check TLS: View SSL certificates
Test routing: Verify external access`),
		},
		"ConfigMaps": {
			title: "ConfigMaps Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`ConfigMaps store configuration data.

Key Bindings:
↑/↓/j/k: Navigate configmaps
enter: View configmap details
d: Delete selected configmaps
r: Refresh
/: Search configmaps
esc: Go back

Common Actions:
View data: See configuration key-value pairs
Edit values: Modify configuration data
Check usage: See which pods use this config`),
		},
		"Secrets": {
			title: "Secrets Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Secrets store sensitive information.

Key Bindings:
↑/↓/j/k: Navigate secrets
enter: View secret details
d: Delete selected secrets
r: Refresh
/: Search secrets
esc: Go back

Secret Types:
Opaque: Generic secret data
TLS: Certificate/key pairs
Docker: Docker registry credentials

Common Actions:
View data: See secret contents (use caution)
Rotate secrets: Update sensitive data
Check usage: See which pods reference this secret`),
		},
		"Jobs": {
			title: "Jobs Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Jobs create and manage batch processing tasks.

Key Bindings:
↑/↓/j/k: Navigate jobs
enter: View job details
d: Delete selected jobs
r: Refresh
/: Search jobs
esc: Go back

Job Status:
Complete: Job finished successfully
Failed: Job failed
Active: Job is running

Common Actions:
View pods: See job execution pods
Check logs: Examine job output
Retry failed jobs: Delete and recreate`),
		},
		"CronJobs": {
			title: "CronJobs Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`CronJobs run jobs on a schedule.

Key Bindings:
↑/↓/j/k: Navigate cronjobs
enter: View cronjob details
d: Delete selected cronjobs
r: Refresh
/: Search cronjobs
esc: Go back

Schedule Format:
Uses standard cron syntax
Example: "0 0 * * *" = daily at midnight

Common Actions:
View schedule: See execution schedule
Check history: View past job executions
Manual trigger: Run job immediately`),
		},
		"Nodes": {
			title: "Nodes Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Nodes are the worker machines in the cluster.

Key Bindings:
↑/↓/j/k: Navigate nodes
enter: View node details
r: Refresh
/: Search nodes
esc: Go back

Node Status:
Ready: Node is healthy and schedulable
NotReady: Node has issues
SchedulingDisabled: Node won't accept new pods

Common Actions:
View capacity: See CPU/memory resources
Check conditions: View node health status
View pods: See pods running on node`),
		},
		"ReplicaSets": {
			title: "ReplicaSets Help",
			content: helpText("ReplicaSets ensure a specified number of pod replicas are running.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate replicasets") +
				helpItem("enter:", "View replicaset details") +
				helpItem("R:", "Restart replicaset") +
				helpItem("d:", "Delete selected replicasets") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search replicasets") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("ReplicaSet Status:") +
				helpItem("Desired:", "Number of desired pods") +
				helpItem("Current:", "Number of current pods") +
				helpItem("Ready:", "Number of ready pods") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("View pods:", "See pods managed by this replicaset") +
				helpItem("Check owner:", "See which deployment owns this replicaset") +
				helpItem("Restart replicaset:", "Press 'R' to trigger a rollout restart") +
				helpItem("Manual scaling:", "Adjust replica count"),
		},
		"StatefulSets": {
			title: "StatefulSets Help",
			content: helpText("StatefulSets manage stateful applications with persistent storage.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate statefulsets") +
				helpItem("enter:", "View statefulset details") +
				helpItem("R:", "Restart statefulset") +
				helpItem("d:", "Delete selected statefulsets") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search statefulsets") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("StatefulSet Status:") +
				helpItem("Ready:", "Shows ready/desired replicas") +
				helpItem("Stable identity:", "Each pod has a stable identity and storage") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("Scale statefulset:", "Change replica count") +
				helpItem("View persistent volumes:", "See attached storage") +
				helpItem("Check pod ordering:", "StatefulSets maintain pod identity") +
				helpItem("Restart statefulset:", "Press 'R' to trigger a rollout restart"),
		},
		"DaemonSets": {
			title: "DaemonSets Help",
			content: helpText("DaemonSets ensure that all (or some) nodes run a copy of a pod.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate daemonsets") +
				helpItem("enter:", "View daemonset details") +
				helpItem("R:", "Restart daemonset") +
				helpItem("d:", "Delete selected daemonsets") +
				helpItem("r:", "Refresh") +
				helpItem("/:", "Search daemonsets") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("DaemonSet Status:") +
				helpItem("Desired:", "Number of desired pods") +
				helpItem("Current:", "Number of current pods") +
				helpItem("Ready:", "Number of ready pods") +
				helpItem("Available:", "Number of available pods") + "\n\n" +

				helpSection("Common Actions:") +
				helpItem("View pods:", "See pods running on each node") +
				helpItem("Check node selectors:", "See which nodes run pods") +
				helpItem("Restart daemonset:", "Press 'R' to trigger a rollout restart") +
				helpItem("Update daemonset:", "Modify pod template"),
		},
		"ResourceList": {
			title: "Resource List Help",
			content: helpText("Resource List shows available Kubernetes resources.") + "\n\n" +
				helpSection("Key Bindings:") +
				helpItem("↑/↓/j/k:", "Navigate resources") +
				helpItem("enter:", "Select resource type") +
				helpItem("/:", "Search resources") +
				helpItem("esc:", "Go back") + "\n\n" +

				helpSection("Resource Categories:") +
				helpItem("Workloads:", "Pods, Deployments, Jobs, etc.") +
				helpItem("Networking:", "Services, Ingresses") +
				helpItem("Configuration:", "ConfigMaps, Secrets") +
				helpItem("Infrastructure:", "Nodes") + "\n\n" +

				helpSection("Quick Navigation:") +
				helpItem("p:", "Pods") +
				helpItem("d:", "Deployments") +
				helpItem("s:", "Services") +
				helpItem("i:", "Ingresses") +
				helpItem("c:", "ConfigMaps") +
				helpItem("e:", "Secrets") +
				helpItem("n:", "Nodes") +
				helpItem("j:", "Jobs") +
				helpItem("k:", "CronJobs") +
				helpItem("m:", "DaemonSets") +
				helpItem("t:", "StatefulSets") +
				helpItem("r:", "ReplicaSets") +
				helpItem("a:", "ServiceAccounts"),
		},
	}

	if help, exists := helpMap[resourceType]; exists {
		return help.title, help.content
	}

	return "Help", helpText(fmt.Sprintf("Help for %s", resourceType)) + "\n\n" +
		helpSection("Key Bindings:") +
		helpItem("↑/↓/j/k:", "Navigate items") +
		helpItem("enter:", "View details") +
		helpItem("d:", "Delete selected items (if supported)") +
		helpItem("r:", "Refresh") +
		helpItem("/:", "Search") +
		helpItem("esc:", "Go back") + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HelpTextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("For more specific help, check the resource documentation.")
}

func (api *PluginAPIImpl) RegisterHelp(resourceType string, title, content string) {
	api.customHelp[resourceType] = struct {
		title   string
		content string
	}{
		title:   title,
		content: content,
	}
	logger.PluginDebug("api", fmt.Sprintf("Registered custom help for resource type: %s", resourceType))
}

// AddClusterTab creates a new cluster tab with the specified kubeconfig, name, and namespace
func (api *PluginAPIImpl) AddClusterTab(kubeconfigPath, clusterName, namespace string) error {
	if api.addClusterTabCallback == nil {
		return fmt.Errorf("add cluster tab callback not set")
	}

	logger.Info(fmt.Sprintf("Plugin API: Adding cluster tab %s with kubeconfig %s and namespace %s", clusterName, kubeconfigPath, namespace))
	return api.addClusterTabCallback(kubeconfigPath, clusterName, namespace)
}

// SetClusterTabs replaces all cluster tabs with the specified clusters
func (api *PluginAPIImpl) SetClusterTabs(clusters []ClusterTabConfig) error {
	if api.setClusterTabsCallback == nil {
		return fmt.Errorf("set cluster tabs callback not set")
	}

	logger.Info(fmt.Sprintf("Plugin API: Setting %d cluster tabs", len(clusters)))
	return api.setClusterTabsCallback(clusters)
}

// GetClusters returns information about all available clusters
func (api *PluginAPIImpl) GetClusters() []ClusterInfo {
	if api.getClustersCallback == nil {
		logger.Warn("Get clusters callback not set, returning empty slice")
		return []ClusterInfo{}
	}

	clusters := api.getClustersCallback()
	logger.Debug(fmt.Sprintf("Plugin API: Retrieved %d clusters", len(clusters)))
	return clusters
}

// SwitchToCluster switches to the specified cluster by ID
func (api *PluginAPIImpl) SwitchToCluster(clusterID string) error {
	if api.switchToClusterCallback == nil {
		return fmt.Errorf("switch to cluster callback not set")
	}

	logger.Info(fmt.Sprintf("Plugin API: Switching to cluster %s", clusterID))
	return api.switchToClusterCallback(clusterID)
}
