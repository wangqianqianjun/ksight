package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EventEmitter interface for abstracting event emission (matches service.EventEmitter)
type EventEmitter interface {
	Emit(event string, data any)
}

// ClusterClientProvider provides access to cluster clients
type ClusterClientProvider interface {
	AddResourceWatcher(clusterID string, gvr schema.GroupVersionResource, namespace string) error
	RemoveResourceWatcher(clusterID string, gvr schema.GroupVersionResource) error
}

// Aggregator manages scheduler state aggregation for multiple clusters
type Aggregator struct {
	clusters map[string]*ClusterAggregation
	emitter  EventEmitter
	provider ClusterClientProvider
	mu       sync.RWMutex
}

// ClusterAggregation manages aggregation for a single cluster
type ClusterAggregation struct {
	clusterID string
	request   SchedulerAggregationRequest
	emitter   EventEmitter

	// Caches (by UID)
	podsByUID     map[string]*PodData
	nodesByUID    map[string]*NodeData
	gpusByUID     map[string]*GPUData
	gpuNodesByUID map[string]*GPUNodeData
	gpuPoolsByUID map[string]*GPUPoolData
	claimsByUID   map[string]*ClaimData
	slicesByUID   map[string]*SliceData
	eventsByUID   map[string]*EventData

	// Indexes for fast lookup
	podsByNode      map[string]map[string]struct{} // nodeName -> set of pod UIDs
	podsByScheduler map[string]map[string]struct{} // schedulerName -> set of pod UIDs
	gpusByNode      map[string]map[string]struct{} // nodeName -> set of GPU UIDs
	gpusByUsedBy    map[string]map[string]struct{} // usedBy -> set of GPU UIDs
	claimsByPod     map[string]map[string]struct{} // podKey (ns/name) -> set of claim UIDs
	eventsByPod     map[string][]*EventData        // podKey -> recent events (max 3)

	// Dirty tracking for delta generation
	dirtyNodes  map[string]struct{}
	dirtyPods   map[string]struct{}
	dirtyClaims map[string]struct{}
	dirtyGPUs   map[string]struct{}

	// Deleted items tracking (for delta removal notifications)
	deletedPods   []PodRef
	deletedClaims []PodRef // using PodRef as ns/name tuple
	deletedSlices []string
	deletedNodes  []string

	// State
	seq        int64
	lastFlush  time.Time
	throttleMs int
	minPayload bool

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
	mu     sync.RWMutex
}

// PodData holds parsed pod information
type PodData struct {
	UID            string
	Namespace      string
	Name           string
	NodeName       string
	SchedulerName  string
	Phase          string
	Annotations    map[string]string
	Labels         map[string]string
	ResourceClaims []string // claim names
	CreationTime   time.Time
	// Resource requests from containers (aggregated)
	GPURequest     string // nvidia.com/gpu request
	MemoryRequest  string // memory request
	CPURequest     string // cpu request
}

// NodeData holds parsed node information
type NodeData struct {
	UID         string
	Name        string
	Labels      map[string]string
	Capacity    map[string]string
	Allocatable map[string]string
	Conditions  map[string]string
}

// GPUData holds parsed GPU (tensor-fusion.ai/v1 GPU) information
type GPUData struct {
	UID         string
	Name        string
	NodeName    string
	Model       string
	UsedBy      string
	Capacity    map[string]string
	Available   map[string]string
	RunningApps []AppRef
}

// GPUNodeData holds parsed GPUNode information
type GPUNodeData struct {
	UID         string
	Name        string
	NodeName    string
	Pool        string
	Phase       string
	ManagedGPUs int32
}

// GPUPoolData holds parsed GPUPool information
type GPUPoolData struct {
	UID              string
	Name             string
	Phase            string
	TotalGPUs        int32
	AvailableVRAM    int64
	AvailableTFlops  float64
	Oversubscription bool
	DRAEnabled       bool
}

// ClaimData holds parsed ResourceClaim information
type ClaimData struct {
	UID           string
	Namespace     string
	Name          string
	ResourceClass string
	Driver        string
	Allocation    map[string]any
	PodRef        *PodRef
}

// SliceData holds parsed ResourceSlice information
type SliceData struct {
	UID      string
	Name     string
	NodeName string
	Driver   string
	Devices  int
}

// EventData holds parsed Event information
type EventData struct {
	UID                 string
	Namespace           string
	Name                string
	Reason              string
	Message             string
	InvolvedObjectKind  string
	InvolvedObjectName  string
	InvolvedObjectNS    string
	ReportingController string
	Timestamp           time.Time
}

// NewAggregator creates a new scheduler aggregator
func NewAggregator(emitter EventEmitter, provider ClusterClientProvider) *Aggregator {
	return &Aggregator{
		clusters: make(map[string]*ClusterAggregation),
		emitter:  emitter,
		provider: provider,
	}
}

