<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Box class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Resources</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled>
            <Filter class="w-4 h-4 mr-2" />
            Filters
          </Button>
          <Button variant="outline" size="sm" disabled>
            <LayoutGrid class="w-4 h-4 mr-2" />
            Group by
          </Button>
        </div>
      </div>
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
            Connect to a cluster to browse Kubernetes resources.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="connectCluster">
              <Plug class="w-4 h-4 mr-2" />
              Connect to cluster
            </Button>
          </div>
        </div>
      </div>

      <!-- Resource Browser (Connected) -->
      <div v-else class="flex h-full">
        <!-- Resource Type Sidebar -->
        <div class="w-64 border-r bg-muted/30 p-4">
          <div class="mb-4">
            <Input placeholder="Search resources..." class="w-full" disabled />
          </div>
          <div class="space-y-1">
            <div class="text-xs font-semibold text-muted-foreground uppercase mb-2">Core Resources</div>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Box class="w-4 h-4 mr-2" />
              Pods
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Server class="w-4 h-4 mr-2" />
              Services
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <FolderOpen class="w-4 h-4 mr-2" />
              ConfigMaps
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Lock class="w-4 h-4 mr-2" />
              Secrets
            </Button>
            <div class="text-xs font-semibold text-muted-foreground uppercase mt-4 mb-2">Workloads</div>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Layers class="w-4 h-4 mr-2" />
              Deployments
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Copy class="w-4 h-4 mr-2" />
              ReplicaSets
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <Database class="w-4 h-4 mr-2" />
              StatefulSets
            </Button>
          </div>
        </div>

        <!-- Main Content Area -->
        <div class="flex-1 flex items-center justify-center">
          <div class="text-center max-w-md">
            <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
              <MousePointer class="w-8 h-8 text-muted-foreground" />
            </div>
            <h2 class="text-xl font-semibold mb-2">Select a resource type</h2>
            <p class="text-muted-foreground">
              Choose a resource type from the sidebar to browse and manage Kubernetes resources.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Box, Filter, LayoutGrid, Unplug, Plug, Server, FolderOpen, Lock,
  Layers, Copy, Database, MousePointer
} from 'lucide-vue-next'
import { useClusterStore } from '@/shared/stores/cluster'
import { useClusterDialog } from '@/shared/composables/useClusterDialog'

const clusterStore = useClusterStore()
const { open: openClusterDialog } = useClusterDialog()

const isConnected = computed(() => clusterStore.activeClusterId !== null)

onMounted(async () => {
  if (clusterStore.clusterList.length === 0) {
    try {
      await clusterStore.loadClusters()
    } catch (e) {
      console.warn('Failed to load clusters:', e)
    }
  }
})

function connectCluster() {
  openClusterDialog()
}
</script>
