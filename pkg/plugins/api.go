package plugins

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

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
	showInputDialogCallback func(title, placeholder, submitCommand, cancelCommand string)
	customHelp              map[string]struct {
		title   string
		content string
	}
}

func NewPluginAPI() *PluginAPIImpl {
	return &PluginAPIImpl{
		currentNamespace:   "default",
		uiManager:          NewUIManager(),
		commandManager:     NewCommandManager(),
		cliArgumentManager: NewCLIArgumentManager(),
		eventManager:       NewEventManager(),
		configManager:      NewConfigManager(),
		resourceRegistry:   NewResourceRegistry(),
		customHelp: make(map[string]struct {
			title   string
			content string
		}),
	}
}

func (api *PluginAPIImpl) GetCurrentNamespace() string {
	return api.currentNamespace
}

func (api *PluginAPIImpl) SetCurrentNamespace(namespace string) {
	logger.Info(fmt.Sprintf("DEBUG: SetCurrentNamespace called with: %s", namespace))
	api.currentNamespace = namespace
	if api.setNamespaceCallback != nil {
		logger.Info("DEBUG: Calling setNamespaceCallback")
		api.setNamespaceCallback(namespace)
	} else {
		logger.Info("DEBUG: setNamespaceCallback is nil")
	}
	api.eventManager.TriggerEvent(EventNamespaceChanged, namespace)
}

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

func (api *PluginAPIImpl) ShowInputDialog(title, placeholder, submitCommand, cancelCommand string) {
	if api.showInputDialogCallback != nil {
		api.showInputDialogCallback(title, placeholder, submitCommand, cancelCommand)
	}
}

