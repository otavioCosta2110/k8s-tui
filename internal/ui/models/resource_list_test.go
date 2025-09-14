package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"strings"
	"testing"
)

func TestNewResourceFactory(t *testing.T) {
	factory := NewResourceFactory()

	if factory == nil {
		t.Error("Expected factory to be non-nil")
	}

	if factory.registry == nil {
		t.Error("Expected registry to be initialized")
	}

	if factory.metadata == nil {
		t.Error("Expected metadata to be initialized")
	}

	if len(factory.validTypes) == 0 {
		t.Error("Expected validTypes to be populated")
	}
}

func TestResourceFactoryGetValidResourceTypes(t *testing.T) {
	factory := NewResourceFactory()
	validTypes := factory.GetValidResourceTypes()

	expectedTypes := []string{
		"Pods", "Deployments", "Services", "Ingresses",
		"ConfigMaps", "Secrets", "ServiceAccounts", "ReplicaSets", "Nodes",
		"Jobs", "CronJobs", "DaemonSets", "StatefulSets",
	}

	if len(validTypes) != len(expectedTypes) {
		t.Errorf("Expected %d valid types, got %d", len(expectedTypes), len(validTypes))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, validType := range validTypes {
			if validType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected resource type '%s' not found in valid types", expectedType)
		}
	}
}

func TestResourceFactoryGetResourceMetadata(t *testing.T) {
	factory := NewResourceFactory()

	metadata, exists := factory.GetResourceMetadata("Pods")
	if !exists {
		t.Error("Expected metadata to exist for Pods")
	}

	if metadata.Name != "Pods" {
		t.Error("Expected metadata name to be 'Pods'")
	}

	if metadata.Description == "" {
		t.Error("Expected metadata description to be non-empty")
	}

	if metadata.Category == "" {
		t.Error("Expected metadata category to be non-empty")
	}

	_, exists = factory.GetResourceMetadata("InvalidResource")
	if exists {
		t.Error("Expected metadata to not exist for invalid resource type")
	}
}

func TestResourceFactoryCreateResource(t *testing.T) {
	factory := NewResourceFactory()

	_, err := factory.CreateResource("InvalidResource", k8s.Client{}, "default")
	if err == nil {
		t.Error("Expected error creating invalid resource")
	}

	if !strings.Contains(err.Error(), "unsupported resource type") {
		t.Errorf("Expected error message to contain 'unsupported resource type', got '%s'", err.Error())
	}

	_, err = factory.CreateResource("", k8s.Client{}, "default")
	if err == nil {
		t.Error("Expected error for empty resource type")
	}
}

func TestResourceFactoryCreateAllResourceTypes(t *testing.T) {
	factory := NewResourceFactory()

	resourceTypes := []string{
		"Pods", "Deployments", "Services", "Ingresses",
		"ConfigMaps", "Secrets", "ReplicaSets",
	}

	for _, resourceType := range resourceTypes {
		t.Run("Registry_"+resourceType, func(t *testing.T) {
			creator, exists := factory.registry[resourceType]
			if !exists {
				t.Errorf("Resource type '%s' not found in registry", resourceType)
			}

			if creator == nil {
				t.Errorf("Resource type '%s' has nil creator function", resourceType)
			}

			metadata, exists := factory.metadata[resourceType]
			if !exists {
				t.Errorf("Resource type '%s' has no metadata", resourceType)
			}

			if metadata.Name != resourceType {
				t.Errorf("Resource type '%s' metadata name mismatch", resourceType)
			}
		})
	}
}

func TestNewResourceList(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	resourceList := NewResourceList(client, "default", "Pods")

	if resourceList.kube.Namespace != "default" {
		t.Error("Expected kube namespace to be 'default'")
	}

	if resourceList.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}

	if resourceList.resourceType != "Pods" {
		t.Error("Expected resourceType to be 'Pods'")
	}
}

func TestResourceListInitComponent(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	resourceList := NewResourceList(client, "default", "Pods")

	if resourceList.resourceType != "Pods" {
		t.Error("Expected resource type to be 'Pods'")
	}

	if resourceList.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
}

func TestResourceListInitComponentInvalidType(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	resourceList := NewResourceList(client, "default", "InvalidResource")

	if resourceList.resourceType != "InvalidResource" {
		t.Error("Expected resource type to be 'InvalidResource'")
	}
}

