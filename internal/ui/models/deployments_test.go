package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewDeployments(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeDeployment {
		t.Error("Expected ResourceType to be ResourceTypeDeployment")
	}
}

func TestDeploymentsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Set mock deployment data
	model.deploymentsInfo = []k8s.DeploymentInfo{
		{
			Name:      "test-deployment",
			Namespace: "default",
			Ready:     "1/1",
			UpToDate:  "1",
			Available: "1",
			Age:       "1h",
		},
	}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 6 {
		t.Error("Expected 6 columns in row")
	}
	if rows[0][1] != "test-deployment" {
		t.Error("Deployment name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Deployment namespace mismatch in row")
	}
	if rows[0][2] != "1/1" {
		t.Error("Deployment ready status mismatch in row")
	}
}

func TestDeploymentsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestDeploymentsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewDeployments(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test that the config is properly set
	if model.config.ResourceType != k8s.ResourceTypeDeployment {
		t.Error("Config ResourceType not set correctly")
	}
	if model.config.Title != "Deployments in test-namespace" {
		t.Error("Config Title not set correctly")
	}
	if len(model.config.Columns) != 6 {
		t.Error("Expected 6 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestDeploymentsModelColumnWidths(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewDeployments(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedWidths := []float64{0.15, 0.25, 0.15, 0.15, 0.09, 0.15}
	if len(model.config.ColumnWidths) != len(expectedWidths) {
		t.Error("ColumnWidths length mismatch")
	}

	for i, expected := range expectedWidths {
		if model.config.ColumnWidths[i] != expected {
			t.Errorf("ColumnWidth[%d] expected %f, got %f", i, expected, model.config.ColumnWidths[i])
		}
	}
}
