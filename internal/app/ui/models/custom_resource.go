package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type customResourceModel struct {
	resourceTypeName string
	resourceData     []types.ResourceData
	k8sClient        *k8s.Client
	namespace        string
	resourceType     string
}

func NewCustomResourceModel(k k8s.Client, namespace string, resourceType string) (*AutoRefreshModel, error) {
	logger.Info("=== Creating Custom Resource Model ===")
	logger.Info(fmt.Sprintf("Resource type: %s, Namespace: %s", resourceType, namespace))

	var customRT plugins.CustomResourceType
	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		logger.Info(fmt.Sprintf("Plugin manager available, checking %d custom resource types", len(pm.GetRegistry().GetCustomResourceTypes())))
		for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
			logger.Info(fmt.Sprintf("Available custom resource: %s (type: %s)", rt.Name, rt.Type))
			if rt.Type == resourceType {
				customRT = rt
				logger.Info(fmt.Sprintf("Found matching custom resource: %s", rt.Name))
				break
			}
		}
		if customRT.Type == "" {
			logger.Error(fmt.Sprintf("No custom resource found for type: %s", resourceType))
			logger.Error("Available custom resources:")
			for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
				logger.Error(fmt.Sprintf("  - %s (type: %s)", rt.Name, rt.Type))
			}
		}
	} else {
		logger.Error("Plugin manager not available")
	}

	if customRT.Type == "" {
		logger.Error(fmt.Sprintf("Custom resource type %s not found", resourceType))
		return nil, fmt.Errorf("custom resource type %s not found", resourceType)
	}

	logger.Debug("Creating custom resource model instance")
	cr := &customResourceModel{
		resourceTypeName: customRT.Name,
		resourceData:     make([]types.ResourceData, 0),
		k8sClient:        &k,
		namespace:        namespace,
		resourceType:     resourceType,
	}
	logger.Debug("Custom resource model instance created")

	logger.Info("Fetching initial data")
	if err := cr.fetchData(); err != nil {
		logger.Error(fmt.Sprintf("Error fetching initial data: %v", err))
		return nil, err
	}
	logger.Info(fmt.Sprintf("Initial data fetched successfully: %d items", len(cr.resourceData)))

	logger.Debug("Creating view model based on display component")
	var viewModel tea.Model

	switch customRT.DisplayComponent.Type {
	case "yaml", "json":
		logger.Info("Creating YAML view for custom resource")
		viewModel = NewCustomResourceYAMLModel(cr, customRT.Name, customRT.Icon, namespace, customRT.DisplayComponent)
	case "chart", "gauge":
		logger.Info("Creating chart view for custom resource")
		viewModel = NewCustomResourceChartModel(cr, customRT.Name, customRT.Icon, namespace, customRT.DisplayComponent)
	case "table":
		logger.Info("Creating table view for custom resource")
		viewModel = NewCustomResourceTableModel(cr, customRT.Name, customRT.Icon, namespace, customRT.DisplayComponent)
	default:
		logger.Info("Creating default text view for custom resource")
		viewModel = NewCustomResourceTextModel(cr, customRT.Name, customRT.Icon, namespace)
	}

	logger.Debug("View model created")

	logger.Debug("Creating auto refresh model")
	if refreshableModel, ok := viewModel.(RefreshableModel); ok {
		result := NewAutoRefreshModel(refreshableModel, customRT.RefreshInterval, &k, customRT.Name)
		logger.Debug("Auto refresh model created successfully")
		return result, nil
	} else {
		logger.Error("View model does not implement RefreshableModel interface")
		return nil, fmt.Errorf("view model does not implement RefreshableModel interface")
	}
}

type CustomResourceTextModel struct {
	crModel      *customResourceModel
	resourceName string
	icon         string
	namespace    string
}

func NewCustomResourceTextModel(cr *customResourceModel, resourceName, icon, namespace string) *CustomResourceTextModel {
	return &CustomResourceTextModel{
		crModel:      cr,
		resourceName: resourceName,
		icon:         icon,
		namespace:    namespace,
	}
}

func (ct *CustomResourceTextModel) Init() tea.Cmd {
	return nil
}

func (ct *CustomResourceTextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		styles.ScreenWidth = msg.Width
		styles.ScreenHeight = msg.Height
	}
	return ct, nil
}

