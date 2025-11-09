package models

import (
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
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
	if len(model.resourceData) != 0 {
		t.Error("Expected resourceData to be empty initially")
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

	podInfo := k8s.PodInfo{
		Name:      "test-pod",
		Namespace: "default",
		Ready:     "1/1",
		Status:    "Running",
		Restarts:  0,
		Age:       "5m",
		Image:     "nginx:latest",
	}
	model.resourceData = []types.ResourceData{PodData{&podInfo}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 7 {
		t.Errorf("Expected 7 columns in row, got %d", len(rows[0]))
	}
	if rows[0][0] != "default" {
		t.Errorf("Pod namespace mismatch in row: expected 'default', got '%s'", rows[0][0])
	}
	if rows[0][1] != "test-pod" {
		t.Errorf("Pod name mismatch in row: expected 'test-pod', got '%s'", rows[0][1])
	}

	if rows[0][2] != "nginx:latest" {
		t.Errorf("Pod image mismatch in row: expected 'nginx:latest', got '%s'", rows[0][2])
	}
	if rows[0][3] != "1/1" {
		t.Errorf("Pod ready status mismatch in row: expected '1/1', got '%s'", rows[0][3])
	}
	if rows[0][4] != "Running" {
		t.Errorf("Pod status mismatch in row: expected 'Running', got '%s'", rows[0][4])
	}
	if rows[0][5] != "0" {
		t.Errorf("Pod restarts mismatch in row: expected '0', got '%s'", rows[0][5])
	}
	if rows[0][6] != "5m" {
		t.Errorf("Pod age mismatch in row: expected '5m', got '%s'", rows[0][6])
	}
}

func TestPodsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewPods(client, "default", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.resourceData = []types.ResourceData{}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for empty data, got %d", len(rows))
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

func TestPodDataGetColumns(t *testing.T) {
	podInfo := k8s.PodInfo{
		Name:      "test-pod",
		Namespace: "default",
		Ready:     "1/1",
		Status:    "Running",
		Restarts:  3,
		Age:       "10m",
		Image:     "nginx:1.21",
	}
	podData := PodData{&podInfo}

	columns := podData.GetColumns()
	if len(columns) != 7 {
		t.Errorf("Expected 7 columns, got %d", len(columns))
	}
	expected := []string{"default", "test-pod", "nginx:1.21", "1/1", "Running", "3", "10m"}
	for i, exp := range expected {
		if columns[i] != exp {
			t.Errorf("Column %d mismatch: expected '%s', got '%s'", i, exp, columns[i])
		}
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
	expectedTitle := customstyles.ResourceIcons["Pods"] + " Pods in test-namespace"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 7 {
		t.Error("Expected 7 columns in config")
	}
}
