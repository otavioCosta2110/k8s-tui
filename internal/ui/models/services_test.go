package models

import (
	"otaviocosta2110/k8s-tui/internal/k8s"
	"testing"
	"time"
)

func TestNewServices(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewServices(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeService {
		t.Error("Expected ResourceType to be ResourceTypeService")
	}
}

func TestServicesModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewServices(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.servicesInfo = []k8s.ServiceInfo{
		{
			Name:       "test-service",
			Namespace:  "default",
			Type:       "ClusterIP",
			ClusterIP:  "10.0.0.1",
			ExternalIP: "<none>",
			Ports:      "80/TCP",
			Age:        "1h",
		},
	}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 7 {
		t.Error("Expected 7 columns in row")
	}
	if rows[0][1] != "test-service" {
		t.Error("Service name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Service namespace mismatch in row")
	}
	if rows[0][2] != "ClusterIP" {
		t.Error("Service type mismatch in row")
	}
	if rows[0][3] != "10.0.0.1" {
		t.Error("Service cluster IP mismatch in row")
	}
	if rows[0][4] != "<none>" {
		t.Error("Service external IP mismatch in row")
	}
	if rows[0][5] != "80/TCP" {
		t.Error("Service ports mismatch in row")
	}
}

func TestServicesModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewServices(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestServicesModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewServices(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeService {
		t.Error("Config ResourceType not set correctly")
	}
	if model.config.Title != "Services in test-namespace" {
		t.Error("Config Title not set correctly")
	}
	if len(model.config.Columns) != 7 {
		t.Error("Expected 7 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestServicesModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewServices(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.servicesInfo = []k8s.ServiceInfo{
		{
			Name:       "test-service-1",
			Namespace:  "default",
			Type:       "ClusterIP",
			ClusterIP:  "10.0.0.1",
			ExternalIP: "<none>",
			Ports:      "80/TCP",
			Age:        "1h",
		},
		{
			Name:       "test-service-2",
			Namespace:  "default",
			Type:       "LoadBalancer",
			ClusterIP:  "10.0.0.2",
			ExternalIP: "192.168.1.100",
			Ports:      "443/TCP",
			Age:        "2h",
		},
	}

	rows := model.dataToRows()
	if len(rows) != 2 {
		t.Error("Expected 2 rows")
	}

	if rows[0][1] != "test-service-1" {
		t.Error("First service name mismatch")
	}
	if rows[0][2] != "ClusterIP" {
		t.Error("First service type mismatch")
	}

	if rows[1][1] != "test-service-2" {
		t.Error("Second service name mismatch")
	}
	if rows[1][2] != "LoadBalancer" {
		t.Error("Second service type mismatch")
	}
}

func TestServicesModelWithDifferentStates(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewServices(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	model.servicesInfo = []k8s.ServiceInfo{
		{
			Name:       "headless-service",
			Namespace:  "default",
			Type:       "ClusterIP",
			ClusterIP:  "<none>",
			ExternalIP: "<none>",
			Ports:      "",
			Age:        "30m",
		},
	}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if rows[0][3] != "<none>" {
		t.Error("Expected <none> for cluster IP")
	}
	if rows[0][4] != "<none>" {
		t.Error("Expected <none> for external IP")
	}
}

func TestNewServiceDetails(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewServiceDetails(client, "default", "test-service")

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.service.Name != "test-service" {
		t.Error("Expected service name to be 'test-service'")
	}
	if model.service.Namespace != "default" {
		t.Error("Expected service namespace to be 'default'")
	}
	if model.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.err != nil {
		t.Error("Expected error to be nil")
	}
}