// StartAggregation starts scheduler aggregation for a cluster
func (a *Aggregator) StartAggregation(request SchedulerAggregationRequest) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.clusters[request.ClusterID]; exists {
		return fmt.Errorf("aggregation already running for cluster %s", request.ClusterID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	throttleMs := request.ThrottleMs
	if throttleMs <= 0 {
		throttleMs = 500 // default
	}

	minPayload := true
	if request.MinPayload != nil {
		minPayload = *request.MinPayload
	}

	ca := &ClusterAggregation{
		clusterID:  request.ClusterID,
		request:    request,
		emitter:    a.emitter,
		throttleMs: throttleMs,
		minPayload: minPayload,

		// Initialize caches
		podsByUID:     make(map[string]*PodData),
		nodesByUID:    make(map[string]*NodeData),
		gpusByUID:     make(map[string]*GPUData),
		gpuNodesByUID: make(map[string]*GPUNodeData),
		gpuPoolsByUID: make(map[string]*GPUPoolData),
		claimsByUID:   make(map[string]*ClaimData),
		slicesByUID:   make(map[string]*SliceData),
		eventsByUID:   make(map[string]*EventData),

		// Initialize indexes
		podsByNode:      make(map[string]map[string]struct{}),
		podsByScheduler: make(map[string]map[string]struct{}),
		gpusByNode:      make(map[string]map[string]struct{}),
		gpusByUsedBy:    make(map[string]map[string]struct{}),
		claimsByPod:     make(map[string]map[string]struct{}),
		eventsByPod:     make(map[string][]*EventData),

		// Initialize dirty sets
		dirtyNodes:  make(map[string]struct{}),
		dirtyPods:   make(map[string]struct{}),
		dirtyClaims: make(map[string]struct{}),
		dirtyGPUs:   make(map[string]struct{}),

		// Initialize deleted tracking slices
		deletedPods:   make([]PodRef, 0),
		deletedClaims: make([]PodRef, 0),
		deletedSlices: make([]string, 0),
		deletedNodes:  make([]string, 0),

		ctx:    ctx,
		cancel: cancel,
	}

	// Register watchers based on request.Include
	if err := ca.registerWatchers(a.provider); err != nil {
		cancel()
		return fmt.Errorf("failed to register watchers: %w", err)
	}

	// Start throttled flush goroutine
	ca.ticker = time.NewTicker(time.Duration(throttleMs) * time.Millisecond)
	go ca.flushLoop()

	a.clusters[request.ClusterID] = ca

	// Generate and emit initial snapshot
	go func() {
		// Wait a bit for initial data to load
		time.Sleep(2 * time.Second)
		snapshot := ca.GenerateSnapshot()
		a.emitter.Emit("scheduler:snapshot", snapshot)
	}()

	return nil
}

// StopAggregation stops scheduler aggregation for a cluster
func (a *Aggregator) StopAggregation(clusterID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	ca, exists := a.clusters[clusterID]
	if !exists {
		return fmt.Errorf("no aggregation running for cluster %s", clusterID)
	}

	ca.stop()
	delete(a.clusters, clusterID)

	return nil
}

// GetSnapshot returns current scheduler snapshot for a cluster
func (a *Aggregator) GetSnapshot(clusterID string) (*SchedulerSnapshot, error) {
	a.mu.RLock()
	ca, exists := a.clusters[clusterID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no aggregation running for cluster %s", clusterID)
	}

	snapshot := ca.GenerateSnapshot()
	return &snapshot, nil
}

// HandleEvent processes a resource event from InformerManager
func (a *Aggregator) HandleEvent(clusterID string, gvr schema.GroupVersionResource, eventType string, obj, oldObj map[string]any) {
	a.mu.RLock()
	ca, exists := a.clusters[clusterID]
	a.mu.RUnlock()

	if !exists {
		return
	}

	ca.handleResourceEvent(gvr, eventType, obj, oldObj)
}

// registerWatchers registers resource watchers based on request configuration
func (ca *ClusterAggregation) registerWatchers(provider ClusterClientProvider) error {
	req := ca.request

	// Always watch pods and nodes
	if req.Include.Pods || req.Include.Nodes {
		if err := provider.AddResourceWatcher(ca.clusterID, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}, ""); err != nil {
			return fmt.Errorf("failed to watch pods: %w", err)
		}

		if err := provider.AddResourceWatcher(ca.clusterID, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "nodes",
		}, ""); err != nil {
			return fmt.Errorf("failed to watch nodes: %w", err)
		}
	}

	// Watch events
	if req.Include.Events {
		if err := provider.AddResourceWatcher(ca.clusterID, schema.GroupVersionResource{
			Group:    "events.k8s.io",
			Version:  "v1",
			Resource: "events",
		}, ""); err != nil {
			// Try fallback to core/v1 events
			provider.AddResourceWatcher(ca.clusterID, schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "events",
			}, "")
		}
	}

	// Watch DRA resources
	if req.Include.DRA {
		draGVRs := []schema.GroupVersionResource{
			{Group: "resource.k8s.io", Version: "v1", Resource: "resourceclaims"},
			{Group: "resource.k8s.io", Version: "v1", Resource: "resourceslices"},
			{Group: "resource.k8s.io", Version: "v1", Resource: "resourceclaimtemplates"},
			{Group: "resource.k8s.io", Version: "v1", Resource: "deviceclasses"},
		}
		for _, gvr := range draGVRs {
			if err := provider.AddResourceWatcher(ca.clusterID, gvr, ""); err != nil {
				// DRA not available, emit warning but continue
				ca.emitter.Emit("scheduler:warning", SchedulerWarning{
					ClusterID:   ca.clusterID,
					GeneratedAt: time.Now(),
					Type:        "DRA_NOT_AVAILABLE",
					Message:     fmt.Sprintf("DRA resource %s not available: %v", gvr.Resource, err),
				})
			}
		}
	}

	// Watch TensorFusion CRDs
	if req.Include.TensorFusion {
		tfGVRs := []schema.GroupVersionResource{
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "gpupools"},
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "gpunodes"},
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "gpus"},
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "gpunodeclaims"},
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "tensorfusionworkloads"},
			{Group: "tensor-fusion.ai", Version: "v1", Resource: "tensorfusionclusters"},
		}
		for _, gvr := range tfGVRs {
			if err := provider.AddResourceWatcher(ca.clusterID, gvr, ""); err != nil {
				// TF CRD not available, emit warning but continue
				ca.emitter.Emit("scheduler:warning", SchedulerWarning{
					ClusterID:   ca.clusterID,
					GeneratedAt: time.Now(),
					Type:        "TF_CRD_NOT_AVAILABLE",
					Message:     fmt.Sprintf("TensorFusion CRD %s not available: %v", gvr.Resource, err),
				})
			}
		}
	}

	return nil
}

