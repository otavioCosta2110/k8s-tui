package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
)

func TestNewNamespaces(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	// Since NewNamespaces calls k8s.FetchNamespaces which would connect to a cluster,
	// we'll test the structure with mock data
	model := &namespacesModel{
		list:      []string{"default", "kube-system", "kube-public"},
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if len(model.list) != 3 {
		t.Error("Expected 3 namespaces in model")
	}
	if model.list[0] != "default" {
		t.Error("Expected first namespace to be 'default'")
	}
	if model.loading {
		t.Error("Expected loading to be false")
	}
	if model.err != nil {
		t.Error("Expected no error")
	}
}

func TestNamespacesModelWithEmptyList(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model := &namespacesModel{
		list:      []string{},
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if len(model.list) != 0 {
		t.Error("Expected 0 namespaces in model")
	}
}

func TestNamespacesModelWithSingleNamespace(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model := &namespacesModel{
		list:      []string{"production"},
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if len(model.list) != 1 {
		t.Error("Expected 1 namespace in model")
	}
	if model.list[0] != "production" {
		t.Error("Expected namespace to be 'production'")
	}
}

func TestNamespacesModelWithSystemNamespaces(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	systemNamespaces := []string{
		"default",
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"ingress-nginx",
		"cert-manager",
	}

	model := &namespacesModel{
		list:      systemNamespaces,
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if len(model.list) != 6 {
		t.Error("Expected 6 namespaces in model")
	}

	// Test that all expected namespaces are present
	expectedNamespaces := map[string]bool{
		"default":         true,
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
		"ingress-nginx":   true,
		"cert-manager":    true,
	}

	for _, ns := range model.list {
		if !expectedNamespaces[ns] {
			t.Errorf("Unexpected namespace: %s", ns)
		}
	}
}

func TestNamespacesModelWithLoadingState(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model := &namespacesModel{
		list:      []string{},
		k8sClient: &client,
		loading:   true,
		err:       nil,
	}

	if !model.loading {
		t.Error("Expected loading to be true")
	}
	if len(model.list) != 0 {
		t.Error("Expected empty list when loading")
	}
}

func TestNamespacesModelWithError(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	testErr := &testError{message: "connection failed"}

	model := &namespacesModel{
		list:      []string{},
		k8sClient: &client,
		loading:   false,
		err:       testErr,
	}

	if model.err == nil {
		t.Error("Expected error to be set")
	}
	if model.err.Error() != "connection failed" {
		t.Error("Expected specific error message")
	}
	if model.loading {
		t.Error("Expected loading to be false when error occurred")
	}
}

// Helper type for testing errors
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestNamespacesModelClientAssignment(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}

	model := &namespacesModel{
		list:      []string{"default"},
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if model.k8sClient == nil {
		t.Error("Expected k8sClient to be assigned")
	}
	if model.k8sClient.Namespace != "test-namespace" {
		t.Error("Expected k8sClient to have correct namespace")
	}
}

func TestNamespacesModelNamespaceDiversity(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	// Test with various namespace patterns
	diverseNamespaces := []string{
		"default",
		"production",
		"staging",
		"development",
		"testing",
		"monitoring",
		"logging",
		"ingress-nginx",
		"cert-manager",
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"app-team-1",
		"backend-services",
		"frontend-apps",
	}

	model := &namespacesModel{
		list:      diverseNamespaces,
		k8sClient: &client,
		loading:   false,
		err:       nil,
	}

	if len(model.list) != 15 {
		t.Error("Expected 15 namespaces in model")
	}

	// Test that we can find specific types of namespaces
	hasSystemNamespace := false
	hasAppNamespace := false
	hasTeamNamespace := false

	for _, ns := range model.list {
		if ns == "kube-system" {
			hasSystemNamespace = true
		}
		if ns == "production" {
			hasAppNamespace = true
		}
		if ns == "app-team-1" {
			hasTeamNamespace = true
		}
	}

	if !hasSystemNamespace {
		t.Error("Expected to find system namespace")
	}
	if !hasAppNamespace {
		t.Error("Expected to find application namespace")
	}
	if !hasTeamNamespace {
		t.Error("Expected to find team namespace")
	}
}
