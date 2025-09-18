package components

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTableModel_Selection(t *testing.T) {
	columns := []table.Column{
		{Title: "Col1", Width: 10},
		{Title: "Col2", Width: 10},
	}
	rows := []table.Row{
		{"val1", "val2"},
		{"val3", "val4"},
	}

	var selectedValue string
	onSelect := func(selected string) tea.Msg {
		selectedValue = selected
		return nil
	}

	tableModel := NewTable(columns, []float64{0.5, 0.5}, rows, "Test", onSelect, 1, nil, nil)

	_, cmd := tableModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Expected a command to be returned")
	}
	cmd()

	if selectedValue != "val2" {
		t.Errorf("Expected selected value to be 'val2', got '%s'", selectedValue)
	}

	tableModel.Update(tea.KeyMsg{Type: tea.KeySpace})

	checked := tableModel.GetCheckedItems()
	if len(checked) != 1 || checked[0] != 0 {
		t.Errorf("Expected one checked item at index 0, got %v", checked)
	}

	tableModel.Table.SetCursor(1)
	tableModel.Update(tea.KeyMsg{Type: tea.KeySpace})

	checked = tableModel.GetCheckedItems()
	if len(checked) != 2 || (checked[0] != 0 && checked[1] != 1) {
		t.Errorf("Expected two checked items at indices 0 and 1, got %v", checked)
	}

	tableModel.Update(tea.KeyMsg{Type: tea.KeySpace})

	checked = tableModel.GetCheckedItems()
	if len(checked) != 1 || checked[0] != 0 {
		t.Errorf("Expected one checked item at index 0, got %v", checked)
	}
}
