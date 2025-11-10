package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTableModel_Search(t *testing.T) {
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "STATUS", Width: 15},
	}

	rows := []table.Row{
		{"nginx-pod", "Running"},
		{"redis-pod", "Pending"},
		{"postgres-pod", "Running"},
		{"nginx-service", "Active"},
	}

	colPercent := []float64{0.6, 0.4}

	model := NewTable(columns, colPercent, rows, "Test", nil, 0, nil, nil)

	// Test initial state
	if model.searchMode {
		t.Error("Search mode should be false initially")
	}

	if model.searchQuery != "" {
		t.Error("Search query should be empty initially")
	}

	// Test activating search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	tableModel := newModel.(*TableModel)

	if !tableModel.searchMode {
		t.Error("Search mode should be activated after pressing '/'")
	}

	// Test typing search query
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("nginx")}
	newModel, _ = tableModel.Update(msg)
	tableModel = newModel.(*TableModel)

	if tableModel.searchQuery != "nginx" {
		t.Errorf("Expected search query 'nginx', got '%s'", tableModel.searchQuery)
	}

	// Check that filtering worked (should only show rows with "nginx")
	if len(tableModel.filteredDataRows) != 2 {
		t.Errorf("Expected 2 filtered rows, got %d", len(tableModel.filteredDataRows))
	}

	// Test exiting search mode with Escape
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = tableModel.Update(msg)
	tableModel = newModel.(*TableModel)

	if tableModel.searchMode {
		t.Error("Search mode should be deactivated after pressing Escape")
	}

	if tableModel.searchQuery != "" {
		t.Error("Search query should be cleared after pressing Escape")
	}

	// Should show all rows again
	if len(tableModel.filteredDataRows) != 4 {
		t.Errorf("Expected 4 rows after clearing search, got %d", len(tableModel.filteredDataRows))
	}
}

func TestTableModel_MatchesSearch(t *testing.T) {
	columns := []table.Column{{Title: "NAME", Width: 20}, {Title: "STATUS", Width: 15}}
	rows := []table.Row{{"nginx-pod", "Running"}}
	colPercent := []float64{0.6, 0.4}

	model := NewTable(columns, colPercent, rows, "Test", nil, 0, nil, nil)

	testRow := rows[0]

	// Test case insensitive matching (matchesSearch expects lowercase query)
	if !model.matchesSearch(testRow, "nginx") {
		t.Error("Should match 'nginx' case insensitively")
	}

	if !model.matchesSearch(testRow, strings.ToLower("NGINX")) {
		t.Error("Should match 'NGINX' case insensitively")
	}

	// Test partial matching
	if !model.matchesSearch(testRow, "ginx") {
		t.Error("Should match partial string 'ginx'")
	}

	// Test non-matching
	if model.matchesSearch(testRow, "redis") {
		t.Error("Should not match 'redis'")
	}

	// Test matching in different columns
	if !model.matchesSearch(testRow, "running") {
		t.Error("Should match 'running' in second column")
	}
}
