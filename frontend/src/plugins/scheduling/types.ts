// Scheduler aggregation types - mirrors backend types

export interface SchedulerAggregationRequest {
  clusterId: string
  namespaces?: string[]
  labelSelector?: string
  throttleMs?: number
  minPayload?: boolean
  include?: {
    pods?: boolean
    nodes?: boolean
    events?: boolean
    dra?: boolean
    tensorFusion?: boolean
  }
}

export interface SchedulerSnapshot {
  clusterId: string
  generatedAt: string
  seq: number
  nodeView: NodeViewSnapshot
  pendingView: PendingViewSnapshot
  draView: DRASnapshot
  migrationView: MigrationSnapshot
  health: SchedulerHealthSnapshot
}

export interface SchedulerDelta {
  clusterId: string
  generatedAt: string
  seq: number
  nodeView?: NodeViewDelta
  pendingView?: PendingDelta
  draView?: DRADelta
  migrationView?: MigrationDelta
}

export interface SchedulerWarning {
  clusterId: string
  generatedAt: string
  type: string
  message: string
  objects?: ObjectRef[]
}

export interface SchedulerHealthSnapshot {
  schedulerStatus?: Record<string, string>
  totalPending: number
  totalNodes: number
  totalGpus: number
  warnings?: string[]
}

// Node View Types

export interface NodeViewSnapshot {
  nodes: NodeRow[]
  summary: NodeSummary
}

export interface NodeRow {
  name: string
  zone?: string
  labels?: Record<string, string>
  capacity?: Record<string, string>
  allocatable?: Record<string, string>
  tensorFusion?: {
    pool?: string
    phase?: string
    managedGpus?: number
  }
  gpu: {
    total: number
    usedBy: Record<string, number>
    devices: GPUDeviceRow[]
  }
  pods?: PodRef[]
}

export interface GPUDeviceRow {
  name: string
  model?: string
  usedBy?: string
  capacity?: Record<string, string>
  available?: Record<string, string>
  runningApps?: AppRef[]
}

export interface NodeSummary {
  totalNodes: number
  readyNodes: number
  totalGpus: number
  gpusByUsedBy?: Record<string, number>
  gpusByModel?: Record<string, number>
  totalTflops?: number
  availableTflops?: number
  totalVram?: number
  availableVram?: number
}

export interface NodeViewDelta {
  upsert?: NodeRow[]
  remove?: string[]
  summary?: NodeSummary
}

// Pending View Types

export interface PendingViewSnapshot {
  pods: PendingPodRow[]
  byScheduler: Record<string, number>
  reasons?: ReasonBucket[]
  eventsAvailable?: boolean
}

export interface PendingPodRow {
  namespace: string
  name: string
  scheduler?: string
  gpuRequest?: string
  pool?: string
  reason?: string
  since?: string
}

export interface ReasonBucket {
  reason: string
  count: number
}

export interface PendingDelta {
  upsert?: PendingPodRow[]
  remove?: PodRef[]
  byScheduler?: Record<string, number>
  reasons?: ReasonBucket[]
}

// DRA View Types

export interface DRASnapshot {
  claims: ClaimRow[]
  slices: SliceRow[]
  stats?: Record<string, number>
  status?: 'available' | 'not_available'
  message?: string
}

export interface ClaimRow {
  namespace: string
  name: string
  allocation?: Record<string, unknown>
  resourceClass?: string
  driver?: string
  podRef?: PodRef
}

export interface SliceRow {
  name: string
  nodeName?: string
  driver: string
  devices: number
}

export interface DRADelta {
  upsertClaims?: ClaimRow[]
  removeClaims?: PodRef[]
  upsertSlices?: SliceRow[]
  removeSlices?: string[]
  stats?: Record<string, number>
}

// Migration View Types

export interface MigrationSnapshot {
  usedByStats: Record<string, number>
  conflicts?: ConflictRow[]
  progressiveMigration: boolean
}

export interface ConflictRow {
  node: string
  pod: PodRef
  usedBy: string
  resourceType: string
}

export interface MigrationDelta {
  usedByStats?: Record<string, number>
  conflicts?: ConflictRow[]
  progressiveMigration?: boolean
}

// Common Types

export interface PodRef {
  namespace: string
  name: string
}

export interface ObjectRef {
  namespace?: string
  name: string
  kind: string
}

export interface AppRef {
  namespace: string
  name: string
  count?: number
}

// Constants

export const UsedByTensorFusion = 'tensor-fusion'
export const UsedByNvidiaDevicePlugin = 'nvidia-device-plugin'
export const UsedByUnknown = 'unknown'

export const SchedulerTensorFusion = 'tensor-fusion-scheduler'
export const SchedulerDefault = 'default-scheduler'