func TestResourceFactoryRegistryIntegrity(t *testing.T) {
	factory := NewResourceFactory()

	for _, resourceType := range factory.validTypes {
		creator, exists := factory.registry[resourceType]
		if !exists {
			t.Errorf("Resource type '%s' missing from registry", resourceType)
		}

		if creator == nil {
			t.Errorf("Resource type '%s' has nil creator function", resourceType)
		}

		metadata, exists := factory.metadata[resourceType]
		if !exists {
			t.Errorf("Resource type '%s' missing from metadata", resourceType)
		}

		if metadata.Name != resourceType {
			t.Errorf("Resource type '%s' metadata name mismatch: expected '%s', got '%s'",
				resourceType, resourceType, metadata.Name)
		}
	}
}

func TestResourceFactoryMetadataCompleteness(t *testing.T) {
	factory := NewResourceFactory()

	for _, resourceType := range factory.validTypes {
		metadata, _ := factory.metadata[resourceType]

		if metadata.Name == "" {
			t.Errorf("Resource type '%s' has empty name", resourceType)
		}

		if metadata.Description == "" {
			t.Errorf("Resource type '%s' has empty description", resourceType)
		}

		if metadata.Category == "" {
			t.Errorf("Resource type '%s' has empty category", resourceType)
		}

		if metadata.HelpText == "" {
			t.Errorf("Resource type '%s' has empty help text", resourceType)
		}
	}
}

func TestResourceFactoryCategories(t *testing.T) {
	factory := NewResourceFactory()

	categories := make(map[string][]string)

	for _, resourceType := range factory.validTypes {
		metadata, _ := factory.metadata[resourceType]
		categories[metadata.Category] = append(categories[metadata.Category], resourceType)
	}

	expectedCategories := []string{"Workloads", "Networking", "Configuration"}

	for _, expectedCategory := range expectedCategories {
		if resources, exists := categories[expectedCategory]; !exists || len(resources) == 0 {
			t.Errorf("Expected category '%s' to have resources, but found: %v", expectedCategory, resources)
		}
	}

	for category, resources := range categories {
		if len(resources) == 0 {
			t.Errorf("Category '%s' has no resources", category)
		}
	}
}

func TestResourceFactoryErrorHandling(t *testing.T) {
	factory := NewResourceFactory()

	_, err := factory.CreateResource("", k8s.Client{}, "default")
	if err == nil {
		t.Error("Expected error for empty resource type")
	}

	_, err = factory.CreateResource("   ", k8s.Client{}, "default")
	if err == nil {
		t.Error("Expected error for whitespace resource type")
	}

	_, err = factory.CreateResource("InvalidResource", k8s.Client{}, "default")
	if err == nil {
		t.Error("Expected error for invalid resource type")
	}
}

func TestResourceFactoryLogging(t *testing.T) {
	factory := NewResourceFactory()

	_, err := factory.CreateResource("InvalidResource", k8s.Client{}, "test-namespace")
	if err == nil {
		t.Error("Expected error for invalid resource type")
	}

	_, err = factory.CreateResource("", k8s.Client{}, "test-namespace")
	if err == nil {
		t.Error("Expected error for empty resource type")
	}
}

func TestResourceListBackwardCompatibility(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	resourceList := NewResourceList(client, "default", "Pods")

	if resourceList.resourceType != "Pods" {
		t.Error("Expected resource type to be 'Pods'")
	}

	if resourceList.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
}

func TestResourceFactoryResourceModelInterface(t *testing.T) {
	factory := NewResourceFactory()

	testCases := []struct {
		resourceType string
		expectError  bool
	}{
		{"Pods", false},
		{"Deployments", false},
		{"Services", false},
		{"Ingresses", false},
		{"ConfigMaps", false},
		{"Secrets", false},
		{"ReplicaSets", false},
		{"InvalidResource", true},
	}

	for _, tc := range testCases {
		t.Run("Interface_"+tc.resourceType, func(t *testing.T) {
			creator, exists := factory.registry[tc.resourceType]

			if tc.expectError {
				if exists {
					t.Errorf("Expected %s to not exist in registry", tc.resourceType)
				}
			} else {
				if !exists {
					t.Errorf("Expected %s to exist in registry", tc.resourceType)
				}
				if creator == nil {
					t.Errorf("Expected %s creator to be non-nil", tc.resourceType)
				}
			}
		})
	}
}
