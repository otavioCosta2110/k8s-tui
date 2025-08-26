package k8s

import ()

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

func NewMetrics(k Client) (Metrics, error) {
	var metrics Metrics
	nm, err := FetchNamespaces(k)
	if err != nil {
		metrics.Error = err
		return metrics, err
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

	return metrics, nil
}
