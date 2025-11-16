package models

import (
	tea "github.com/charmbracelet/bubbletea"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"strings"
	"testing"
)

func TestNewSizeCheckModel(t *testing.T) {
	model := NewSizeCheckModel()

	if model == nil {
		t.Fatal("Expected non-nil SizeCheckModel")
	}

	if model.checkPeriod == 0 {
		t.Error("Expected non-zero check period")
	}
}

func TestSizeCheckModelInit(t *testing.T) {
	model := NewSizeCheckModel()
	cmd := model.Init()

	if cmd == nil {
		t.Error("Expected non-nil command from Init")
	}
}

func TestSizeCheckModelUpdate(t *testing.T) {
	model := NewSizeCheckModel()

	// Test WindowSizeMsg update
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, _ := model.Update(msg)

	updatedModel, ok := updated.(*SizeCheckModel)
	if !ok {
		t.Fatal("Expected SizeCheckModel from Update")
	}

	if updatedModel.width != 100 {
		t.Errorf("Expected width 100, got %d", updatedModel.width)
	}

	if updatedModel.height != 50 {
		t.Errorf("Expected height 50, got %d", updatedModel.height)
	}
}

func TestSizeCheckModelIsSizeValid(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		{"Valid size", 100, 50, true},
		{"Minimum valid size", 80, 24, true},
		{"Too narrow", 79, 24, false},
		{"Too short", 80, 23, false},
		{"Too small both", 40, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSizeCheckModel()
			model.width = tt.width
			model.height = tt.height

			result := model.IsSizeValid()
			if result != tt.expected {
				t.Errorf("Expected %v for size %dx%d, got %v", tt.expected, tt.width, tt.height, result)
			}
		})
	}
}

func TestCheckTerminalSize(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		{"Valid size", 100, 50, true},
		{"Minimum valid size", 80, 24, true},
		{"Too narrow", 79, 24, false},
		{"Too short", 80, 23, false},
		{"Too small both", 40, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckTerminalSize(tt.width, tt.height)
			if result != tt.expected {
				t.Errorf("Expected %v for size %dx%d, got %v", tt.expected, tt.width, tt.height, result)
			}
		})
	}
}

func TestSizeCheckModelView(t *testing.T) {
	// Initialize colors for testing
	_ = customstyles.InitColors()

	model := NewSizeCheckModel()

	// Test view with valid size
	model.width = 100
	model.height = 50
	view := model.View()
	if view != "" {
		t.Error("Expected empty view for valid size")
	}

	// Test view with invalid size
	model.width = 40
	model.height = 10
	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view for invalid size")
	}

	// Check that view contains expected elements (regardless of colors)
	if !contains(view, "Terminal Size Too Small") {
		t.Error("Expected view to contain size warning")
	}

	if !contains(view, "40x10") {
		t.Error("Expected view to contain current size")
	}

	if !contains(view, "80x24") {
		t.Error("Expected view to contain minimum size")
	}

	// Check for instruction text (may be split across lines due to styling)
	if !contains(view, "Please resize") && !contains(view, "resize your terminal") {
		t.Errorf("Expected view to contain resize instruction. View content: %s", view)
	}

	// Check for quit option
	if !contains(view, "Press Esc") && !contains(view, "Esc or Q") {
		t.Errorf("Expected view to contain quit option. View content: %s", view)
	}
}

func TestSizeCheckConstants(t *testing.T) {
	if MinTerminalWidth <= 0 {
		t.Error("Expected positive MinTerminalWidth")
	}

	if MinTerminalHeight <= 0 {
		t.Error("Expected positive MinTerminalHeight")
	}

	if MinTerminalWidth < 80 {
		t.Error("Expected MinTerminalWidth to be at least 80")
	}

	if MinTerminalHeight < 24 {
		t.Error("Expected MinTerminalHeight to be at least 24")
	}
}

func TestSizeCheckModelKeyHandling(t *testing.T) {
	model := NewSizeCheckModel()

	// Set invalid size to ensure size check screen is active
	model.width = 40
	model.height = 10

	// Test Esc key
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Expected quit command for Esc key")
	}

	// Test 'q' key
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("Expected quit command for 'q' key")
	}

	// Test 'Q' key
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd == nil {
		t.Error("Expected quit command for 'Q' key")
	}

	// Test that other keys don't quit when size is invalid
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("Did not expect any command for 'x' key")
	}

	// Test that keys don't quit when size is valid
	model.width = 100
	model.height = 50
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("Did not expect any command when terminal size is valid")
	}
}

func TestSizeCheckModelWhitespace(t *testing.T) {
	// Initialize colors for testing
	_ = customstyles.InitColors()

	model := NewSizeCheckModel()
	model.width = 40
	model.height = 10

	view := model.View()

	// Check that view doesn't have excessive whitespace lines
	lines := strings.Split(view, "\n")
	emptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLines++
		}
	}

	// Should have minimal empty lines (just for spacing, not filling entire terminal)
	if emptyLines > 8 {
		t.Errorf("Too many empty lines in view: %d. View content:\n%s", emptyLines, view)
	}

	// Check that view doesn't contain the problematic pattern
	if strings.Contains(view, "%%%%%%%%") {
		t.Error("View contains problematic whitespace pattern")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
