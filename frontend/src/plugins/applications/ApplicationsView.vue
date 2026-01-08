<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Layers class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Applications</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" :disabled="pods.length === 0">
            <Filter class="w-4 h-4 mr-2" />
            Filters
          </Button>
          <Button variant="outline" size="sm" :disabled="pods.length === 0">
            <LayoutGrid class="w-4 h-4 mr-2" />
            Group by
          </Button>
        </div>
      </div>
    </div>

    <!-- Summary Row (only shown when connected) -->
    <div v-if="isConnected && !loading" class="px-4 py-2 border-b bg-muted/30 flex items-center gap-4 text-sm">
      <Badge variant="secondary">
        <Box class="w-3 h-3 mr-1" />
        {{ stats.apps }} Apps
      </Badge>
      <Badge variant="secondary">
        <Container class="w-3 h-3 mr-1" />
        {{ stats.pods }} Pods
      </Badge>
      <Badge variant="secondary">
        <FolderOpen class="w-3 h-3 mr-1" />
        {{ stats.namespaces }} Namespaces
      </Badge>
      <Badge v-if="stats.unhealthy > 0" variant="destructive">
        <AlertCircle class="w-3 h-3 mr-1" />
        {{ stats.unhealthy }} Unhealthy
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
            Connect to a cluster to view applications and workloads.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="connectCluster">
              <Plug class="w-4 h-4 mr-2" />
              Connect to cluster
            </Button>
            <Button variant="outline" @click="openSettings">
              Open settings
            </Button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-else-if="loading" class="flex items-center justify-center h-full">
        <div class="text-center">
          <RefreshCw class="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
          <p class="text-muted-foreground">Loading applications...</p>
          <p class="text-xs text-muted-foreground mt-1">This may take a few seconds</p>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="loadError" class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <h2 class="text-xl font-semibold mb-2">Failed to load applications</h2>
          <p class="text-muted-foreground mb-6">{{ loadError }}</p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="fetchPods">Retry</Button>
            <Button variant="outline" @click="openSettings">Open settings</Button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="pods.length === 0" class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Inbox class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No applications found</h2>
          <p class="text-muted-foreground mb-6">
            This cluster has no running workloads yet.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button>
              <Plus class="w-4 h-4 mr-2" />
              Create workload
            </Button>
            <Button variant="outline" @click="openTemplates">
              Open templates
            </Button>
          </div>
        </div>
      </div>

      <!-- Data Table (when pods exist) -->
      <div v-else class="p-4">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Container class="w-5 h-5" />
              Workloads
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pod</TableHead>
                  <TableHead>Namespace</TableHead>
                  <TableHead>Node</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Age</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="pod in pods" :key="`${pod.namespace}/${pod.name}`">
                  <TableCell class="font-medium">{{ pod.name }}</TableCell>
                  <TableCell>{{ pod.namespace }}</TableCell>
                  <TableCell>{{ pod.nodeName || '-' }}</TableCell>
                  <TableCell>
                    <Badge :variant="getStatusVariant(pod.phase)">
                      {{ pod.phase }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{ formatAge(pod.age) }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-1">
                      <Button variant="ghost" size="sm" title="Logs">
                        <FileText class="w-4 h-4" />
                      </Button>
                      <Button variant="ghost" size="sm" title="Shell">
                        <Terminal class="w-4 h-4" />
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
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Layers, LayoutGrid, Filter, Box, Container, FolderOpen, AlertCircle,
  Unplug, Plug, RefreshCw, Inbox, Plus, FileText, Terminal, MoreHorizontal
} from 'lucide-vue-next'
import { useClusterStore } from '@/shared/stores/cluster'
import { useClusterDialog } from '@/shared/composables/useClusterDialog'
import { k8s } from '@/lib/k8s-sdk'

const router = useRouter()
const clusterStore = useClusterStore()
const { open: openClusterDialog } = useClusterDialog()

const loading = ref(false)
const loadError = ref<string | null>(null)
const pods = ref<any[]>([])

const isConnected = computed(() => clusterStore.activeClusterId !== null)

const stats = computed(() => ({
  apps: new Set(pods.value.map(p => p.name.split('-').slice(0, -1).join('-'))).size,
  pods: pods.value.length,
  namespaces: new Set(pods.value.map(p => p.namespace)).size,
  unhealthy: pods.value.filter(p => p.phase !== 'Running').length
}))

async function fetchPods() {
  if (!clusterStore.activeClusterId) return

  loading.value = true
  loadError.value = null
  try {
    pods.value = await k8s.getPods(clusterStore.activeClusterId)
  } catch (e) {
    console.error('Failed to fetch pods:', e)
    loadError.value = e instanceof Error ? e.message : String(e)
    pods.value = []
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

  // Fetch pods if connected
  if (clusterStore.activeClusterId) {
    await fetchPods()
  }
})

// Watch for cluster connection changes
watch(() => clusterStore.activeClusterId, async (newClusterId) => {
  if (newClusterId) {
    await fetchPods()
  } else {
    pods.value = []
  }
})

function connectCluster() {
  openClusterDialog()
}

function openSettings() {
  // TODO: Open settings
}

function openTemplates() {
  router.push('/plugins/templates')
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

function getStatusVariant(phase: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (phase?.toLowerCase()) {
    case 'running':
      return 'default'
    case 'pending':
      return 'secondary'
    case 'failed':
    case 'error':
      return 'destructive'
    default:
      return 'outline'
  }
}
</script>
