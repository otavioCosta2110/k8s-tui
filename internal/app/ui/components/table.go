package components

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateActionsMsg struct {
	key    string
	action func() tea.Cmd
}

type fetchResultMsg struct {
	rows []table.Row
	err  error
}

type TableModel struct {
	Table            table.Model
	OnSelected       func(selected string) tea.Msg
	selectColumn     int
	loading          bool
	initialized      bool
	colPercent       []float64
	checkedRows      map[int]bool
	refreshInterval  time.Duration
	lastRefresh      time.Time
	refreshFunc      func() ([]table.Row, error)
	updateActions    map[string]func() tea.Cmd
	spinner          SpinnerModel
	searchMode       bool
	searchQuery      string
	allDataRows      []table.Row
	filteredDataRows []table.Row
}

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

	normalizedDataWeights := normalizeColumnPercentages(colPercent)
	newColPercent := make([]float64, len(normalizedDataWeights)+1)
	newColPercent[0] = 0
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
		Table:            t,
		OnSelected:       onSelect,
		selectColumn:     selectColumn + 1,
		colPercent:       newColPercent,
		loading:          len(rows) == 0,
		initialized:      false,
		checkedRows:      make(map[int]bool),
		refreshInterval:  5 * time.Second,
		refreshFunc:      refreshFunc,
		lastRefresh:      time.Now(),
		updateActions:    updateActions,
		spinner:          NewSpinner("Loading resources..."),
		searchMode:       false,
		searchQuery:      "",
		allDataRows:      rows,
		filteredDataRows: rows,
	}
}

func (m *TableModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Init()}

	if m.loading && m.refreshFunc != nil {
		cmds = append(cmds, m.fetchDataCmd())
	}

	return tea.Batch(cmds...)
}

func (m *TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fetchResultMsg:
		if msg.err != nil {
			return m, nil
		}
		m.UpdateRows(msg.rows)
		m.loading = false
		return m, nil
	case tea.WindowSizeMsg:
		m.updateColumnWidths(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			// In search mode, only handle search-specific keys
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEscape:
				m.searchMode = false
				if msg.Type == tea.KeyEscape {
					m.searchQuery = ""
					m.applySearchFilter()
				}
				return m, nil
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.applySearchFilter()
				}
				return m, nil
			case tea.KeyRunes:
				for _, r := range msg.Runes {
					if r >= 32 && r <= 126 {
						m.searchQuery += string(r)
					}
				}
				m.applySearchFilter()
				return m, nil
			default:
				return m, nil
			}
		} else {
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
				if string(msg.Runes) == "/" {
					m.searchMode = true
					m.searchQuery = ""
					return m, nil
				}
			case tea.KeyEnter:
				if !m.loading && m.OnSelected != nil {
					if len(m.Table.SelectedRow()) > 0 {
						selected := m.Table.SelectedRow()[m.selectColumn]
						if strings.Contains(selected, " ") {
							parts := strings.SplitN(selected, " ", 2)
							if len(parts) == 2 && len(parts[0]) > 0 {
								firstPart := parts[0]
								if len(firstPart) > 1 || (len(firstPart) == 1 && firstPart[0] > 127) {
									selected = parts[1]
								}
							}
						}
						return m, func() tea.Msg {
							return m.OnSelected(selected)
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)

	return m, tea.Batch(cmd, spinnerCmd)
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
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Panic in Table View: %v", r))
		}
	}()
	logger.Debug("Table View called")
	if m.loading {
		return m.spinner.CenteredView(styles.ScreenWidth, styles.ScreenHeight)
	}

	if styles.ScreenWidth < 10 || styles.ScreenHeight < 5 {
		return "Initializing..."
	}

	m.updateColumnWidths(styles.ScreenWidth)

	// Reserve space for search line at bottom if search is active or has query
	searchHeight := 0
	if m.searchMode || m.searchQuery != "" {
		searchHeight = 1
	}

	tableHeight := styles.ScreenHeight - searchHeight
	sumWidths := 0
	for _, col := range m.Table.Columns() {
		sumWidths += col.Width
	}
	m.Table.SetHeight(tableHeight)
	tableWidth := styles.ScreenWidth
	if sumWidths > styles.ScreenWidth {
		tableWidth = sumWidths
	}
	m.Table.SetWidth(tableWidth)

	if len(m.Table.Rows()) == 0 {
		// Create a single row with a "No data available" message
		noDataRow := table.Row{""}
		for i := 1; i < len(m.Table.Columns()); i++ {
			if i == 1 { // Put the message in the first data column
				noDataRow = append(noDataRow, "No data available")
			} else {
				noDataRow = append(noDataRow, "")
			}
		}

		// Temporarily set the no data row
		originalRows := m.Table.Rows()
		m.Table.SetRows([]table.Row{noDataRow})
		tableView := m.Table.View()
		m.Table.SetRows(originalRows) // Restore original empty rows

		// Add search line at bottom if search is active or has query
		if m.searchMode || m.searchQuery != "" {
			searchLine := m.getSearchLine()
			return lipgloss.JoinVertical(lipgloss.Left, tableView, searchLine)
		}
		return tableView
	}

	logger.Debug("About to call m.Table.View()")
	tableView := m.Table.View()
	logger.Debug("m.Table.View() returned")

	// Add search line at bottom if search is active or has query
	if m.searchMode || m.searchQuery != "" {
		searchLine := m.getSearchLine()
		return lipgloss.JoinVertical(lipgloss.Left, tableView, searchLine)
	}

	return tableView
}

