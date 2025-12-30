package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
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
	LoadInitialData(clusterID string, gvr schema.GroupVersionResource) ([]map[string]any, string, error)
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
	podsByUID        map[string]*PodData
	nodesByUID       map[string]*NodeData
	gpusByUID        map[string]*GPUData
	gpuNodesByUID    map[string]*GPUNodeData
	gpuPoolsByUID    map[string]*GPUPoolData
	claimsByUID      map[string]*ClaimData
	slicesByUID      map[string]*SliceData
	eventsByUID      map[string]*EventData
	deploymentsByUID map[string]*DeploymentData

	// Indexes for fast lookup
	podsByNode      map[string]map[string]struct{} // nodeName -> set of pod UIDs
	podsByScheduler map[string]map[string]struct{} // schedulerName -> set of pod UIDs
	gpusByNode      map[string]map[string]struct{} // nodeName -> set of GPU UIDs
	gpusByUsedBy    map[string]map[string]struct{} // usedBy -> set of GPU UIDs
	claimsByPod     map[string]map[string]struct{} // podKey (ns/name) -> set of claim UIDs
	eventsByPod     map[string][]*EventData        // podKey -> recent events (max 3)
	eventUIDOrder   []string

	// Dirty tracking for delta generation
	dirtyNodes     map[string]struct{}
	dirtyPods      map[string]struct{}
	dirtyClaims    map[string]struct{}
	dirtyGPUs      map[string]struct{}
	dirtySlices    map[string]struct{}
	dirtyMigration bool

	// Deleted items tracking (for delta removal notifications)
	deletedPods   []PodRef
	deletedClaims []PodRef // using PodRef as ns/name tuple
	deletedSlices []string
	deletedNodes  []string

	// State
	seq           int64
	lastFlush     time.Time
	throttleMs    int
	minPayload    bool
	maxDeltaBytes int
	maxEvents     int
	forceSnapshot bool

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
	mu     sync.RWMutex

	// Filters
	labelSelector labels.Selector
	namespaceSet  map[string]struct{}

	// Watch tracking
	watchedGVRs []schema.GroupVersionResource

	// Availability flags
	draAvailable    bool
	tfAvailable     bool
	eventsAvailable bool

	// Health tracking
	warnings []string
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
	GPURequest    string // nvidia.com/gpu request
	MemoryRequest string // memory request
	CPURequest    string // cpu request
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

// DeploymentData holds parsed Deployment information needed for migration mode detection
type DeploymentData struct {
	UID                  string
	Namespace            string
	Name                 string
	Labels               map[string]string
	ProgressiveMigration *bool
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
	if _, exists := a.clusters[request.ClusterID]; exists {
		a.mu.Unlock()
		return fmt.Errorf("aggregation already running for cluster %s", request.ClusterID)
	}
	a.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	throttleMs := request.ThrottleMs
	if throttleMs <= 0 {
		throttleMs = 500 // default
	}

	minPayload := true
	if request.MinPayload != nil {
		minPayload = *request.MinPayload
	}

	var selector labels.Selector
	if request.LabelSelector != "" {
		parsed, err := labels.Parse(request.LabelSelector)
		if err != nil {
			return fmt.Errorf("invalid labelSelector %q: %w", request.LabelSelector, err)
		}
		selector = parsed
	}

	namespaceSet := make(map[string]struct{})
	for _, ns := range request.Namespaces {
		if ns != "" {
			namespaceSet[ns] = struct{}{}
		}
	}

	ca := &ClusterAggregation{
		clusterID:  request.ClusterID,
		request:    request,
		emitter:    a.emitter,
		throttleMs: throttleMs,
		minPayload: minPayload,

		// Initialize caches
		podsByUID:        make(map[string]*PodData),
		nodesByUID:       make(map[string]*NodeData),
		gpusByUID:        make(map[string]*GPUData),
		gpuNodesByUID:    make(map[string]*GPUNodeData),
		gpuPoolsByUID:    make(map[string]*GPUPoolData),
		claimsByUID:      make(map[string]*ClaimData),
		slicesByUID:      make(map[string]*SliceData),
		eventsByUID:      make(map[string]*EventData),
		deploymentsByUID: make(map[string]*DeploymentData),

		// Initialize indexes
		podsByNode:      make(map[string]map[string]struct{}),
		podsByScheduler: make(map[string]map[string]struct{}),
		gpusByNode:      make(map[string]map[string]struct{}),
		gpusByUsedBy:    make(map[string]map[string]struct{}),
		claimsByPod:     make(map[string]map[string]struct{}),
		eventsByPod:     make(map[string][]*EventData),
		eventUIDOrder:   make([]string, 0, 10000),

		// Initialize dirty sets
		dirtyNodes:  make(map[string]struct{}),
		dirtyPods:   make(map[string]struct{}),
		dirtyClaims: make(map[string]struct{}),
		dirtyGPUs:   make(map[string]struct{}),
		dirtySlices: make(map[string]struct{}),

		// Initialize deleted tracking slices
		deletedPods:   make([]PodRef, 0),
		deletedClaims: make([]PodRef, 0),
		deletedSlices: make([]string, 0),
		deletedNodes:  make([]string, 0),

		ctx:    ctx,
		cancel: cancel,

		labelSelector: selector,
		namespaceSet:  namespaceSet,
		maxDeltaBytes: 1_000_000,
		maxEvents:     10000,
		draAvailable:  true,
		tfAvailable:   true,
		warnings:      make([]string, 0, 10),
	}

	// Register watchers based on request.Include
	if err := ca.registerWatchers(a.provider); err != nil {
		cancel()
		return fmt.Errorf("failed to register watchers: %w", err)
	}

	a.mu.Lock()
	a.clusters[request.ClusterID] = ca
	a.mu.Unlock()

	// Load cached data before emitting the first snapshot
	if err := ca.loadInitialData(a.provider); err != nil {
		ca.recordWarning(fmt.Sprintf("initial data load failed: %v", err))
	}

	// Start throttled flush goroutine
	ca.ticker = time.NewTicker(time.Duration(throttleMs) * time.Millisecond)
	go ca.flushLoop()

	// Generate and emit initial snapshot
	go func() {
		snapshot := ca.GenerateSnapshot()
		a.emitter.Emit("scheduler:snapshot", snapshot)
	}()

	return nil
}

