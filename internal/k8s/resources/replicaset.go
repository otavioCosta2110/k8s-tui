package k8s

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/format"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReplicaSetInfo struct {
	Namespace string
	Name      string
	Desired   string
	Current   string
	Ready     string
	Age       string
	Raw       *appsv1.ReplicaSet
	Client    Client
}

func NewReplicaSet(name, namespace string, k Client) *ReplicaSetInfo {
	return &ReplicaSetInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func (r *ReplicaSetInfo) Fetch() error {
	replicaSet, err := r.Client.Clientset.AppsV1().ReplicaSets(r.Namespace).Get(
		context.Background(),
		r.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to get replicaset: %v", err)
	}
	r.Raw = replicaSet
	return nil
}

func FetchReplicaSetList(client Client, namespace string) ([]string, error) {
	rs, err := client.Clientset.AppsV1().ReplicaSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch replicasets: %v", err)
	}

	replicaSetNames := make([]string, 0, len(rs.Items))
	for _, replicaSet := range rs.Items {
		replicaSetNames = append(replicaSetNames, replicaSet.Name)
	}

	return replicaSetNames, nil
}

func GetReplicaSetsTableData(client Client, namespace string) ([]ReplicaSetInfo, error) {
	replicaSets, err := client.Clientset.AppsV1().ReplicaSets(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets: %v", err)
	}

	var replicaSetInfos []ReplicaSetInfo
	for _, replicaSet := range replicaSets.Items {
		freshReplicaSet, err := client.Clientset.AppsV1().ReplicaSets(namespace).Get(
			context.Background(),
			replicaSet.Name,
			metav1.GetOptions{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get fresh status for replicaset %s: %v", replicaSet.Name, err)
		}

		status := freshReplicaSet.Status
		spec := freshReplicaSet.Spec

		var desiredReplicas int32
		if spec.Replicas != nil {
			desiredReplicas = *spec.Replicas
		}

		readyStr := fmt.Sprintf("%d/%d", status.ReadyReplicas, desiredReplicas)

		replicaSetInfos = append(replicaSetInfos, ReplicaSetInfo{
			Namespace: freshReplicaSet.Namespace,
			Name:      freshReplicaSet.Name,
			Desired:   fmt.Sprintf("%d", desiredReplicas),
			Current:   fmt.Sprintf("%d", status.Replicas),
			Ready:     readyStr,
			Age:       format.FormatAge(freshReplicaSet.CreationTimestamp.Time),
			Raw:       freshReplicaSet.DeepCopy(),
			Client:    client,
		})
	}

	return replicaSetInfos, nil
}

func (r *ReplicaSetInfo) GetPods() ([]PodInfo, error) {
	selector, err := r.GetLabelSelector()
	if err != nil {
		return nil, err
	}
	pods, err := FetchPods(r.Client, r.Namespace, selector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (r *ReplicaSetInfo) GetLabelSelector() (string, error) {
	if r.Raw == nil {
		return "", fmt.Errorf("replicaset raw data not available")
	}

	if r.Raw.Spec.Selector == nil {
		return "", fmt.Errorf("replicaset has no selector")
	}

	
	requirements, err := metav1.LabelSelectorAsSelector(r.Raw.Spec.Selector)
	if err != nil {
		return "", fmt.Errorf("failed to convert label selector: %v", err)
	}

	return requirements.String(), nil
}

func DeleteReplicaSet(client Client, namespace string, replicaSetName string) error {
	err := client.Clientset.AppsV1().ReplicaSets(namespace).Delete(context.Background(), replicaSetName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete replicaset %s: %v", replicaSetName, err)
	}
	return nil
}
