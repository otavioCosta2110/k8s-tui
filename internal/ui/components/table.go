package components

import (
	global "otaviocosta2110/k8s-tui/internal"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableModel struct {
	Table        table.Model
	OnSelected   func(selected string) tea.Msg
	selectColumn int
	loading      bool
	initialized  bool
	colPercent   []float64
	checkedRows  map[int]bool 
}

type loadedTableMsg struct{}

func NewTable(columns []table.Column, colPercent []float64, rows []table.Row, title string, onSelect func(selected string) tea.Msg, selectColumn int) *TableModel {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(global.Colors.Pink)).
		BorderBottom(true).
		Bold(true)

	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color(global.Colors.Pink)).
		Bold(true)

	checkboxColumn := table.Column{Title: "✓", Width: 3}
	columns = append([]table.Column{checkboxColumn}, columns...)
	
	checkboxPercent := 0.03 
	newColPercent := make([]float64, len(colPercent)+1)
	newColPercent[0] = checkboxPercent
	for i, p := range colPercent {
		newColPercent[i+1] = p * (1 - checkboxPercent)
	}

	newRows := make([]table.Row, len(rows))
	for i, row := range rows {
		newRows[i] = append(table.Row{"[ ]"}, row...)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(newRows),
		table.WithFocused(true),
	)

	t.SetStyles(styles)

	return &TableModel{
		Table:        t,
		OnSelected:   onSelect,
		selectColumn: selectColumn + 1, 
		colPercent:   newColPercent,
		loading:      false,
		initialized:  false,
		checkedRows:  make(map[int]bool),
	}
}

func (m *TableModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return loadedTableMsg{}
	})
}

func (m *TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedTableMsg:
		m.loading = false
		m.initialized = true
		return m, nil
	case tea.WindowSizeMsg:
		m.updateColumnWidths(msg.Width)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == " " && !m.loading { 
			selectedIdx := m.Table.Cursor()
			m.toggleCheckbox(selectedIdx)
			return m, nil
		}
		if msg.String() == "enter" && !m.loading && m.OnSelected != nil {
			if len(m.Table.SelectedRow()) > 0 {
				selected := m.Table.SelectedRow()[m.selectColumn]
				return m, func() tea.Msg {
					return m.OnSelected(selected)
				}
			}
		}
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m *TableModel) toggleCheckbox(rowIdx int) {
	rows := m.Table.Rows()
	if rowIdx < 0 || rowIdx >= len(rows) {
		return
	}

	m.checkedRows[rowIdx] = !m.checkedRows[rowIdx]

	if m.checkedRows[rowIdx] {
		rows[rowIdx][0] = "[✓]"
	} else {
		rows[rowIdx][0] = "[ ]"
	}

	m.Table.SetRows(rows)
}

func (m *TableModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading...")
	}

	m.updateColumnWidths(global.ScreenWidth)
	m.Table.SetHeight(global.ScreenHeight - len(m.Table.Columns()))

	return m.Table.View()
}

func (m *TableModel) updateColumnWidths(totalWidth int) {
	columns := m.Table.Columns()
	for i := range columns {
		width := int(float64(totalWidth) * m.colPercent[i])
		columns[i].Width = width
	}
	m.Table.SetColumns(columns)
}

func (m *TableModel) GetCheckedItems() []int {
	var checked []int
	for idx, isChecked := range m.checkedRows {
		if isChecked {
			checked = append(checked, idx)
		}
	}
	return checked
}

func NewColumn(title string, percent float64) table.Column {
	return table.Column{
		Title: title,
		Width: 0,
	}
}

func NewRow(values ...string) table.Row {
	return values
}

func (m *TableModel) UpdateRows(rows []table.Row) {
	newRows := make([]table.Row, len(rows))
	for i, row := range rows {
		if m.checkedRows[i] {
			newRows[i] = append(table.Row{"[✓]"}, row...)
		} else {
			newRows[i] = append(table.Row{"[ ]"}, row...)
		}
	}
	m.Table.SetRows(newRows)
}

func (m *TableModel) UpdateColumns(columns []table.Column) {
	columns = append([]table.Column{{Title: "✓", Width: 3}}, columns...)
	m.Table.SetColumns(columns)
}
