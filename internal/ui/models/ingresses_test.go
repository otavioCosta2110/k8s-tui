package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"
	"testing"
	"time"
)

func TestNewIngresses(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeIngress {
		t.Error("Expected ResourceType to be ResourceTypeIngress")
	}
}

func TestIngressesModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	ingressesInfo := []k8s.IngressInfo{
		{
			Name:      "test-ingress",
			Namespace: "default",
			Class:     "nginx",
			Hosts:     "example.com",
			Address:   "192.168.1.100",
			Ports:     "80",
			Age:       "1h",
		},
	}

	model.resourceData = []types.ResourceData{IngressData{&ingressesInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 7 {
		t.Error("Expected 7 columns in row")
	}
	if rows[0][1] != "test-ingress" {
		t.Error("Ingress name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Ingress namespace mismatch in row")
	}
	if rows[0][2] != "nginx" {
		t.Error("Ingress class mismatch in row")
	}
	if rows[0][3] != "example.com" {
		t.Error("Ingress hosts mismatch in row")
	}
	if rows[0][4] != "192.168.1.100" {
		t.Error("Ingress address mismatch in row")
	}
	if rows[0][5] != "80" {
		t.Error("Ingress ports mismatch in row")
	}
}

func TestIngressesModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestIngressesModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewIngresses(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeIngress {
		t.Error("Config ResourceType not set correctly")
	}
	expectedTitle := customstyles.ResourceIcons["Ingresses"] + " Ingresses in test-namespace"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 7 {
		t.Error("Expected 7 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestIngressesModelColumnWidths(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedWidths := []float64{0.13, 0.23, 0.13, 0.13, 0.13, 0.13, 0.03}
	if len(model.config.ColumnWidths) != len(expectedWidths) {
		t.Error("ColumnWidths length mismatch")
	}

	for i, expected := range expectedWidths {
		if model.config.ColumnWidths[i] != expected {
			t.Errorf("ColumnWidth[%d] expected %f, got %f", i, expected, model.config.ColumnWidths[i])
		}
	}
}

func TestIngressesModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	ingressesInfo := []k8s.IngressInfo{
		{
			Name:      "test-ingress-1",
			Namespace: "default",
			Class:     "nginx",
			Hosts:     "example.com",
			Address:   "192.168.1.100",
			Ports:     "80",
			Age:       "1h",
		},
		{
			Name:      "test-ingress-2",
			Namespace: "default",
			Class:     "traefik",
			Hosts:     "api.example.com",
			Address:   "192.168.1.101",
			Ports:     "443",
			Age:       "2h",
		},
	}

	model.resourceData = []types.ResourceData{
		IngressData{&ingressesInfo[0]},
		IngressData{&ingressesInfo[1]},
	}

	rows := model.dataToRows()
	if len(rows) != 2 {
		t.Error("Expected 2 rows")
	}

	if rows[0][1] != "test-ingress-1" {
		t.Error("First ingress name mismatch")
	}
	if rows[0][2] != "nginx" {
		t.Error("First ingress class mismatch")
	}

	if rows[1][1] != "test-ingress-2" {
		t.Error("Second ingress name mismatch")
	}
	if rows[1][2] != "traefik" {
		t.Error("Second ingress class mismatch")
	}
}

func TestIngressesModelWithDifferentStates(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewIngresses(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	ingressesInfo := []k8s.IngressInfo{
		{
			Name:      "no-class-ingress",
			Namespace: "default",
			Class:     "",
			Hosts:     "",
			Address:   "",
			Ports:     "",
			Age:       "30m",
		},
	}

	model.resourceData = []types.ResourceData{IngressData{&ingressesInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if rows[0][2] != "" {
		t.Error("Expected empty class for ingress with no class")
	}
	if rows[0][3] != "" {
		t.Error("Expected empty hosts for ingress with no hosts")
	}
}

func TestNewIngressDetails(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewIngressDetails(client, "default", "test-ingress")

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.ingress.Name != "test-ingress" {
		t.Error("Expected ingress name to be 'test-ingress'")
	}
	if model.ingress.Namespace != "default" {
		t.Error("Expected ingress namespace to be 'default'")
	}
	if model.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.err != nil {
		t.Error("Expected error to be nil")
	}
}
