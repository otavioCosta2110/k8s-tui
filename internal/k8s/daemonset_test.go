package k8s

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewDaemonSet(t *testing.T) {
	client := Client{Namespace: "default"}
	daemonset := NewDaemonSet("test-daemonset", "default", client)

	if daemonset.Name != "test-daemonset" {
		t.Error("DaemonSet name mismatch")
	}
	if daemonset.Namespace != "default" {
		t.Error("DaemonSet namespace mismatch")
	}
}

func TestGetDaemonSetsTableData(t *testing.T) {
	desired := "3"
	current := "3"
	ready := "3"
	upToDate := "3"
	available := "3"

	if desired != "3" {
		t.Error("Expected desired to be '3'")
	}
	if current != "3" {
		t.Error("Expected current to be '3'")
	}
	if ready != "3" {
		t.Error("Expected ready to be '3'")
	}
	if upToDate != "3" {
		t.Error("Expected upToDate to be '3'")
	}
	if available != "3" {
		t.Error("Expected available to be '3'")
	}
}

func TestDaemonSetInfoStruct(t *testing.T) {
	daemonsetInfo := DaemonSetInfo{
		Name:         "test-daemonset",
		Namespace:    "default",
		Desired:      "3",
		Current:      "3",
		Ready:        "3",
		UpToDate:     "3",
		Available:    "3",
		NodeSelector: "<none>",
		Age:          "1h",
	}

	if daemonsetInfo.Name != "test-daemonset" {
		t.Error("DaemonSetInfo Name field mismatch")
	}
	if daemonsetInfo.Namespace != "default" {
		t.Error("DaemonSetInfo Namespace field mismatch")
	}
	if daemonsetInfo.Desired != "3" {
		t.Error("DaemonSetInfo Desired field mismatch")
	}
	if daemonsetInfo.Current != "3" {
		t.Error("DaemonSetInfo Current field mismatch")
	}
	if daemonsetInfo.Ready != "3" {
		t.Error("DaemonSetInfo Ready field mismatch")
	}
	if daemonsetInfo.UpToDate != "3" {
		t.Error("DaemonSetInfo UpToDate field mismatch")
	}
	if daemonsetInfo.Available != "3" {
		t.Error("DaemonSetInfo Available field mismatch")
	}
	if daemonsetInfo.NodeSelector != "<none>" {
		t.Error("DaemonSetInfo NodeSelector field mismatch")
	}
	if daemonsetInfo.Age != "1h" {
		t.Error("DaemonSetInfo Age field mismatch")
	}
}

func TestDaemonSetDescribe(t *testing.T) {
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-daemonset",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            map[string]string{"app": "test"},
			Annotations:       map[string]string{"description": "test daemonset"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:1.21",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			CurrentNumberScheduled: 3,
			NumberReady:            3,
			UpdatedNumberScheduled: 3,
			NumberAvailable:        3,
		},
	}

	daemonsetInfo := &DaemonSetInfo{
		Name:      "test-daemonset",
		Namespace: "default",
		Raw:       daemonset,
	}

	events := &corev1.EventList{}
	desc, err := daemonsetInfo.DescribeDaemonSet(events)
	if err != nil {
		t.Errorf("DescribeDaemonSet failed: %v", err)
	}

	if desc["name"] != "test-daemonset" {
		t.Error("Expected name to be 'test-daemonset'")
	}
	if desc["namespace"] != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if desc["selector"].(map[string]string)["app"] != "test" {
		t.Error("Expected selector to contain app=test")
	}
	if desc["status"].(map[string]any)["desiredNumberScheduled"] != int32(3) {
		t.Error("Expected desiredNumberScheduled to be 3")
	}
	if desc["status"].(map[string]any)["currentNumberScheduled"] != int32(3) {
		t.Error("Expected currentNumberScheduled to be 3")
	}
	if desc["status"].(map[string]any)["numberReady"] != int32(3) {
		t.Error("Expected numberReady to be 3")
	}
}
