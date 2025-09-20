package k8s

import (
	"bytes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestNewPod(t *testing.T) {
	client := Client{Namespace: "default"}
	pod := NewPod("test-pod", "default", client)

	if pod.Name != "test-pod" {
		t.Error("Pod name mismatch")
	}
	if pod.Namespace != "default" {
		t.Error("Pod namespace mismatch")
	}
	if pod.Client != nil {
		t.Error("Pod client should be nil in mock setup")
	}
}

func TestFormatTime(t *testing.T) {
	testTime := metav1.Time{Time: time.Now()}
	formatted := formatTime(testTime)
	if formatted == "" {
		t.Error("Expected formatted time to be non-empty")
	}

	specificTime := metav1.Time{Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	formatted = formatTime(specificTime)
	if formatted == "" {
		t.Error("Expected formatted time to be non-empty for specific time")
	}
}

func TestBorderedWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := newBorderedWriter(&buf)
	if writer == nil {
		t.Error("Expected borderedWriter to be non-nil")
	}

	testData := []byte("test output")
	n, err := writer.Write(testData)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n <= len(testData) {
		t.Errorf("Expected to write more than %d bytes (styled output), wrote %d", len(testData), n)
	}

	if buf.Len() <= len(testData) {
		t.Error("Expected styled output to be longer than input")
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestPodWithMockData(t *testing.T) {
	client := Client{Namespace: "test-namespace"}
	pod := NewPod("mock-pod", "test-namespace", client)

	pod.Raw = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mock-pod",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	if pod.Raw.Name != "mock-pod" {
		t.Error("Pod raw data name mismatch")
	}
	if pod.Raw.Namespace != "test-namespace" {
		t.Error("Pod raw data namespace mismatch")
	}
	if pod.Raw.Status.Phase != corev1.PodRunning {
		t.Error("Pod status mismatch")
	}
}

func TestPodInfo(t *testing.T) {
	podInfo := PodInfo{
		Name:      "test-pod",
		Namespace: "default",
		Ready:     "1/1",
		Status:    "Running",
		Restarts:  0,
		Age:       "5m",
	}

	if podInfo.Name != "test-pod" {
		t.Error("PodInfo name mismatch")
	}
	if podInfo.Status != "Running" {
		t.Error("PodInfo status mismatch")
	}
	if podInfo.Restarts != 0 {
		t.Error("PodInfo restarts mismatch")
	}
}
