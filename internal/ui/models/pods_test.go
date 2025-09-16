package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewPods(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewPods(client, "default", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if len(model.podsInfo) != 2 {
		t.Error("Expected 2 pods in model")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
}

func TestPodsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}

	model, err := NewPods(client, "default", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 6 {
		t.Error("Expected 6 columns in row")
	}
	if rows[0][1] != "test-pod" {
		t.Error("Pod name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Pod namespace mismatch in row")
	}
	if rows[0][2] != "1/1" {
		t.Error("Pod ready status mismatch in row")
	}
}

func TestPodsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewPods(client, "default", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestResourceConfig(t *testing.T) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypePod,
		Title:           "Test Pods",
		RefreshInterval: 5 * time.Second,
	}

	if config.ResourceType != k8s.ResourceTypePod {
		t.Error("ResourceType mismatch")
	}
	if config.Title != "Test Pods" {
		t.Error("Title mismatch")
	}
	if config.RefreshInterval != 5*time.Second {
		t.Error("RefreshInterval mismatch")
	}
}

func TestPodsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}

	model, err := NewPods(client, "test-namespace", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypePod {
		t.Error("Config ResourceType not set correctly")
	}
	if model.config.Title != "Pods in test-namespace" {
		t.Error("Config Title not set correctly")
	}
	if len(model.config.Columns) != 6 {
		t.Error("Expected 6 columns in config")
	}
}
