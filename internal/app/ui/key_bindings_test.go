package ui

import (
	"testing"

	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
)

func TestGetKeyBinding(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{"quit binding", "quit", "q"},
		{"help binding", "help", "?"},
		{"refresh binding", "refresh", "r"},
		{"back binding", "back", "["},
		{"forward binding", "forward", "]"},
		{"new_tab binding", "new_tab", "ctrl+t"},
		{"close_tab binding", "close_tab", "ctrl+w"},
		{"tab_next binding", "tab_next", "right"},
		{"tab_prev binding", "tab_prev", "left"},
		{"quick_nav binding", "quick_nav", "g"},
		{"cluster_prev binding", "cluster_prev", "ctrl+left"},
		{"cluster_next binding", "cluster_next", "ctrl+right"},
		{"add_cluster binding", "add_cluster", "ctrl+n"},
		{"remove_cluster binding", "remove_cluster", "ctrl+x"},
		{"unknown action", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal AppModel for testing
			model := &AppModel{
				config: config.AppConfig{
					KeyBindings: make(map[string]string),
				},
			}

			result := model.GetKeyBinding(tt.action)
			if result != tt.expected {
				t.Errorf("GetKeyBinding(%s) = %v, want %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestGetKeyBindingWithCustomConfig(t *testing.T) {
	// Test with custom key bindings
	model := &AppModel{
		config: config.AppConfig{
			KeyBindings: map[string]string{
				"quit":           "custom_q",
				"remove_cluster": "custom_ctrl+x",
			},
		},
	}

	tests := []struct {
		action   string
		expected string
	}{
		{"quit", "custom_q"},
		{"remove_cluster", "custom_ctrl+x"},
		{"help", "?"}, // Should use default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := model.GetKeyBinding(tt.action)
			if result != tt.expected {
				t.Errorf("GetKeyBinding(%s) = %v, want %v", tt.action, result, tt.expected)
			}
		})
	}
}
