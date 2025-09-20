package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"time"

	"gopkg.in/yaml.v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobInfo struct {
	Namespace    string
	Name         string
	Schedule     string
	Suspend      string
	Active       string
	LastSchedule string
	Age          string
	Raw          *batchv1.CronJob
	Client       Client
}

func NewCronJob(name, namespace string, k Client) *CronJobInfo {
	return &CronJobInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchCronJobList(client Client, namespace string) ([]string, error) {
	cronjobs, err := client.Clientset.BatchV1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cronjobs: %v", err)
	}

	cronjobNames := make([]string, 0, len(cronjobs.Items))
	for _, cronjob := range cronjobs.Items {
		cronjobNames = append(cronjobNames, cronjob.Name)
	}

	return cronjobNames, nil
}

func GetCronJobsTableData(client Client, namespace string) ([]CronJobInfo, error) {
	cronjobs, err := client.Clientset.BatchV1().CronJobs(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list cronjobs: %v", err)
	}

	var cronjobInfos []CronJobInfo
	for _, cronjob := range cronjobs.Items {
		suspend := "False"
		if cronjob.Spec.Suspend != nil && *cronjob.Spec.Suspend {
			suspend = "True"
		}

		active := fmt.Sprintf("%d", len(cronjob.Status.Active))

		lastSchedule := "<none>"
		if cronjob.Status.LastScheduleTime != nil {
			lastSchedule = utils.FormatAge(cronjob.Status.LastScheduleTime.Time)
		}

		cronjobInfos = append(cronjobInfos, CronJobInfo{
			Namespace:    cronjob.Namespace,
			Name:         cronjob.Name,
			Schedule:     cronjob.Spec.Schedule,
			Suspend:      suspend,
			Active:       active,
			LastSchedule: lastSchedule,
			Age:          utils.FormatAge(cronjob.CreationTimestamp.Time),
			Raw:          cronjob.DeepCopy(),
			Client:       client,
		})
	}

	return cronjobInfos, nil
}

func (cj *CronJobInfo) Fetch() error {
	cronjob, err := cj.Client.Clientset.BatchV1().CronJobs(cj.Namespace).Get(context.Background(), cj.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob: %v", err)
	}
	cj.Raw = cronjob
	return nil
}

func (cj *CronJobInfo) Describe() (string, error) {
	if cj.Raw == nil {
		if err := cj.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch cronjob: %v", err)
		}
	}

	events, err := cj.Client.Clientset.CoreV1().Events(cj.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=CronJob", cj.Name, cj.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get cronjob events: %v", err)
	}

	data, err := cj.DescribeCronJob(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe cronjob: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cronjob to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (cj *CronJobInfo) DescribeCronJob(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        cj.Name,
		"namespace":   cj.Namespace,
		"labels":      cj.Raw.Labels,
		"annotations": cj.Raw.Annotations,
		"created":     formatTime(cj.Raw.CreationTimestamp),
	}

	desc["schedule"] = cj.Raw.Spec.Schedule

	if cj.Raw.Spec.Suspend != nil {
		desc["suspend"] = *cj.Raw.Spec.Suspend
	}

	if cj.Raw.Spec.ConcurrencyPolicy != "" {
		desc["concurrencyPolicy"] = string(cj.Raw.Spec.ConcurrencyPolicy)
	}

	if cj.Raw.Spec.StartingDeadlineSeconds != nil {
		desc["startingDeadlineSeconds"] = *cj.Raw.Spec.StartingDeadlineSeconds
	}

	status := map[string]any{
		"active": len(cj.Raw.Status.Active),
	}

	if cj.Raw.Status.LastScheduleTime != nil {
		status["lastSchedule"] = formatTime(*cj.Raw.Status.LastScheduleTime)
	}

	desc["status"] = status

	if cj.Raw.Spec.JobTemplate.Spec.Completions != nil {
		desc["completions"] = *cj.Raw.Spec.JobTemplate.Spec.Completions
	}

	if cj.Raw.Spec.JobTemplate.Spec.Parallelism != nil {
		desc["parallelism"] = *cj.Raw.Spec.JobTemplate.Spec.Parallelism
	}

	if len(cj.Raw.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
		containers := make([]map[string]any, 0, len(cj.Raw.Spec.JobTemplate.Spec.Template.Spec.Containers))
		for _, container := range cj.Raw.Spec.JobTemplate.Spec.Template.Spec.Containers {
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

func DeleteCronJob(client Client, namespace string, cronjobName string) error {
	err := client.Clientset.BatchV1().CronJobs(namespace).Delete(context.Background(), cronjobName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete cronjob %s: %v", cronjobName, err)
	}
	return nil
}
