package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type customResourceModel struct {
	*GenericResourceModel
	resourceTypeName string
}

func NewCustomResourceModel(k k8s.Client, namespace string, resourceType string) (*AutoRefreshModel, error) {
	logger.Debug("=== Creating Custom Resource Model ===")
	logger.Debug(fmt.Sprintf("Resource type: %s, Namespace: %s", resourceType, namespace))

	// Get custom resource type info from plugins
	var customRT plugins.CustomResourceType
	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		logger.Debug(fmt.Sprintf("Plugin manager available, checking %d custom resource types", len(pm.GetRegistry().GetCustomResourceTypes())))
		for _, rt := range pm.GetRegistry().GetCustomResourceTypes() {
			logger.Debug(fmt.Sprintf("Checking custom resource: %s (type: %s)", rt.Name, rt.Type))
			if rt.Type == resourceType {
				customRT = rt
				logger.Debug(fmt.Sprintf("Found matching custom resource: %s", rt.Name))
				break
			}
		}
	}

	// Calculate column widths dynamically based on number of columns
	numColumns := len(customRT.Columns)
	logger.Debug("Number of columns: " + fmt.Sprintf("%d", numColumns))
	var columnWidths []float64
	if numColumns == 0 {
		columnWidths = []float64{1.0}
		logger.Debug("No columns defined, using default width")
	} else {
		width := 1.0 / float64(numColumns)
		columnWidths = make([]float64, numColumns)
		for i := range columnWidths {
			columnWidths[i] = width - 0.01 // Slightly reduce to avoid overflow
		}
		logger.Debug(fmt.Sprintf("Calculated %d column widths: %v", numColumns, columnWidths))
	}

	config := ResourceConfig{
		ResourceType:    k8s.ResourceType(resourceType),
		Title:           customstyles.ResourceIcons["ResourceList"] + " " + customRT.Name + " in " + namespace,
		ColumnWidths:    columnWidths,
		RefreshInterval: customRT.RefreshInterval,
		Columns:         customRT.Columns,
	}

	logger.Debug(fmt.Sprintf("Config created with %d columns and %d widths", len(config.Columns), len(config.ColumnWidths)))

	logger.Debug("Creating custom resource model instance")
	cr := &customResourceModel{
		GenericResourceModel: NewGenericResourceModel(k, namespace, config),
		resourceTypeName:     customRT.Name,
	}
	logger.Debug("Custom resource model instance created")

	logger.Debug("Fetching initial data")
	if err := cr.fetchData(); err != nil {
		logger.Error(fmt.Sprintf("Error fetching initial data: %v", err))
		return nil, err
	}
	logger.Debug("Initial data fetched successfully")

	onSelect := func(selected string) tea.Msg {
		logger.Debug(fmt.Sprintf("Custom resource item selected: %s", selected))
		// For now, just show an error that details view is not implemented
		return components.NavigateMsg{
			Error:   fmt.Errorf("Custom resource details view not yet implemented"),
			Cluster: k,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		logger.Debug("Fetching data for table refresh")
		if err := cr.fetchData(); err != nil {
			logger.Error(fmt.Sprintf("Error fetching data for refresh: %v", err))
			return nil, err
		}
		rows := cr.dataToRows()
		logger.Debug(fmt.Sprintf("Data refresh completed, %d rows returned", len(rows)))
		return rows, nil
	}

	logger.Debug(fmt.Sprintf("Creating table with %d columns, %d widths, %d initial rows", len(cr.config.Columns), len(cr.config.ColumnWidths), len(cr.dataToRows())))
	tableModel := ui.NewTable(cr.config.Columns, cr.config.ColumnWidths, cr.dataToRows(), cr.config.Title, onSelect, 1, fetchFunc, nil)
	logger.Debug("Table model created")

	actions := map[string]func() tea.Cmd{
		"d": cr.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)
	logger.Debug("Table actions set")

	logger.Debug("Creating auto refresh model")
	result := NewAutoRefreshModel(tableModel, cr.refreshInterval, cr.k8sClient, customRT.Name)
	logger.Debug("Auto refresh model created successfully")

	return result, nil
}

func (cr *customResourceModel) fetchData() error {
	logger.Debug(fmt.Sprintf("Fetching data for resource type: %s, namespace: %s", cr.resourceType, cr.namespace))

	if pm := plugins.GetGlobalPluginManager(); pm != nil {
		logger.Debug("Plugin manager available, calling GetCustomResourceData")
		data, err := pm.GetCustomResourceData(*cr.k8sClient, string(cr.resourceType), cr.namespace)
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