// stop stops the cluster aggregation
func (ca *ClusterAggregation) stop() {
	ca.cancel()
	if ca.ticker != nil {
		ca.ticker.Stop()
	}
}

// flushLoop runs the throttled delta emission loop
func (ca *ClusterAggregation) flushLoop() {
	for {
		select {
		case <-ca.ctx.Done():
			return
		case <-ca.ticker.C:
			ca.flushDelta()
		}
	}
}

// flushDelta generates and emits delta if there are changes
func (ca *ClusterAggregation) flushDelta() {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	// Check if there are any dirty or deleted items
	hasDirty := len(ca.dirtyNodes) > 0 || len(ca.dirtyPods) > 0 ||
		len(ca.dirtyClaims) > 0 || len(ca.dirtyGPUs) > 0
	hasDeleted := len(ca.deletedPods) > 0 || len(ca.deletedClaims) > 0 ||
		len(ca.deletedSlices) > 0 || len(ca.deletedNodes) > 0

	if !hasDirty && !hasDeleted {
		return
	}

	ca.seq++
	delta := ca.generateDelta()

	// Emit warnings for detected conflicts
	ca.emitConflictWarnings()

	// Clear dirty sets
	ca.dirtyNodes = make(map[string]struct{})
	ca.dirtyPods = make(map[string]struct{})
	ca.dirtyClaims = make(map[string]struct{})
	ca.dirtyGPUs = make(map[string]struct{})

	// Clear deleted tracking slices
	ca.deletedPods = ca.deletedPods[:0]
	ca.deletedClaims = ca.deletedClaims[:0]
	ca.deletedSlices = ca.deletedSlices[:0]
	ca.deletedNodes = ca.deletedNodes[:0]

	ca.lastFlush = time.Now()

	ca.emitter.Emit("scheduler:delta", delta)
}

