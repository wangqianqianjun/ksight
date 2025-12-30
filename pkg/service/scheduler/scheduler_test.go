package scheduler

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type emittedEvent struct {
	name string
	data any
}

type fakeEmitter struct {
	events []emittedEvent
}

func (f *fakeEmitter) Emit(event string, data any) {
	f.events = append(f.events, emittedEvent{name: event, data: data})
}

func newTestClusterAggregation(minPayload bool) (*ClusterAggregation, *fakeEmitter) {
	emitter := &fakeEmitter{}
	ca := &ClusterAggregation{
		clusterID:     "cluster-1",
		request:       SchedulerAggregationRequest{ClusterID: "cluster-1"},
		emitter:       emitter,
		throttleMs:    500,
		minPayload:    minPayload,
		maxDeltaBytes: 1_000_000,
		maxEvents:     10000,
		draAvailable:  true,
		tfAvailable:   true,
		namespaceSet:  make(map[string]struct{}),

		podsByUID:        make(map[string]*PodData),
		nodesByUID:       make(map[string]*NodeData),
		gpusByUID:        make(map[string]*GPUData),
		gpuNodesByUID:    make(map[string]*GPUNodeData),
		gpuPoolsByUID:    make(map[string]*GPUPoolData),
		claimsByUID:      make(map[string]*ClaimData),
		slicesByUID:      make(map[string]*SliceData),
		eventsByUID:      make(map[string]*EventData),
		deploymentsByUID: make(map[string]*DeploymentData),

		podsByNode:      make(map[string]map[string]struct{}),
		podsByScheduler: make(map[string]map[string]struct{}),
		gpusByNode:      make(map[string]map[string]struct{}),
		gpusByUsedBy:    make(map[string]map[string]struct{}),
		claimsByPod:     make(map[string]map[string]struct{}),
		eventsByPod:     make(map[string][]*EventData),
		eventUIDOrder:   make([]string, 0),

		dirtyNodes:  make(map[string]struct{}),
		dirtyPods:   make(map[string]struct{}),
		dirtyClaims: make(map[string]struct{}),
		dirtyGPUs:   make(map[string]struct{}),
		dirtySlices: make(map[string]struct{}),

		deletedPods:   make([]PodRef, 0),
		deletedClaims: make([]PodRef, 0),
		deletedSlices: make([]string, 0),
		deletedNodes:  make([]string, 0),

		warnings: make([]string, 0),
	}
	return ca, emitter
}

