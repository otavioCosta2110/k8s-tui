package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"slices"
	"testing"
	"time"
)

func TestNewReplicaSets(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewReplicaSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeReplicaSet {
		t.Error("Expected ResourceType to be ResourceTypeReplicaSet")
	}
}

func TestReplicaSetsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewReplicaSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	replicasetsInfo := []k8s.ReplicaSetInfo{
		{
			Name:      "test-replicaset",
			Namespace: "default",
			Desired:   "3",
			Current:   "3",
			Ready:     "3",
			Age:       "2h",
		},
	}

	model.resourceData = []ResourceData{ReplicaSetData{&replicasetsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 6 {
		t.Error("Expected 6 columns in row")
	}
	if rows[0][1] != "test-replicaset" {
		t.Error("ReplicaSet name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("ReplicaSet namespace mismatch in row")
	}
	if rows[0][2] != "3" {
		t.Error("ReplicaSet desired count mismatch in row")
	}
}

func TestReplicaSetsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewReplicaSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestReplicaSetsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewReplicaSets(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeReplicaSet {
		t.Error("Config ResourceType not set correctly")
	}
	if model.config.Title != "ReplicaSets in test-namespace" {
		t.Error("Config Title not set correctly")
	}
	if len(model.config.Columns) != 6 {
		t.Error("Expected 6 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestReplicaSetsModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewReplicaSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	replicasetsInfo := []k8s.ReplicaSetInfo{
		{
			Name:      "replicaset-1",
			Namespace: "default",
			Desired:   "3",
			Current:   "3",
			Ready:     "3",
			Age:       "1h",
		},
		{
			Name:      "replicaset-2",
			Namespace: "default",
			Desired:   "5",
			Current:   "4",
			Ready:     "4",
			Age:       "2h",
		},
	}

	model.resourceData = []ResourceData{
		ReplicaSetData{&replicasetsInfo[0]},
		ReplicaSetData{&replicasetsInfo[1]},
	}

	rows := model.dataToRows()
	if len(rows) != 2 {
		t.Error("Expected 2 rows")
	}

	expectedNames := []string{"replicaset-1", "replicaset-2"}
	for i, row := range rows {
		found := slices.Contains(expectedNames, row[1])
		if !found {
			t.Errorf("ReplicaSet name %s not found in row %d", row[1], i)
		}
	}
}

func TestReplicaSetsModelWithDifferentStates(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewReplicaSets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	replicasetsInfo := []k8s.ReplicaSetInfo{
		{
			Name:      "healthy-rs",
			Namespace: "default",
			Desired:   "3",
			Current:   "3",
			Ready:     "3",
			Age:       "1h",
		},
		{
			Name:      "scaling-rs",
			Namespace: "default",
			Desired:   "5",
			Current:   "2",
			Ready:     "2",
			Age:       "30m",
		},
		{
			Name:      "failed-rs",
			Namespace: "default",
			Desired:   "2",
			Current:   "0",
			Ready:     "0",
			Age:       "10m",
		},
	}

	model.resourceData = []ResourceData{
		ReplicaSetData{&replicasetsInfo[0]},
		ReplicaSetData{&replicasetsInfo[1]},
		ReplicaSetData{&replicasetsInfo[2]},
	}

	rows := model.dataToRows()
	if len(rows) != 3 {
		t.Error("Expected 3 rows")
	}

	for _, row := range rows {
		switch row[1] {
		case "healthy-rs":
			if row[2] != "3" || row[3] != "3" || row[4] != "3" {
				t.Errorf("Healthy ReplicaSet data incorrect: %v", row)
			}
		case "scaling-rs":
			if row[2] != "5" || row[3] != "2" || row[4] != "2" {
				t.Errorf("Scaling ReplicaSet data incorrect: %v", row)
			}
		case "failed-rs":
			if row[2] != "2" || row[3] != "0" || row[4] != "0" {
				t.Errorf("Failed ReplicaSet data incorrect: %v", row)
			}
		default:
			t.Errorf("Unexpected ReplicaSet name: %s", row[1])
		}
	}
}