// emitConflictWarnings checks for conflicts and emits warning events
func (ca *ClusterAggregation) emitConflictWarnings() {
	// Check for GPU usedBy mismatches
	for _, pod := range ca.podsByUID {
		if pod.NodeName == "" {
			continue
		}

		gpuUIDs := ca.gpusByNode[pod.NodeName]
		for uid := range gpuUIDs {
			gpu := ca.gpusByUID[uid]
			if gpu == nil {
				continue
			}

			// Check if pod is using this GPU but usedBy doesn't match expected
			for _, app := range gpu.RunningApps {
				if app.Namespace == pod.Namespace && app.Name == pod.Name {
					expectedUsedBy := UsedByTensorFusion
					if pod.SchedulerName != SchedulerTensorFusion {
						expectedUsedBy = UsedByNvidiaDevicePlugin
					}

					if gpu.UsedBy != "" && gpu.UsedBy != expectedUsedBy {
						ca.emitter.Emit("scheduler:warning", SchedulerWarning{
							ClusterID:   ca.clusterID,
							GeneratedAt: time.Now(),
							Type:        WarningGPUUsedByMismatch,
							Message:     fmt.Sprintf("GPU %s usedBy=%s but pod %s/%s scheduled by %s", gpu.Name, gpu.UsedBy, pod.Namespace, pod.Name, pod.SchedulerName),
							Objects: []ObjectRef{
								{Kind: "Pod", Namespace: pod.Namespace, Name: pod.Name},
								{Kind: "GPU", Name: gpu.Name},
							},
						})
					}
				}
			}
		}

		// Check for scheduler mismatch: pod using TF scheduler but requesting nvidia.com/gpu
		if pod.SchedulerName == SchedulerTensorFusion && pod.GPURequest != "" {
			// Pod scheduled by TF but has native GPU request - potential misconfiguration
			if pod.Annotations == nil || pod.Annotations[TFAnnotationGPUPool] == "" {
				ca.emitter.Emit("scheduler:warning", SchedulerWarning{
					ClusterID:   ca.clusterID,
					GeneratedAt: time.Now(),
					Type:        WarningSchedulerMismatch,
					Message:     fmt.Sprintf("Pod %s/%s uses tensor-fusion-scheduler but has nvidia.com/gpu request without TF annotations", pod.Namespace, pod.Name),
					Objects: []ObjectRef{
						{Kind: "Pod", Namespace: pod.Namespace, Name: pod.Name},
					},
				})
			}
		}
	}
}

// GenerateSnapshot generates a full scheduler snapshot
func (ca *ClusterAggregation) GenerateSnapshot() SchedulerSnapshot {
	ca.mu.Lock()
	ca.seq++
	seq := ca.seq
	ca.mu.Unlock()

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	return SchedulerSnapshot{
		ClusterID:     ca.clusterID,
		GeneratedAt:   time.Now(),
		Seq:           seq,
		NodeView:      ca.generateNodeView(),
		PendingView:   ca.generatePendingView(),
		DRAView:       ca.generateDRAView(),
		MigrationView: ca.generateMigrationView(),
		Health:        ca.generateHealthSnapshot(),
	}
}

// generateDelta generates a delta update
func (ca *ClusterAggregation) generateDelta() SchedulerDelta {
	delta := SchedulerDelta{
		ClusterID:   ca.clusterID,
		GeneratedAt: time.Now(),
		Seq:         ca.seq,
	}

	// Generate view deltas based on dirty sets
	if len(ca.dirtyNodes) > 0 || len(ca.dirtyGPUs) > 0 {
		nodeViewDelta := ca.generateNodeViewDelta()
		delta.NodeView = &nodeViewDelta
	}

	if len(ca.dirtyPods) > 0 {
		pendingDelta := ca.generatePendingViewDelta()
		delta.PendingView = &pendingDelta
	}

	if len(ca.dirtyClaims) > 0 {
		draDelta := ca.generateDRAViewDelta()
		delta.DRAView = &draDelta
	}

	if len(ca.dirtyGPUs) > 0 {
		migrationDelta := ca.generateMigrationViewDelta()
		delta.MigrationView = &migrationDelta
	}

	return delta
}

// generateHealthSnapshot generates the health snapshot
func (ca *ClusterAggregation) generateHealthSnapshot() SchedulerHealthSnapshot {
	pendingCount := 0
	for _, pod := range ca.podsByUID {
		if pod.Phase == "Pending" && pod.NodeName == "" {
			pendingCount++
		}
	}

	return SchedulerHealthSnapshot{
		TotalPending: pendingCount,
		TotalNodes:   len(ca.nodesByUID),
		TotalGPUs:    len(ca.gpusByUID),
	}
}

// handleResourceEvent processes incoming resource events
func (ca *ClusterAggregation) handleResourceEvent(gvr schema.GroupVersionResource, eventType string, obj, oldObj map[string]any) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	switch gvr.Resource {
	case "pods":
		ca.handlePodEvent(eventType, obj, oldObj)
	case "nodes":
		ca.handleNodeEvent(eventType, obj, oldObj)
	case "events":
		ca.handleEventEvent(eventType, obj)
	case "resourceclaims":
		ca.handleClaimEvent(eventType, obj, oldObj)
	case "resourceslices":
		ca.handleSliceEvent(eventType, obj, oldObj)
	case "gpus":
		ca.handleGPUEvent(eventType, obj, oldObj)
	case "gpunodes":
		ca.handleGPUNodeEvent(eventType, obj, oldObj)
	case "gpupools":
		ca.handleGPUPoolEvent(eventType, obj, oldObj)
	}
}

