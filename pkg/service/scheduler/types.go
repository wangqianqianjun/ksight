package scheduler

import "time"

// AggregationIncludeConfig specifies which resources to include in aggregation
type AggregationIncludeConfig struct {
	Pods         bool `json:"pods,omitempty"`
	Nodes        bool `json:"nodes,omitempty"`
	Events       bool `json:"events,omitempty"`
	DRA          bool `json:"dra,omitempty"`          // claims/slices/deviceclasses
	TensorFusion bool `json:"tensorFusion,omitempty"` // TF CRDs
}

// SchedulerAggregationRequest represents a request to start scheduler aggregation
type SchedulerAggregationRequest struct {
	ClusterID     string                   `json:"clusterId"`
	Namespaces    []string                 `json:"namespaces,omitempty"`    // Empty means all namespaces
	LabelSelector string                   `json:"labelSelector,omitempty"` // Optional label filter
	ThrottleMs    int                      `json:"throttleMs,omitempty"`    // Default 500
	MinPayload    *bool                    `json:"minPayload,omitempty"`    // Default true, minimal fields mode
	Include       AggregationIncludeConfig `json:"include,omitempty"`
}

// SchedulerSnapshot represents a full snapshot of scheduler state
type SchedulerSnapshot struct {
	ClusterID     string                  `json:"clusterId"`
	GeneratedAt   time.Time               `json:"generatedAt"`
	Seq           int64                   `json:"seq"`
	NodeView      NodeViewSnapshot        `json:"nodeView"`
	PendingView   PendingViewSnapshot     `json:"pendingView"`
	DRAView       DRASnapshot             `json:"draView"`
	MigrationView MigrationSnapshot       `json:"migrationView"`
	Health        SchedulerHealthSnapshot `json:"health"`
}

// SchedulerDelta represents incremental updates
type SchedulerDelta struct {
	ClusterID     string          `json:"clusterId"`
	GeneratedAt   time.Time       `json:"generatedAt"`
	Seq           int64           `json:"seq"`
	NodeView      *NodeViewDelta  `json:"nodeView,omitempty"`
	PendingView   *PendingDelta   `json:"pendingView,omitempty"`
	DRAView       *DRADelta       `json:"draView,omitempty"`
	MigrationView *MigrationDelta `json:"migrationView,omitempty"`
}

// SchedulerWarning represents detected conflicts/anomalies
type SchedulerWarning struct {
	ClusterID   string      `json:"clusterId"`
	GeneratedAt time.Time   `json:"generatedAt"`
	Type        string      `json:"type"` // GPU_USED_BY_MISMATCH, SCHEDULER_MISMATCH, DELTA_TRUNCATED
	Message     string      `json:"message"`
	Objects     []ObjectRef `json:"objects,omitempty"`
}

// SchedulerHealthSnapshot represents overall scheduler health
type SchedulerHealthSnapshot struct {
	SchedulerStatus map[string]string `json:"schedulerStatus,omitempty"` // scheduler name -> status
	TotalPending    int               `json:"totalPending"`
	TotalNodes      int               `json:"totalNodes"`
	TotalGPUs       int               `json:"totalGpus"`
	Warnings        []string          `json:"warnings,omitempty"`
}

// ----------------- Node View -----------------

// NodeViewSnapshot represents full node view state
type NodeViewSnapshot struct {
	Nodes   []NodeRow   `json:"nodes"`
	Summary NodeSummary `json:"summary"`
}

// NodeTensorFusionInfo holds TensorFusion-specific node info
type NodeTensorFusionInfo struct {
	Pool        string `json:"pool,omitempty"`
	Phase       string `json:"phase,omitempty"`       // Pending/Migrating/Running
	ManagedGPUs int32  `json:"managedGpus,omitempty"` // from GPUNode.status.managedGPUs
}

// NodeGPUInfo holds GPU-related node info
type NodeGPUInfo struct {
	Total   int            `json:"total"`
	UsedBy  map[string]int `json:"usedBy"` // tensor-fusion / nvidia-device-plugin / unknown
	Devices []GPUDeviceRow `json:"devices"`
}

// NodeRow represents a single node in the view
type NodeRow struct {
	Name         string               `json:"name"`
	Zone         string               `json:"zone,omitempty"`
	Labels       map[string]string    `json:"labels,omitempty"`
	Capacity     map[string]string    `json:"capacity,omitempty"`
	Allocatable  map[string]string    `json:"allocatable,omitempty"`
	TensorFusion NodeTensorFusionInfo `json:"tensorFusion,omitempty"`
	GPU          NodeGPUInfo          `json:"gpu"`
	Pods         []PodRef             `json:"pods,omitempty"`
}

