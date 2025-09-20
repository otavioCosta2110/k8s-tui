package plugins

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// Plugin defines the interface that all plugins must implement
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

	// Columns defines the table columns for this resource type
	Columns []table.Column

	// RefreshInterval defines how often to refresh this resource type
	RefreshInterval time.Duration

	// Namespaced indicates if this resource type is namespaced
	Namespaced bool
}

// UIExtension defines a UI extension provided by a plugin
type UIExtension struct {
	// Name is the name of the extension
	Name string

	// Type is the type of extension (e.g., "menu_item", "toolbar_button")
	Type string

	// Handler is the function to call when the extension is activated
	Handler func() tea.Cmd

	// KeyBinding is the key binding for this extension
	KeyBinding string
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

// GetCustomResourceTypes returns all custom resource types from all plugins
func (pr *PluginRegistry) GetCustomResourceTypes() []CustomResourceType {
	var types []CustomResourceType
	for _, plugin := range pr.resourcePlugins {
		types = append(types, plugin.GetResourceTypes()...)
	}
	return types
}
