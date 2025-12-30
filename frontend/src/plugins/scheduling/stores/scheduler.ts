import { defineStore } from 'pinia'
import { k8s, wsManager } from '@/lib/k8s-sdk'
import type {
  SchedulerSnapshot,
  SchedulerDelta,
  SchedulerWarning,
  SchedulerAggregationRequest
} from '../types'

export const useSchedulerStore = defineStore('scheduler', () => {
  // State
  const clusterId = ref<string | null>(null)
  const snapshot = ref<SchedulerSnapshot | null>(null)
  const warnings = ref<SchedulerWarning[]>([])
  const lastSeq = ref<number>(0)
  const isAggregating = ref(false)
  const error = ref<string | null>(null)

  // Event unsubscribe functions
  let unsubSnapshot: (() => void) | null = null
  let unsubDelta: (() => void) | null = null
  let unsubWarning: (() => void) | null = null

  // Actions
  async function startAggregation(request: SchedulerAggregationRequest) {
    if (isAggregating.value && clusterId.value === request.clusterId) {
      return // Already aggregating for this cluster
    }

    // Stop previous aggregation if any
    if (isAggregating.value && clusterId.value) {
      await stopAggregation()
    }

    clusterId.value = request.clusterId
    error.value = null

    try {
      // Subscribe to events
      subscribeToEvents()

      // Start backend aggregation via HTTP API
      await k8s.startSchedulerAggregation(request)
      isAggregating.value = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      unsubscribeFromEvents()
      throw e
    }
  }

  async function stopAggregation() {
    if (!clusterId.value) return

    try {
      await k8s.stopSchedulerAggregation(clusterId.value)
    } catch (e) {
      console.error('Failed to stop aggregation:', e)
    } finally {
      unsubscribeFromEvents()
      isAggregating.value = false
      snapshot.value = null
      lastSeq.value = 0
    }
  }

  async function refreshSnapshot() {
    if (!clusterId.value) return

    try {
      const newSnapshot = await k8s.getSchedulerSnapshot(clusterId.value)
      snapshot.value = newSnapshot
      lastSeq.value = newSnapshot.seq
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      throw e
    }
  }

  function subscribeToEvents() {
    unsubSnapshot = wsManager.on('scheduler:snapshot', (data: SchedulerSnapshot) => {
      if (data?.clusterId === clusterId.value) {
        snapshot.value = data
        lastSeq.value = data.seq
      }
    })

    unsubDelta = wsManager.on('scheduler:delta', (data: SchedulerDelta) => {
      if (data?.clusterId === clusterId.value) {
        // Check for seq continuity
        if (data.seq !== lastSeq.value + 1) {
          // Seq gap detected, request full snapshot
          refreshSnapshot()
          return
        }

        applyDelta(data)
        lastSeq.value = data.seq
      }
    })

    unsubWarning = wsManager.on('scheduler:warning', (data: SchedulerWarning) => {
      if (data?.clusterId === clusterId.value) {
        warnings.value.push(data)
        // Keep only last 100 warnings
        if (warnings.value.length > 100) {
          warnings.value = warnings.value.slice(-100)
        }
      }
    })
  }

  function unsubscribeFromEvents() {
    unsubSnapshot?.()
    unsubDelta?.()
    unsubWarning?.()
    unsubSnapshot = null
    unsubDelta = null
    unsubWarning = null
  }

  function applyDelta(delta: SchedulerDelta) {
    if (!snapshot.value) return

    // Apply node view delta
    if (delta.nodeView) {
      const nodeView = snapshot.value.nodeView

      // Remove nodes
      if (delta.nodeView.remove) {
        const removeSet = new Set(delta.nodeView.remove)
        nodeView.nodes = nodeView.nodes.filter(n => !removeSet.has(n.name))
      }

      // Upsert nodes
      if (delta.nodeView.upsert) {
        for (const upsertNode of delta.nodeView.upsert) {
          const idx = nodeView.nodes.findIndex(n => n.name === upsertNode.name)
          if (idx >= 0) {
            nodeView.nodes[idx] = upsertNode
          } else {
            nodeView.nodes.push(upsertNode)
          }
        }
      }

      // Update summary
      if (delta.nodeView.summary) {
        nodeView.summary = delta.nodeView.summary
      }
    }

    // Apply pending view delta
    if (delta.pendingView) {
      const pendingView = snapshot.value.pendingView

      // Remove pods
      if (delta.pendingView.remove) {
        const removeKeys = new Set(delta.pendingView.remove.map(p => `${p.namespace}/${p.name}`))
        pendingView.pods = pendingView.pods.filter(
          p => !removeKeys.has(`${p.namespace}/${p.name}`)
        )
      }

      // Upsert pods
      if (delta.pendingView.upsert) {
        for (const upsertPod of delta.pendingView.upsert) {
          const idx = pendingView.pods.findIndex(
            p => p.namespace === upsertPod.namespace && p.name === upsertPod.name
          )
          if (idx >= 0) {
            pendingView.pods[idx] = upsertPod
          } else {
            pendingView.pods.push(upsertPod)
          }
        }
      }

      // Update aggregates
      if (delta.pendingView.byScheduler) {
        pendingView.byScheduler = delta.pendingView.byScheduler
      }
      if (delta.pendingView.reasons) {
        pendingView.reasons = delta.pendingView.reasons
      }
    }

    // Apply DRA view delta
    if (delta.draView) {
      const draView = snapshot.value.draView

      // Claims
      if (delta.draView.removeClaims) {
        const removeKeys = new Set(delta.draView.removeClaims.map(p => `${p.namespace}/${p.name}`))
        draView.claims = draView.claims.filter(
          c => !removeKeys.has(`${c.namespace}/${c.name}`)
        )
      }
      if (delta.draView.upsertClaims) {
        for (const claim of delta.draView.upsertClaims) {
          const idx = draView.claims.findIndex(
            c => c.namespace === claim.namespace && c.name === claim.name
          )
          if (idx >= 0) {
            draView.claims[idx] = claim
          } else {
            draView.claims.push(claim)
          }
        }
      }

      // Slices
      if (delta.draView.removeSlices) {
        const removeSet = new Set(delta.draView.removeSlices)
        draView.slices = draView.slices.filter(s => !removeSet.has(s.name))
      }
      if (delta.draView.upsertSlices) {
        for (const slice of delta.draView.upsertSlices) {
          const idx = draView.slices.findIndex(s => s.name === slice.name)
          if (idx >= 0) {
            draView.slices[idx] = slice
          } else {
            draView.slices.push(slice)
          }
        }
      }

      // Stats
      if (delta.draView.stats) {
        draView.stats = delta.draView.stats
      }
    }

    // Apply migration view delta
    if (delta.migrationView) {
      const migrationView = snapshot.value.migrationView

      if (delta.migrationView.usedByStats) {
        migrationView.usedByStats = delta.migrationView.usedByStats
      }
      if (delta.migrationView.conflicts) {
        migrationView.conflicts = delta.migrationView.conflicts
      }
      if (delta.migrationView.progressiveMigration !== undefined) {
        migrationView.progressiveMigration = delta.migrationView.progressiveMigration
      }
    }

    // Update health
    snapshot.value.health.totalPending = snapshot.value.pendingView.pods.length
    snapshot.value.health.totalNodes = snapshot.value.nodeView.nodes.length
    snapshot.value.health.totalGpus = snapshot.value.nodeView.summary.totalGpus
  }

  function clearWarnings() {
    warnings.value = []
  }

  function setClusterId(id: string) {
    clusterId.value = id
  }

  return {
    // State
    clusterId,
    snapshot,
    warnings,
    lastSeq,
    isAggregating,
    error,

    // Actions
    startAggregation,
    stopAggregation,
    refreshSnapshot,
    clearWarnings,
    setClusterId
  }
})
