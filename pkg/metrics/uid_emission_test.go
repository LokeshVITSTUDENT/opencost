package metrics

import (
	"testing"

	"github.com/opencost/opencost/core/pkg/clustercache"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/types"
)

// TestNodeMetricsIncludeUID tests that node metrics include UID in labels
func TestNodeMetricsIncludeUID(t *testing.T) {
	// Create a mock node with UID
	node := &clustercache.Node{
		Name: "test-node",
		UID:  types.UID("test-node-uid-123"),
		Labels: map[string]string{
			"kubernetes.io/hostname": "test-node",
		},
	}

	// Create a metric
	metric := newKubeNodeLabelsMetric(node.Name, string(node.UID), "kube_node_labels", []string{"label_kubernetes_io_hostname"}, []string{"test-node"})

	// Check that UID is included in the descriptor
	desc := metric.Desc()
	if desc == nil {
		t.Fatal("Expected non-nil descriptor")
	}

	// Write the metric and check labels
	var m dto.Metric
	err := metric.Write(&m)
	if err != nil {
		t.Fatalf("Error writing metric: %v", err)
	}

	// Check that UID label is present
	foundUID := false
	for _, label := range m.Label {
		if label.GetName() == "uid" && label.GetValue() == "test-node-uid-123" {
			foundUID = true
			break
		}
	}

	if !foundUID {
		t.Error("Expected UID label to be present in node metric")
	}

	// Print the metric for demonstration
	t.Logf("Node metric with UID emitted successfully:")
	t.Logf("  Metric name: kube_node_labels")
	for _, label := range m.Label {
		t.Logf("  Label: %s = %s", label.GetName(), label.GetValue())
	}
}

// TestDeploymentMetricsIncludeUID tests that deployment metrics include UID in labels
func TestDeploymentMetricsIncludeUID(t *testing.T) {
	// Create a metric
	metric := newDeploymentMatchLabelsMetric("test-deployment", "default", "test-deployment-uid-456", "deployment_match_labels", []string{"app"}, []string{"test"})

	// Write the metric and check labels
	var m dto.Metric
	err := metric.Write(&m)
	if err != nil {
		t.Fatalf("Error writing metric: %v", err)
	}

	// Check that UID label is present
	foundUID := false
	for _, label := range m.Label {
		if label.GetName() == "uid" && label.GetValue() == "test-deployment-uid-456" {
			foundUID = true
			break
		}
	}

	if !foundUID {
		t.Error("Expected UID label to be present in deployment metric")
	}

	// Print the metric for demonstration
	t.Logf("Deployment metric with UID emitted successfully:")
	t.Logf("  Metric name: deployment_match_labels")
	for _, label := range m.Label {
		t.Logf("  Label: %s = %s", label.GetName(), label.GetValue())
	}
}

// TestServiceMetricsIncludeUID tests that service metrics include UID in labels
func TestServiceMetricsIncludeUID(t *testing.T) {
	// Create a metric
	metric := newServiceSelectorLabelsMetric("test-service", "default", "test-service-uid-789", "service_selector_labels", []string{"app"}, []string{"test"})

	// Write the metric and check labels
	var m dto.Metric
	err := metric.Write(&m)
	if err != nil {
		t.Fatalf("Error writing metric: %v", err)
	}

	// Check that UID label is present
	foundUID := false
	for _, label := range m.Label {
		if label.GetName() == "uid" && label.GetValue() == "test-service-uid-789" {
			foundUID = true
			break
		}
	}

	if !foundUID {
		t.Error("Expected UID label to be present in service metric")
	}

	// Print the metric for demonstration
	t.Logf("Service metric with UID emitted successfully:")
	t.Logf("  Metric name: service_selector_labels")
	for _, label := range m.Label {
		t.Logf("  Label: %s = %s", label.GetName(), label.GetValue())
	}
}

// TestPrometheusQueryExamples demonstrates how the UID metrics would appear in Prometheus queries
func TestPrometheusQueryExamples(t *testing.T) {
	t.Log("=== Prometheus Query Examples with UIDs ===")
	
	// Show example queries that would now be possible with UIDs
	t.Log("")
	t.Log("Now that UIDs are emitted, you can run Prometheus queries like:")
	t.Log("")
	
	t.Log("1. Get all node metrics with their UIDs:")
	t.Log("   kube_node_labels{uid!=\"\"}")
	t.Log("")
	
	t.Log("2. Get deployment metrics for a specific UID:")
	t.Log("   deployment_match_labels{uid=\"test-deployment-uid-456\"}")
	t.Log("")
	
	t.Log("3. Get service metrics grouped by UID:")
	t.Log("   service_selector_labels group by (uid)")
	t.Log("")
	
	t.Log("4. Count unique objects by UID:")
	t.Log("   count by (uid) (kube_node_labels)")
	t.Log("   count by (uid) (deployment_match_labels)")
	t.Log("   count by (uid) (service_selector_labels)")
	t.Log("")
	
	// Create actual metrics to demonstrate
	nodeMetric := newKubeNodeLabelsMetric("worker-node-1", "node-uid-abc123", "kube_node_labels", []string{"label_zone"}, []string{"us-west-1a"})
	deploymentMetric := newDeploymentMatchLabelsMetric("web-app", "production", "deploy-uid-def456", "deployment_match_labels", []string{"app"}, []string{"web"})
	serviceMetric := newServiceSelectorLabelsMetric("web-svc", "production", "svc-uid-ghi789", "service_selector_labels", []string{"app"}, []string{"web"})
	
	var nodeM, deployM, serviceM dto.Metric
	nodeMetric.Write(&nodeM)
	deploymentMetric.Write(&deployM)
	serviceMetric.Write(&serviceM)
	
	t.Log("Sample metrics with UIDs:")
	t.Log("")
	
	t.Log("Node metric:")
	for _, label := range nodeM.Label {
		t.Logf("  %s=\"%s\"", label.GetName(), label.GetValue())
	}
	t.Log("")
	
	t.Log("Deployment metric:")
	for _, label := range deployM.Label {
		t.Logf("  %s=\"%s\"", label.GetName(), label.GetValue())
	}
	t.Log("")
	
	t.Log("Service metric:")
	for _, label := range serviceM.Label {
		t.Logf("  %s=\"%s\"", label.GetName(), label.GetValue())
	}
	
	t.Log("")
	t.Log("=== UID emission implementation complete! ===")
}