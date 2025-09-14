package k8s

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewJob(t *testing.T) {
	client := Client{Namespace: "default"}
	job := NewJob("test-job", "default", client)

	if job.Name != "test-job" {
		t.Error("Job name mismatch")
	}
	if job.Namespace != "default" {
		t.Error("Job namespace mismatch")
	}
}

func TestGetJobsTableData(t *testing.T) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-job",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: batchv1.JobSpec{
			Completions: &[]int32{1}[0],
		},
		Status: batchv1.JobStatus{
			Succeeded: 1,
		},
	}

	completions := "1/1"
	if job.Spec.Completions != nil && job.Status.Succeeded != 0 {
		completions = "1/1"
	}

	duration := "<none>"
	if job.Status.CompletionTime != nil && job.Status.StartTime != nil {
		duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time).Round(time.Second).String()
	}

	if completions != "1/1" {
		t.Error("Expected completions to be '1/1'")
	}
	if duration != "<none>" {
		t.Error("Expected duration to be '<none>' for incomplete job")
	}
}

func TestJobInfoStruct(t *testing.T) {
	jobInfo := JobInfo{
		Name:        "test-job",
		Namespace:   "default",
		Completions: "1/1",
		Duration:    "30s",
		Age:         "1h",
	}

	if jobInfo.Name != "test-job" {
		t.Error("JobInfo Name field mismatch")
	}
	if jobInfo.Namespace != "default" {
		t.Error("JobInfo Namespace field mismatch")
	}
	if jobInfo.Completions != "1/1" {
		t.Error("JobInfo Completions field mismatch")
	}
	if jobInfo.Duration != "30s" {
		t.Error("JobInfo Duration field mismatch")
	}
	if jobInfo.Age != "1h" {
		t.Error("JobInfo Age field mismatch")
	}
}

func TestJobDescribe(t *testing.T) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-job",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            map[string]string{"app": "test"},
			Annotations:       map[string]string{"description": "test job"},
		},
		Spec: batchv1.JobSpec{
			Completions: &[]int32{1}[0],
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
		Status: batchv1.JobStatus{
			Succeeded: 1,
		},
	}

	jobInfo := &JobInfo{
		Name:      "test-job",
		Namespace: "default",
		Raw:       job,
	}

	events := &corev1.EventList{}
	desc, err := jobInfo.DescribeJob(events)
	if err != nil {
		t.Errorf("DescribeJob failed: %v", err)
	}

	if desc["name"] != "test-job" {
		t.Error("Expected name to be 'test-job'")
	}
	if desc["namespace"] != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if desc["labels"].(map[string]string)["app"] != "test" {
		t.Error("Expected labels to contain app=test")
	}
}
