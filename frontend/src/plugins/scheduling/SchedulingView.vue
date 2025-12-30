<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-6 border-b">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <Activity class="w-5 h-5" />
            <h1 class="text-xl font-semibold">GPU Scheduling</h1>
          </div>
          <p class="text-xs text-muted-foreground">
            Real-time scheduling health, DRA allocation, and migration conflicts.
          </p>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2 rounded-lg border bg-background p-1.5">
            <Button
              :variant="activeView === 'nodes' ? 'secondary' : 'ghost'"
              size="sm"
              @click="activeView = 'nodes'"
            >
              <Server class="w-4 h-4 mr-2" />
              Nodes
            </Button>
            <Button
              :variant="activeView === 'pending' ? 'secondary' : 'ghost'"
              size="sm"
              @click="activeView = 'pending'"
            >
              <Clock class="w-4 h-4 mr-2" />
              Pending
              <Badge v-if="pendingCount > 0" variant="destructive" class="ml-2">
                {{ pendingCount }}
              </Badge>
            </Button>
            <Button
              :variant="activeView === 'dra' ? 'secondary' : 'ghost'"
              size="sm"
              @click="activeView = 'dra'"
            >
              <Layers class="w-4 h-4 mr-2" />
              DRA
            </Button>
            <Button
              :variant="activeView === 'migration' ? 'secondary' : 'ghost'"
              size="sm"
              @click="activeView = 'migration'"
            >
              <GitMerge class="w-4 h-4 mr-2" />
              Migration
            </Button>
          </div>
          <Button
            variant="ghost"
            size="sm"
            @click="refreshSnapshot"
            title="Refresh scheduler data"
            aria-label="Refresh scheduler data"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          </Button>
          <div class="flex items-center gap-2 rounded-full border bg-muted/40 px-2 py-1 text-xs text-muted-foreground">
            <Switch v-model="mockMode" class="scale-75" />
            <span>Mock data</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-auto p-6">
      <!-- Loading State -->
      <div v-if="loading && !snapshot" class="flex items-center justify-center h-full">
        <div class="text-center">
          <RefreshCw class="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Loading scheduler data...</p>
        </div>
      </div>

      <!-- No Cluster Selected -->
      <div v-else-if="!clusterId && !mockMode" class="flex items-center justify-center h-full">
        <div class="text-center">
          <Server class="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Select a cluster to view scheduling data</p>
        </div>
      </div>

      <div v-else-if="snapshot" class="space-y-6">
        <Card v-if="schedulerStore.error">
          <CardContent class="pt-4">
            <Alert variant="destructive">
              <AlertTriangle class="w-4 h-4" />
              <AlertTitle>Scheduler data unavailable</AlertTitle>
              <AlertDescription class="flex items-center justify-between gap-4">
                <span>{{ schedulerStore.error }}</span>
                <Button size="sm" variant="outline" @click="refreshSnapshot">
                  Retry
                </Button>
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-2">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <CardTitle>Status overview</CardTitle>
              <div class="flex items-center gap-2 text-xs text-muted-foreground">
                <Clock class="w-4 h-4" />
                <span>Updated {{ formatTimestamp(snapshot.generatedAt) }}</span>
                <Badge v-if="mockMode" variant="outline" class="text-[10px]">Mock</Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
              <div class="rounded-lg border bg-background px-3 py-3">
                <div class="flex items-center justify-between">
                  <div class="text-xs text-muted-foreground">Pending pods</div>
                  <Clock class="w-4 h-4 text-muted-foreground" />
                </div>
                <div class="mt-2 text-2xl font-semibold" :class="pendingCount > 0 ? 'text-amber-600' : ''">
                  {{ pendingCount }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ pendingCount > 0 ? 'Queue pressure' : 'No pending pods' }}
                </div>
              </div>
              <div class="rounded-lg border bg-background px-3 py-3">
                <div class="flex items-center justify-between">
                  <div class="text-xs text-muted-foreground">Failed schedules</div>
                  <AlertTriangle class="w-4 h-4 text-muted-foreground" />
                </div>
                <div class="mt-2 text-2xl font-semibold" :class="failedSchedules > 0 ? 'text-destructive' : ''">
                  {{ failedSchedules }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ failedSchedules > 0 ? 'Scheduling errors detected' : 'No scheduling failures' }}
                </div>
              </div>
              <div class="rounded-lg border bg-background px-3 py-3">
                <div class="flex items-center justify-between">
                  <div class="text-xs text-muted-foreground">DRA claims</div>
                  <Layers class="w-4 h-4" :class="draNotAvailable ? 'text-amber-500' : 'text-muted-foreground'" />
                </div>
                <div class="mt-2 text-2xl font-semibold" :class="draNotAvailable ? 'text-amber-600' : ''">
                  {{ draNotAvailable ? '-' : draClaims }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ draNotAvailable ? 'DRA not available' : (draClaims > 0 ? 'Active device claims' : 'No active claims') }}
                </div>
              </div>
              <div class="rounded-lg border bg-background px-3 py-3">
                <div class="flex items-center justify-between">
                  <div class="text-xs text-muted-foreground">Conflicts</div>
                  <GitMerge class="w-4 h-4 text-muted-foreground" />
                </div>
                <div class="mt-2 text-2xl font-semibold" :class="conflictCount > 0 ? 'text-destructive' : ''">
                  {{ conflictCount }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ conflictCount > 0 ? 'Migration mismatches' : 'No migration conflicts' }}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card v-if="warningCount > 0" class="border-amber-200 bg-amber-50">
          <CardContent class="pt-4 flex flex-wrap items-center justify-between gap-3 text-amber-900">
            <div class="flex items-start gap-2">
              <AlertTriangle class="w-4 h-4 mt-0.5" />
              <div>
                <div class="font-medium">Scheduler warnings</div>
                <div class="text-xs text-amber-700">
                  {{ warningCount }} warning{{ warningCount > 1 ? 's' : '' }} detected. Check scheduling mismatches and conflicts.
                </div>
              </div>
            </div>
            <Button size="sm" variant="outline" class="border-amber-200" @click="showWarnings = true">
              View details
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ activeViewLabel }}</CardTitle>
            <CardDescription>{{ activeViewDescription }}</CardDescription>
          </CardHeader>
          <CardContent>
            <KeepAlive>
              <NodeView v-if="activeView === 'nodes'" :data="snapshot.nodeView" />
              <PendingView
                v-else-if="activeView === 'pending'"
                :data="snapshot.pendingView"
                :last-updated="snapshot.generatedAt"
              />
              <DRAView v-else-if="activeView === 'dra'" :data="snapshot.draView" />
              <MigrationView v-else-if="activeView === 'migration'" :data="snapshot.migrationView" />
            </KeepAlive>
          </CardContent>
        </Card>
      </div>
    </div>

    <Sheet :open="showWarnings" @update:open="showWarnings = $event">
      <SheetContent side="right" class="w-[400px] sm:w-[560px]">
        <SheetHeader>
          <SheetTitle class="flex items-center gap-2">
            <AlertTriangle class="w-5 h-5 text-amber-600" />
            Scheduler Warnings
          </SheetTitle>
          <SheetDescription>
            Recent warning events from the scheduler aggregation.
          </SheetDescription>
        </SheetHeader>
        <div class="mt-4 space-y-3">
          <div
            v-for="{ warning, count } in deduplicatedWarnings"
            :key="`${warning.type}-${warning.message}`"
            class="rounded-md border p-3 bg-muted/30"
          >
            <div class="flex items-center justify-between text-sm">
              <div class="flex items-center gap-2">
                <span class="font-medium">{{ warning.type }}</span>
                <Badge v-if="count > 1" variant="secondary" class="text-xs">
                  {{ count }}x
                </Badge>
              </div>
              <span class="text-xs text-muted-foreground">{{ formatTimestamp(warning.generatedAt) }}</span>
            </div>
            <p class="text-sm text-muted-foreground mt-1">{{ warning.message }}</p>
            <div v-if="warning.objects?.length" class="mt-2 flex flex-wrap gap-2">
              <Badge v-for="obj in warning.objects" :key="`${obj.kind}-${obj.name}`" variant="outline">
                {{ obj.kind }} {{ obj.namespace ? `${obj.namespace}/` : '' }}{{ obj.name }}
              </Badge>
            </div>
          </div>
          <div v-if="deduplicatedWarnings.length === 0" class="text-sm text-muted-foreground">
            No warnings to display.
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Activity, Server, Clock, Layers, GitMerge, RefreshCw, AlertTriangle } from 'lucide-vue-next'
import NodeView from './views/NodeView.vue'
import PendingView from './views/PendingView.vue'
import DRAView from './views/DRAView.vue'
import MigrationView from './views/MigrationView.vue'
import { useSchedulerStore } from './stores/scheduler'
import { useClusterStore } from '@/shared/stores/cluster'
import { createMockSchedulerSnapshot, createMockSchedulerWarnings } from './mockData'

