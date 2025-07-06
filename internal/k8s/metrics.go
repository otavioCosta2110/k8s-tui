package k8s

import (
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

func NewMetrics(k Client) Metrics {
	var metrics Metrics
	nm, err := FetchNamespaces(k)
	if err != nil {
		metrics.Error = err
	}
	metrics = Metrics{
		Error:              nil,
		PodsNumber:         0,
		NodesNumber:        0,
		NamespacesNumber:   len(nm),
		DeploymentsNumber:  0,
		ServicesNumber:     0,
		ReplicaSetsNumber:  0,
		StatefulSetsNumber: 0,
		DaemonSetsNumber:   0,
		JobsNumber:         0,
	}

	return metrics
}
