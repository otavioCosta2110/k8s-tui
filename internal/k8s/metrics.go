package k8s

import (
	"sync"
	"time"
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

type MetricsLoader struct {
	client   Client
	metrics  Metrics
	mutex    sync.RWMutex
	loading  bool
	stopChan chan struct{}
}

func NewMetricsLoader(k Client) *MetricsLoader {
	return &MetricsLoader{
		client:   k,
		stopChan: make(chan struct{}),
	}
}

func (ml *MetricsLoader) Start() {
	ml.mutex.Lock()
	if ml.loading {
		ml.mutex.Unlock()
		return
	}
	ml.loading = true
	ml.mutex.Unlock()

	go ml.loadMetrics()
}

func (ml *MetricsLoader) Stop() {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	if ml.loading {
		close(ml.stopChan)
		ml.loading = false
	}
}

func (ml *MetricsLoader) GetMetrics() Metrics {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.metrics
}

func (ml *MetricsLoader) IsLoading() bool {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.loading
}

func (ml *MetricsLoader) loadMetrics() {
	defer func() {
		ml.mutex.Lock()
		ml.loading = false
		ml.mutex.Unlock()
	}()

	select {
	case <-ml.stopChan:
		return
	default:
		ml.loadCriticalMetrics()
	}

	select {
	case <-ml.stopChan:
		return
	default:
		ml.loadNamespaceMetrics()
	}

	ml.mutex.Lock()
	ml.metrics.LastUpdated = time.Now()
	ml.mutex.Unlock()
}

func (ml *MetricsLoader) loadCriticalMetrics() {
	namespaces, err := FetchNamespaces(ml.client)
	ml.mutex.Lock()
	if err != nil {
		ml.metrics.Error = err
		ml.mutex.Unlock()
		return
	}
	ml.metrics.NamespacesNumber = len(namespaces)
	ml.mutex.Unlock()

	nodes, err := FetchNodeList(ml.client)
	ml.mutex.Lock()
	if err == nil {
		ml.metrics.NodesNumber = len(nodes)
	}
	ml.mutex.Unlock()
}

func (ml *MetricsLoader) loadNamespaceMetrics() {
	namespaces, err := FetchNamespaces(ml.client)
	if err != nil {
		ml.mutex.Lock()
		ml.metrics.Error = err
		ml.mutex.Unlock()
		return
	}

	totalPods := 0
	totalDeployments := 0
	totalServices := 0
	totalReplicaSets := 0

	batchSize := 5
	for i := 0; i < len(namespaces); i += batchSize {
		select {
		case <-ml.stopChan:
			return
		default:
		}

		end := i + batchSize
		if end > len(namespaces) {
			end = len(namespaces)
		}

		batch := namespaces[i:end]
		ml.loadBatchMetrics(batch, &totalPods, &totalDeployments, &totalServices, &totalReplicaSets)
	}

	ml.mutex.Lock()
	ml.metrics.PodsNumber = totalPods
	ml.metrics.DeploymentsNumber = totalDeployments
	ml.metrics.ServicesNumber = totalServices
	ml.metrics.ReplicaSetsNumber = totalReplicaSets
	ml.mutex.Unlock()
}

func (ml *MetricsLoader) loadBatchMetrics(namespaces []string, totalPods, totalDeployments, totalServices, totalReplicaSets *int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, namespace := range namespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()

			select {
			case <-ml.stopChan:
				return
			default:
			}

			if pods, err := FetchPods(ml.client, ns, ""); err == nil {
				mu.Lock()
				*totalPods += len(pods)
				mu.Unlock()
			}

			if deployments, err := FetchDeploymentList(ml.client, ns); err == nil {
				mu.Lock()
				*totalDeployments += len(deployments)
				mu.Unlock()
			}

			if services, err := FetchServiceList(ml.client, ns); err == nil {
				mu.Lock()
				*totalServices += len(services)
				mu.Unlock()
			}

			if replicaSets, err := FetchReplicaSetList(ml.client, ns); err == nil {
				mu.Lock()
				*totalReplicaSets += len(replicaSets)
				mu.Unlock()
			}
		}(namespace)
	}

	wg.Wait()
}

func NewMetrics(k Client) (Metrics, error) {
	loader := NewMetricsLoader(k)
	loader.Start()

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return loader.GetMetrics(), nil
		case <-ticker.C:
			if !loader.IsLoading() {
				return loader.GetMetrics(), nil
			}
		}
	}
}