const schedulerStore = useSchedulerStore()
const clusterStore = useClusterStore()

const activeView = ref<'nodes' | 'pending' | 'dra' | 'migration'>('nodes')
const loading = ref(false)
const showWarnings = ref(false)

// Get current cluster ID from cluster store
const clusterId = computed(() => clusterStore.activeClusterId)
const snapshot = computed(() => schedulerStore.snapshot)
const pendingCount = computed(() => snapshot.value?.pendingView?.pods?.length ?? 0)
const failedSchedules = computed(() => snapshot.value?.pendingView?.pods?.filter(pod => !!pod.reason).length ?? 0)
const draClaims = computed(() => snapshot.value?.draView?.claims?.length ?? 0)
const draNotAvailable = computed(() => snapshot.value?.draView?.status === 'not_available')
const conflictCount = computed(() => snapshot.value?.migrationView?.conflicts?.length ?? 0)
const warnings = computed(() => schedulerStore.warnings)
const warningCount = computed(() => warnings.value.length)
// Deduplicate warnings by type+message, showing count for duplicates
const deduplicatedWarnings = computed(() => {
  const grouped = new Map<string, { warning: typeof warnings.value[0]; count: number }>()
  for (const warning of warnings.value) {
    const key = `${warning.type}:${warning.message}`
    const existing = grouped.get(key)
    if (existing) {
      existing.count++
      // Keep the most recent timestamp
      if (warning.generatedAt > existing.warning.generatedAt) {
        existing.warning = warning
      }
    } else {
      grouped.set(key, { warning, count: 1 })
    }
  }
  return Array.from(grouped.values())
})
const activeViewLabel = computed(() => {
  switch (activeView.value) {
    case 'nodes':
      return 'Nodes'
    case 'pending':
      return 'Pending queue'
    case 'dra':
      return 'DRA allocation'
    case 'migration':
      return 'Migration status'
    default:
      return 'Overview'
  }
})
const activeViewDescription = computed(() => {
  switch (activeView.value) {
    case 'nodes':
      return 'GPU usage, TensorFusion state, and per-node capacity.'
    case 'pending':
      return 'Pending pods grouped by scheduler with reasons and age.'
    case 'dra':
      return 'ResourceClaim to ResourceSlice to device mapping.'
    case 'migration':
      return 'UsedBy distribution and conflicting assignments.'
    default:
      return 'Scheduler overview.'
  }
})
const mockMode = computed({
  get: () => schedulerStore.mockMode,
  set: (value: boolean) => {
    if (value) {
      enableMockData()
    } else {
      disableMockData()
    }
  }
})

