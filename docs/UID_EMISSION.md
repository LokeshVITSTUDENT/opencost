# OpenCost UID Emission Implementation

This document describes the implementation of UID emission for Kubernetes objects in OpenCost metrics.

## Overview

This implementation adds support for emitting Kubernetes object UIDs in Prometheus metrics for Node, Deployment, and Service objects. This fulfills the requirement of the LFX Mentorship coding challenge to begin emitting UIDs for any 3 Kubernetes objects.

## Implementation Details

### Core Changes

1. **Modified core clustercache structs** (`core/pkg/clustercache/clustercache.go`):
   - Added `UID types.UID` field to `Node`, `Deployment`, and `Service` structs
   - Updated corresponding `Transform*` functions to preserve UIDs from Kubernetes objects

2. **Updated metric collectors** to include UID in labels:
   - `pkg/metrics/nodemetrics.go` - Node metrics now include UID
   - `pkg/metrics/deploymentmetrics.go` - Deployment metrics now include UID
   - `pkg/metrics/servicemetrics.go` - Service metrics now include UID

### Metrics Enhanced

1. **Node Metrics**: `kube_node_labels`
   - Now includes `uid` label with the node's Kubernetes UID
   - Example: `kube_node_labels{node="worker-1", uid="node-uid-123456", ...}`

2. **Deployment Metrics**: `deployment_match_labels`
   - Now includes `uid` label with the deployment's Kubernetes UID
   - Example: `deployment_match_labels{deployment="webapp", namespace="prod", uid="deploy-uid-789abc", ...}`

3. **Service Metrics**: `service_selector_labels`
   - Now includes `uid` label with the service's Kubernetes UID
   - Example: `service_selector_labels{service="webapp-svc", namespace="prod", uid="svc-uid-def456", ...}`

## Usage Examples

With UID emission enabled, you can now run Prometheus queries like:

```promql
# Get all node metrics with their UIDs
kube_node_labels{uid!=""}

# Get deployment metrics for a specific UID
deployment_match_labels{uid="deployment-uid-123"}

# Get service metrics grouped by UID
service_selector_labels group by (uid)

# Count unique objects by UID
count by (uid) (kube_node_labels)
count by (uid) (deployment_match_labels)
count by (uid) (service_selector_labels)
```

## Testing

Comprehensive tests have been added in `pkg/metrics/uid_emission_test.go` to verify:
- UID labels are properly included in metrics
- Metric structures correctly handle UID fields
- Example Prometheus queries that would be possible with UIDs

## Backward Compatibility

This implementation maintains full backward compatibility:
- Existing metrics continue to work as before
- No breaking changes to existing APIs
- UID is added as an additional label without affecting existing functionality

## Future Work

This implementation provides the foundation for the next generation of OpenCost features mentioned in the problem statement:
- Hierarchical organization of Kubernetes objects by UID
- GPU resource tracking with UID-based grouping
- Compressed protobuf data objects with UID-based hierarchy
- Integration tests comparing UID-based metrics with existing objects

## Files Modified

- `core/pkg/clustercache/clustercache.go` - Added UID fields to core structs
- `pkg/metrics/nodemetrics.go` - Enhanced node metrics with UID
- `pkg/metrics/deploymentmetrics.go` - Enhanced deployment metrics with UID
- `pkg/metrics/servicemetrics.go` - Enhanced service metrics with UID
- `pkg/metrics/uid_emission_test.go` - Added comprehensive tests