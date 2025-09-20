package plugins

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string

	// Version returns the version of the plugin
	Version() string

	// Description returns a description of what the plugin does
	Description() string

	// Initialize is called when the plugin is loaded
	Initialize() error

	// Shutdown is called when the plugin is unloaded
	Shutdown() error
}

type ResourcePlugin interface {
	Plugin

	// GetResourceTypes returns the custom resource types provided by this plugin
	GetResourceTypes() []CustomResourceType

	// GetResourceData fetches data for a custom resource type
	GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error)

	// DeleteResource deletes a custom resource
	DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error

	// GetResourceInfo gets information about a specific custom resource
	GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error)
}

type UIPlugin interface {
	Plugin

	// GetUIExtensions returns UI extensions provided by this plugin
	GetUIExtensions() []UIExtension
}

type CustomResourceType struct {
	// Name is the display name of the resource type
	Name string

	// Type is the internal type identifier
	Type string

	// Icon is the icon to display for this resource type
	Icon string

	// DisplayComponent defines how this resource should be rendered
	DisplayComponent DisplayComponent

	// Columns defines the table columns for this resource type (legacy support)
	Columns []table.Column

	// RefreshInterval defines how often to refresh this resource type
	RefreshInterval time.Duration

	// Namespaced indicates if this resource type is namespaced
	Namespaced bool

	// Category for grouping similar resources
	Category string

	// Description of what this resource type shows
	Description string
}

type DisplayComponent struct {
	// Type is the component type (table, yaml, text, chart, gauge, etc.)
	Type string

	// Config contains component-specific configuration
	Config map[string]interface{}

	// Style contains styling information
	Style ComponentStyle
}

type ComponentStyle struct {
	// Width in characters (0 = auto)
	Width int

	// Height in lines (0 = auto)
	Height int

	// Border style
	Border string

	// Colors
	ForegroundColor string
	BackgroundColor string
	BorderColor     string
}

type UIInjectionPoint struct {
	// Location where to inject (header, footer, sidebar, status_bar, notifications)
	Location string

	// Position within the location (left, right, center, top, bottom)
	Position string

	// Priority for ordering (higher = more prominent)
	Priority int

	// Component to render at this location
	Component DisplayComponent

	// Data source for dynamic content
	DataSource string

	// UpdateInterval in seconds (0 = static)
	UpdateInterval int
}

type Interaction struct {
	// Type of interaction (button, menu, keybinding, hover)
	Type string

	// Label for display
	Label string

	// KeyBinding for keyboard activation
	KeyBinding string

	// Handler function to call
	Handler func() tea.Cmd

	// Context when this interaction is available
	Context string

	// Enabled indicates if this interaction is currently available
	Enabled bool

	// Tooltip for hover help
	Tooltip string
}

type UIExtension struct {
	// Name is the name of the extension
	Name string

	// Type is the type of extension (e.g., "menu_item", "toolbar_button", "ui_injection")
	Type string

	// Handler is the function to call when the extension is activated
	Handler func() tea.Cmd

	// KeyBinding is the key binding for this extension
	KeyBinding string

	// InjectionPoints where this extension should be rendered
	InjectionPoints []UIInjectionPoint

	// Interactions available for this extension
	Interactions []Interaction

	// Dependencies on other plugins or features
	Dependencies []string
}

type PluginRegistry struct {
	resourcePlugins []ResourcePlugin
	uiPlugins       []UIPlugin
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		resourcePlugins: make([]ResourcePlugin, 0),
		uiPlugins:       make([]UIPlugin, 0),
	}
}

func (pr *PluginRegistry) RegisterResourcePlugin(plugin ResourcePlugin) {
	pr.resourcePlugins = append(pr.resourcePlugins, plugin)
}

func (pr *PluginRegistry) RegisterUIPlugin(plugin UIPlugin) {
	pr.uiPlugins = append(pr.uiPlugins, plugin)
}

func (pr *PluginRegistry) GetResourcePlugins() []ResourcePlugin {
	return pr.resourcePlugins
}

