package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/types"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

func TestNewGenericResourceModel(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	namespace := "test-namespace"
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypePod,
		Title:           "Test Resources",
		ColumnWidths:    []float64{0.5, 0.5},
		RefreshInterval: 10 * time.Second,
	}

	model := NewGenericResourceModel(client, namespace, config)

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != namespace {
		t.Error("Expected namespace to be set correctly")
	}
	if model.resourceType != k8s.ResourceTypePod {
		t.Error("Expected resourceType to be set correctly")
	}
	if model.refreshInterval != 10*time.Second {
		t.Error("Expected refreshInterval to be set correctly")
	}
	if model.config.Title != "Test Resources" {
		t.Error("Expected config title to be set correctly")
	}
	if model.loading {
		t.Error("Expected loading to be false initially")
	}
	if model.err != nil {
		t.Error("Expected no initial error")
	}
}

func TestGenericResourceModelGetters(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	config := ResourceConfig{
		ResourceType: k8s.ResourceTypeDeployment,
	}

	model := NewGenericResourceModel(client, "test-namespace", config)

	if model.GetResourceType() != k8s.ResourceTypeDeployment {
		t.Error("GetResourceType should return correct resource type")
	}
	if model.GetNamespace() != "test-namespace" {
		t.Error("GetNamespace should return correct namespace")
	}
}

func TestGenericResourceModelWithDifferentResourceTypes(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	testCases := []struct {
		name         string
		resourceType k8s.ResourceType
		expected     k8s.ResourceType
	}{
		{"Pod", k8s.ResourceTypePod, k8s.ResourceTypePod},
		{"Deployment", k8s.ResourceTypeDeployment, k8s.ResourceTypeDeployment},
		{"Service", k8s.ResourceTypeService, k8s.ResourceTypeService},
		{"ConfigMap", k8s.ResourceTypeConfigMap, k8s.ResourceTypeConfigMap},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := ResourceConfig{ResourceType: tc.resourceType}
			model := NewGenericResourceModel(client, "test-namespace", config)

			if model.GetResourceType() != tc.expected {
				t.Errorf("Expected resource type %v, got %v", tc.expected, model.GetResourceType())
			}
		})
	}
}

func TestGenericResourceModelWithEmptyConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	config := ResourceConfig{}

	model := NewGenericResourceModel(client, "test-namespace", config)

	if model.config.ResourceType != "" {
		t.Error("Expected empty resource type")
	}
	if model.config.Title != "" {
		t.Error("Expected empty title")
	}
	if model.config.RefreshInterval != 0 {
		t.Error("Expected zero refresh interval")
	}
}

func TestGenericResourceModelWithNilClient(t *testing.T) {
	var client k8s.Client
	config := ResourceConfig{ResourceType: k8s.ResourceTypePod}

	model := NewGenericResourceModel(client, "test-namespace", config)

	if model.k8sClient == nil {
		t.Error("Expected k8sClient to be set even with nil client")
	}
}

func TestGenericResourceModelStateManagement(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	config := ResourceConfig{ResourceType: k8s.ResourceTypePod}

	model := NewGenericResourceModel(client, "test-namespace", config)

	if model.loading {
		t.Error("Expected initial loading state to be false")
	}
	if model.err != nil {
		t.Error("Expected initial error to be nil")
	}
	if len(model.resourceData) != 0 {
		t.Error("Expected initial resourceData to be empty")
	}

	model.loading = true
	if !model.loading {
		t.Error("Expected loading state to be changeable")
	}

	testErr := &testError{message: "test error"}
	model.err = testErr
	if model.err == nil {
		t.Error("Expected error to be settable")
	}
	if model.err.Error() != "test error" {
		t.Error("Expected error message to be preserved")
	}
}

func TestGenericResourceModelWithResourceData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	config := ResourceConfig{ResourceType: k8s.ResourceTypePod}

	model := NewGenericResourceModel(client, "test-namespace", config)

	mockData := []types.ResourceData{
		&mockResourceData{name: "resource1", namespace: "test-namespace"},
		&mockResourceData{name: "resource2", namespace: "test-namespace"},
	}

	model.resourceData = mockData

	if len(model.resourceData) != 2 {
		t.Error("Expected 2 resource data items")
	}
	if model.resourceData[0].GetName() != "resource1" {
		t.Error("Expected first resource to have correct name")
	}
	if model.resourceData[1].GetNamespace() != "test-namespace" {
		t.Error("Expected second resource to have correct namespace")
	}
}

func TestGenericResourceModelNamespaceVariations(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	testNamespaces := []string{
		"default",
		"production",
		"staging",
		"development",
		"testing",
		"kube-system",
		"kube-public",
		"ingress-nginx",
		"cert-manager",
		"monitoring",
		"logging",
		"",
	}

	for _, ns := range testNamespaces {
		t.Run("namespace_"+ns, func(t *testing.T) {
			config := ResourceConfig{ResourceType: k8s.ResourceTypePod}
			model := NewGenericResourceModel(client, ns, config)

			if model.GetNamespace() != ns {
				t.Errorf("Expected namespace to be '%s', got '%s'", ns, model.GetNamespace())
			}
		})
	}
}

type mockResourceData struct {
	name      string
	namespace string
}

func (m *mockResourceData) GetName() string {
	return m.name
}

func (m *mockResourceData) GetNamespace() string {
	return m.namespace
}

func (m *mockResourceData) GetColumns() table.Row {
	return table.Row{m.name, m.namespace}
}
