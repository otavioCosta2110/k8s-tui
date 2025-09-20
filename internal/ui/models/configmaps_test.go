package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"
	"slices"
	"testing"
	"time"
)

func TestNewConfigmaps(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{Name: "test-configmap", Namespace: "default", Data: "2", Age: "1h"},
	}

	model, err := NewConfigmaps(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	model.cms = configmaps
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

	model, err := NewConfigmaps(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.resourceData = []types.ResourceData{ConfigMapData{&configmaps[0]}}

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
	model, err := NewConfigmaps(client, "default")
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

	model, err := NewConfigmaps(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeConfigMap {
		t.Error("Config ResourceType not set correctly")
	}
	expectedTitle := customstyles.ResourceIcons["ConfigMaps"] + " ConfigMaps in test-namespace"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 4 {
		t.Error("Expected 4 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestConfigmapsModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	configmaps := []k8s.Configmap{
		{Name: "configmap-1", Namespace: "default", Data: "2", Age: "1h"},
		{Name: "configmap-2", Namespace: "default", Data: "1", Age: "2h"},
		{Name: "configmap-3", Namespace: "default", Data: "5", Age: "30m"},
	}

	model, err := NewConfigmaps(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.resourceData = []types.ResourceData{
		ConfigMapData{&configmaps[0]},
		ConfigMapData{&configmaps[1]},
		ConfigMapData{&configmaps[2]},
	}

	rows := model.dataToRows()
	if len(rows) != 3 {
		t.Error("Expected 3 rows")
	}

	expectedNames := []string{"configmap-1", "configmap-2", "configmap-3"}
	for i, row := range rows {
		found := slices.Contains(expectedNames, row[1])
		if !found {
			t.Errorf("ConfigMap name %s not found in row %d", row[1], i)
		}
	}
}
