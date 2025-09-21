package k8s

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentInfo_GetLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		deployment  *appsv1.Deployment
		expectError bool
		expected    string
	}{
		{
			name:        "nil deployment",
			deployment:  nil,
			expectError: true,
		},
		{
			name: "deployment with nil selector",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.DeploymentSpec{
					Selector: nil,
				},
			},
			expectError: true,
		},
		{
			name: "deployment with simple selector",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
				},
			},
			expectError: false,
			expected:    "app=test",
		},
		{
			name: "deployment with complex selector",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":     "test",
							"version": "v1",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "environment",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"prod", "staging"},
							},
						},
					},
				},
			},
			expectError: false,
			expected:    "app=test,environment in (prod,staging),version=v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentInfo := &DeploymentInfo{
				Name:      "test",
				Namespace: "default",
				Raw:       tt.deployment,
			}

			selector, err := deploymentInfo.GetLabelSelector()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if selector != tt.expected {
					t.Errorf("Expected selector %q, got %q", tt.expected, selector)
				}
			}
		})
	}
}

func TestDeploymentInfo_Fetch(t *testing.T) {
	// Create a fake deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      3,
			ReadyReplicas: 3,
		},
	}

	// Create fake clientset
	fakeClientset := fake.NewSimpleClientset(deployment)

	// Create client
	client := Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	// Create deployment info
	deploymentInfo := &DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Client:    client,
	}

	// Test Fetch
	err := deploymentInfo.Fetch()
	if err != nil {
		t.Errorf("Fetch failed: %v", err)
	}

	// Verify Raw is populated
	if deploymentInfo.Raw == nil {
		t.Error("Expected Raw to be populated after Fetch")
	}

	if deploymentInfo.Raw.Name != "test-deployment" {
		t.Errorf("Expected name 'test-deployment', got %q", deploymentInfo.Raw.Name)
	}

	if deploymentInfo.Raw.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got %q", deploymentInfo.Raw.Namespace)
	}
}

func TestDeploymentInfo_GetPods(t *testing.T) {
	// Create a deployment with selector
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}

	// Create matching and non-matching pods
	matchingPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "matching-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	nonMatchingPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "non-matching-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "other",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	// Create fake clientset with deployment and pods
	fakeClientset := fake.NewSimpleClientset(deployment, matchingPod, nonMatchingPod)

	// Create client
	client := Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	// Create deployment info with Raw already set
	deploymentInfo := &DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Client:    client,
		Raw:       deployment,
	}

	// Test GetPods
	pods, err := deploymentInfo.GetPods()
	if err != nil {
		t.Errorf("GetPods failed: %v", err)
	}

	// Should only return the matching pod
	if len(pods) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods))
	}

	if len(pods) > 0 && pods[0].Name != "matching-pod" {
		t.Errorf("Expected pod 'matching-pod', got %q", pods[0].Name)
	}
}

func TestDeploymentInfo_GetPods_NoSelector(t *testing.T) {
	// Create a deployment without selector
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: nil, // No selector
		},
	}

	// Create client
	client := Client{
		Clientset: fake.NewSimpleClientset(deployment),
		Namespace: "default",
	}

	// Create deployment info
	deploymentInfo := &DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Client:    client,
		Raw:       deployment,
	}

	// Test GetPods - should fail
	_, err := deploymentInfo.GetPods()
	if err == nil {
		t.Error("Expected error when deployment has no selector")
	}
}

func TestDeploymentInfo_GetPods_NoRawData(t *testing.T) {
	// Create client
	client := Client{
		Clientset: fake.NewSimpleClientset(),
		Namespace: "default",
	}

	// Create deployment info without Raw data
	deploymentInfo := &DeploymentInfo{
		Name:      "test-deployment",
		Namespace: "default",
		Client:    client,
		Raw:       nil, // No raw data
	}

	// Test GetPods - should fail
	_, err := deploymentInfo.GetPods()
	if err == nil {
		t.Error("Expected error when deployment has no raw data")
	}
}

// Helper function
func int32Ptr(i int32) *int32 {
	return &i
}