func (ct *CustomResourceTextModel) View() string {
	if ct.crModel.resourceData == nil || len(ct.crModel.resourceData) == 0 {
		return ct.renderEmptyState()
	}

	return ct.renderData()
}

func (ct *CustomResourceTextModel) renderEmptyState() string {
	style := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#888888")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := ct.icon + " " + ct.resourceName
	if ct.namespace != "" && ct.namespace != "default" {
		title += " in " + ct.namespace
	}

	return style.Render(title + "\n\nNo data available")
}

func (ct *CustomResourceTextModel) renderData() string {
	var content []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := ct.icon + " " + ct.resourceName
	if ct.namespace != "" && ct.namespace != "default" {
		title += " in " + ct.namespace
	}
	content = append(content, titleStyle.Render(title))
	content = append(content, "")

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A1EFD3")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	for _, item := range ct.crModel.resourceData {
		columns := item.GetColumns()
		if len(columns) >= 4 {
			name := columns[0]
			status := columns[2]
			value := columns[3]

			line := fmt.Sprintf("• %s: %s (%s)", name, value, status)
			content = append(content, itemStyle.Render(line))
		}
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	containerStyle := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Padding(1, 2).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return containerStyle.Render(fullContent)
}

func (ct *CustomResourceTextModel) Refresh() (tea.Model, tea.Cmd) {
	if err := ct.crModel.fetchData(); err != nil {
		logger.PluginError(ct.crModel.resourceTypeName, fmt.Sprintf("Error refreshing data: %v", err))
	}
	return ct, nil
}

func (cr *customResourceModel) fetchData() error {
	logger.Debug(fmt.Sprintf("Fetching data for resource type: %s, namespace: %s", cr.resourceType, cr.namespace))

	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		logger.Debug("Plugin manager available, calling GetCustomResourceData")
		data, err := pm.GetCustomResourceData(*cr.k8sClient, cr.resourceType, cr.namespace)
		if err != nil {
			logger.Error(fmt.Sprintf("Error from GetCustomResourceData: %v", err))
			return err
		}
		cr.resourceData = data
		logger.Debug(fmt.Sprintf("Data fetched successfully, %d items", len(data)))
	} else {
		logger.Debug("Plugin manager not available")
	}
	return nil
}

func (cr *customResourceModel) Init() tea.Cmd {
	return nil
}

func (cr *customResourceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return cr, nil
}

func (cr *customResourceModel) View() string {
	return "Custom Resource View"
}

func (cr *customResourceModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	return cr, nil
}

type CustomResourceYAMLModel struct {
	crModel      *customResourceModel
	resourceName string
	icon         string
	namespace    string
	displayComp  plugins.DisplayComponent
}

func NewCustomResourceYAMLModel(cr *customResourceModel, resourceName, icon, namespace string, displayComp plugins.DisplayComponent) *CustomResourceYAMLModel {
	return &CustomResourceYAMLModel{
		crModel:      cr,
		resourceName: resourceName,
		icon:         icon,
		namespace:    namespace,
		displayComp:  displayComp,
	}
}

func (cy *CustomResourceYAMLModel) Init() tea.Cmd {
	return nil
}

func (cy *CustomResourceYAMLModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		styles.ScreenWidth = msg.Width
		styles.ScreenHeight = msg.Height
	}
	return cy, nil
}

func (cy *CustomResourceYAMLModel) View() string {
	if cy.crModel.resourceData == nil || len(cy.crModel.resourceData) == 0 {
		return cy.renderEmptyState()
	}

	return cy.renderYAML()
}

func (cy *CustomResourceYAMLModel) renderEmptyState() string {
	style := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#888888")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := cy.icon + " " + cy.resourceName
	if cy.namespace != "" && cy.namespace != "default" {
		title += " in " + cy.namespace
	}

	return style.Render(title + "\n\nNo data available")
}

