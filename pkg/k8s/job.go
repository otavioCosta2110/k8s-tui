package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/utils"
	"time"

	"gopkg.in/yaml.v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobInfo struct {
	Namespace   string
	Name        string
	Completions string
	Duration    string
	Age         string
	Raw         *batchv1.Job
	Client      Client
}

func NewJob(name, namespace string, k Client) *JobInfo {
	return &JobInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchJobList(client Client, namespace string) ([]string, error) {
	jobs, err := client.Clientset.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %v", err)
	}

	jobNames := make([]string, 0, len(jobs.Items))
	for _, job := range jobs.Items {
		jobNames = append(jobNames, job.Name)
	}

	return jobNames, nil
}

func GetJobsTableData(client Client, namespace string) ([]JobInfo, error) {
	jobs, err := client.Clientset.BatchV1().Jobs(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %v", err)
	}

	var jobInfos []JobInfo
	for _, job := range jobs.Items {
		completions := "1/1"
		if job.Spec.Completions != nil && job.Status.Succeeded != 0 {
			completions = fmt.Sprintf("%d/%d", job.Status.Succeeded, *job.Spec.Completions)
		}

		duration := "<none>"
		if job.Status.CompletionTime != nil && job.Status.StartTime != nil {
			duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time).Round(time.Second).String()
		}

		jobInfos = append(jobInfos, JobInfo{
			Namespace:   job.Namespace,
			Name:        job.Name,
			Completions: completions,
			Duration:    duration,
			Age:         utils.FormatAge(job.CreationTimestamp.Time),
			Raw:         job.DeepCopy(),
			Client:      client,
		})
	}

	return jobInfos, nil
}

func (j *JobInfo) Fetch() error {
	job, err := j.Client.Clientset.BatchV1().Jobs(j.Namespace).Get(context.Background(), j.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get job: %v", err)
	}
	j.Raw = job
	return nil
}

func (j *JobInfo) Describe() (string, error) {
	if j.Raw == nil {
		if err := j.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch job: %v", err)
		}
	}

	events, err := j.Client.Clientset.CoreV1().Events(j.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Job", j.Name, j.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get job events: %v", err)
	}

	data, err := j.DescribeJob(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe job: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (j *JobInfo) DescribeJob(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        j.Name,
		"namespace":   j.Namespace,
		"labels":      j.Raw.Labels,
		"annotations": j.Raw.Annotations,
		"created":     formatTime(j.Raw.CreationTimestamp),
	}

	if j.Raw.Spec.Completions != nil {
		desc["completions"] = *j.Raw.Spec.Completions
	}

	if j.Raw.Spec.Parallelism != nil {
		desc["parallelism"] = *j.Raw.Spec.Parallelism
	}

	desc["status"] = map[string]any{
		"active":    j.Raw.Status.Active,
		"succeeded": j.Raw.Status.Succeeded,
		"failed":    j.Raw.Status.Failed,
	}

	if j.Raw.Status.StartTime != nil {
		desc["startTime"] = formatTime(*j.Raw.Status.StartTime)
	}

	if j.Raw.Status.CompletionTime != nil {
		desc["completionTime"] = formatTime(*j.Raw.Status.CompletionTime)
	}

	if len(j.Raw.Spec.Template.Spec.Containers) > 0 {
		containers := make([]map[string]any, 0, len(j.Raw.Spec.Template.Spec.Containers))
		for _, container := range j.Raw.Spec.Template.Spec.Containers {
			containerDesc := map[string]any{
				"name":  container.Name,
				"image": container.Image,
			}
			if len(container.Command) > 0 {
				containerDesc["command"] = container.Command
			}
			if len(container.Args) > 0 {
				containerDesc["args"] = container.Args
			}
			containers = append(containers, containerDesc)
		}
		desc["containers"] = containers
	}

	if len(events.Items) > 0 {
		eventList := make([]Event, 0)
		for _, event := range events.Items {
			age := time.Since(event.LastTimestamp.Time).Round(time.Second)
			eventList = append(eventList, Event{
				Type:    event.Type,
				Reason:  event.Reason,
				Age:     age.String(),
				From:    event.Source.Component,
				Message: event.Message,
			})
		}
		desc["events"] = eventList
	}

	return desc, nil
}

func DeleteJob(client Client, namespace string, jobName string) error {
	err := client.Clientset.BatchV1().Jobs(namespace).Delete(context.Background(), jobName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete job %s: %v", jobName, err)
	}
	return nil
}
