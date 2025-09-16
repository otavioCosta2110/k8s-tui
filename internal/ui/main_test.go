package ui

import (
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	"strings"
	"testing"
)

func TestGetResourceTypeFromKey(t *testing.T) {
	appModel := &AppModel{}

	tests := []struct {
		key         string
		expected    string
		description string
	}{
		{"p", "Pods", "Pods mapping"},
		{"d", "Deployments", "Deployments mapping"},
		{"s", "Services", "Services mapping"},
		{"i", "Ingresses", "Ingresses mapping"},
		{"c", "ConfigMaps", "ConfigMaps mapping"},
		{"e", "Secrets", "Secrets mapping"},
		{"a", "ServiceAccounts", "ServiceAccounts mapping"},
		{"r", "ReplicaSets", "ReplicaSets mapping"},
		{"n", "Nodes", "Nodes mapping"},
		{"j", "Jobs", "Jobs mapping"},
		{"k", "CronJobs", "CronJobs mapping"},
		{"m", "DaemonSets", "DaemonSets mapping"},
		{"t", "StatefulSets", "StatefulSets mapping"},
		{"l", "ResourceList", "ResourceList mapping"},
		{"x", "", "Invalid key returns empty string"},
		{"", "", "Empty key returns empty string"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := appModel.getResourceTypeFromKey(test.key)
			if result != test.expected {
				t.Errorf("getResourceTypeFromKey(%q) = %q, expected %q", test.key, result, test.expected)
			}
		})
	}
}

func TestGetBreadcrumbTrail(t *testing.T) {
	global.ScreenWidth = 120

	client := &k8s.Client{
		Namespace:      "test-namespace",
		KubeconfigPath: "/home/user/.kube/config",
	}

	appModel := &AppModel{
		kube: *client,
	}

	tests := []struct {
		breadcrumbTrail []string
		expectedParts   []string
		description     string
	}{
		{
			breadcrumbTrail: []string{"Resource List"},
			expectedParts:   []string{"config", "test-namespace", "Resource List"},
			description:     "Resource List breadcrumb",
		},
		{
			breadcrumbTrail: []string{"Resource List", "Pods"},
			expectedParts:   []string{"config", "test-namespace", "Resource List", "Pods"},
			description:     "Resource List with Pods breadcrumb",
		},
		{
			breadcrumbTrail: []string{"Deployments"},
			expectedParts:   []string{"config", "test-namespace", "Deployments"},
			description:     "Direct resource breadcrumb",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			appModel.breadcrumbTrail = test.breadcrumbTrail
			result := appModel.getBreadcrumbTrail()

			for _, part := range test.expectedParts {
				if !strings.Contains(result, part) {
					t.Errorf("getBreadcrumbTrail() = %q, expected to contain %q", result, part)
				}
			}

			parts := strings.Split(result, " > ")
			if len(parts) != len(test.expectedParts) {
				t.Errorf("getBreadcrumbTrail() = %q, expected %d parts but got %d", result, len(test.expectedParts), len(parts))
			}
		})
	}
}

func TestIsCurrentScreenResourceType(t *testing.T) {
	appModel := &AppModel{}

	tests := []struct {
		breadcrumbTrail []string
		resourceType    string
		expected        bool
		description     string
	}{
		{
			breadcrumbTrail: []string{"Resource List"},
			resourceType:    "ResourceList",
			expected:        true,
			description:     "ResourceList on Resource List screen",
		},
		{
			breadcrumbTrail: []string{"Pods"},
			resourceType:    "Pods",
			expected:        true,
			description:     "Pods on Pods screen",
		},
		{
			breadcrumbTrail: []string{"Resource List"},
			resourceType:    "Pods",
			expected:        false,
			description:     "Pods not on Resource List screen",
		},
		{
			breadcrumbTrail: []string{},
			resourceType:    "ResourceList",
			expected:        false,
			description:     "ResourceList not detected with empty breadcrumb",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			appModel.breadcrumbTrail = test.breadcrumbTrail
			result := appModel.isCurrentScreenResourceType(test.resourceType)
			if result != test.expected {
				t.Errorf("isCurrentScreenResourceType(%q) = %v, expected %v", test.resourceType, result, test.expected)
			}
		})
	}
}

func TestListModelFooterText(t *testing.T) {
	listModel := &components.ListModel{}
	testFooterText := "Test breadcrumb: config > default > Resource List"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetFooterText panicked: %v", r)
		}
	}()

	listModel.SetFooterText(testFooterText)
}

func TestInitializeInitialBreadcrumb(t *testing.T) {
	client := &k8s.Client{
		Namespace:      "test-namespace",
		KubeconfigPath: "/home/user/.kube/config",
	}

	appModel := &AppModel{
		kube: *client,
	}

	resourceList := components.NewList([]string{"Pods", "Deployments"}, "Resource Types", nil)
	appModel.initializeInitialBreadcrumb(resourceList)

	expectedTrail := []string{"Resource List"}
	if len(appModel.breadcrumbTrail) != len(expectedTrail) {
		t.Errorf("Expected breadcrumb trail length %d, got %d", len(expectedTrail), len(appModel.breadcrumbTrail))
	}

	for i, expected := range expectedTrail {
		if i >= len(appModel.breadcrumbTrail) || appModel.breadcrumbTrail[i] != expected {
			t.Errorf("Expected breadcrumb[%d] = %q, got %q", i, expected, appModel.breadcrumbTrail[i])
		}
	}

	kubeconfigList := components.NewList([]string{"config1", "config2"}, "Kubeconfigs", nil)
	appModel.initializeInitialBreadcrumb(kubeconfigList)

	if len(appModel.breadcrumbTrail) != 0 {
		t.Errorf("Expected empty breadcrumb trail for kubeconfig list, got %v", appModel.breadcrumbTrail)
	}
}

func TestDuplicateNavigationPrevention(t *testing.T) {
	client := &k8s.Client{
		Namespace:      "test-namespace",
		KubeconfigPath: "/home/user/.kube/config",
	}

	appModel := &AppModel{
		kube:            *client,
		breadcrumbTrail: []string{"Resource List"},
	}

	result := appModel.isCurrentScreenResourceType("ResourceList")
	if !result {
		t.Errorf("Expected to detect that we're already on ResourceList, but got false")
	}

	result = appModel.isCurrentScreenResourceType("Pods")
	if result {
		t.Errorf("Expected to detect that we're NOT on Pods, but got true")
	}
}