func (cy *CustomResourceYAMLModel) renderYAML() string {
	var content []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := cy.icon + " " + cy.resourceName
	if cy.namespace != "" && cy.namespace != "default" {
		title += " in " + cy.namespace
	}
	content = append(content, titleStyle.Render(title))
	content = append(content, "")

	for _, item := range cy.crModel.resourceData {
		if luaData, ok := item.(*plugins.LuaResourceData); ok {
			fields := luaData.GetFields()
			yamlData, err := yaml.Marshal(fields)
			if err == nil {
				yamlStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#A1EFD3")).
					Background(lipgloss.Color(customstyles.BackgroundColor))
				content = append(content, yamlStyle.Render(string(yamlData)))
				content = append(content, "")
			}
		}
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	containerStyle := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Padding(1, 2).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return containerStyle.Render(fullContent)
}

func (cy *CustomResourceYAMLModel) Refresh() (tea.Model, tea.Cmd) {
	if err := cy.crModel.fetchData(); err != nil {
		logger.PluginError(cy.crModel.resourceTypeName, fmt.Sprintf("Error refreshing data: %v", err))
	}
	return cy, nil
}

type CustomResourceChartModel struct {
	crModel      *customResourceModel
	resourceName string
	icon         string
	namespace    string
	displayComp  plugins.DisplayComponent
}

func NewCustomResourceChartModel(cr *customResourceModel, resourceName, icon, namespace string, displayComp plugins.DisplayComponent) *CustomResourceChartModel {
	return &CustomResourceChartModel{
		crModel:      cr,
		resourceName: resourceName,
		icon:         icon,
		namespace:    namespace,
		displayComp:  displayComp,
	}
}

func (cc *CustomResourceChartModel) Init() tea.Cmd {
	return nil
}

func (cc *CustomResourceChartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		styles.ScreenWidth = msg.Width
		styles.ScreenHeight = msg.Height
	}
	return cc, nil
}

func (cc *CustomResourceChartModel) View() string {
	if cc.crModel.resourceData == nil || len(cc.crModel.resourceData) == 0 {
		return cc.renderEmptyState()
	}

	return cc.renderChart()
}

func (cc *CustomResourceChartModel) renderEmptyState() string {
	style := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#888888")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := cc.icon + " " + cc.resourceName
	if cc.namespace != "" && cc.namespace != "default" {
		title += " in " + cc.namespace
	}

	return style.Render(title + "\n\nNo data available")
}

