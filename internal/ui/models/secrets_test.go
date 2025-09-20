package models

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSecrets(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewSecrets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.namespace != "default" {
		t.Error("Expected namespace to be 'default'")
	}
	if model.config.ResourceType != k8s.ResourceTypeSecret {
		t.Error("Expected ResourceType to be ResourceTypeSecret")
	}
}

func TestSecretsModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewSecrets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	secretsInfo := []k8s.SecretInfo{
		{
			Name:      "test-secret",
			Namespace: "default",
			Type:      "Opaque",
			Data:      "2",
			Age:       "1h",
		},
	}

	model.resourceData = []types.ResourceData{SecretData{&secretsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 5 {
		t.Error("Expected 5 columns in row")
	}
	if rows[0][1] != "test-secret" {
		t.Error("Secret name mismatch in row")
	}
	if rows[0][0] != "default" {
		t.Error("Secret namespace mismatch in row")
	}
	if rows[0][2] != "Opaque" {
		t.Error("Secret type mismatch in row")
	}
	if rows[0][3] != "2" {
		t.Error("Secret data count mismatch in row")
	}
}

func TestSecretsModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewSecrets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestSecretsModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "test-namespace"}
	model, err := NewSecrets(client, "test-namespace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeSecret {
		t.Error("Config ResourceType not set correctly")
	}
	expectedTitle := customstyles.ResourceIcons["Secrets"] + " Secrets in test-namespace"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 5 {
		t.Error("Expected 5 columns in config")
	}
	if model.config.RefreshInterval != 5*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestSecretsModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewSecrets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	secretsInfo := []k8s.SecretInfo{
		{
			Name:      "test-secret-1",
			Namespace: "default",
			Type:      "Opaque",
			Data:      "2",
			Age:       "1h",
		},
		{
			Name:      "test-secret-2",
			Namespace: "default",
			Type:      "kubernetes.io/tls",
			Data:      "3",
			Age:       "2h",
		},
	}

	model.resourceData = []types.ResourceData{
		SecretData{&secretsInfo[0]},
		SecretData{&secretsInfo[1]},
	}

	rows := model.dataToRows()
	if len(rows) != 2 {
		t.Error("Expected 2 rows")
	}

	if rows[0][1] != "test-secret-1" {
		t.Error("First secret name mismatch")
	}
	if rows[0][2] != "Opaque" {
		t.Error("First secret type mismatch")
	}

	if rows[1][1] != "test-secret-2" {
		t.Error("Second secret name mismatch")
	}
	if rows[1][2] != "kubernetes.io/tls" {
		t.Error("Second secret type mismatch")
	}
}

func TestSecretsModelWithDifferentStates(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewSecrets(client, "default")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	secretsInfo := []k8s.SecretInfo{
		{
			Name:      "empty-secret",
			Namespace: "default",
			Type:      "Opaque",
			Data:      "0",
			Age:       "30m",
		},
	}

	model.resourceData = []types.ResourceData{SecretData{&secretsInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if rows[0][3] != "0" {
		t.Error("Expected 0 data keys for empty secret")
	}
}

func TestNewSecretDetails(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.secret.Name != "test-secret" {
		t.Error("Expected secret name to be 'test-secret'")
	}
	if model.secret.Namespace != "default" {
		t.Error("Expected secret namespace to be 'default'")
	}
	if model.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.err != nil {
		t.Error("Expected error to be nil")
	}
	if model.showValues != false {
		t.Error("Expected showValues to be false by default")
	}
}

func TestSecretDetailsModelInitComponent(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	model.secret.Raw = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("secret123"),
		},
	}

	teaModel, err := model.InitComponent(&client)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.yamlViewer == nil {
		t.Error("Expected yamlViewer to be initialized")
	}

	if teaModel != model {
		t.Error("Expected returned model to be the same as our model")
	}
}

func TestSecretDetailsModelView(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	model.secret.Raw = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("secret123"),
		},
	}

	desc, err := model.secret.DescribeWithVisibility(false)
	if err != nil {
		t.Errorf("DescribeWithVisibility failed: %v", err)
		return
	}
	t.Logf("Description: %s", desc)

	_, err = model.InitComponent(&client)
	if err != nil {
		t.Errorf("Expected no error during init, got %v", err)
		return
	}

	if model.yamlViewer == nil {
		t.Error("Expected yamlViewer to be initialized")
		return
	}

	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	_, _ = model.yamlViewer.Update(windowMsg)

	if model.showValues != false {
		t.Error("Expected showValues to be false initially")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	updatedModel, _ := model.Update(msg)

	if secretModel, ok := updatedModel.(*secretDetailsModel); ok {
		if secretModel.showValues != true {
			t.Error("Expected showValues to be true after pressing 'v'")
		}
	} else {
		t.Error("Expected updated model to be of type *secretDetailsModel")
	}
}

func TestSecretDetailsModelTitleUpdate(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	model.secret.Raw = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("secret123"),
		},
	}

	_, err := model.InitComponent(&client)
	if err != nil {
		t.Errorf("Expected no error during init, got %v", err)
		return
	}

	t.Logf("YAMLViewer after init: %v", model.yamlViewer)

	if model.showValues != false {
		t.Error("Expected showValues to be false initially")
	}

	desc, err := model.secret.DescribeWithVisibility(false)
	if err != nil {
		t.Errorf("Expected no error getting description, got %v", err)
	}
	if !containsString(desc, "dataKeys") {
		t.Error("Expected description to contain 'dataKeys' when showValues is false")
	}

	desc, err = model.secret.DescribeWithVisibility(true)
	if err != nil {
		t.Errorf("Expected no error getting description, got %v", err)
	}
	if !containsString(desc, "data:") {
		t.Error("Expected description to contain 'data:' when showValues is true")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	updatedModel, _ := model.Update(msg)

	if secretModel, ok := updatedModel.(*secretDetailsModel); ok {
		if secretModel.showValues != true {
			t.Error("Expected showValues to be true after pressing 'v'")
		}
	} else {
		t.Error("Expected updated model to be of type *secretDetailsModel")
	}
}

func TestSecretDetailsModelErrorHandling(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	model.err = fmt.Errorf("test error")

	view := model.View()
	if !containsString(view, "Error: test error") {
		t.Error("Expected view to contain error message")
	}
}

func TestSecretDetailsModelQuit(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewSecretDetails(client, "default", "test-secret")

	model.secret.Raw = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
		},
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
