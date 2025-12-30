<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <LayoutDashboard class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Custom Boards</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button size="sm">
            <Plus class="w-4 h-4 mr-2" />
            New Board
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
            Connect to a cluster to create custom dashboards with metrics and operations.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="connectCluster">
              <Plug class="w-4 h-4 mr-2" />
              Connect to cluster
            </Button>
          </div>
        </div>
      </div>

      <!-- Empty State (Connected but no boards) -->
      <div v-else class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Inbox class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No boards yet</h2>
          <p class="text-muted-foreground mb-6">
            Create custom dashboards to monitor your cluster with panels, metrics, and quick operations.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button>
              <Plus class="w-4 h-4 mr-2" />
              Create board
            </Button>
            <Button variant="outline">
              Use template
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { LayoutDashboard, Plus, Unplug, Plug, Inbox } from 'lucide-vue-next'
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