func (cc *CustomResourceChartModel) renderChart() string {
	var content []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	title := cc.icon + " " + cc.resourceName
	if cc.namespace != "" && cc.namespace != "default" {
		title += " in " + cc.namespace
	}
	content = append(content, titleStyle.Render(title))
	content = append(content, "")

	for _, item := range cc.crModel.resourceData {
		columns := item.GetColumns()
		if len(columns) >= 3 {
			name := columns[0]
			status := columns[1]
			age := columns[2]

			bar := strings.Repeat("█", 20) 
			line := fmt.Sprintf("%s: %s (%s, %s)", name, bar, status, age)

			chartStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A1EFD3")).
				Background(lipgloss.Color(customstyles.BackgroundColor))
			content = append(content, chartStyle.Render(line))
		}
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	containerStyle := lipgloss.NewStyle().
		Width(styles.ScreenWidth).
		Height(styles.ScreenHeight).
		Padding(1, 2).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return containerStyle.Render(fullContent)
}

func (cc *CustomResourceChartModel) Refresh() (tea.Model, tea.Cmd) {
	if err := cc.crModel.fetchData(); err != nil {
		logger.PluginError(cc.crModel.resourceTypeName, fmt.Sprintf("Error refreshing data: %v", err))
	}
	return cc, nil
}

type CustomResourceTableModel struct {
	tableModel  *components.TableModel
	crModel     *customResourceModel
	displayComp plugins.DisplayComponent
}

func NewCustomResourceTableModel(cr *customResourceModel, resourceName, icon, namespace string, displayComp plugins.DisplayComponent) *CustomResourceTableModel {
	logger.Info(fmt.Sprintf("Creating table model for %s with %d items", resourceName, len(cr.resourceData)))

	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Namespace", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Age", Width: 10},
	}

	var colWidths []float64

	if configWidths, ok := displayComp.Config["ColumnWidths"].([]interface{}); ok && len(configWidths) > 0 {
		numWidths := len(configWidths)
		if numWidths > len(columns) {
			numWidths = len(columns)
		}
		colWidths = make([]float64, numWidths)
		for i := 0; i < numWidths; i++ {
			if width, ok := configWidths[i].(float64); ok {
				colWidths[i] = width
			}
		}
		if len(colWidths) < len(columns) {
			remainingWidth := 1.0
			for _, w := range colWidths {
				remainingWidth -= w
			}
			remainingCols := len(columns) - len(colWidths)
			if remainingCols > 0 {
				avgWidth := remainingWidth / float64(remainingCols)
				for len(colWidths) < len(columns) {
					colWidths = append(colWidths, avgWidth)
				}
			}
		}
		logger.Info(fmt.Sprintf("Using custom column widths from config: %v", colWidths))
	} else {
		defaultWidth := 1.0 / float64(len(columns))
		colWidths = make([]float64, len(columns))
		for i := range colWidths {
			colWidths[i] = defaultWidth
		}
		logger.Info("Using default column widths")
	}

	rows := make([]table.Row, 0, len(cr.resourceData))
	if len(cr.resourceData) == 0 {
		rows = append(rows, table.Row{"No data available", "", "", ""})
		logger.Warn("No resource data available, adding placeholder row")
	} else {
		for i, item := range cr.resourceData {
			if item == nil {
				logger.Warn(fmt.Sprintf("Item %d is nil, skipping", i))
				continue
			}
			itemColumns := item.GetColumns()
			if len(itemColumns) >= 4 {
				rows = append(rows, itemColumns[:4]) 
				logger.Debug(fmt.Sprintf("Row %d: %v", i, itemColumns[:4]))
			} else if len(itemColumns) > 0 {
				paddedRow := make(table.Row, 4)
				copy(paddedRow, itemColumns)
				for j := len(itemColumns); j < 4; j++ {
					paddedRow[j] = ""
				}
				rows = append(rows, paddedRow)
				logger.Debug(fmt.Sprintf("Row %d (padded): %v", i, paddedRow))
			} else {
				rows = append(rows, table.Row{"N/A", "N/A", "N/A", "N/A"})
				logger.Warn(fmt.Sprintf("Item %d has no columns, using fallback", i))
			}
		}
	}

	title := icon + " " + resourceName
	if namespace != "" && namespace != "default" {
		title += " in " + namespace
	}

	logger.Info(fmt.Sprintf("Creating table with %d columns and %d rows", len(columns), len(rows)))

	tableModel := components.NewTable(
		columns,
		colWidths,
		rows,
		title,
		nil, 
		0,   
		func() ([]table.Row, error) {
			logger.Debug("Refreshing table data")
			if err := cr.fetchData(); err != nil {
				logger.Error(fmt.Sprintf("Error refreshing data: %v", err))
				return nil, err
			}
			newRows := make([]table.Row, 0, len(cr.resourceData))
			for i, item := range cr.resourceData {
				if item == nil {
					logger.Warn(fmt.Sprintf("Item %d is nil during refresh, skipping", i))
					continue
				}
				columns := item.GetColumns()
				if len(columns) > 0 {
					newRows = append(newRows, columns)
				} else {
					logger.Warn(fmt.Sprintf("Item %d has no columns during refresh, using fallback", i))
					newRows = append(newRows, table.Row{"N/A", "N/A", "N/A", "N/A"})
				}
			}
			if len(newRows) == 0 {
				newRows = append(newRows, table.Row{"No data available", "", "", ""})
			}
			logger.Debug(fmt.Sprintf("Refreshed to %d rows", len(newRows)))
			return newRows, nil
		},
		nil, 
	)

	logger.Info("Table model created successfully")

	return &CustomResourceTableModel{
		tableModel:  tableModel,
		crModel:     cr,
		displayComp: displayComp,
	}
}

func (ct *CustomResourceTableModel) Init() tea.Cmd {
	return ct.tableModel.Init()
}

func (ct *CustomResourceTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedModel, cmd := ct.tableModel.Update(msg)
	if tm, ok := updatedModel.(*components.TableModel); ok {
		ct.tableModel = tm
	}
	return ct, cmd
}

func (ct *CustomResourceTableModel) View() string {
	return ct.tableModel.View()
}

func (ct *CustomResourceTableModel) Refresh() (tea.Model, tea.Cmd) {
	return ct.tableModel.Refresh()
}
