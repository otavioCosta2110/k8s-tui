package models

import (
	"fmt"
	styles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	"time"

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
	Loading            bool
	LastUpdated        time.Time
}

type MetricsManager struct {
	loader   *k8s.MetricsLoader
	metrics  Metrics
	lastLoad time.Time
}

var metricsManager *MetricsManager

func (m Metrics) GetMetrics() Metrics {
	return m
}

func NewMetricsManager(k k8s.Client) *MetricsManager {
	if metricsManager == nil {
		metricsManager = &MetricsManager{
			loader: k8s.NewMetricsLoader(k),
		}
		metricsManager.loader.Start()
	}
	return metricsManager
}

func (mm *MetricsManager) GetMetrics() Metrics {
	if time.Since(mm.lastLoad) > 30*time.Second {
		mm.loader.Start()
		mm.lastLoad = time.Now()
	}

	k8sMetrics := mm.loader.GetMetrics()

	return Metrics{
		PodsNumber:         k8sMetrics.PodsNumber,
		NodesNumber:        k8sMetrics.NodesNumber,
		NamespacesNumber:   k8sMetrics.NamespacesNumber,
		DeploymentsNumber:  k8sMetrics.DeploymentsNumber,
		ServicesNumber:     k8sMetrics.ServicesNumber,
		ReplicaSetsNumber:  k8sMetrics.ReplicaSetsNumber,
		StatefulSetsNumber: k8sMetrics.StatefulSetsNumber,
		DaemonSetsNumber:   k8sMetrics.DaemonSetsNumber,
		JobsNumber:         k8sMetrics.JobsNumber,
		Error:              k8sMetrics.Error,
		Loading:            mm.loader.IsLoading(),
		LastUpdated:        k8sMetrics.LastUpdated,
	}
}

func (mm *MetricsManager) Stop() {
	if mm.loader != nil {
		mm.loader.Stop()
	}
}

func NewMetrics(k k8s.Client) (Metrics, error) {
	manager := NewMetricsManager(k)
	return manager.GetMetrics(), nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	metricStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(customstyles.BackgroundColor))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1EFD3")).
			Background(lipgloss.Color(customstyles.BackgroundColor))

	sectionStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.BackgroundColor))
)

func (m Metrics) ViewMetrics() string {
	columnWidth := (styles.ScreenWidth) / 3

	sectionStyle = sectionStyle.
		Width(columnWidth).
		Padding(0, 2)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Italic(true).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	formatMetric := func(label string, value int, loading bool) string {
		if loading && value == 0 {
			return metricStyle.Render(label+":") + " " + loadingStyle.Render("Loading...")
		}
		return metricStyle.Render(label+":") + " " + valueStyle.Render(fmt.Sprint(value))
	}

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
		formatMetric("Pods", m.PodsNumber, m.Loading),
		formatMetric("Nodes", m.NodesNumber, m.Loading),
		formatMetric("Namespaces", m.NamespacesNumber, m.Loading),
	)

	workloadsSection := createSection(
		"Workloads",
		formatMetric("Deployments", m.DeploymentsNumber, m.Loading),
		formatMetric("ReplicaSets", m.ReplicaSetsNumber, m.Loading),
		formatMetric("StatefulSets", m.StatefulSetsNumber, m.Loading),
	)

	servicesSection := createSection(
		"Services & Jobs",
		formatMetric("Services", m.ServicesNumber, m.Loading),
		formatMetric("DaemonSets", m.DaemonSetsNumber, m.Loading),
		formatMetric("Jobs", m.JobsNumber, m.Loading),
	)

	return lipgloss.PlaceHorizontal(
		styles.ScreenWidth,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			clusterSection,
			workloadsSection,
			servicesSection,
		),
	)
}
