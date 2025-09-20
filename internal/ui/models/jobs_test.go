package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"testing"
	"time"
)

func TestNewJobs(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewJobs(client, "default")
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

func TestJobsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	jobsInfo := []k8s.JobInfo{
		{
			Name:        "test-job",
			Namespace:   "default",
			Completions: "1/1",
			Duration:    "30s",
			Age:         "1h",
		},
	}

	model, err := NewJobs(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.resourceData = []types.ResourceData{JobData{&jobsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	expectedRow := []string{"default", "test-job", "1/1", "30s", "1h"}
	for i, expected := range expectedRow {
		if rows[0][i] != expected {
			t.Errorf("Expected row[%d] to be '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

func TestJobsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewJobs(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeJob {
		t.Error("Expected ResourceType to be ResourceTypeJob")
	}

	expectedTitle := customstyles.ResourceIcons["Jobs"] + " Jobs in default"
	if model.config.Title != expectedTitle {
		t.Errorf("Expected title to be '%s', got '%s'", expectedTitle, model.config.Title)
	}

	if len(model.config.Columns) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(model.config.Columns))
	}

	if len(model.config.ColumnWidths) != 5 {
		t.Errorf("Expected 5 column widths, got %d", len(model.config.ColumnWidths))
	}

	expectedColumns := []string{"NAMESPACE", "NAME", "COMPLETIONS", "DURATION", "AGE"}
	for i, expected := range expectedColumns {
		if model.config.Columns[i].Title != expected {
			t.Errorf("Expected column[%d] title to be '%s', got '%s'", i, expected, model.config.Columns[i].Title)
		}
	}

	if model.refreshInterval != 5*time.Second {
		t.Error("Expected refresh interval to be 5 seconds")
	}
}
