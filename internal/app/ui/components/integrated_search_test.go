package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
)

func TestIntegratedSearchFunctionality(t *testing.T) {
	// Set up screen dimensions to avoid "Initializing..." state
	styles.ScreenWidth = 80
	styles.ScreenHeight = 24

	// Test data
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "STATUS", Width: 15},
		{Title: "NAMESPACE", Width: 15},
	}

	rows := []table.Row{
		{"nginx-pod", "Running", "default"},
		{"redis-pod", "Pending", "default"},
		{"postgres-pod", "Running", "database"},
		{"nginx-service", "Active", "default"},
		{"redis-service", "Active", "cache"},
	}

	colPercent := []float64{0.4, 0.3, 0.3}

	// Create table model
	model := NewTable(columns, colPercent, rows, "Test Resources", nil, 0, nil, nil)

	// Test 1: Initial state - no search mode
	if model.searchMode {
		t.Error("Search mode should be false initially")
	}

	if len(model.filteredDataRows) != 5 {
		t.Errorf("Expected 5 rows initially, got %d", len(model.filteredDataRows))
	}

	// Test 2: Activate search mode with "/"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ := model.Update(msg)
	tableModel := updatedModel.(*TableModel)

	if !tableModel.searchMode {
		t.Error("Search mode should be activated after pressing '/'")
	}

	if tableModel.searchQuery != "" {
		t.Errorf("Search query should be empty initially, got '%s'", tableModel.searchQuery)
	}

	// Test 3: Type search query "nginx"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("nginx")}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	if tableModel.searchQuery != "nginx" {
		t.Errorf("Expected search query 'nginx', got '%s'", tableModel.searchQuery)
	}

	// Should only show rows containing "nginx"
	if len(tableModel.filteredDataRows) != 2 {
		t.Errorf("Expected 2 filtered rows for 'nginx', got %d", len(tableModel.filteredDataRows))
	}

	// Verify the filtered rows contain nginx
	for _, row := range tableModel.filteredDataRows {
		if !strings.Contains(strings.ToLower(row[0]), "nginx") &&
			!strings.Contains(strings.ToLower(row[1]), "nginx") &&
			!strings.Contains(strings.ToLower(row[2]), "nginx") {
			t.Errorf("Row %v should not appear in nginx search results", row)
		}
	}

	// Test 4: Add more characters "-service"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-service")}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	if tableModel.searchQuery != "nginx-service" {
		t.Errorf("Expected search query 'nginx-service', got '%s'", tableModel.searchQuery)
	}

	// Should only show nginx-service
	if len(tableModel.filteredDataRows) != 1 {
		t.Errorf("Expected 1 filtered row for 'nginx-service', got %d", len(tableModel.filteredDataRows))
	}

	if tableModel.filteredDataRows[0][0] != "nginx-service" {
		t.Errorf("Expected 'nginx-service', got '%s'", tableModel.filteredDataRows[0][0])
	}

	// Test 5: Clear search with Escape
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	if tableModel.searchMode {
		t.Error("Search mode should be deactivated after pressing Escape")
	}

	if tableModel.searchQuery != "" {
		t.Error("Search query should be cleared after pressing Escape")
	}

	// Should show all rows again
	if len(tableModel.filteredDataRows) != 5 {
		t.Errorf("Expected 5 rows after clearing search, got %d", len(tableModel.filteredDataRows))
	}

	// Test 6: Search in different columns - search for "database"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("database")}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	// Should find postgres-pod in database namespace
	if len(tableModel.filteredDataRows) != 1 {
		t.Errorf("Expected 1 filtered row for 'database', got %d", len(tableModel.filteredDataRows))
	}

	if tableModel.filteredDataRows[0][0] != "postgres-pod" {
		t.Errorf("Expected 'postgres-pod', got '%s'", tableModel.filteredDataRows[0][0])
	}

	// Test 7: Case insensitive search - start fresh search
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	// Activate search mode again
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	// Type "RUNNING" in uppercase
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("RUNNING")}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	// Should find all running pods regardless of case
	if len(tableModel.filteredDataRows) != 2 {
		t.Errorf("Expected 2 filtered rows for 'RUNNING', got %d", len(tableModel.filteredDataRows))
	}

	// Test 8: Verify search line appears at bottom and stays visible
	view := tableModel.View()

	// Should contain search line at bottom (not top)
	if !strings.Contains(view, "Search: RUNNING") {
		t.Error("View should contain search line with query")
	}

	// Search line should appear after table content, not before
	lines := strings.Split(view, "\n")
	searchLineIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "Search: RUNNING") {
			searchLineIndex = i
			break
		}
	}

	if searchLineIndex == -1 {
		t.Error("Search line not found in view")
	} else if searchLineIndex < len(lines)-3 { // Allow some padding for bottom
		t.Errorf("Search line should appear at bottom of view, found at line %d of %d", searchLineIndex, len(lines))
	}

	// Should show cursor when in search mode
	if !strings.Contains(view, "Search: RUNNING█") {
		t.Error("Should show cursor when in search mode")
	}

	// Test 9: Exit search mode but keep filter visible
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = tableModel.Update(msg)
	tableModel = updatedModel.(*TableModel)

	// Should no longer be in search mode but query should remain for display
	if tableModel.searchMode {
		t.Error("Should not be in search mode after Enter")
	}

	if tableModel.searchQuery != "RUNNING" {
		t.Errorf("Search query should remain 'RUNNING', got '%s'", tableModel.searchQuery)
	}

	// Search line should still be visible since query is not empty
	viewAfterExit := tableModel.View()
	if !strings.Contains(viewAfterExit, "Search: RUNNING") {
		t.Error("Search line should remain visible after exiting search mode")
	}

	// Should not have cursor since not in search mode
	if strings.Contains(viewAfterExit, "█") {
		t.Error("Should not show cursor when not in search mode")
	}
}
