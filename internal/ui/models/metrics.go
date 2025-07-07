package models

import (
	"fmt"
	global "otaviocosta2110/k8s-tui/internal"
	"otaviocosta2110/k8s-tui/internal/k8s"

	"github.com/charmbracelet/lipgloss"
)

type Metrics struct {
	PodsNumber         int
	NodesNumber        int
	NamespacesNumber   int
	DeploymentsNumber  int
	ServicesNumber     int
	ReplicaSetsNumber  int
	StatefulSetsNumber int
	DaemonSetsNumber   int
	JobsNumber         int
	Error              error
}

func (m Metrics) GetMetrics() Metrics {
	return m
}

func NewMetrics(k k8s.Client) (Metrics, error) {
	metrics, err := k8s.NewMetrics(k)

	returnedMetrics := Metrics{
		PodsNumber:         metrics.PodsNumber,
		NodesNumber:        metrics.NodesNumber,
		NamespacesNumber:   metrics.NamespacesNumber,
		DeploymentsNumber:  metrics.DeploymentsNumber,
		ServicesNumber:     metrics.ServicesNumber,
		ReplicaSetsNumber:  metrics.ReplicaSetsNumber,
		StatefulSetsNumber: metrics.StatefulSetsNumber,
		DaemonSetsNumber:   metrics.DaemonSetsNumber,
		JobsNumber:         metrics.JobsNumber,
		Error:              metrics.Error,
	}


	if err != nil {
		metrics.Error = err
		return returnedMetrics, err
	}

	return returnedMetrics, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	metricStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1EFD3"))

	sectionStyle = lipgloss.NewStyle()
)

func (m Metrics) ViewMetrics() string {
	columnWidth := (global.ScreenWidth - global.Margin) / 3

	sectionStyle = sectionStyle.
		Width(columnWidth).
		Padding(0, 2)

	createSection := func(title string, metrics ...string) string {
		content := []string{titleStyle.Render(title)}
		for _, m := range metrics {
			content = append(content, m)
		}
		return sectionStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left, content...),
		)
	}

	clusterSection := createSection(
		"Cluster Metrics",
		metricStyle.Render("Pods:")+" "+valueStyle.Render(fmt.Sprint(m.PodsNumber)),
		metricStyle.Render("Nodes:")+" "+valueStyle.Render(fmt.Sprint(m.NodesNumber)),
		metricStyle.Render("Namespaces:")+" "+valueStyle.Render(fmt.Sprint(m.NamespacesNumber)),
	)

	workloadsSection := createSection(
		"Workloads",
		metricStyle.Render("Deployments:")+" "+valueStyle.Render(fmt.Sprint(m.DeploymentsNumber)),
		metricStyle.Render("ReplicaSets:")+" "+valueStyle.Render(fmt.Sprint(m.ReplicaSetsNumber)),
		metricStyle.Render("StatefulSets:")+" "+valueStyle.Render(fmt.Sprint(m.StatefulSetsNumber)),
	)

	servicesSection := createSection(
		"Services & Jobs",
		metricStyle.Render("Services:")+" "+valueStyle.Render(fmt.Sprint(m.ServicesNumber)),
		metricStyle.Render("DaemonSets:")+" "+valueStyle.Render(fmt.Sprint(m.DaemonSetsNumber)),
		metricStyle.Render("Jobs:")+" "+valueStyle.Render(fmt.Sprint(m.JobsNumber)),
	)

	return lipgloss.PlaceHorizontal(
		global.ScreenWidth,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			clusterSection,
			workloadsSection,
			servicesSection,
		),
	)
}