// StopAggregation stops scheduler aggregation for a cluster
func (a *Aggregator) StopAggregation(clusterID string) error {
	a.mu.Lock()
	ca, exists := a.clusters[clusterID]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("no aggregation running for cluster %s", clusterID)
	}
	delete(a.clusters, clusterID)
	a.mu.Unlock()

	ca.stop()

	var removeErr error
	for _, gvr := range ca.watchedGVRs {
		if err := a.provider.RemoveResourceWatcher(clusterID, gvr); err != nil && removeErr == nil {
			removeErr = err
		}
	}

	return removeErr
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
		if err := ca.addWatcher(provider, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}, ""); err != nil {
			return fmt.Errorf("failed to watch pods: %w", err)
		}

		if err := ca.addWatcher(provider, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "nodes",
		}, ""); err != nil {
			return fmt.Errorf("failed to watch nodes: %w", err)
		}
	}

	// Watch events
	if req.Include.Events {
		ca.eventsAvailable = true // assume available, will be set to false if both fail
		if err := ca.addWatcher(provider, schema.GroupVersionResource{
			Group:    "events.k8s.io",
			Version:  "v1",
			Resource: "events",
		}, ""); err != nil {
			// Try fallback to core/v1 events
			if err2 := ca.addWatcher(provider, schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "events",
			}, ""); err2 != nil {
				ca.eventsAvailable = false
				ca.recordWarning("Events API not available")
			}
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
			if err := ca.addWatcher(provider, gvr, ""); err != nil {
				ca.draAvailable = false
				// DRA not available, emit warning but continue
				ca.emitter.Emit("scheduler:warning", SchedulerWarning{
					ClusterID:   ca.clusterID,
					GeneratedAt: time.Now(),
					Type:        "DRA_NOT_AVAILABLE",
					Message:     fmt.Sprintf("DRA resource %s not available: %v", gvr.Resource, err),
				})
				ca.recordWarning(fmt.Sprintf("DRA resource %s not available", gvr.Resource))
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
			if err := ca.addWatcher(provider, gvr, ""); err != nil {
				ca.tfAvailable = false
				// TF CRD not available, emit warning but continue
				ca.emitter.Emit("scheduler:warning", SchedulerWarning{
					ClusterID:   ca.clusterID,
					GeneratedAt: time.Now(),
					Type:        "TF_CRD_NOT_AVAILABLE",
					Message:     fmt.Sprintf("TensorFusion CRD %s not available: %v", gvr.Resource, err),
				})
				ca.recordWarning(fmt.Sprintf("TensorFusion CRD %s not available", gvr.Resource))
			}
		}

		// Watch deployments for progressive migration mode detection
		if err := ca.addWatcher(provider, schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}, ""); err != nil {
			ca.emitter.Emit("scheduler:warning", SchedulerWarning{
				ClusterID:   ca.clusterID,
				GeneratedAt: time.Now(),
				Type:        "TF_DEPLOYMENT_NOT_AVAILABLE",
				Message:     fmt.Sprintf("TensorFusion deployment watch not available: %v", err),
			})
			ca.recordWarning("TensorFusion deployment watch not available")
		}
	}

	return nil
}

