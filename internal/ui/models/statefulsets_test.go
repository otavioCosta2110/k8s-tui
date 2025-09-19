package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"testing"
	"time"
)

func TestNewStatefulSets(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewStatefulSets(client, "default")
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

func TestStatefulSetsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	statefulsetsInfo := []k8s.StatefulSetInfo{
		{
			Name:      "test-statefulset",
			Namespace: "default",
			Replicas:  "3",
			Ready:     "3/3",
			Age:       "2h",
		},
	}

	model, err := NewStatefulSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.resourceData = []ResourceData{StatefulSetData{&statefulsetsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	expectedRow := []string{"default", "test-statefulset", "3/3", "2h"}
	for i, expected := range expectedRow {
		if rows[0][i] != expected {
			t.Errorf("Expected row[%d] to be '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

func TestStatefulSetsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewStatefulSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeStatefulSet {
		t.Error("Expected ResourceType to be ResourceTypeStatefulSet")
	}

	expectedTitle := customstyles.ResourceIcons["StatefulSets"] + " StatefulSets in default"
	if model.config.Title != expectedTitle {
		t.Errorf("Expected title to be '%s', got '%s'", expectedTitle, model.config.Title)
	}

	if len(model.config.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(model.config.Columns))
	}

	if len(model.config.ColumnWidths) != 4 {
		t.Errorf("Expected 4 column widths, got %d", len(model.config.ColumnWidths))
	}

	expectedColumns := []string{"NAMESPACE", "NAME", "READY", "AGE"}
	for i, expected := range expectedColumns {
		if model.config.Columns[i].Title != expected {
			t.Errorf("Expected column[%d] title to be '%s', got '%s'", i, expected, model.config.Columns[i].Title)
		}
	}

	if model.refreshInterval != 5*time.Second {
		t.Error("Expected refresh interval to be 5 seconds")
	}
}
