package main

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

// ExamplePlugin demonstrates a custom resource plugin
type ExamplePlugin struct {
	name    string
	version string
}

// Plugin is the exported plugin instance
var Plugin ExamplePlugin

func init() {
	Plugin = ExamplePlugin{
		name:    "example-plugin",
		version: "1.0.0",
	}
}

func (p ExamplePlugin) Name() string {
	return p.name
}

func (p ExamplePlugin) Version() string {
	return p.version
}

func (p ExamplePlugin) Description() string {
	return "Example plugin demonstrating custom resource functionality"
}

func (p ExamplePlugin) Initialize() error {
	fmt.Printf("Example plugin %s v%s initialized\n", p.name, p.version)
	return nil
}

func (p ExamplePlugin) Shutdown() error {
	fmt.Printf("Example plugin %s v%s shutting down\n", p.name, p.version)
	return nil
}

func (p ExamplePlugin) GetResourceTypes() []plugins.CustomResourceType {
	return []plugins.CustomResourceType{
		{
			Name: "ExampleResources",
			Type: "exampleresource",
			Icon: "🔧",
			Columns: []table.Column{
				{Title: "Name", Width: 10},
				{Title: "Namespace", Width: 10},
				{Title: "Status", Width: 10},
				{Title: "Age", Width: 10},
			},
			RefreshInterval: 10 * time.Second,
			Namespaced:      true,
		},
	}
}

func (p ExamplePlugin) GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
	if resourceType != "exampleresource" {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Return some example data
	return []types.ResourceData{
		&ExampleResourceData{
			name:      "example-resource-1",
			namespace: namespace,
			status:    "Running",
			age:       "5m",
		},
		&ExampleResourceData{
			name:      "example-resource-2",
			namespace: namespace,
			status:    "Pending",
			age:       "2m",
		},
	}, nil
}

func (p ExamplePlugin) DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error {
	if resourceType != "exampleresource" {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	fmt.Printf("Deleting example resource %s/%s\n", namespace, name)
	return nil
}

func (p ExamplePlugin) GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error) {
	if resourceType != "exampleresource" {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return &k8s.ResourceInfo{
		Name:      name,
		Namespace: namespace,
		Kind:      k8s.ResourceType(resourceType),
		Age:       "5m",
	}, nil
}

func (p ExamplePlugin) GetUIExtensions() []plugins.UIExtension {
	return []plugins.UIExtension{}
}

// ExampleResourceData implements the ResourceData interface
type ExampleResourceData struct {
	name      string
	namespace string
	status    string
	age       string
}

func (erd *ExampleResourceData) GetName() string {
	return erd.name
}

func (erd *ExampleResourceData) GetNamespace() string {
	return erd.namespace
}

func (erd *ExampleResourceData) GetColumns() table.Row {
	return table.Row{
		erd.name,
		erd.namespace,
		erd.status,
		erd.age,
	}
}