func (ca *ClusterAggregation) addWatcher(provider ClusterClientProvider, gvr schema.GroupVersionResource, namespace string) error {
	if err := provider.AddResourceWatcher(ca.clusterID, gvr, namespace); err != nil {
		return err
	}
	ca.watchedGVRs = append(ca.watchedGVRs, gvr)
	return nil
}

func (ca *ClusterAggregation) loadInitialData(provider ClusterClientProvider) error {
	var firstErr error
	for _, gvr := range ca.watchedGVRs {
		resources, _, err := provider.LoadInitialData(ca.clusterID, gvr)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, obj := range resources {
			ca.handleResourceEvent(gvr, "ADDED", obj, nil)
		}
	}
	return firstErr
}

func (ca *ClusterAggregation) recordWarning(message string) {
	if message == "" {
		return
	}
	if len(ca.warnings) >= 10 {
		ca.warnings = ca.warnings[1:]
	}
	ca.warnings = append(ca.warnings, message)
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

	if ca.forceSnapshot {
		snapshot := ca.generateSnapshotLocked()
		ca.forceSnapshot = false
		ca.clearDirtyState()
		ca.emitter.Emit("scheduler:snapshot", snapshot)
		return
	}

	// Check if there are any dirty or deleted items
	hasDirty := len(ca.dirtyNodes) > 0 || len(ca.dirtyPods) > 0 ||
		len(ca.dirtyClaims) > 0 || len(ca.dirtyGPUs) > 0 || len(ca.dirtySlices) > 0 || ca.dirtyMigration
	hasDeleted := len(ca.deletedPods) > 0 || len(ca.deletedClaims) > 0 ||
		len(ca.deletedSlices) > 0 || len(ca.deletedNodes) > 0

	if !hasDirty && !hasDeleted {
		return
	}

	ca.seq++
	delta := ca.generateDelta()

	if ca.maxDeltaBytes > 0 {
		if size := ca.deltaSize(delta); size > ca.maxDeltaBytes {
			ca.emitter.Emit("scheduler:warning", SchedulerWarning{
				ClusterID:   ca.clusterID,
				GeneratedAt: time.Now(),
				Type:        WarningDeltaTruncated,
				Message:     fmt.Sprintf("scheduler delta exceeded %d bytes (size=%d)", ca.maxDeltaBytes, size),
			})
			ca.recordWarning("delta truncated; snapshot scheduled")
			ca.forceSnapshot = true
			ca.clearDirtyState()
			return
		}
	}

	// Emit warnings for detected conflicts
	ca.emitConflictWarnings()

	ca.clearDirtyState()

	ca.lastFlush = time.Now()

	ca.emitter.Emit("scheduler:delta", delta)
}

func (ca *ClusterAggregation) clearDirtyState() {
	ca.dirtyNodes = make(map[string]struct{})
	ca.dirtyPods = make(map[string]struct{})
	ca.dirtyClaims = make(map[string]struct{})
	ca.dirtyGPUs = make(map[string]struct{})
	ca.dirtySlices = make(map[string]struct{})
	ca.dirtyMigration = false

	ca.deletedPods = ca.deletedPods[:0]
	ca.deletedClaims = ca.deletedClaims[:0]
	ca.deletedSlices = ca.deletedSlices[:0]
	ca.deletedNodes = ca.deletedNodes[:0]
}

func (ca *ClusterAggregation) deltaSize(delta SchedulerDelta) int {
	data, err := json.Marshal(delta)
	if err != nil {
		return 0
	}
	return len(data)
}

