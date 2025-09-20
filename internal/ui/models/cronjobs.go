package models

import (
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type cronjobsModel struct {
	*GenericResourceModel
	cronjobsInfo []k8s.CronJobInfo
}

func NewCronJobs(k k8s.Client, namespace string) (*cronjobsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeCronJob,
		Title:           customstyles.ResourceIcons["CronJobs"] + " CronJobs in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.20, 0.10, 0.05, 0.1, 0.07},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("SCHEDULE", 0),
			components.NewColumn("SUSPEND", 0),
			components.NewColumn("ACTIVE", 0),
			components.NewColumn("LAST SCHEDULE", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &cronjobsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (cj *cronjobsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	cj.k8sClient = k

	if err := cj.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		cronjobDetails, err := NewCronJobDetails(*k, cj.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: cronjobDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := cj.fetchData(); err != nil {
			return nil, err
		}
		return cj.dataToRows(), nil
	}

	tableModel := ui.NewTable(cj.config.Columns, cj.config.ColumnWidths, cj.dataToRows(), cj.config.Title, onSelect, 1, fetchFunc, nil)

	actions := map[string]func() tea.Cmd{
		"d": cj.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, cj.refreshInterval, cj.k8sClient, "CronJobs"), nil
}

func (cj *cronjobsModel) fetchData() error {
	cronjobInfo, err := k8s.GetCronJobsTableData(*cj.k8sClient, cj.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch cronjobs: %v", err)
	}
	cj.cronjobsInfo = cronjobInfo

	cj.resourceData = make([]types.ResourceData, len(cronjobInfo))
	for idx, cronjob := range cronjobInfo {
		cj.resourceData[idx] = CronJobData{&cronjob}
	}

	return nil
}
