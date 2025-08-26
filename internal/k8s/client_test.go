package k8s

import (
	"testing"
	"time"
)

func TestResourceTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant ResourceType
		expected string
	}{
		{"Pod", ResourceTypePod, "pod"},
		{"Deployment", ResourceTypeDeployment, "deployment"},
		{"ReplicaSet", ResourceTypeReplicaSet, "replicaset"},
		{"ConfigMap", ResourceTypeConfigMap, "configmap"},
		{"Service", ResourceTypeService, "service"},
		{"ServiceAccount", ResourceTypeServiceAccount, "serviceaccount"},
		{"Ingress", ResourceTypeIngress, "ingress"},
		{"Secret", ResourceTypeSecret, "secret"},
		{"Node", ResourceTypeNode, "node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.constant))
			}
		})
	}
}

func TestResourceInfoStruct(t *testing.T) {
	testTime := time.Now()
	info := ResourceInfo{
		Name:      "test-pod",
		Namespace: "default",
		Kind:      ResourceTypePod,
		Age:       "1h",
		CreatedAt: testTime,
	}

	if info.Name != "test-pod" {
		t.Error("ResourceInfo Name field mismatch")
	}
	if info.Namespace != "default" {
		t.Error("ResourceInfo Namespace field mismatch")
	}
	if info.Kind != ResourceTypePod {
		t.Error("ResourceInfo Kind field mismatch")
	}
	if info.Age != "1h" {
		t.Error("ResourceInfo Age field mismatch")
	}
	if info.CreatedAt != testTime {
		t.Error("ResourceInfo CreatedAt field mismatch")
	}
}

func TestResourceManagerInterface(t *testing.T) {
	if ResourceTypePod != "pod" {
		t.Error("ResourceManager interface constants should be accessible")
	}
}

func TestClientStruct(t *testing.T) {
	client := &Client{
		Namespace: "test-namespace",
	}

	if client.Namespace != "test-namespace" {
		t.Error("Client namespace not set correctly")
	}
}