func (api *PluginAPIImpl) SetShowInputDialogCallback(callback func(title, placeholder, submitCommand, cancelCommand string)) {
	api.showInputDialogCallback = callback
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
			content: lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pods are the smallest deployable units in Kubernetes.") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Key Bindings:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ↑/↓/j/k:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Navigate pods") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• enter:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("View pod details") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• d:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Delete selected pods") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• r:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Refresh") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• /:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Search pods") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• esc:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Go back") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pod Status:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Running:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pod is running successfully") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Pending:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pod is being scheduled") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Failed:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pod has failed") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Succeeded:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pod completed successfully") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Common Actions:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• View logs:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Enter on a pod to see details") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Delete pod:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Select with space, then press 'd'") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Refresh:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Press 'r' to update the list"),
		},
		"Deployments": {
			title: "Deployments Help",
			content: lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Deployments manage the deployment and scaling of applications.") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Key Bindings:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ↑/↓/j/k:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Navigate deployments") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• enter:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("View deployment details") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• d:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Delete selected deployments") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• r:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Refresh") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• /:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Search deployments") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• esc:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Go back") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Deployment Status:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Ready:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Shows ready/desired replicas") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Updated:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Shows updated replicas") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Available:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Shows available replicas") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Common Actions:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Scale deployment:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Enter to view details and scale") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Update image:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Use deployment details view") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• View pods:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("See associated pods in details"),
		},
		"Services": {
			title: "Services Help",
			content: lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Services expose applications running on pods.") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Key Bindings:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ↑/↓/j/k:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Navigate services") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• enter:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("View service details") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• d:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Delete selected services") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• r:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Refresh") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• /:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Search services") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• esc:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Go back") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Service Types:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ClusterIP:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Internal cluster access") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• NodePort:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("External access via node port") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• LoadBalancer:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Cloud load balancer") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ExternalName:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("DNS alias") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Common Actions:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• View endpoints:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("See pods backing the service") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Check connectivity:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Use service details"),
		},
		"Ingresses": {
			title: "Ingresses Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Ingresses manage external access to services.

Key Bindings:
• ↑/↓/j/k: Navigate ingresses
• enter: View ingress details
• d: Delete selected ingresses
• r: Refresh
• /: Search ingresses
• esc: Go back

Common Actions:
• View rules: See routing rules
• Check TLS: View SSL certificates
• Test routing: Verify external access`),
		},
		"ConfigMaps": {
			title: "ConfigMaps Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`ConfigMaps store configuration data.

Key Bindings:
• ↑/↓/j/k: Navigate configmaps
• enter: View configmap details
• d: Delete selected configmaps
• r: Refresh
• /: Search configmaps
• esc: Go back

Common Actions:
• View data: See configuration key-value pairs
• Edit values: Modify configuration data
• Check usage: See which pods use this config`),
		},
		"Secrets": {
			title: "Secrets Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Secrets store sensitive information.

Key Bindings:
• ↑/↓/j/k: Navigate secrets
• enter: View secret details
• d: Delete selected secrets
• r: Refresh
• /: Search secrets
• esc: Go back

Secret Types:
• Opaque: Generic secret data
• TLS: Certificate/key pairs
• Docker: Docker registry credentials

Common Actions:
• View data: See secret contents (use caution)
• Rotate secrets: Update sensitive data
• Check usage: See which pods reference this secret`),
		},
		"Jobs": {
			title: "Jobs Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Jobs create and manage batch processing tasks.

Key Bindings:
• ↑/↓/j/k: Navigate jobs
• enter: View job details
• d: Delete selected jobs
• r: Refresh
• /: Search jobs
• esc: Go back

Job Status:
• Complete: Job finished successfully
• Failed: Job failed
• Active: Job is running

Common Actions:
• View pods: See job execution pods
• Check logs: Examine job output
• Retry failed jobs: Delete and recreate`),
		},
		"CronJobs": {
			title: "CronJobs Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`CronJobs run jobs on a schedule.

Key Bindings:
• ↑/↓/j/k: Navigate cronjobs
• enter: View cronjob details
• d: Delete selected cronjobs
• r: Refresh
• /: Search cronjobs
• esc: Go back

Schedule Format:
• Uses standard cron syntax
• Example: "0 0 * * *" = daily at midnight

Common Actions:
• View schedule: See execution schedule
• Check history: View past job executions
• Manual trigger: Run job immediately`),
		},
		"Nodes": {
			title: "Nodes Help",
			content: lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(`Nodes are the worker machines in the cluster.

Key Bindings:
• ↑/↓/j/k: Navigate nodes
• enter: View node details
• r: Refresh
• /: Search nodes
• esc: Go back

Node Status:
• Ready: Node is healthy and schedulable
• NotReady: Node has issues
• SchedulingDisabled: Node won't accept new pods

Common Actions:
• View capacity: See CPU/memory resources
• Check conditions: View node health status
• View pods: See pods running on node`),
		},
		"ResourceList": {
			title: "Resource List Help",
			content: lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Resource List shows available Kubernetes resources.") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Key Bindings:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ↑/↓/j/k:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Navigate resources") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• enter:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Select resource type") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• /:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Search resources") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• esc:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Go back") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Resource Categories:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Workloads:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pods, Deployments, Jobs, etc.") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Networking:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Services, Ingresses") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Configuration:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("ConfigMaps, Secrets") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• Infrastructure:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Nodes") + "\n\n" +

				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Quick Navigation:") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• p:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Pods") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• d:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Deployments") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• s:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Services") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• i:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Ingresses") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• c:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("ConfigMaps") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• e:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Secrets") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• n:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Nodes") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• j:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Jobs") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• k:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("CronJobs") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• m:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("DaemonSets") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• t:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("StatefulSets") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• r:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("ReplicaSets") + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• a:") +
				lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("ServiceAccounts"),
		},
	}

	if help, exists := helpMap[resourceType]; exists {
		return help.title, help.content
	}

	return "Help", lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render(fmt.Sprintf("Help for %s", resourceType)) + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.HeaderColor)).Bold(true).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Key Bindings:") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• ↑/↓/j/k:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Navigate items") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• enter:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("View details") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• d:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Delete selected items (if supported)") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• r:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Refresh") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• /:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Search") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLKeyColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("• esc:") +
		lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(customstyles.YAMLValueColor)).Background(lipgloss.Color(customstyles.BackgroundColor)).Render("Go back") + "\n\n" +
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