// GPUDeviceRow represents a single GPU device
type GPUDeviceRow struct {
	Name        string            `json:"name"`
	Model       string            `json:"model,omitempty"`
	UsedBy      string            `json:"usedBy,omitempty"` // tensor-fusion / nvidia-device-plugin
	Capacity    map[string]string `json:"capacity,omitempty"`
	Available   map[string]string `json:"available,omitempty"`
	RunningApps []AppRef          `json:"runningApps,omitempty"`
}

// NodeSummary represents aggregated node statistics
type NodeSummary struct {
	TotalNodes      int            `json:"totalNodes"`
	ReadyNodes      int            `json:"readyNodes"`
	TotalGPUs       int            `json:"totalGpus"`
	GPUsByUsedBy    map[string]int `json:"gpusByUsedBy,omitempty"`
	GPUsByModel     map[string]int `json:"gpusByModel,omitempty"`
	TotalTFlops     float64        `json:"totalTflops,omitempty"`
	AvailableTFlops float64        `json:"availableTflops,omitempty"`
	TotalVRAM       int64          `json:"totalVram,omitempty"`       // in bytes
	AvailableVRAM   int64          `json:"availableVram,omitempty"` // in bytes
}

// NodeViewDelta represents incremental node view updates
type NodeViewDelta struct {
	Upsert  []NodeRow    `json:"upsert,omitempty"`
	Remove  []string     `json:"remove,omitempty"` // node names
	Summary *NodeSummary `json:"summary,omitempty"`
}

// ----------------- Pending View -----------------

// PendingViewSnapshot represents full pending pod view state
type PendingViewSnapshot struct {
	Pods        []PendingPodRow  `json:"pods"`
	ByScheduler map[string]int   `json:"byScheduler"`
	Reasons     []ReasonBucket   `json:"reasons,omitempty"`
}

// PendingPodRow represents a pending pod
type PendingPodRow struct {
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Scheduler  string `json:"scheduler,omitempty"`
	GPURequest string `json:"gpuRequest,omitempty"` // from annotations or resource requests
	Pool       string `json:"pool,omitempty"`       // tensor-fusion.ai/gpupool annotation
	Reason     string `json:"reason,omitempty"`     // from events
	Since      string `json:"since,omitempty"`      // duration string
}

// ReasonBucket aggregates failure reasons
type ReasonBucket struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

// PendingDelta represents incremental pending view updates
type PendingDelta struct {
	Upsert      []PendingPodRow `json:"upsert,omitempty"`
	Remove      []PodRef        `json:"remove,omitempty"`
	ByScheduler map[string]int  `json:"byScheduler,omitempty"`
	Reasons     []ReasonBucket  `json:"reasons,omitempty"`
}

// ----------------- DRA View -----------------

// DRASnapshot represents full DRA allocation view state
type DRASnapshot struct {
	Claims []ClaimRow     `json:"claims"`
	Slices []SliceRow     `json:"slices"`
	Stats  map[string]int `json:"stats,omitempty"` // allocated/unallocated/pending counts
}

// ClaimRow represents a ResourceClaim
type ClaimRow struct {
	Namespace     string         `json:"namespace"`
	Name          string         `json:"name"`
	AllocationRaw map[string]any `json:"allocation,omitempty"` // status.allocation
	ResourceClass string         `json:"resourceClass,omitempty"`
	Driver        string         `json:"driver,omitempty"`
	PodRef        *PodRef        `json:"podRef,omitempty"` // owning pod if any
}

// SliceRow represents a ResourceSlice
type SliceRow struct {
	Name     string `json:"name"`
	NodeName string `json:"nodeName,omitempty"`
	Driver   string `json:"driver"`
	Devices  int    `json:"devices"` // count of devices in this slice
}

// DRADelta represents incremental DRA view updates
type DRADelta struct {
	UpsertClaims []ClaimRow     `json:"upsertClaims,omitempty"`
	RemoveClaims []PodRef       `json:"removeClaims,omitempty"` // namespace/name as PodRef
	UpsertSlices []SliceRow     `json:"upsertSlices,omitempty"`
	RemoveSlices []string       `json:"removeSlices,omitempty"` // slice names
	Stats        map[string]int `json:"stats,omitempty"`
}

