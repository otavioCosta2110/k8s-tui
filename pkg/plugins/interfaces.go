package plugins

import (
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type Plugin interface {
	
	Name() string

	
	Version() string

	
	Description() string

	
	Initialize() error

	
	Shutdown() error
}

type ResourcePlugin interface {
	Plugin

	
	GetResourceTypes() []CustomResourceType

	
	GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error)

	
	DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error

	
	GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error)
}

type UIPlugin interface {
	Plugin

	
	GetUIExtensions() []UIExtension
}

type CustomResourceType struct {
	
	Name string

	
	Type string

	
	Icon string

	
	DisplayComponent DisplayComponent

	
	Columns []table.Column

	
	RefreshInterval time.Duration

	
	Namespaced bool

	
	Category string

	
	Description string
}

type DisplayComponent struct {
	
	Type string

	
	Config map[string]interface{}

	
	Style ComponentStyle
}

type ComponentStyle struct {
	
	Width int

	
	Height int

	
	Border string

	
	ForegroundColor string
	BackgroundColor string
	BorderColor     string
}

type UIInjectionPoint struct {
	
	Location string

	
	Position string

	
	Priority int

	
	Component DisplayComponent

	
	DataSource string

	
	UpdateInterval int
}

type Interaction struct {
	
	Type string

	
	Label string

	
	KeyBinding string

	
	Handler func() tea.Cmd

	
	Context string

	
	Enabled bool

	
	Tooltip string
}

type UIExtension struct {
	
	Name string

	
	Type string

	
	Handler func() tea.Cmd

	
	KeyBinding string

	
	InjectionPoints []UIInjectionPoint

	
	Interactions []Interaction

	
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

	
	Setup(opts map[string]interface{}) error

	
	Config() map[string]interface{}

	
	Commands() []PluginCommand

	
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
	
	GetCurrentNamespace() string

	
	SetStatusMessage(message string)

	
	AddHeaderComponent(component UIInjectionPoint)

	
	AddFooterComponent(component UIInjectionPoint)

	
	RegisterCommand(name, description string, handler func(args []string) (string, error))

	
	ExecuteCommand(name string, args []string) (string, error)

	
	GetConfig(key string) interface{}

	
	SetConfig(key string, value interface{})

	
	GetClient() k8s.Client

	
	SetClient(client k8s.Client)

	
	GetPods(namespace string, selector ...string) ([]k8s.PodInfo, error)
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

	
	DescribePod(namespace, name string) (string, error)
	DescribeService(namespace, name string) (string, error)
	DescribeDeployment(namespace, name string) (string, error)
	DescribeConfigMap(namespace, name string) (string, error)
	DescribeSecret(namespace, name string) (string, error)
	DescribeIngress(namespace, name string) (string, error)
	DescribeJob(namespace, name string) (string, error)
	DescribeCronJob(namespace, name string) (string, error)
	DescribeDaemonSet(namespace, name string) (string, error)
	DescribeStatefulSet(namespace, name string) (string, error)
	DescribeReplicaSet(namespace, name string) (string, error)
	DescribeNode(name string) (string, error)
	DescribeServiceAccount(namespace, name string) (string, error)

	
	RegisterResourceHandler(resourceType k8s.ResourceType, handler ResourceHandler)
	GetSupportedResourceTypes() []k8s.ResourceType
	GetResourceHandler(resourceType k8s.ResourceType) (ResourceHandler, bool)

	
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
