<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Server class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Nodes</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled>
            <Cpu class="w-4 h-4 mr-2" />
            CPU
          </Button>
          <Button variant="outline" size="sm" disabled>
            <HardDrive class="w-4 h-4 mr-2" />
            Memory
          </Button>
          <Button variant="outline" size="sm" disabled>
            <PlayCircle class="w-4 h-4 mr-2" />
            Simulate Schedule
          </Button>
        </div>
      </div>
    </div>

    <!-- Summary Row (only shown when connected and has data) -->
    <div v-if="isConnected && !loading && nodes.length > 0" class="px-4 py-2 border-b bg-muted/30 flex items-center gap-4 text-sm">
      <Badge variant="secondary">
        <Server class="w-3 h-3 mr-1" />
        {{ nodes.length }} Nodes
      </Badge>
      <Badge variant="secondary">
        <CheckCircle class="w-3 h-3 mr-1" />
        {{ readyCount }} Ready
      </Badge>
      <Badge v-if="notReadyCount > 0" variant="destructive">
        <AlertCircle class="w-3 h-3 mr-1" />
        {{ notReadyCount }} Not Ready
      </Badge>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-auto">
      <!-- Disconnected State -->
      <div v-if="!isConnected" class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Unplug class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No cluster connected</h2>
          <p class="text-muted-foreground mb-6">
            Connect to a cluster to view and manage nodes.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="connectCluster">
              <Plug class="w-4 h-4 mr-2" />
              Connect to cluster
            </Button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-else-if="loading" class="flex items-center justify-center h-full">
        <div class="text-center">
          <RefreshCw class="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Loading nodes...</p>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="nodes.length === 0" class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Inbox class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No nodes found</h2>
          <p class="text-muted-foreground mb-6">
            This cluster has no nodes. Check your cluster configuration.
          </p>
          <Button variant="outline" @click="fetchNodes">
            <RefreshCw class="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      <!-- Data Table -->
      <div v-else class="p-4">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Server class="w-5 h-5" />
              Cluster Nodes
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Node</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Roles</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Age</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="node in nodes" :key="node.name">
                  <TableCell class="font-medium">{{ node.name }}</TableCell>
                  <TableCell>
                    <Badge :variant="node.ready ? 'default' : 'destructive'">
                      {{ node.ready ? 'Ready' : 'Not Ready' }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{ node.roles || '-' }}</TableCell>
                  <TableCell>{{ node.version || '-' }}</TableCell>
                  <TableCell>{{ formatAge(node.age) }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-1">
                      <Button variant="ghost" size="sm" title="Shell">
                        <Terminal class="w-4 h-4" />
                      </Button>
                      <Button variant="ghost" size="sm" title="Cordon">
                        <Ban class="w-4 h-4" />
                      </Button>
                      <Button variant="ghost" size="sm" title="Details">
                        <MoreHorizontal class="w-4 h-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Server, Cpu, HardDrive, PlayCircle, Unplug, Plug, RefreshCw, Inbox,
  CheckCircle, AlertCircle, Terminal, Ban, MoreHorizontal
} from 'lucide-vue-next'
import { useClusterStore } from '@/shared/stores/cluster'
import { useClusterDialog } from '@/shared/composables/useClusterDialog'

const clusterStore = useClusterStore()
const { open: openClusterDialog } = useClusterDialog()

const loading = ref(false)
const nodes = ref<any[]>([])

const isConnected = computed(() => clusterStore.activeClusterId !== null)
const readyCount = computed(() => nodes.value.filter(n => n.ready).length)
const notReadyCount = computed(() => nodes.value.filter(n => !n.ready).length)

async function fetchNodes() {
  if (!clusterStore.activeClusterId) return

  loading.value = true
  try {
    const response = await fetch(`/api/clusters/${clusterStore.activeClusterId}/nodes`)
    if (response.ok) {
      nodes.value = await response.json() || []
    } else {
      console.error('Failed to fetch nodes:', await response.text())
      nodes.value = []
    }
  } catch (e) {
    console.error('Failed to fetch nodes:', e)
    nodes.value = []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (clusterStore.clusterList.length === 0) {
    try {
      await clusterStore.loadClusters()
    } catch (e) {
      console.warn('Failed to load clusters:', e)
    }
  }

  if (clusterStore.activeClusterId) {
    await fetchNodes()
  }
})

watch(() => clusterStore.activeClusterId, async (newClusterId) => {
  if (newClusterId) {
    await fetchNodes()
  } else {
    nodes.value = []
  }
})

function connectCluster() {
  openClusterDialog()
}

function formatAge(timestamp?: string): string {
  if (!timestamp) return '-'

  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffDay > 0) return `${diffDay}d`
  if (diffHour > 0) return `${diffHour}h`
  if (diffMin > 0) return `${diffMin}m`
  return `${diffSec}s`
}
</script>
