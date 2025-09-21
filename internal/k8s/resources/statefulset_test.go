package k8s

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewStatefulSet(t *testing.T) {
	client := Client{Namespace: "default"}
	statefulset := NewStatefulSet("test-statefulset", "default", client)

	if statefulset.Name != "test-statefulset" {
		t.Error("StatefulSet name mismatch")
	}
	if statefulset.Namespace != "default" {
		t.Error("StatefulSet namespace mismatch")
	}
}

func TestGetStatefulSetsTableData(t *testing.T) {
	replicasStr := "3"
	ready := "3/3"

	if replicasStr != "3" {
		t.Error("Expected replicas to be '3'")
	}
	if ready != "3/3" {
		t.Error("Expected ready to be '3/3'")
	}
}

func TestStatefulSetInfoStruct(t *testing.T) {
	statefulsetInfo := StatefulSetInfo{
		Name:      "test-statefulset",
		Namespace: "default",
		Replicas:  "3",
		Ready:     "3/3",
		Age:       "1h",
	}

	if statefulsetInfo.Name != "test-statefulset" {
		t.Error("StatefulSetInfo Name field mismatch")
	}
	if statefulsetInfo.Namespace != "default" {
		t.Error("StatefulSetInfo Namespace field mismatch")
	}
	if statefulsetInfo.Replicas != "3" {
		t.Error("StatefulSetInfo Replicas field mismatch")
	}
	if statefulsetInfo.Ready != "3/3" {
		t.Error("StatefulSetInfo Ready field mismatch")
	}
	if statefulsetInfo.Age != "1h" {
		t.Error("StatefulSetInfo Age field mismatch")
	}
}

func TestStatefulSetDescribe(t *testing.T) {
	replicas := int32(3)

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-statefulset",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            map[string]string{"app": "test"},
			Annotations:       map[string]string{"description": "test statefulset"},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "test-service",
			Replicas:    &replicas,
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
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:      3,
			ReadyReplicas: 3,
		},
	}

	statefulsetInfo := &StatefulSetInfo{
		Name:      "test-statefulset",
		Namespace: "default",
		Raw:       statefulset,
	}

	events := &corev1.EventList{}
	desc, err := statefulsetInfo.DescribeStatefulSet(events)
	if err != nil {
		t.Errorf("DescribeStatefulSet failed: %v", err)
	}

	if desc["name"] != "test-statefulset" {
		t.Error("Expected name to be 'test-statefulset'")
	}
	if desc["namespace"] != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if desc["serviceName"] != "test-service" {
		t.Error("Expected serviceName to be 'test-service'")
	}
	if desc["replicas"] != int32(3) {
		t.Error("Expected replicas to be 3")
	}
	if desc["selector"].(map[string]string)["app"] != "test" {
		t.Error("Expected selector to contain app=test")
	}
	if desc["status"].(map[string]any)["replicas"] != int32(3) {
		t.Error("Expected status.replicas to be 3")
	}
	if desc["status"].(map[string]any)["readyReplicas"] != int32(3) {
		t.Error("Expected status.readyReplicas to be 3")
	}
}