// ----------------- Migration View -----------------

// MigrationSnapshot represents full migration/coexistence view state
type MigrationSnapshot struct {
	UsedByStats          map[string]int `json:"usedByStats"` // tensor-fusion / nvidia-device-plugin counts
	Conflicts            []ConflictRow  `json:"conflicts,omitempty"`
	ProgressiveMigration bool           `json:"progressiveMigration"` // from TF controller env
}

// ConflictRow represents a GPU usage conflict
type ConflictRow struct {
	Node         string `json:"node"`
	Pod          PodRef `json:"pod"`
	UsedBy       string `json:"usedBy"`       // actual GPU usedBy
	ResourceType string `json:"resourceType"` // nvidia.com/gpu | dra | tensor-fusion
}

// MigrationDelta represents incremental migration view updates
type MigrationDelta struct {
	UsedByStats          map[string]int `json:"usedByStats,omitempty"`
	Conflicts            []ConflictRow  `json:"conflicts,omitempty"`
	ProgressiveMigration *bool          `json:"progressiveMigration,omitempty"`
}

// ----------------- Common References -----------------

// PodRef references a pod by namespace and name
type PodRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ObjectRef references any Kubernetes object
type ObjectRef struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
}

// AppRef references an application running on GPU
type AppRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Count     int    `json:"count,omitempty"` // number of instances
}

// ----------------- GVR Constants -----------------

// Well-known GVRs for scheduler aggregation
const (
	// Core resources
	PodsGVR   = "v1/pods"
	NodesGVR  = "v1/nodes"
	EventsGVR = "events.k8s.io/v1/events"

	// DRA resources (K8S 1.31+)
	ResourceClaimsGVR         = "resource.k8s.io/v1/resourceclaims"
	ResourceSlicesGVR         = "resource.k8s.io/v1/resourceslices"
	ResourceClaimTemplatesGVR = "resource.k8s.io/v1/resourceclaimtemplates"
	DeviceClassesGVR          = "resource.k8s.io/v1/deviceclasses"

	// TensorFusion CRDs
	GPUPoolsGVR              = "tensor-fusion.ai/v1/gpupools"
	GPUNodesGVR              = "tensor-fusion.ai/v1/gpunodes"
	GPUsGVR                  = "tensor-fusion.ai/v1/gpus"
	GPUNodeClaimsGVR         = "tensor-fusion.ai/v1/gpunodeclaims"
	TensorFusionWorkloadsGVR = "tensor-fusion.ai/v1/tensorfusionworkloads"
	TensorFusionClustersGVR  = "tensor-fusion.ai/v1/tensorfusionclusters"
)

// TensorFusion annotations
const (
	TFAnnotationGPUPool        = "tensor-fusion.ai/gpupool"
	TFAnnotationGPUCount       = "tensor-fusion.ai/gpu-count"
	TFAnnotationTFlopsRequest  = "tensor-fusion.ai/tflops-request"
	TFAnnotationVRAMRequest    = "tensor-fusion.ai/vram-request"
	TFAnnotationQoS            = "tensor-fusion.ai/qos"
	TFAnnotationDRAEnabled     = "tensor-fusion.ai/dra-enabled"
	TFAnnotationDRACEL         = "tensor-fusion.ai/dra-cel-expression"
	TFAnnotationGPUModel       = "tensor-fusion.ai/gpu-model"
)

// GPU UsedBy values
const (
	UsedByTensorFusion       = "tensor-fusion"
	UsedByNvidiaDevicePlugin = "nvidia-device-plugin"
	UsedByUnknown            = "unknown"
)

// Scheduler names
const (
	SchedulerTensorFusion = "tensor-fusion-scheduler"
	SchedulerDefault      = "default-scheduler"
)

// Event reasons
const (
	ReasonFailedScheduling           = "FailedScheduling"
	ReasonScheduled                  = "Scheduled"
	ReasonGPUQuotaOrCapacityNotEnough = "GPUQuotaOrCapacityNotEnough"
	ReasonScheduleWithNativeGPU      = "ScheduleWithNativeGPU"
)

// Warning types
const (
	WarningGPUUsedByMismatch = "GPU_USED_BY_MISMATCH"
	WarningSchedulerMismatch = "SCHEDULER_MISMATCH"
	WarningDeltaTruncated    = "DELTA_TRUNCATED"
)