// emitConflictWarnings checks for conflicts and emits warning events
func (ca *ClusterAggregation) emitConflictWarnings() {
	// Check for GPU usedBy mismatches
	for _, pod := range ca.podsByUID {
		if pod.NodeName == "" {
			continue
		}

		resourceType := determineResourceType(pod)
		expectedUsedBy := expectedUsedByForResource(resourceType)

		gpuUIDs := ca.gpusByNode[pod.NodeName]
		for uid := range gpuUIDs {
			gpu := ca.gpusByUID[uid]
			if gpu == nil {
				continue
			}

			// Check if pod is using this GPU but usedBy doesn't match expected
			for _, app := range gpu.RunningApps {
				if app.Namespace == pod.Namespace && app.Name == pod.Name {
					if expectedUsedBy != "" && gpu.UsedBy != "" && gpu.UsedBy != expectedUsedBy {
						ca.emitter.Emit("scheduler:warning", SchedulerWarning{
							ClusterID:   ca.clusterID,
							GeneratedAt: time.Now(),
							Type:        WarningGPUUsedByMismatch,
							Message:     fmt.Sprintf("GPU %s usedBy=%s but pod %s/%s expects %s", gpu.Name, gpu.UsedBy, pod.Namespace, pod.Name, expectedUsedBy),
							Objects: []ObjectRef{
								{Kind: "Pod", Namespace: pod.Namespace, Name: pod.Name},
								{Kind: "GPU", Name: gpu.Name},
							},
						})
						ca.recordWarning("gpu usedBy mismatch detected")
					}
				}
			}
		}

		// Check for scheduler mismatch: pod using TF scheduler but requesting native GPU resources
		if pod.SchedulerName == SchedulerTensorFusion && resourceType == "nvidia.com/gpu" {
			ca.emitter.Emit("scheduler:warning", SchedulerWarning{
				ClusterID:   ca.clusterID,
				GeneratedAt: time.Now(),
				Type:        WarningSchedulerMismatch,
				Message:     fmt.Sprintf("Pod %s/%s uses tensor-fusion-scheduler with native GPU requests", pod.Namespace, pod.Name),
				Objects: []ObjectRef{
					{Kind: "Pod", Namespace: pod.Namespace, Name: pod.Name},
				},
			})
			ca.recordWarning("scheduler mismatch detected")
		}
	}
}

// GenerateSnapshot generates a full scheduler snapshot
func (ca *ClusterAggregation) GenerateSnapshot() SchedulerSnapshot {
	ca.mu.Lock()
	snapshot := ca.generateSnapshotLocked()
	ca.mu.Unlock()
	return snapshot
}

func (ca *ClusterAggregation) generateSnapshotLocked() SchedulerSnapshot {
	ca.seq++
	return SchedulerSnapshot{
		ClusterID:     ca.clusterID,
		GeneratedAt:   time.Now(),
		Seq:           ca.seq,
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

	if len(ca.dirtyClaims) > 0 || len(ca.dirtySlices) > 0 || len(ca.deletedClaims) > 0 || len(ca.deletedSlices) > 0 {
		draDelta := ca.generateDRAViewDelta()
		delta.DRAView = &draDelta
	}

	if len(ca.dirtyGPUs) > 0 || len(ca.dirtyPods) > 0 || ca.dirtyMigration {
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
		Warnings:     append([]string(nil), ca.warnings...),
	}
}

// handleResourceEvent processes incoming resource events
func (ca *ClusterAggregation) handleResourceEvent(gvr schema.GroupVersionResource, eventType string, obj, oldObj map[string]any) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	if gvr.Resource != "events" {
		filteredType, filteredObj, ok := ca.filterResourceEvent(eventType, obj, oldObj)
		if !ok {
			return
		}
		eventType = filteredType
		obj = filteredObj
		oldObj = nil
	}

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
	case "deployments":
		ca.handleDeploymentEvent(eventType, obj, oldObj)
	}
}

func (ca *ClusterAggregation) filterResourceEvent(eventType string, obj, oldObj map[string]any) (string, map[string]any, bool) {
	if ca.matchesResourceFilter(obj) {
		return eventType, obj, true
	}
	if oldObj != nil && ca.matchesResourceFilter(oldObj) {
		return "DELETED", oldObj, true
	}
	return "", nil, false
}

func (ca *ClusterAggregation) matchesResourceFilter(obj map[string]any) bool {
	if obj == nil {
		return false
	}
	namespace := getStringField(obj, "metadata", "namespace")
	if !ca.namespaceAllowed(namespace) {
		return false
	}
	if namespace == "" || ca.labelSelector == nil || ca.labelSelector.Empty() {
		return true
	}
	labelsMap := getStringMapField(getMapField(obj, "metadata"), "labels")
	return ca.labelSelector.Matches(labels.Set(labelsMap))
}

func (ca *ClusterAggregation) namespaceAllowed(namespace string) bool {
	if len(ca.namespaceSet) == 0 || namespace == "" {
		return true
	}
	_, ok := ca.namespaceSet[namespace]
	return ok
}

func (ca *ClusterAggregation) shouldIncludeEvent(event *EventData) bool {
	if event == nil {
		return false
	}
	if !ca.namespaceAllowed(event.Namespace) {
		return false
	}
	if ca.labelSelector == nil || ca.labelSelector.Empty() {
		return true
	}
	if event.InvolvedObjectKind != "Pod" {
		return false
	}
	pod := ca.findPodByName(event.InvolvedObjectNS, event.InvolvedObjectName)
	if pod == nil {
		return false
	}
	return ca.labelSelector.Matches(labels.Set(pod.Labels))
}

func (ca *ClusterAggregation) findPodByName(namespace, name string) *PodData {
	for _, pod := range ca.podsByUID {
		if pod.Namespace == namespace && pod.Name == name {
			return pod
		}
	}
	return nil
}
