package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeInfo struct {
	Name    string
	Roles   string
	Status  string
	Version string
	CPU     string
	Memory  string
	Pods    string
	Age     string
	Raw     *corev1.Node
	Client  Client
}

func NewNode(name string, k Client) *NodeInfo {
	return &NodeInfo{
		Name:   name,
		Client: k,
	}
}

func FetchNodeList(client Client) ([]string, error) {
	nodes, err := client.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %v", err)
	}

	nodeNames := make([]string, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames, nil
}

func GetNodesTableData(client Client) ([]NodeInfo, error) {
	nodes, err := client.Clientset.CoreV1().Nodes().List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	var nodeInfos []NodeInfo
	for _, node := range nodes.Items {
		roles := getNodeRoles(&node)

		status := getNodeStatus(&node)

		version := node.Status.NodeInfo.KubeletVersion

		cpu, memory := getNodeResources(&node)

		pods := "N/A" 

		nodeInfos = append(nodeInfos, NodeInfo{
			Name:    node.Name,
			Roles:   roles,
			Status:  status,
			Version: version,
			CPU:     cpu,
			Memory:  memory,
			Pods:    pods,
			Age:     utils.FormatAge(node.CreationTimestamp.Time),
			Raw:     node.DeepCopy(),
			Client:  client,
		})
	}

	return nodeInfos, nil
}

func getNodeRoles(node *corev1.Node) string {
	var roles []string

	if node.Labels != nil {
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			roles = append(roles, "master")
		}
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			roles = append(roles, "control-plane")
		}
		if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
			roles = append(roles, "worker")
		}
	}

	if len(roles) == 0 {
		return "<none>"
	}

	return fmt.Sprintf("%v", roles)
}

func getNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			} else {
				return "NotReady"
			}
		}
	}
	return "Unknown"
}

func getNodeResources(node *corev1.Node) (cpu, memory string) {
	if node.Status.Capacity != nil {
		if cpuQty, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			cpu = cpuQty.String()
		}
		if memQty, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			memory = utils.FormatBytes(memQty.String())
		}
	}
	return cpu, memory
}

func (n *NodeInfo) Fetch() error {
	node, err := n.Client.Clientset.CoreV1().Nodes().Get(context.Background(), n.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node: %v", err)
	}
	n.Raw = node
	return nil
}

func (n *NodeInfo) Describe() (string, error) {
	if n.Raw == nil {
		if err := n.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch node: %v", err)
		}
	}

	events, err := n.Client.Clientset.CoreV1().Events("").List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Node", n.Name),
	})
	if err != nil {
		events = &corev1.EventList{Items: []corev1.Event{}}
	}

	data, err := n.DescribeNode(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe node: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal node to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (n *NodeInfo) DescribeNode(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        n.Name,
		"labels":      n.Raw.Labels,
		"annotations": n.Raw.Annotations,
		"created":     formatTime(n.Raw.CreationTimestamp),
	}

	if n.Raw.Status.NodeInfo.MachineID != "" {
		desc["machineID"] = n.Raw.Status.NodeInfo.MachineID
	}
	if n.Raw.Status.NodeInfo.SystemUUID != "" {
		desc["systemUUID"] = n.Raw.Status.NodeInfo.SystemUUID
	}
	if n.Raw.Status.NodeInfo.BootID != "" {
		desc["bootID"] = n.Raw.Status.NodeInfo.BootID
	}
	desc["kernelVersion"] = n.Raw.Status.NodeInfo.KernelVersion
	desc["osImage"] = n.Raw.Status.NodeInfo.OSImage
	desc["containerRuntimeVersion"] = n.Raw.Status.NodeInfo.ContainerRuntimeVersion
	desc["kubeletVersion"] = n.Raw.Status.NodeInfo.KubeletVersion
	desc["kubeProxyVersion"] = n.Raw.Status.NodeInfo.KubeProxyVersion

	if len(n.Raw.Status.Addresses) > 0 {
		addresses := make([]map[string]string, 0, len(n.Raw.Status.Addresses))
		for _, addr := range n.Raw.Status.Addresses {
			addresses = append(addresses, map[string]string{
				"type":    string(addr.Type),
				"address": addr.Address,
			})
		}
		desc["addresses"] = addresses
	}

	if n.Raw.Status.Capacity != nil {
		capacity := make(map[string]string)
		for resource, quantity := range n.Raw.Status.Capacity {
			capacity[string(resource)] = quantity.String()
		}
		desc["capacity"] = capacity
	}

	if n.Raw.Status.Allocatable != nil {
		allocatable := make(map[string]string)
		for resource, quantity := range n.Raw.Status.Allocatable {
			allocatable[string(resource)] = quantity.String()
		}
		desc["allocatable"] = allocatable
	}

	if len(n.Raw.Status.Conditions) > 0 {
		conditions := make([]map[string]any, 0, len(n.Raw.Status.Conditions))
		for _, condition := range n.Raw.Status.Conditions {
			conditions = append(conditions, map[string]any{
				"type":               string(condition.Type),
				"status":             string(condition.Status),
				"reason":             condition.Reason,
				"message":            condition.Message,
				"lastTransitionTime": formatTime(condition.LastTransitionTime),
			})
		}
		desc["conditions"] = conditions
	}

	roles := getNodeRoles(n.Raw)
	if roles != "<none>" {
		desc["roles"] = roles
	}

	if len(events.Items) > 0 {
		eventList := make([]Event, 0)
		for _, event := range events.Items {
			age := time.Since(event.LastTimestamp.Time).Round(time.Second)
			eventList = append(eventList, Event{
				Type:    event.Type,
				Reason:  event.Reason,
				Age:     age.String(),
				From:    event.Source.Component,
				Message: event.Message,
			})
		}
		desc["events"] = eventList
	}

	return desc, nil
}

func DeleteNode(client Client, nodeName string) error {
	err := client.Clientset.CoreV1().Nodes().Delete(context.Background(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete node %s: %v", nodeName, err)
	}
	return nil
}
