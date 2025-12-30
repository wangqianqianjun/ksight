<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Activity class="w-5 h-5" />
          <h1 class="text-xl font-semibold">GPU Scheduling</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button
            :variant="activeView === 'nodes' ? 'default' : 'outline'"
            size="sm"
            @click="activeView = 'nodes'"
          >
            <Server class="w-4 h-4 mr-2" />
            Nodes
          </Button>
          <Button
            :variant="activeView === 'pending' ? 'default' : 'outline'"
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
            :variant="activeView === 'dra' ? 'default' : 'outline'"
            size="sm"
            @click="activeView = 'dra'"
          >
            <Layers class="w-4 h-4 mr-2" />
            DRA
          </Button>
          <Button
            :variant="activeView === 'migration' ? 'default' : 'outline'"
            size="sm"
            @click="activeView = 'migration'"
          >
            <GitMerge class="w-4 h-4 mr-2" />
            Migration
          </Button>
          <div class="border-l ml-2 pl-2">
            <Button variant="ghost" size="sm" @click="refreshSnapshot">
              <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Health Summary Bar -->
    <div v-if="snapshot" class="px-4 py-2 border-b bg-muted/30 flex items-center gap-6 text-sm">
      <div class="flex items-center gap-2">
        <Server class="w-4 h-4 text-muted-foreground" />
        <span class="text-muted-foreground">Nodes:</span>
        <span class="font-medium">{{ snapshot.health.totalNodes }}</span>
      </div>
      <div class="flex items-center gap-2">
        <Cpu class="w-4 h-4 text-muted-foreground" />
        <span class="text-muted-foreground">GPUs:</span>
        <span class="font-medium">{{ snapshot.health.totalGpus }}</span>
      </div>
      <div class="flex items-center gap-2">
        <Clock class="w-4 h-4" :class="snapshot.health.totalPending > 0 ? 'text-yellow-500' : 'text-muted-foreground'" />
        <span class="text-muted-foreground">Pending:</span>
        <span class="font-medium" :class="snapshot.health.totalPending > 0 ? 'text-yellow-500' : ''">
          {{ snapshot.health.totalPending }}
        </span>
      </div>
      <div v-if="snapshot.health.warnings?.length" class="flex items-center gap-2 text-destructive">
        <AlertTriangle class="w-4 h-4" />
        <span>{{ snapshot.health.warnings.length }} warnings</span>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-auto p-4">
      <!-- Loading State -->
      <div v-if="loading && !snapshot" class="flex items-center justify-center h-full">
        <div class="text-center">
          <RefreshCw class="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Loading scheduler data...</p>
        </div>
      </div>

      <!-- No Cluster Selected -->
      <div v-else-if="!clusterId" class="flex items-center justify-center h-full">
        <div class="text-center">
          <Server class="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Select a cluster to view scheduling data</p>
        </div>
      </div>

      <!-- Views -->
      <template v-else-if="snapshot">
        <NodeView v-if="activeView === 'nodes'" :data="snapshot.nodeView" />
        <PendingView v-else-if="activeView === 'pending'" :data="snapshot.pendingView" />
        <DRAView v-else-if="activeView === 'dra'" :data="snapshot.draView" />
        <MigrationView v-else-if="activeView === 'migration'" :data="snapshot.migrationView" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Activity, Server, Clock, Layers, GitMerge, RefreshCw, Cpu, AlertTriangle } from 'lucide-vue-next'
import NodeView from './views/NodeView.vue'
import PendingView from './views/PendingView.vue'
import DRAView from './views/DRAView.vue'
import MigrationView from './views/MigrationView.vue'
import { useSchedulerStore } from './stores/scheduler'
import { useClusterStore } from '@/shared/stores/cluster'

const schedulerStore = useSchedulerStore()
const clusterStore = useClusterStore()

const activeView = ref<'nodes' | 'pending' | 'dra' | 'migration'>('nodes')
const loading = ref(false)

// Get current cluster ID from cluster store
const clusterId = computed(() => clusterStore.activeClusterId)
const snapshot = computed(() => schedulerStore.snapshot)
const pendingCount = computed(() => snapshot.value?.health?.totalPending ?? 0)

// Start aggregation when component mounts
onMounted(async () => {
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
})

// Watch for cluster changes
watch(clusterId, async (newClusterId, oldClusterId) => {
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
  if (clusterId.value) {
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
</script>
