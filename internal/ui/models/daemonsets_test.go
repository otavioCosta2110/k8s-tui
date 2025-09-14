package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewDaemonSets(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewDaemonSets(client, "default")
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

func TestDaemonSetsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	daemonsetsInfo := []k8s.DaemonSetInfo{
		{
			Name:         "test-daemonset",
			Namespace:    "default",
			Desired:      "3",
			Current:      "3",
			Ready:        "3",
			UpToDate:     "3",
			Available:    "3",
			NodeSelector: "ssd=true",
			Age:          "1h",
		},
	}

	model, err := NewDaemonSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.daemonsetsInfo = daemonsetsInfo

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	expectedRow := []string{"default", "test-daemonset", "3", "3", "3", "3", "3", "ssd=true", "1h"}
	for i, expected := range expectedRow {
		if rows[0][i] != expected {
			t.Errorf("Expected row[%d] to be '%s', got '%s'", i, expected, rows[0][i])
		}
	}
}

func TestDaemonSetsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewDaemonSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeDaemonSet {
		t.Error("Expected ResourceType to be ResourceTypeDaemonSet")
	}

	if model.config.Title != "DaemonSets in default" {
		t.Error("Expected title to be 'DaemonSets in default'")
	}

	if len(model.config.Columns) != 9 {
		t.Errorf("Expected 9 columns, got %d", len(model.config.Columns))
	}

	if len(model.config.ColumnWidths) != 9 {
		t.Errorf("Expected 9 column widths, got %d", len(model.config.ColumnWidths))
	}

	expectedColumns := []string{"NAMESPACE", "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "NODE SELECTOR", "AGE"}
	for i, expected := range expectedColumns {
		if model.config.Columns[i].Title != expected {
			t.Errorf("Expected column[%d] title to be '%s', got '%s'", i, expected, model.config.Columns[i].Title)
		}
	}

	if model.refreshInterval != 5*time.Second {
		t.Error("Expected refresh interval to be 5 seconds")
	}
}
