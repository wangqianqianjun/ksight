<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <PlayCircle class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Operations</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" :disabled="!isConnected">
            <Filter class="w-4 h-4 mr-2" />
            Filters
          </Button>
          <Button size="sm" :disabled="!isConnected">
            <Plus class="w-4 h-4 mr-2" />
            New Operation
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
            Connect to a cluster to run operations and automation scripts.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button @click="connectCluster">
              <Plug class="w-4 h-4 mr-2" />
              Connect to cluster
            </Button>
          </div>
        </div>
      </div>

      <!-- Empty State (Connected but no operations) -->
      <div v-else class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Inbox class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No operations yet</h2>
          <p class="text-muted-foreground mb-6">
            Create TypeScript operations to automate Kubernetes tasks and troubleshooting workflows.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button>
              <Plus class="w-4 h-4 mr-2" />
              Create operation
            </Button>
            <Button variant="outline">
              Browse templates
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
import { PlayCircle, Filter, Plus, Unplug, Plug, Inbox } from 'lucide-vue-next'
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
