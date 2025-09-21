package models

import (
	"testing"
	"time"

	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewDeployments(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeDeployment {
		t.Error("Expected ResourceType to be ResourceTypeDeployment")
	}
}

func TestDeploymentsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	deploymentsInfo := []k8s.DeploymentInfo{
		{
			Name:      "test-deployment",
			Namespace: "default",
			Ready:     "1/1",
			UpToDate:  "1",
			Available: "1",
			Age:       "1h",
		},
	}

	model.resourceData = []types.ResourceData{DeploymentData{&deploymentsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 6 {
		t.Error("Expected 6 columns in row")
	}
	if rows[0][1] != "test-deployment" {
		t.Error("Deployment name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Deployment namespace mismatch in row")
	}
	if rows[0][2] != "1/1" {
		t.Error("Deployment ready status mismatch in row")
	}
}

func TestDeploymentsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestDeploymentSelectionWithPodFiltering(t *testing.T) {

	matchingPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	nonMatchingPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "default",
			Labels: map[string]string{
				"app": "other-app",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	deploymentRaw := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
	}

	fakeClientset := fake.NewSimpleClientset(deploymentRaw, matchingPod, nonMatchingPod)

	client := k8s.Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	deployment := &k8s.DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Client:    client,
		Raw:       deploymentRaw,
	}

	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Fatalf("Failed to create deployment model: %v", err)
	}

	model.deploymentsInfo = []k8s.DeploymentInfo{*deployment}
	model.resourceData = []types.ResourceData{DeploymentData{deployment}}


	selector, err := deployment.GetLabelSelector()
	if err != nil {
		t.Fatalf("GetLabelSelector failed: %v", err)
	}

	expectedSelector := "app=test-app"
	if selector != expectedSelector {
		t.Errorf("Expected selector %q, got %q", expectedSelector, selector)
	}

	pods, err := deployment.GetPods()
	if err != nil {
		t.Fatalf("GetPods failed: %v", err)
	}

	if len(pods) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods))
	}

	if len(pods) > 0 && pods[0].Name != "pod-1" {
		t.Errorf("Expected pod 'pod-1', got %q", pods[0].Name)
	}
}

func TestDeploymentsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewDeployments(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeDeployment {
		t.Error("Config ResourceType not set correctly")
	}
	expectedTitle := customstyles.ResourceIcons["Deployments"] + " Deployments in test-namespace"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 6 {
		t.Error("Expected 6 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}
