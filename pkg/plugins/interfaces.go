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

// Legacy plugin interfaces - DEPRECATED
// These will be removed in a future version. Use NeovimStylePlugin instead.

// ResourcePlugin defines the interface for plugins that provide custom resources
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

// UIPlugin defines the interface for plugins that extend the UI
type UIPlugin interface {
	Plugin

	// GetUIExtensions returns UI extensions provided by this plugin
	GetUIExtensions() []UIExtension
}

// CustomResourceType defines a custom resource type provided by a plugin
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

// DisplayComponent defines how plugin data should be rendered
type DisplayComponent struct {
	// Type is the component type (table, yaml, text, chart, gauge, etc.)
	Type string

	// Config contains component-specific configuration
	Config map[string]interface{}

	// Style contains styling information
	Style ComponentStyle
}

// ComponentStyle defines styling for display components
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

// UIInjectionPoint defines where in the UI a plugin can inject content
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

// Interaction defines user interactions available for plugin content
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

// UIExtension defines a UI extension provided by a plugin
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

// PluginRegistry manages loaded plugins
type PluginRegistry struct {
	resourcePlugins []ResourcePlugin
	uiPlugins       []UIPlugin
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		resourcePlugins: make([]ResourcePlugin, 0),
		uiPlugins:       make([]UIPlugin, 0),
	}
}

// RegisterResourcePlugin registers a resource plugin
func (pr *PluginRegistry) RegisterResourcePlugin(plugin ResourcePlugin) {
	pr.resourcePlugins = append(pr.resourcePlugins, plugin)
}

// RegisterUIPlugin registers a UI plugin
func (pr *PluginRegistry) RegisterUIPlugin(plugin UIPlugin) {
	pr.uiPlugins = append(pr.uiPlugins, plugin)
}

// GetResourcePlugins returns all registered resource plugins
func (pr *PluginRegistry) GetResourcePlugins() []ResourcePlugin {
	return pr.resourcePlugins
}

// GetUIPlugins returns all registered UI plugins
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

// NeovimStylePlugin defines a plugin that follows Neovim-style architecture
type NeovimStylePlugin interface {
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

// PluginCommand represents a command that a plugin provides
type PluginCommand struct {
	Name        string
	Description string
	Handler     func(args []string) (string, error)
}

// PluginHook represents a hook that a plugin can register for
type PluginHook struct {
	Event   string
	Handler func(data interface{}) error
}

// PluginEvent represents events that can be triggered in the application
type PluginEvent string

const (
	EventAppStarted       PluginEvent = "app_started"
	EventAppShutdown      PluginEvent = "app_shutdown"
	EventNamespaceChanged PluginEvent = "namespace_changed"
	EventResourceSelected PluginEvent = "resource_selected"
	EventUIUpdate         PluginEvent = "ui_update"
)

// PluginAPI provides methods for plugins to interact with the application
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
}
