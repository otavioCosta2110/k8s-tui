package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewConfigmaps(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{Name: "test-configmap", Namespace: "default", Data: "2", Age: "1h"},
	}

	model, err := NewConfigmaps(client, "default", configmaps)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if len(model.cms) != 1 {
		t.Error("Expected 1 configmap in model")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeConfigMap {
		t.Error("Expected ResourceType to be ResourceTypeConfigMap")
	}
}

func TestConfigmapsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{
			Name:      "test-configmap",
			Namespace: "default",
			Data:      "2",
			Age:       "1h",
		},
	}

	model, err := NewConfigmaps(client, "default", configmaps)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 4 {
		t.Error("Expected 4 columns in row")
	}
	if rows[0][1] != "test-configmap" {
		t.Error("ConfigMap name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("ConfigMap namespace mismatch in row")
	}
	if rows[0][2] != "2" {
		t.Error("ConfigMap data count mismatch in row")
	}
}

func TestConfigmapsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewConfigmaps(client, "default", []k8s.Configmap{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestConfigmapsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	configmaps := []k8s.Configmap{
		{Name: "test-configmap", Namespace: "test-namespace", Data: "1", Age: "30m"},
	}

	model, err := NewConfigmaps(client, "test-namespace", configmaps)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeConfigMap {
		t.Error("Config ResourceType not set correctly")
	}
	if model.config.Title != "ConfigMaps in test-namespace" {
		t.Error("Config Title not set correctly")
	}
	if len(model.config.Columns) != 4 {
		t.Error("Expected 4 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestConfigmapsModelColumnWidths(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{Name: "test-configmap", Namespace: "default", Data: "1", Age: "1h"},
	}

	model, err := NewConfigmaps(client, "default", configmaps)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedWidths := []float64{0.30, 0.30, 0.16, 0.20}
	if len(model.config.ColumnWidths) != len(expectedWidths) {
		t.Error("ColumnWidths length mismatch")
	}

	for i, expected := range expectedWidths {
		if model.config.ColumnWidths[i] != expected {
			t.Errorf("ColumnWidth[%d] expected %f, got %f", i, expected, model.config.ColumnWidths[i])
		}
	}
}

func TestConfigmapsModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{Name: "configmap-1", Namespace: "default", Data: "2", Age: "1h"},
		{Name: "configmap-2", Namespace: "default", Data: "1", Age: "2h"},
		{Name: "configmap-3", Namespace: "default", Data: "5", Age: "30m"},
	}

	model, err := NewConfigmaps(client, "default", configmaps)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(model.cms) != 3 {
		t.Error("Expected 3 configmaps in model")
	}

	rows := model.dataToRows()
	if len(rows) != 3 {
		t.Error("Expected 3 rows")
	}

	expectedNames := []string{"configmap-1", "configmap-2", "configmap-3"}
	for i, row := range rows {
		found := false
		for _, expectedName := range expectedNames {
			if row[1] == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ConfigMap name %s not found in row %d", row[1], i)
		}
	}
}
