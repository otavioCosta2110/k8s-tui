package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s"
	"testing"
)

func TestNewMainModel(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	namespace := "test-namespace"

	model := NewMainModel(client, namespace)

	if model.kube.Namespace != "default" {
		t.Error("Expected kube client namespace to be 'default'")
	}
	if model.namespace != "test-namespace" {
		t.Error("Expected namespace to be 'test-namespace'")
	}
}

func TestMainModelWithEmptyNamespace(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewMainModel(client, "")

	if model.namespace != "" {
		t.Error("Expected namespace to be empty")
	}
	if model.kube.Namespace != "default" {
		t.Error("Expected kube client namespace to be 'default'")
	}
}

func TestMainModelWithDifferentClients(t *testing.T) {
	testCases := []struct {
		name     string
		client   k8s.Client
		expected string
	}{
		{
			name:     "Default client",
			client:   k8s.Client{Namespace: "default"},
			expected: "default",
		},
		{
			name:     "Production client",
			client:   k8s.Client{Namespace: "production"},
			expected: "production",
		},
		{
			name:     "Empty client",
			client:   k8s.Client{Namespace: ""},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewMainModel(tc.client, "test-namespace")
			if model.kube.Namespace != tc.expected {
				t.Errorf("Expected kube client namespace to be '%s', got '%s'", tc.expected, model.kube.Namespace)
			}
		})
	}
}

func TestMainModelInit(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewMainModel(client, "test-namespace")

	cmd := model.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

func TestMainModelUpdate(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewMainModel(client, "test-namespace")

	msg := "test message"
	updatedModel, cmd := model.Update(msg)

	if updatedModel != model {
		t.Error("Expected model to remain unchanged")
	}
	if cmd != nil {
		t.Error("Expected cmd to be nil")
	}
}

func TestMainModelView(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewMainModel(client, "test-namespace")

	view := model.View()
	if view != "" {
		t.Error("Expected View to return empty string")
	}
}

func TestMainModelWithVariousNamespaces(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	testNamespaces := []string{
		"default",
		"production",
		"staging",
		"development",
		"testing",
		"monitoring",
		"kube-system",
		"",
	}

	for _, ns := range testNamespaces {
		t.Run("namespace_"+ns, func(t *testing.T) {
			model := NewMainModel(client, ns)
			if model.namespace != ns {
				t.Errorf("Expected namespace to be '%s', got '%s'", ns, model.namespace)
			}
		})
	}
}

func TestMainModelClientIsolation(t *testing.T) {
	client1 := k8s.Client{Namespace: "client1"}
	client2 := k8s.Client{Namespace: "client2"}

	model1 := NewMainModel(client1, "namespace1")
	model2 := NewMainModel(client2, "namespace2")

	if model1.kube.Namespace != "client1" {
		t.Error("Model1 should have client1 namespace")
	}
	if model2.kube.Namespace != "client2" {
		t.Error("Model2 should have client2 namespace")
	}
	if model1.namespace != "namespace1" {
		t.Error("Model1 should have namespace1")
	}
	if model2.namespace != "namespace2" {
		t.Error("Model2 should have namespace2")
	}
}

func TestMainModelStateConsistency(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	originalNamespace := "original-namespace"

	model := NewMainModel(client, originalNamespace)

	cmd1 := model.Init()
	cmd2 := model.Init()

	if cmd1 != nil || cmd2 != nil {
		t.Error("Expected both Init calls to return nil")
	}

	view1 := model.View()
	view2 := model.View()

	if view1 != view2 {
		t.Error("Expected View to be consistent")
	}

	if model.namespace != originalNamespace {
		t.Error("Expected namespace to remain unchanged")
	}
	if model.kube.Namespace != "default" {
		t.Error("Expected kube namespace to remain unchanged")
	}
}