// Start aggregation when component mounts
onMounted(async () => {
  if (!mockMode.value) {
    // Load clusters if not already loaded
    if (clusterStore.clusterList.length === 0) {
      try {
        await clusterStore.loadClusters()
      } catch (e) {
        console.warn('Failed to load clusters:', e)
      }
    }

    if (clusterId.value) {
      await startAggregation()
    }
  }
})

// Watch for cluster changes
watch(clusterId, async (newClusterId, oldClusterId) => {
  if (mockMode.value) return
  // Stop previous aggregation
  if (oldClusterId) {
    await schedulerStore.stopAggregation()
  }

  // Start new aggregation
  if (newClusterId) {
    await startAggregation()
  }
})

// Cleanup on unmount
onUnmounted(() => {
  if (clusterId.value && !mockMode.value) {
    schedulerStore.stopAggregation()
  }
})

async function startAggregation() {
  if (!clusterId.value) return
  loading.value = true
  try {
    await schedulerStore.startAggregation({
      clusterId: clusterId.value,
      include: {
        pods: true,
        nodes: true,
        events: true,
        dra: true,
        tensorFusion: true
      }
    })
  } catch (e) {
    console.error('Failed to start aggregation:', e)
  } finally {
    loading.value = false
  }
}

async function refreshSnapshot() {
  if (mockMode.value) {
    schedulerStore.setMockSnapshot(createMockSchedulerSnapshot(), createMockSchedulerWarnings())
    return
  }
  if (!clusterId.value) return
  loading.value = true
  try {
    await schedulerStore.refreshSnapshot()
  } catch (e) {
    console.error('Failed to refresh snapshot:', e)
  } finally {
    loading.value = false
  }
}

async function enableMockData() {
  if (schedulerStore.isAggregating) {
    await schedulerStore.stopAggregation()
  }
  schedulerStore.setMockSnapshot(createMockSchedulerSnapshot(), createMockSchedulerWarnings())
}

async function disableMockData() {
  schedulerStore.clearMockSnapshot()
  if (clusterId.value) {
    await startAggregation()
  }
}

function formatTimestamp(value?: string): string {
  if (!value) return 'unknown'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString()
}
</script>
