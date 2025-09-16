package models

import (
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"
	"otaviocosta2110/k8s-tui/internal/ui/components"
	ui "otaviocosta2110/k8s-tui/internal/ui/components"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type jobsModel struct {
	*GenericResourceModel
	jobsInfo []k8s.JobInfo
}

func NewJobs(k k8s.Client, namespace string) (*jobsModel, error) {
	config := ResourceConfig{
		ResourceType:    k8s.ResourceTypeJob,
		Title:           "Jobs in " + namespace,
		ColumnWidths:    []float64{0.15, 0.25, 0.20, 0.20, 0.13},
		RefreshInterval: 5 * time.Second,
		Columns: []table.Column{
			components.NewColumn("NAMESPACE", 0),
			components.NewColumn("NAME", 0),
			components.NewColumn("COMPLETIONS", 0),
			components.NewColumn("DURATION", 0),
			components.NewColumn("AGE", 0),
		},
	}

	genericModel := NewGenericResourceModel(k, namespace, config)

	model := &jobsModel{
		GenericResourceModel: genericModel,
	}

	return model, nil
}

func (j *jobsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	j.k8sClient = k

	if err := j.fetchData(); err != nil {
		return nil, err
	}

	onSelect := func(selected string) tea.Msg {
		jobDetails, err := NewJobDetails(*k, j.namespace, selected).InitComponent(k)
		if err != nil {
			return components.NavigateMsg{
				Error:   err,
				Cluster: *k,
			}
		}
		return components.NavigateMsg{
			NewScreen: jobDetails,
		}
	}

	fetchFunc := func() ([]table.Row, error) {
		if err := j.fetchData(); err != nil {
			return nil, err
		}
		return j.dataToRows(), nil
	}

	tableModel := ui.NewTable(j.config.Columns, j.config.ColumnWidths, j.dataToRows(), j.config.Title, onSelect, 1, fetchFunc, nil, "")

	actions := map[string]func() tea.Cmd{
		"d": j.createDeleteAction(tableModel),
	}
	tableModel.SetUpdateActions(actions)

	return NewAutoRefreshModel(tableModel, j.refreshInterval, j.k8sClient, "Jobs"), nil
}

func (j *jobsModel) fetchData() error {
	jobInfo, err := k8s.GetJobsTableData(*j.k8sClient, j.namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %v", err)
	}
	j.jobsInfo = jobInfo

	j.resourceData = make([]ResourceData, len(jobInfo))
	for idx, job := range jobInfo {
		j.resourceData[idx] = JobData{&job}
	}

	return nil
}

func (j *jobsModel) dataToRows() []table.Row {
	rows := make([]table.Row, len(j.jobsInfo))
	for idx, job := range j.jobsInfo {
		rows[idx] = table.Row{
			job.Namespace,
			job.Name,
			job.Completions,
			job.Duration,
			job.Age,
		}
	}
	return rows
}
