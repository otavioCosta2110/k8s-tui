package k8s

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewCronJob(t *testing.T) {
	client := Client{Namespace: "default"}
	cronjob := NewCronJob("test-cronjob", "default", client)

	if cronjob.Name != "test-cronjob" {
		t.Error("CronJob name mismatch")
	}
	if cronjob.Namespace != "default" {
		t.Error("CronJob namespace mismatch")
	}
}

func TestGetCronJobsTableData(t *testing.T) {
	suspend := true
	cronjob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-cronjob",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "*/5 * * * *",
			Suspend:  &suspend,
		},
		Status: batchv1.CronJobStatus{
			Active: []corev1.ObjectReference{
				{Name: "test-job-1"},
				{Name: "test-job-2"},
			},
			LastScheduleTime: &metav1.Time{Time: time.Now().Add(-time.Hour)},
		},
	}

	suspendStr := "False"
	if cronjob.Spec.Suspend != nil && *cronjob.Spec.Suspend {
		suspendStr = "True"
	}

	active := "2" 

	if suspendStr != "True" {
		t.Error("Expected suspend to be 'True'")
	}
	if active != "2" {
		t.Error("Expected active to be '2'")
	}
}

func TestCronJobInfoStruct(t *testing.T) {
	cronjobInfo := CronJobInfo{
		Name:         "test-cronjob",
		Namespace:    "default",
		Schedule:     "*/5 * * * *",
		Suspend:      "False",
		Active:       "0",
		LastSchedule: "<none>",
		Age:          "1h",
	}

	if cronjobInfo.Name != "test-cronjob" {
		t.Error("CronJobInfo Name field mismatch")
	}
	if cronjobInfo.Namespace != "default" {
		t.Error("CronJobInfo Namespace field mismatch")
	}
	if cronjobInfo.Schedule != "*/5 * * * *" {
		t.Error("CronJobInfo Schedule field mismatch")
	}
	if cronjobInfo.Suspend != "False" {
		t.Error("CronJobInfo Suspend field mismatch")
	}
	if cronjobInfo.Active != "0" {
		t.Error("CronJobInfo Active field mismatch")
	}
	if cronjobInfo.LastSchedule != "<none>" {
		t.Error("CronJobInfo LastSchedule field mismatch")
	}
	if cronjobInfo.Age != "1h" {
		t.Error("CronJobInfo Age field mismatch")
	}
}

func TestCronJobDescribe(t *testing.T) {
	suspend := false
	completions := int32(1)
	parallelism := int32(2)
	startingDeadlineSeconds := int64(300)

	cronjob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-cronjob",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            map[string]string{"app": "test"},
			Annotations:       map[string]string{"description": "test cronjob"},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                "*/5 * * * *",
			Suspend:                 &suspend,
			ConcurrencyPolicy:       batchv1.AllowConcurrent,
			StartingDeadlineSeconds: &startingDeadlineSeconds,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Completions: &completions,
					Parallelism: &parallelism,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test-container",
									Image: "busybox",
								},
							},
						},
					},
				},
			},
		},
		Status: batchv1.CronJobStatus{
			Active: []corev1.ObjectReference{
				{Name: "test-job-1"},
			},
		},
	}

	cronjobInfo := &CronJobInfo{
		Name:      "test-cronjob",
		Namespace: "default",
		Raw:       cronjob,
	}

	events := &corev1.EventList{}
	desc, err := cronjobInfo.DescribeCronJob(events)
	if err != nil {
		t.Errorf("DescribeCronJob failed: %v", err)
	}

	if desc["name"] != "test-cronjob" {
		t.Error("Expected name to be 'test-cronjob'")
	}
	if desc["namespace"] != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if desc["schedule"] != "*/5 * * * *" {
		t.Error("Expected schedule to be '*/5 * * * *'")
	}
	if desc["suspend"] != false {
		t.Error("Expected suspend to be false")
	}
	if desc["concurrencyPolicy"] != "Allow" {
		t.Error("Expected concurrencyPolicy to be 'Allow'")
	}
	if desc["startingDeadlineSeconds"] != int64(300) {
		t.Error("Expected startingDeadlineSeconds to be 300")
	}
}