func (m *TableModel) getSearchLine() string {
	cursor := ""
	if m.searchMode {
		cursor = "█"
	}

	searchText := fmt.Sprintf("Search: %s%s", m.searchQuery, cursor)
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(styles.ScreenWidth).
		Padding(0, 1)

	searchLine := searchStyle.Render(searchText)

	return searchLine
}

func (m *TableModel) updateColumnWidths(totalWidth int) {
	logger.Debug(fmt.Sprintf("updateColumnWidths with width %d", totalWidth))
	columns := m.Table.Columns()
	widths := make([]int, len(columns))

	checkboxWidth := 3
	widths[0] = checkboxWidth
	remainingWidth := totalWidth + checkboxWidth
	totalAssigned := checkboxWidth + len(columns)*2

	for i := 1; i < len(columns); i++ {
		width := int(float64(remainingWidth) * m.colPercent[i])
		widths[i] = width
		totalAssigned += width
	}

	if len(widths) > 1 {
		widths[len(widths)-1] += totalWidth - totalAssigned
	}
	for i := range columns {
		minWidth := 3
		if i > 0 {
			minWidth = len(columns[i].Title) + 2
		}
		if widths[i] < minWidth {
			widths[i] = minWidth
		}
		columns[i].Width = widths[i]
	}
	logger.Debug(fmt.Sprintf("final widths: %v", widths))
	m.Table.SetColumns(columns)
}

func normalizeColumnPercentages(percentages []float64) []float64 {
	if len(percentages) == 0 {
		return percentages
	}

	total := 0.0
	for _, p := range percentages {
		total += p
	}

	if total == 0 {

		evenPercent := 1.0 / float64(len(percentages))
		normalized := make([]float64, len(percentages))
		for i := range normalized {
			normalized[i] = evenPercent
		}
		return normalized
	}

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
	m.allDataRows = rows
	m.filteredDataRows = rows
	m.applySearchFilter()
}

func (m *TableModel) UpdateColumns(columns []table.Column) {
	columns = append([]table.Column{{Title: "✓", Width: 3}}, columns...)
	m.Table.SetColumns(columns)
}

func (m *TableModel) refreshData() tea.Cmd {
	return m.fetchDataCmd()
}

func (t *TableModel) Refresh() (tea.Model, tea.Cmd) {
	if t.refreshFunc == nil {
		return t, nil
	}

	return t, t.fetchDataCmd()
}

func (t *TableModel) fetchDataCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		rows, err := t.refreshFunc()
		return fetchResultMsg{rows: rows, err: err}
	})
}

func (m *TableModel) applySearchFilter() {
	if m.searchQuery == "" {
		m.filteredDataRows = m.allDataRows
	} else {
		query := strings.ToLower(m.searchQuery)
		var filtered []table.Row
		for _, row := range m.allDataRows {
			if m.matchesSearch(row, query) {
				filtered = append(filtered, row)
			}
		}
		m.filteredDataRows = filtered
	}

	newRows := make([]table.Row, len(m.filteredDataRows))
	for i, row := range m.filteredDataRows {
		// Check if this row was originally checked
		originalIndex := m.findOriginalRowIndex(row)
		if originalIndex >= 0 && m.checkedRows[originalIndex] {
			newRows[i] = append(table.Row{"🗹"}, row...)
		} else {
			newRows[i] = append(table.Row{"▢"}, row...)
		}
	}
	m.Table.SetRows(newRows)
}

func (m *TableModel) findOriginalRowIndex(targetRow table.Row) int {
	for i, row := range m.allDataRows {
		if m.rowsEqual(row, targetRow) {
			return i
		}
	}
	return -1
}

func (m *TableModel) rowsEqual(row1, row2 table.Row) bool {
	if len(row1) != len(row2) {
		return false
	}
	for i := range row1 {
		if row1[i] != row2[i] {
			return false
		}
	}
	return true
}

func (m *TableModel) matchesSearch(row table.Row, query string) bool {
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), query) {
			return true
		}
	}
	return false
}