func (pr *PluginRegistry) GetUIPlugins() []UIPlugin {
	return pr.uiPlugins
}

func (pr *PluginRegistry) GetCustomResourceTypes() []CustomResourceType {
	var types []CustomResourceType
	for _, plugin := range pr.resourcePlugins {
		types = append(types, plugin.GetResourceTypes()...)
	}
	return types
}

type PluginmanagerStylePlugin interface {
	Plugin

	// Setup is called to configure the plugin with user options
	Setup(opts map[string]interface{}) error

	// Config returns the default configuration for this plugin
	Config() map[string]interface{}

	// Commands returns the commands this plugin provides
	Commands() []PluginCommand

	// Hooks returns the hooks this plugin registers for
	Hooks() []PluginHook
}

type PluginCommand struct {
	Name        string
	Description string
	Handler     func(args []string) (string, error)
}

type PluginHook struct {
	Event   string
	Handler func(data interface{}) error
}

type PluginEvent string

const (
	EventAppStarted       PluginEvent = "app_started"
	EventAppShutdown      PluginEvent = "app_shutdown"
	EventNamespaceChanged PluginEvent = "namespace_changed"
	EventResourceSelected PluginEvent = "resource_selected"
	EventUIUpdate         PluginEvent = "ui_update"
)

type PluginAPI interface {
	// GetCurrentNamespace returns the current namespace
	GetCurrentNamespace() string

	// SetStatusMessage sets a status message in the UI
	SetStatusMessage(message string)

	// AddHeaderComponent adds a component to the header
	AddHeaderComponent(component UIInjectionPoint)

	// AddFooterComponent adds a component to the footer
	AddFooterComponent(component UIInjectionPoint)

	// RegisterCommand registers a new command
	RegisterCommand(name, description string, handler func(args []string) (string, error))

	// ExecuteCommand executes a command by name
	ExecuteCommand(name string, args []string) (string, error)

	// GetConfig gets a configuration value
	GetConfig(key string) interface{}

	// SetConfig sets a configuration value
	SetConfig(key string, value interface{})

	// GetClient gets the Kubernetes client
	GetClient() k8s.Client

	// SetClient sets the Kubernetes client
	SetClient(client k8s.Client)

	// Kubernetes resource API methods
	GetPods(namespace string) ([]k8s.PodInfo, error)
	GetServices(namespace string) ([]k8s.ServiceInfo, error)
	GetDeployments(namespace string) ([]k8s.DeploymentInfo, error)
	GetConfigMaps(namespace string) ([]k8s.Configmap, error)
	GetSecrets(namespace string) ([]k8s.SecretInfo, error)
	GetIngresses(namespace string) ([]k8s.IngressInfo, error)
	GetJobs(namespace string) ([]k8s.JobInfo, error)
	GetCronJobs(namespace string) ([]k8s.CronJobInfo, error)
	GetDaemonSets(namespace string) ([]k8s.DaemonSetInfo, error)
	GetStatefulSets(namespace string) ([]k8s.StatefulSetInfo, error)
	GetReplicaSets(namespace string) ([]k8s.ReplicaSetInfo, error)
	GetNodes() ([]k8s.NodeInfo, error)
	GetNamespaces() ([]string, error)
	GetServiceAccounts(namespace string) ([]k8s.ServiceAccountInfo, error)

	// Delete methods
	DeletePod(namespace, name string) error
	DeleteService(namespace, name string) error
	DeleteDeployment(namespace, name string) error
	DeleteConfigMap(namespace, name string) error
	DeleteSecret(namespace, name string) error
	DeleteIngress(namespace, name string) error
	DeleteJob(namespace, name string) error
	DeleteCronJob(namespace, name string) error
	DeleteDaemonSet(namespace, name string) error
	DeleteStatefulSet(namespace, name string) error
	DeleteReplicaSet(namespace, name string) error
	DeleteServiceAccount(namespace, name string) error
}
