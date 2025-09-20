package plugins

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

type PluginAPIImpl struct {
	currentNamespace string
	headerComponents []UIInjectionPoint
	footerComponents []UIInjectionPoint
	commands         map[string]PluginCommand
	config           map[string]interface{}
	eventHandlers    map[PluginEvent][]func(data interface{}) error
	client           k8s.Client
}

func NewPluginAPI() *PluginAPIImpl {
	return &PluginAPIImpl{
		currentNamespace: "default",
		headerComponents: make([]UIInjectionPoint, 0),
		footerComponents: make([]UIInjectionPoint, 0),
		commands:         make(map[string]PluginCommand),
		config:           make(map[string]any),
		eventHandlers:    make(map[PluginEvent][]func(data any) error),
	}
}

func (api *PluginAPIImpl) GetCurrentNamespace() string {
	return api.currentNamespace
}

func (api *PluginAPIImpl) SetCurrentNamespace(namespace string) {
	api.currentNamespace = namespace
	api.TriggerEvent(EventNamespaceChanged, namespace)
}

func (api *PluginAPIImpl) SetStatusMessage(message string) {
	logger.Info(fmt.Sprintf("📢 Plugin Status: %s", message))
	// In a real implementation, this would update the UI status bar
}

func (api *PluginAPIImpl) AddHeaderComponent(component UIInjectionPoint) {
	api.headerComponents = append(api.headerComponents, component)
	logger.PluginDebug("api", fmt.Sprintf("Added header component: %s", component.Component.Config["content"]))
}

func (api *PluginAPIImpl) AddFooterComponent(component UIInjectionPoint) {
	api.footerComponents = append(api.footerComponents, component)
	logger.PluginDebug("api", fmt.Sprintf("Added footer component: %s", component.Component.Config["content"]))
}

func (api *PluginAPIImpl) GetHeaderComponents() []UIInjectionPoint {
	return api.headerComponents
}

func (api *PluginAPIImpl) GetFooterComponents() []UIInjectionPoint {
	return api.footerComponents
}

func (api *PluginAPIImpl) RegisterCommand(name, description string, handler func(args []string) (string, error)) {
	api.commands[name] = PluginCommand{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	logger.PluginDebug("api", fmt.Sprintf("Registered command: %s - %s", name, description))
}

func (api *PluginAPIImpl) ExecuteCommand(name string, args []string) (string, error) {
	if cmd, exists := api.commands[name]; exists {
		return cmd.Handler(args)
	}
	return "", fmt.Errorf("command not found: %s", name)
}

func (api *PluginAPIImpl) GetConfig(key string) any {
	return api.config[key]
}

func (api *PluginAPIImpl) SetConfig(key string, value any) {
	api.config[key] = value
	logger.PluginDebug("api", fmt.Sprintf("Set config %s = %v", key, value))
}

func (api *PluginAPIImpl) RegisterEventHandler(event PluginEvent, handler func(data any) error) {
	api.eventHandlers[event] = append(api.eventHandlers[event], handler)
}

func (api *PluginAPIImpl) TriggerEvent(event PluginEvent, data any) {
	if handlers, exists := api.eventHandlers[event]; exists {
		for _, handler := range handlers {
			if err := handler(data); err != nil {
				logger.PluginError("api", fmt.Sprintf("Error in event handler for %s: %v", event, err))
			}
		}
	}
}

func (api *PluginAPIImpl) GetCommands() map[string]PluginCommand {
	return api.commands
}

func (api *PluginAPIImpl) GetClient() k8s.Client {
	return api.client
}

func (api *PluginAPIImpl) SetClient(client k8s.Client) {
	api.client = client
}
