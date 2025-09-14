package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewCronJobs(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewCronJobs(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
}

func TestCronJobsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	cronjobsInfo := []k8s.CronJobInfo{
		{
			Name:         "test-cronjob",
			Namespace:    "default",
			Schedule:     "*/5 * * * *",
			Suspend:      "False",
			Active:       "1",
			LastSchedule: "1h ago",
			Age:          "2h",
		},
	}

	model, err := NewCronJobs(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.cronjobsInfo = cronjobsInfo

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	expectedRow := []string{"default", "test-cronjob", "*/5 * * * *", "False", "1", "1h ago", "2h"}
	for i, expected := range expectedRow {
		if rows[0][i] != expected {
			t.Errorf("Expected row[%d] to be '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

func TestCronJobsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewCronJobs(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeCronJob {
		t.Error("Expected ResourceType to be ResourceTypeCronJob")
	}

	if model.config.Title != "CronJobs in default" {
		t.Error("Expected title to be 'CronJobs in default'")
	}

	if len(model.config.Columns) != 7 {
		t.Errorf("Expected 7 columns, got %d", len(model.config.Columns))
	}

	if len(model.config.ColumnWidths) != 7 {
		t.Errorf("Expected 7 column widths, got %d", len(model.config.ColumnWidths))
	}

	expectedColumns := []string{"NAMESPACE", "NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST SCHEDULE", "AGE"}
	for i, expected := range expectedColumns {
		if model.config.Columns[i].Title != expected {
			t.Errorf("Expected column[%d] title to be '%s', got '%s'", i, expected, model.config.Columns[i].Title)
		}
	}

	if model.refreshInterval != 5*time.Second {
		t.Error("Expected refresh interval to be 5 seconds")
	}
}
