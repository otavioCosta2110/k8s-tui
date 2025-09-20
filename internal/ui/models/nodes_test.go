package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"testing"
	"time"
)

func TestNewNodes(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.config.ResourceType != k8s.ResourceTypeNode {
		t.Error("Expected ResourceType to be ResourceTypeNode")
	}
}

func TestNodesModelDataToRows(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	nodesInfo := []k8s.NodeInfo{
		{
			Name:    "test-node",
			Status:  "Ready",
			Roles:   "worker",
			Version: "v1.28.0",
			CPU:     "4",
			Memory:  "8Gi",
			Pods:    "10",
			Age:     "1d",
		},
	}

	model.resourceData = []types.ResourceData{NodeData{&nodesInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if len(rows[0]) != 8 {
		t.Error("Expected 8 columns in row")
	}
	if rows[0][0] != "test-node" {
		t.Error("Node name mismatch in row")
	}
	if rows[0][1] != "Ready" {
		t.Error("Node status mismatch in row")
	}
	if rows[0][2] != "worker" {
		t.Error("Node roles mismatch in row")
	}
}

func TestNodesModelWithEmptyData(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	rows := model.dataToRows()
	if len(rows) != 0 {
		t.Error("Expected 0 rows for empty data")
	}
}

func TestNodesModelConfig(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if model.config.ResourceType != k8s.ResourceTypeNode {
		t.Error("Config ResourceType not set correctly")
	}
	expectedTitle := customstyles.ResourceIcons["Nodes"] + " Nodes in cluster"
	if model.config.Title != expectedTitle {
		t.Errorf("Config Title not set correctly, expected %s, got %s", expectedTitle, model.config.Title)
	}
	if len(model.config.Columns) != 8 {
		t.Error("Expected 8 columns in config")
	}
	if model.config.RefreshInterval != 10*time.Second {
		t.Error("Config RefreshInterval not set correctly")
	}
}

func TestNodesModelWithMultipleItems(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	nodesInfo := []k8s.NodeInfo{
		{
			Name:    "node-1",
			Status:  "Ready",
			Roles:   "master",
			Version: "v1.28.0",
			CPU:     "8",
			Memory:  "16Gi",
			Pods:    "5",
			Age:     "7d",
		},
		{
			Name:    "node-2",
			Status:  "Ready",
			Roles:   "worker",
			Version: "v1.28.0",
			CPU:     "4",
			Memory:  "8Gi",
			Pods:    "15",
			Age:     "3d",
		},
	}

	model.resourceData = []types.ResourceData{
		NodeData{&nodesInfo[0]},
		NodeData{&nodesInfo[1]},
	}

	rows := model.dataToRows()
	if len(rows) != 2 {
		t.Error("Expected 2 rows")
	}

	if rows[0][0] != "node-1" {
		t.Error("First node name mismatch")
	}
	if rows[0][2] != "master" {
		t.Error("First node roles mismatch")
	}

	if rows[1][0] != "node-2" {
		t.Error("Second node name mismatch")
	}
	if rows[1][2] != "worker" {
		t.Error("Second node roles mismatch")
	}
}

func TestNodesModelWithDifferentStates(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model, err := NewNodes(client, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	nodesInfo := []k8s.NodeInfo{
		{
			Name:    "bare-node",
			Status:  "NotReady",
			Roles:   "<none>",
			Version: "v1.27.0",
			CPU:     "2",
			Memory:  "4Gi",
			Pods:    "0",
			Age:     "1h",
		},
	}

	model.resourceData = []types.ResourceData{NodeData{&nodesInfo[0]}}

	rows := model.dataToRows()
	if len(rows) != 1 {
		t.Error("Expected 1 row")
	}
	if rows[0][1] != "NotReady" {
		t.Error("Expected NotReady status")
	}
	if rows[0][2] != "<none>" {
		t.Error("Expected <none> for roles")
	}
}

func TestNewNodeDetails(t *testing.T) {
	client := k8s.Client{Namespace: "default"}
	model := NewNodeDetails(client, "test-node")

	if model == nil {
		t.Error("Expected model to be non-nil")
	}
	if model.node.Name != "test-node" {
		t.Error("Expected node name to be 'test-node'")
	}
	if model.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.err != nil {
		t.Error("Expected error to be nil")
	}
}

func TestNodeDataGetName(t *testing.T) {
	nodeInfo := &k8s.NodeInfo{
		Name: "test-node",
	}

	nodeData := NodeData{nodeInfo}

	if nodeData.GetName() != "test-node" {
		t.Error("Expected GetName to return 'test-node'")
	}
}

func TestNodeDataGetNamespace(t *testing.T) {
	nodeInfo := &k8s.NodeInfo{
		Name: "test-node",
	}

	nodeData := NodeData{nodeInfo}

	if nodeData.GetNamespace() != "" {
		t.Error("Expected GetNamespace to return empty string for nodes")
	}
}

func TestNodeDataGetColumns(t *testing.T) {
	nodeInfo := &k8s.NodeInfo{
		Name:    "test-node",
		Status:  "Ready",
		Roles:   "worker",
		Version: "v1.28.0",
		CPU:     "4",
		Memory:  "8Gi",
		Pods:    "10",
		Age:     "1d",
	}

	nodeData := NodeData{nodeInfo}

	columns := nodeData.GetColumns()
	if len(columns) != 8 {
		t.Errorf("Expected 8 columns, got %d", len(columns))
	}

	if columns[0] != "test-node" {
		t.Error("First column should be node name")
	}
	if columns[1] != "Ready" {
		t.Error("Second column should be status")
	}
	if columns[2] != "worker" {
		t.Error("Third column should be roles")
	}
}
