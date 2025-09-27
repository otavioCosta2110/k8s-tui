package components

import (
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateActionsMsg struct {
	key    string
	action func() tea.Cmd
}

type TableModel struct {
	Table           table.Model
	OnSelected      func(selected string) tea.Msg
	selectColumn    int
	loading         bool
	initialized     bool
	colPercent      []float64
	checkedRows     map[int]bool
	refreshInterval time.Duration
	lastRefresh     time.Time
	refreshFunc     func() ([]table.Row, error)
	updateActions   map[string]func() tea.Cmd
}

type loadedTableMsg struct{}

func NewTable(columns []table.Column, colPercent []float64, rows []table.Row, title string, onSelect func(selected string) tea.Msg, selectColumn int, refreshFunc func() ([]table.Row, error), updateActions map[string]func() tea.Cmd) *TableModel {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderBottom(true).
		BorderForeground(lipgloss.Color(customstyles.TextColor)).
		BorderStyle(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color(customstyles.TextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor))

	styles.Cell = styles.Cell.Foreground(lipgloss.Color(customstyles.TextColor)).Background(lipgloss.Color(customstyles.BackgroundColor))
	styles.Selected = customstyles.SelectedStyle().Foreground(lipgloss.Color(customstyles.SelectionForeground)).Background(lipgloss.Color(customstyles.SelectionBackground))
	checkboxColumn := table.Column{Title: "✓", Width: 3}
	columns = append([]table.Column{checkboxColumn}, columns...)

	// colPercent are weights for the data columns, normalize them
	normalizedDataWeights := normalizeColumnPercentages(colPercent)
	newColPercent := make([]float64, len(normalizedDataWeights)+1)
	newColPercent[0] = 0 // checkbox has fixed width
	for i, p := range normalizedDataWeights {
		newColPercent[i+1] = p
	}

	newRows := make([]table.Row, len(rows))
	for i, row := range rows {
		newRows[i] = append(table.Row{"▢"}, row...)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(newRows),
		table.WithFocused(true),
	)

	t.SetStyles(styles)

	return &TableModel{
		Table:           t,
		OnSelected:      onSelect,
		selectColumn:    selectColumn + 1,
		colPercent:      newColPercent,
		loading:         false,
		initialized:     false,
		checkedRows:     make(map[int]bool),
		refreshInterval: 5 * time.Second,
		refreshFunc:     refreshFunc,
		lastRefresh:     time.Now(),
		updateActions:   updateActions,
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
		switch msg.Type {
		case tea.KeySpace:
			if !m.loading {
				selectedIdx := m.Table.Cursor()
				m.toggleCheckbox(selectedIdx)
				return m, nil
			}
		case tea.KeyRunes:
			if action, exists := m.updateActions[string(msg.Runes)]; exists {
				cmd := action()
				m.refreshData()
				return m, cmd
			}
			if string(msg.Runes) == "r" {
				return m, m.refreshData()
			}
		case tea.KeyEnter:
			if !m.loading && m.OnSelected != nil {
				if len(m.Table.SelectedRow()) > 0 {
					selected := m.Table.SelectedRow()[m.selectColumn]
					return m, func() tea.Msg {
						return m.OnSelected(selected)
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m *TableModel) SetUpdateActions(actions map[string]func() tea.Cmd) {
	m.updateActions = actions
}

func (m *TableModel) toggleCheckbox(rowIdx int) {
	rows := m.Table.Rows()
	if rowIdx < 0 || rowIdx >= len(rows) {
		return
	}

	m.checkedRows[rowIdx] = !m.checkedRows[rowIdx]

	if m.checkedRows[rowIdx] {
		rows[rowIdx][0] = "🗹"
	} else {
		rows[rowIdx][0] = "▢"
	}

	m.Table.SetRows(rows)
}

func (m *TableModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Loading...")
	}

	m.updateColumnWidths(styles.ScreenWidth)

	tableHeight := styles.ScreenHeight + 1
	m.Table.SetHeight(tableHeight)
	m.Table.SetWidth(styles.ScreenWidth)

	tableView := m.Table.View()

	return tableView
}

func (m *TableModel) updateColumnWidths(totalWidth int) {
	columns := m.Table.Columns()
	widths := make([]int, len(columns))
	// Set checkbox to fixed width
	checkboxWidth := 3
	widths[0] = checkboxWidth
	remainingWidth := totalWidth + checkboxWidth
	totalAssigned := checkboxWidth + len(columns)*2
	// Distribute remaining width among data columns based on weights
	for i := 1; i < len(columns); i++ {
		width := int(float64(remainingWidth) * m.colPercent[i])
		widths[i] = width
		totalAssigned += width
	}
	// Adjust the last column to fill the remaining width
	if len(widths) > 1 {
		widths[len(widths)-1] += totalWidth - totalAssigned
	}
	for i := range columns {
		columns[i].Width = widths[i]
	}
	m.Table.SetColumns(columns)
}

// normalizeColumnPercentages ensures that column percentages add up to 1.0
func normalizeColumnPercentages(percentages []float64) []float64 {
	if len(percentages) == 0 {
		return percentages
	}

	total := 0.0
	for _, p := range percentages {
		total += p
	}

	if total == 0 {
		// If all percentages are 0, distribute evenly
		evenPercent := 1.0 / float64(len(percentages))
		normalized := make([]float64, len(percentages))
		for i := range normalized {
			normalized[i] = evenPercent
		}
		return normalized
	}

	// Normalize to sum to 1.0
	normalized := make([]float64, len(percentages))
	for i, p := range percentages {
		normalized[i] = p / total
	}
	return normalized
}

func (m *TableModel) GetCheckedItems() []int {
	if len(m.checkedRows) == 0 {
		return []int{m.Table.Cursor()}
	}
	var checked []int
	for idx, isChecked := range m.checkedRows {
		if isChecked {
			checked = append(checked, idx)
		}
	}
	sort.Ints(checked)
	return checked
}

func (m *TableModel) ClearCheckedItems() {
	m.checkedRows = make(map[int]bool)
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
			newRows[i] = append(table.Row{"🗹"}, row...)
		} else {
			newRows[i] = append(table.Row{"▢"}, row...)
		}
	}
	m.Table.SetRows(newRows)
}

func (m *TableModel) UpdateColumns(columns []table.Column) {
	columns = append([]table.Column{{Title: "✓", Width: 3}}, columns...)
	m.Table.SetColumns(columns)
}

func (m *TableModel) refreshData() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.refreshFunc()
		if err != nil {
			return err
		}
		m.UpdateRows(rows)
		return nil
	}
}

func (t *TableModel) Refresh() (tea.Model, tea.Cmd) {
	if t.refreshFunc == nil {
		return t, nil
	}

	rows, err := t.refreshFunc()
	if err != nil {
		return t, nil
	}

	t.UpdateRows(rows)
	return t, nil
}