func TestFilterResourceEventNamespaceAndLabel(t *testing.T) {
	ca, _ := newTestClusterAggregation(true)
	ca.namespaceSet["ns-1"] = struct{}{}
	selector, err := labels.Parse("app=test")
	if err != nil {
		t.Fatalf("failed to parse label selector: %v", err)
	}
	ca.labelSelector = selector

	podObj := map[string]any{
		"metadata": map[string]any{
			"uid":               "pod-1",
			"namespace":         "ns-1",
			"name":              "pod-1",
			"labels":            map[string]any{"app": "test"},
			"creationTimestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		"spec": map[string]any{
			"schedulerName": "default-scheduler",
			"containers": []any{
				map[string]any{
					"resources": map[string]any{
						"requests": map[string]any{
							"nvidia.com/gpu": "1",
						},
					},
				},
			},
		},
		"status": map[string]any{
			"phase": "Pending",
		},
	}

	ca.handleResourceEvent(schema.GroupVersionResource{Resource: "pods"}, "ADDED", podObj, nil)
	if _, ok := ca.podsByUID["pod-1"]; !ok {
		t.Fatalf("expected pod to be stored when namespace/label match")
	}

	nonMatching := map[string]any{
		"metadata": map[string]any{
			"uid":               "pod-2",
			"namespace":         "ns-2",
			"name":              "pod-2",
			"labels":            map[string]any{"app": "test"},
			"creationTimestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		"spec": map[string]any{},
		"status": map[string]any{
			"phase": "Pending",
		},
	}
	ca.handleResourceEvent(schema.GroupVersionResource{Resource: "pods"}, "ADDED", nonMatching, nil)
	if _, ok := ca.podsByUID["pod-2"]; ok {
		t.Fatalf("did not expect pod outside namespace filter to be stored")
	}
}

func TestNodeViewMinPayload(t *testing.T) {
	ca, _ := newTestClusterAggregation(true)
	ca.nodesByUID["node-1"] = &NodeData{
		UID:  "node-1",
		Name: "node-1",
		Labels: map[string]string{
			"topology.kubernetes.io/zone": "zone-a",
			"kubernetes.io/hostname":      "node-1",
			"extra":                       "ignore",
		},
		Capacity: map[string]string{
			"cpu": "8",
		},
		Allocatable: map[string]string{
			"nvidia.com/gpu": "2",
		},
		Conditions: map[string]string{"Ready": "True"},
	}
	ca.gpusByUID["gpu-1"] = &GPUData{
		UID:       "gpu-1",
		Name:      "gpu-1",
		NodeName:  "node-1",
		Model:     "A100",
		UsedBy:    UsedByTensorFusion,
		Capacity:  map[string]string{"tflops": "100"},
		Available: map[string]string{"vram": "80Gi"},
		RunningApps: []AppRef{
			{Namespace: "ns-1", Name: "pod-1"},
		},
	}
	ca.addToIndex(ca.gpusByNode, "node-1", "gpu-1")

	snapshot := ca.generateNodeView()
	if len(snapshot.Nodes) != 1 {
		t.Fatalf("expected 1 node row, got %d", len(snapshot.Nodes))
	}
	row := snapshot.Nodes[0]
	if row.Capacity != nil || row.Allocatable != nil {
		t.Fatalf("expected capacity/allocatable to be omitted in minPayload mode")
	}
	if row.Labels != nil {
		if _, ok := row.Labels["extra"]; ok {
			t.Fatalf("expected non-essential labels to be filtered out in minPayload mode")
		}
	}
	if len(row.GPU.Devices) != 1 {
		t.Fatalf("expected 1 GPU device, got %d", len(row.GPU.Devices))
	}
	device := row.GPU.Devices[0]
	if device.Capacity != nil || device.Available != nil || len(device.RunningApps) > 0 {
		t.Fatalf("expected GPU device capacity/available/runningApps to be omitted in minPayload mode")
	}
	if snapshot.Summary.GPUsByModel["A100"] != 1 {
		t.Fatalf("expected GPUsByModel to include A100 count")
	}
}

func TestPendingViewMinPayload(t *testing.T) {
	pod := &PodData{
		UID:           "pod-1",
		Namespace:     "ns-1",
		Name:          "pod-1",
		SchedulerName: SchedulerDefault,
		Phase:         "Pending",
		CreationTime:  time.Now().Add(-10 * time.Minute),
		GPURequest:    "1",
	}

	ca, _ := newTestClusterAggregation(true)
	row := ca.buildPendingPodRow(pod)
	if row.Since != "" {
		t.Fatalf("expected Since to be empty in minPayload mode")
	}

	ca, _ = newTestClusterAggregation(false)
	row = ca.buildPendingPodRow(pod)
	if row.Since == "" {
		t.Fatalf("expected Since to be set when minPayload is false")
	}
}

func TestMigrationConflictDetection(t *testing.T) {
	ca, _ := newTestClusterAggregation(true)
	pod := &PodData{
		UID:           "pod-1",
		Namespace:     "ns-1",
		Name:          "pod-1",
		NodeName:      "node-1",
		SchedulerName: SchedulerDefault,
		GPURequest:    "1",
	}
	ca.podsByUID[pod.UID] = pod

	gpu := &GPUData{
		UID:      "gpu-1",
		Name:     "gpu-1",
		NodeName: "node-1",
		UsedBy:   UsedByTensorFusion,
		RunningApps: []AppRef{
			{Namespace: "ns-1", Name: "pod-1"},
		},
	}
	ca.gpusByUID[gpu.UID] = gpu
	ca.addToIndex(ca.gpusByNode, "node-1", gpu.UID)

	view := ca.generateMigrationView()
	if len(view.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(view.Conflicts))
	}
	if view.Conflicts[0].ResourceType != "nvidia.com/gpu" {
		t.Fatalf("expected conflict resource type to be nvidia.com/gpu, got %s", view.Conflicts[0].ResourceType)
	}
}

func TestProgressiveMigrationDetection(t *testing.T) {
	ca, _ := newTestClusterAggregation(true)
	value := true
	ca.deploymentsByUID["dep-1"] = &DeploymentData{
		UID:       "dep-1",
		Namespace: "tf-system",
		Name:      "tf-controller",
		Labels: map[string]string{
			"tensor-fusion.ai/component":  "operator",
			"app.kubernetes.io/component": "controller",
		},
		ProgressiveMigration: &value,
	}

	view := ca.generateMigrationView()
	if !view.ProgressiveMigration {
		t.Fatalf("expected progressive migration to be true")
	}
}

func TestDeltaTruncationForcesSnapshot(t *testing.T) {
	ca, emitter := newTestClusterAggregation(true)
	ca.maxDeltaBytes = 1

	ca.nodesByUID["node-1"] = &NodeData{
		UID:  "node-1",
		Name: "node-1",
	}
	ca.dirtyNodes["node-1"] = struct{}{}

	ca.flushDelta()

	for _, evt := range emitter.events {
		if evt.name == "scheduler:delta" {
			t.Fatalf("did not expect delta emission when payload is truncated")
		}
	}

	foundWarning := false
	for _, evt := range emitter.events {
		if evt.name == "scheduler:warning" {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Fatalf("expected warning when delta is truncated")
	}

	emitter.events = nil
	ca.flushDelta()

	foundSnapshot := false
	for _, evt := range emitter.events {
		if evt.name == "scheduler:snapshot" {
			foundSnapshot = true
			break
		}
	}
	if !foundSnapshot {
		t.Fatalf("expected snapshot emission after delta truncation")
	}
}
