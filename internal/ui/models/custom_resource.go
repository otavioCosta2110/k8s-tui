package models

import (
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/plugins"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	customstyles "otaviocosta2110/k8s-tui/internal/ui/custom_styles"
	"otaviocosta2110/k8s-tui/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type customResourceModel struct {
	*GenericResourceModel
	resourceTypeName string
}

func NewCustomResourceModel(k k8s.Client, namespace string, resourceType string) (*AutoRefreshModel, error) {
	utils.WriteStringNewLine("debug.log", "=== Creating Custom Resource Model ===")
	utils.WriteStringNewLine("debug.log", fmt.Sprintf("Resource type: %s, Namespace: %s", resourceType, namespace))

	// Get custom resource type info from plugins
	var customRT plugins.CustomResourceType
	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Plugin manager available, checking %d custom resource types", len(pm.GetRegistry().GetCustomResourceTypes())))
		for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Checking custom resource: %s (type: %s)", rt.Name, rt.Type))
			if rt.Type == resourceType {
				customRT = rt
				utils.WriteStringNewLine("debug.log", fmt.Sprintf("Found matching custom resource: %s", rt.Name))
				break
			}
		}
	}

	// Calculate column widths dynamically based on number of columns
	numColumns := len(customRT.Columns)
	utils.WriteStringNewLine("debug.log", "Number of columns: "+fmt.Sprintf("%d", numColumns))
	var columnWidths []float64
	if numColumns == 0 {
		columnWidths = []float64{1.0}
		utils.WriteStringNewLine("debug.log", "No columns defined, using default width")
	} else {
		width := 1.0 / float64(numColumns)
		columnWidths = make([]float64, numColumns)
		for i := range columnWidths {
			columnWidths[i] = width - 0.01 // Slightly reduce to avoid overflow
		}
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Calculated %d column widths: %v", numColumns, columnWidths))
	}

	config := ResourceConfig{
		ResourceType:    k8s.ResourceType(resourceType),
		Title:           customstyles.ResourceIcons["ResourceList"] + " " + customRT.Name + " in " + namespace,
		ColumnWidths:    columnWidths,
		RefreshInterval: customRT.RefreshInterval,
		Columns:         customRT.Columns,
	}

	utils.WriteStringNewLine("debug.log", fmt.Sprintf("Config created with %d columns and %d widths", len(config.Columns), len(config.ColumnWidths)))

	utils.WriteStringNewLine("debug.log", "Creating custom resource model instance")
	cr := &customResourceModel{
		GenericResourceModel: NewGenericResourceModel(k, namespace, config),
		resourceTypeName:     customRT.Name,
	}
	utils.WriteStringNewLine("debug.log", "Custom resource model instance created")

	utils.WriteStringNewLine("debug.log", "Fetching initial data")
	if err := cr.fetchData(); err != nil {
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Error fetching initial data: %v", err))
		return nil, err
	}
	utils.WriteStringNewLine("debug.log", "Initial data fetched successfully")

	onSelect := func(selected string) tea.Msg {
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Custom resource item selected: %s", selected))
		// For now, just show an error that details view is not implemented
		return components.NavigateMsg{
			Error:   fmt.Errorf("Custom resource details view not yet implemented"),
			Cluster: k,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		utils.WriteStringNewLine("debug.log", "Fetching data for table refresh")
		if err := cr.fetchData(); err != nil {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Error fetching data for refresh: %v", err))
			return nil, err
		}
		rows := cr.dataToRows()
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Data refresh completed, %d rows returned", len(rows)))
		return rows, nil
	}

	utils.WriteStringNewLine("debug.log", fmt.Sprintf("Creating table with %d columns, %d widths, %d initial rows", len(cr.config.Columns), len(cr.config.ColumnWidths), len(cr.dataToRows())))
	tableModel := ui.NewTable(cr.config.Columns, cr.config.ColumnWidths, cr.dataToRows(), cr.config.Title, onSelect, 1, fetchFunc, nil)
	utils.WriteStringNewLine("debug.log", "Table model created")

	actions := map[string]func() tea.Cmd{
		"d": cr.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)
	utils.WriteStringNewLine("debug.log", "Table actions set")

	utils.WriteStringNewLine("debug.log", "Creating auto refresh model")
	result := NewAutoRefreshModel(tableModel, cr.refreshInterval, cr.k8sClient, customRT.Name)
	utils.WriteStringNewLine("debug.log", "Auto refresh model created successfully")

	return result, nil
}

func (cr *customResourceModel) fetchData() error {
	utils.WriteStringNewLine("debug.log", fmt.Sprintf("Fetching data for resource type: %s, namespace: %s", cr.resourceType, cr.namespace))

	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		utils.WriteStringNewLine("debug.log", "Plugin manager available, calling GetCustomResourceData")
		data, err := pm.GetCustomResourceData(*cr.k8sClient, string(cr.resourceType), cr.namespace)
		if err != nil {
			utils.WriteStringNewLine("debug.log", fmt.Sprintf("Error from GetCustomResourceData: %v", err))
			return err
		}
		cr.resourceData = data
		utils.WriteStringNewLine("debug.log", fmt.Sprintf("Data fetched successfully, %d items", len(data)))
	} else {
		utils.WriteStringNewLine("debug.log", "Plugin manager not available")
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
