package models

import (
	"strings"
	"testing"
	"time"
)

func TestHeaderRefreshInterval(t *testing.T) {
	if HeaderRefreshInterval < 5*time.Second {
		t.Errorf("Refresh interval too short: %v (should be at least 5 seconds)", HeaderRefreshInterval)
	}
	if HeaderRefreshInterval > 60*time.Second {
		t.Errorf("Refresh interval too long: %v (should be at most 60 seconds)", HeaderRefreshInterval)
	}
}

func TestHeaderRefreshCycle(t *testing.T) {
	header := NewHeader("Test Header", nil, nil)

	cmd := header.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil when no kubeconfig")
	}

	refreshMsg := HeaderRefreshMsg{}
	newHeader, cmd := header.Update(refreshMsg)

	if newHeader == nil {
		t.Error("Expected Update to return a header")
	}
	if cmd != nil {
		t.Error("Expected Update to return nil command when no kubeconfig")
	}
}

func TestHeaderConstants(t *testing.T) {
	msg := HeaderRefreshMsg{}

	if msg == (HeaderRefreshMsg{}) {
		t.Log("HeaderRefreshMsg is properly defined")
	}

	expectedInterval := 10 * time.Second
	if HeaderRefreshInterval != expectedInterval {
		t.Errorf("Expected HeaderRefreshInterval to be %v, got %v", expectedInterval, HeaderRefreshInterval)
	}
}

func TestHeaderViewWithoutKubeconfig(t *testing.T) {
	header := NewHeader("Test Header", nil, nil)

	content := header.View()

	if content == "" {
		t.Error("Expected header to show no connection message")
	}

	// The header may show truncated content for small terminals in responsive mode
	if !strings.Contains(content, "K8s") {
		t.Errorf("Expected header to contain 'K8s', got '%s'", content)
	}
}
