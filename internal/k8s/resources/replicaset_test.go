package k8s

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReplicaSetInfo_GetLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		replicaSet  *appsv1.ReplicaSet
		expectError bool
		expected    string
	}{
		{
			name:        "nil replicaset",
			replicaSet:  nil,
			expectError: true,
		},
		{
			name: "replicaset with nil selector",
			replicaSet: &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.ReplicaSetSpec{
					Selector: nil,
				},
			},
			expectError: true,
		},
		{
			name: "replicaset with simple selector",
			replicaSet: &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.ReplicaSetSpec{
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
			name: "replicaset with complex selector",
			replicaSet: &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: appsv1.ReplicaSetSpec{
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
			replicaSetInfo := &ReplicaSetInfo{
				Name:      "test",
				Namespace: "default",
				Raw:       tt.replicaSet,
			}

			selector, err := replicaSetInfo.GetLabelSelector()

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

func TestReplicaSetInfo_Fetch(t *testing.T) {
	// Create a fake replicaset
	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-replicaset",
			Namespace: "default",
		},
		Spec: appsv1.ReplicaSetSpec{
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
		Status: appsv1.ReplicaSetStatus{
			Replicas:      3,
			ReadyReplicas: 3,
		},
	}

	// Create fake clientset
	fakeClientset := fake.NewSimpleClientset(replicaSet)

	// Create client
	client := Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	// Create replicaset info
	replicaSetInfo := &ReplicaSetInfo{
		Name:      "test-replicaset",
		Namespace: "default",
		Client:    client,
	}

	// Test Fetch
	err := replicaSetInfo.Fetch()
	if err != nil {
		t.Errorf("Fetch failed: %v", err)
	}

	// Verify Raw is populated
	if replicaSetInfo.Raw == nil {
		t.Error("Expected Raw to be populated after Fetch")
	}

	if replicaSetInfo.Raw.Name != "test-replicaset" {
		t.Errorf("Expected name 'test-replicaset', got %q", replicaSetInfo.Raw.Name)
	}

	if replicaSetInfo.Raw.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got %q", replicaSetInfo.Raw.Namespace)
	}
}

func TestReplicaSetInfo_GetPods(t *testing.T) {
	// Create a replicaset with selector
	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-replicaset",
			Namespace: "default",
		},
		Spec: appsv1.ReplicaSetSpec{
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

	// Create fake clientset with replicaset and pods
	fakeClientset := fake.NewSimpleClientset(replicaSet, matchingPod, nonMatchingPod)

	// Create client
	client := Client{
		Clientset: fakeClientset,
		Namespace: "default",
	}

	// Create replicaset info with Raw already set
	replicaSetInfo := &ReplicaSetInfo{
		Name:      "test-replicaset",
		Namespace: "default",
		Client:    client,
		Raw:       replicaSet,
	}

	// Test GetPods
	pods, err := replicaSetInfo.GetPods()
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

func TestReplicaSetInfo_GetPods_NoSelector(t *testing.T) {
	// Create a replicaset without selector
	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-replicaset",
			Namespace: "default",
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: nil, // No selector
		},
	}

	// Create client
	client := Client{
		Clientset: fake.NewSimpleClientset(replicaSet),
		Namespace: "default",
	}

	// Create replicaset info
	replicaSetInfo := &ReplicaSetInfo{
		Name:      "test-replicaset",
		Namespace: "default",
		Client:    client,
		Raw:       replicaSet,
	}

	// Test GetPods - should fail
	_, err := replicaSetInfo.GetPods()
	if err == nil {
		t.Error("Expected error when replicaset has no selector")
	}
}

func TestReplicaSetInfo_GetPods_NoRawData(t *testing.T) {
	// Create client
	client := Client{
		Clientset: fake.NewSimpleClientset(),
		Namespace: "default",
	}

	// Create replicaset info without Raw data
	replicaSetInfo := &ReplicaSetInfo{
		Name:      "test-replicaset",
		Namespace: "default",
		Client:    client,
		Raw:       nil, // No raw data
	}

	// Test GetPods - should fail
	_, err := replicaSetInfo.GetPods()
	if err == nil {
		t.Error("Expected error when replicaset has no raw data")
	}
}
