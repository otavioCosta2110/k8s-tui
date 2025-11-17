package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"testing"
)

func TestPodGetResourceUsage(t *testing.T) {
	tests := []struct {
		name      string
		podName   string
		namespace string
	}{
		{
			name:      "Test basic pod resource usage without metrics client",
			podName:   "test-pod",
			namespace: "default",
		},
		{
			name:      "Test basic pod resource usage with nil metrics client",
			podName:   "test-pod-2",
			namespace: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with nil metrics client (should fall back to basic resource usage)
			pod := &Pod{
				Name:          tt.podName,
				Namespace:     tt.namespace,
				MetricsClient: nil, // No metrics client available
				Raw: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "test-container",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("128Mi"),
									},
								},
							},
						},
					},
				},
			}

			metrics, err := pod.GetResourceUsage()

			// Should not return error when falling back to basic resource usage
			if err != nil {
				t.Fatalf("Unexpected error getting resource usage: %v", err)
			}

			if metrics == nil {
				t.Fatal("Expected metrics, got nil")
			}

			if metrics.Name != tt.podName {
				t.Errorf("Expected pod name %s, got %s", tt.podName, metrics.Name)
			}
			if metrics.Namespace != tt.namespace {
				t.Errorf("Expected namespace %s, got %s", tt.namespace, metrics.Namespace)
			}

			// Should have one container with resource requests
			if len(metrics.Containers) != 1 {
				t.Errorf("Expected 1 container, got %d", len(metrics.Containers))
			}

			if metrics.Containers[0].Name != "test-container" {
				t.Errorf("Expected container name test-container, got %s", metrics.Containers[0].Name)
			}

			if metrics.Containers[0].CPU != "100m" {
				t.Errorf("Expected CPU 100m, got %s", metrics.Containers[0].CPU)
			}

			if metrics.Containers[0].Memory != "128.0MiB" {
				t.Errorf("Expected Memory 128.0MiB, got %s", metrics.Containers[0].Memory)
			}
		})
	}
}

func TestPodBasicResourceUsage(t *testing.T) {
	// Create a mock pod with minimal required data
	pod := &Pod{
		Name:      "test-pod",
		Namespace: "default",
		Raw: &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "test-container",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		},
	}

	metrics, err := pod.getBasicResourceUsage()
	if err != nil {
		t.Fatalf("Unexpected error getting basic resource usage: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics, got nil")
	}

	if metrics.Name != "test-pod" {
		t.Errorf("Expected pod name test-pod, got %s", metrics.Name)
	}

	if metrics.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", metrics.Namespace)
	}

	if len(metrics.Containers) != 1 {
		t.Errorf("Expected 1 container, got %d", len(metrics.Containers))
	}

	if metrics.Containers[0].Name != "test-container" {
		t.Errorf("Expected container name test-container, got %s", metrics.Containers[0].Name)
	}

	if metrics.Containers[0].CPU != "100m" {
		t.Errorf("Expected CPU 100m, got %s", metrics.Containers[0].CPU)
	}

	if metrics.Containers[0].Memory != "128.0MiB" {
		t.Errorf("Expected Memory 128.0MiB, got %s", metrics.Containers[0].Memory)
	}
}

func TestPodConvertMetricsAPIResponse(t *testing.T) {
	pod := &Pod{
		Name:      "test-pod",
		Namespace: "default",
	}

	// Create mock metrics API response
	podMetrics := &metricsv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Containers: []metricsv1beta1.ContainerMetrics{
			{
				Name: "container-1",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			},
			{
				Name: "container-2",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
	}

	result := pod.convertMetricsAPIResponse(podMetrics)

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Name != "test-pod" {
		t.Errorf("Expected pod name test-pod, got %s", result.Name)
	}

	if result.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", result.Namespace)
	}

	if len(result.Containers) != 2 {
		t.Errorf("Expected 2 containers, got %d", len(result.Containers))
	}

	// Check first container
	if result.Containers[0].Name != "container-1" {
		t.Errorf("Expected first container name container-1, got %s", result.Containers[0].Name)
	}
	if result.Containers[0].CPU != "50m" {
		t.Errorf("Expected first container CPU 50m, got %s", result.Containers[0].CPU)
	}
	if result.Containers[0].Memory != "64.0MiB" {
		t.Errorf("Expected first container Memory 64.0MiB, got %s", result.Containers[0].Memory)
	}

	// Check second container
	if result.Containers[1].Name != "container-2" {
		t.Errorf("Expected second container name container-2, got %s", result.Containers[1].Name)
	}
	if result.Containers[1].CPU != "100m" {
		t.Errorf("Expected second container CPU 100m, got %s", result.Containers[1].CPU)
	}
	if result.Containers[1].Memory != "128.0MiB" {
		t.Errorf("Expected second container Memory 128.0MiB, got %s", result.Containers[1].Memory)
	}
}
