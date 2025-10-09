package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/models"
)

func (m *AppModel) createModelForBreadcrumb(breadcrumb []string) (tea.Model, error) {
	if len(breadcrumb) < 2 {
		return nil, fmt.Errorf("breadcrumb too short: %v", breadcrumb)
	}

	if breadcrumb[0] != "Resource List" {
		return nil, fmt.Errorf("invalid breadcrumb start: %s", breadcrumb[0])
	}

	resourceType := breadcrumb[1]

	if len(breadcrumb) == 2 {
		return models.NewResourceList(m.kube, m.config.DefaultNamespace, resourceType).InitComponent(m.kube)
	}

	if len(breadcrumb) >= 3 {
		resourceName := breadcrumb[2]
		return m.createResourceDetailsModel(resourceType, resourceName)
	}

	return nil, fmt.Errorf("unable to determine model type for breadcrumb: %v", breadcrumb)
}

func (m *AppModel) createResourceDetailsModel(resourceType, resourceName string) (tea.Model, error) {
	switch resourceType {
	case "Pods", "pods":
		return models.NewPodDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Services", "services":
		return models.NewServiceDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "ConfigMaps", "configmaps":
		return models.NewConfigmapDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Secrets", "secrets":
		return models.NewSecretDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Ingresses", "ingresses":
		return models.NewIngressDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Jobs", "jobs":
		return models.NewJobDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "CronJobs", "cronjobs":
		return models.NewCronJobDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "DaemonSets", "daemonsets":
		return models.NewDaemonSetDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "StatefulSets", "statefulsets":
		return models.NewStatefulSetDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "ServiceAccounts", "serviceaccounts":
		return models.NewServiceAccountDetails(m.kube, m.config.DefaultNamespace, resourceName).InitComponent(&m.kube)
	case "Nodes", "nodes":
		return models.NewNodeDetails(m.kube, resourceName).InitComponent(&m.kube)
	default:
		return nil, fmt.Errorf("unsupported resource type for details: %s", resourceType)
	}
}
