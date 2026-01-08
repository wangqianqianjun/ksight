<template>
  <Sheet :open="isOpen" @update:open="handleOpenChange">
    <SheetContent side="right" class="w-[400px] sm:w-[540px]">
      <SheetHeader>
        <SheetTitle class="flex items-center gap-2">
          <Server class="w-5 h-5" />
          Connect to Cluster
        </SheetTitle>
        <SheetDescription>
          Connect to a Kubernetes cluster using kubeconfig or by selecting from available clusters.
        </SheetDescription>
      </SheetHeader>

      <div class="py-6 space-y-6">
        <!-- Connection Method Tabs -->
        <Tabs v-model="connectionMethod" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="existing">Existing Clusters</TabsTrigger>
            <TabsTrigger value="kubeconfig">Kubeconfig</TabsTrigger>
          </TabsList>

          <!-- Existing Clusters Tab -->
          <TabsContent value="existing" class="space-y-4">
            <div v-if="loading" class="flex items-center justify-center py-8">
              <RefreshCw class="w-6 h-6 animate-spin text-muted-foreground" />
            </div>

            <div v-else-if="clusters.length === 0" class="text-center py-8">
              <div class="w-12 h-12 rounded-full bg-muted flex items-center justify-center mx-auto mb-3">
                <Inbox class="w-6 h-6 text-muted-foreground" />
              </div>
              <p class="text-sm text-muted-foreground">No clusters available</p>
              <p class="text-xs text-muted-foreground mt-1">Add a cluster using kubeconfig</p>
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="cluster in clusters"
                :key="cluster.id"
                class="flex items-center gap-3 p-3 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                :class="{ 'border-primary bg-primary/5': selectedCluster === cluster.id }"
                @click="selectedCluster = cluster.id"
              >
                <div class="w-2 h-2 rounded-full" :class="cluster.connected ? 'bg-green-500' : 'bg-gray-400'" />
                <Server class="w-4 h-4 text-muted-foreground" />
                <div class="flex-1">
                  <div class="font-medium text-sm">{{ cluster.name }}</div>
                  <div class="text-xs text-muted-foreground">{{ cluster.context || 'default' }}</div>
                </div>
                <CheckCircle v-if="selectedCluster === cluster.id" class="w-4 h-4 text-primary" />
              </div>
            </div>
          </TabsContent>

          <!-- Kubeconfig Tab -->
          <TabsContent value="kubeconfig" class="space-y-4">
            <div class="space-y-2">
              <Label for="cluster-name">Cluster Name</Label>
              <Input
                id="cluster-name"
                v-model="clusterName"
                placeholder="my-cluster"
              />
            </div>

            <div class="space-y-2">
              <Label for="kubeconfig">Kubeconfig</Label>
              <Textarea
                id="kubeconfig"
                v-model="kubeconfig"
                placeholder="Paste your kubeconfig YAML here..."
                class="min-h-[200px] font-mono text-xs"
              />
            </div>

            <div class="space-y-2">
              <Label for="context">Context (optional)</Label>
              <Input
                id="context"
                v-model="context"
                placeholder="Leave empty for default context"
              />
            </div>

            <Button variant="outline" size="sm" @click="loadDefaultKubeconfig" :disabled="loadingDefault">
              <RefreshCw v-if="loadingDefault" class="w-4 h-4 mr-2 animate-spin" />
              <Upload v-else class="w-4 h-4 mr-2" />
              Load from ~/.kube/config
            </Button>
          </TabsContent>
        </Tabs>

        <!-- Error Message -->
        <div v-if="error" class="p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
          <p class="text-sm text-destructive">{{ error }}</p>
        </div>
      </div>

      <SheetFooter>
        <Button variant="outline" @click="close">Cancel</Button>
        <Button @click="connect" :disabled="!canConnect || connecting">
          <RefreshCw v-if="connecting" class="w-4 h-4 mr-2 animate-spin" />
          <Plug v-else class="w-4 h-4 mr-2" />
          {{ connecting ? 'Connecting...' : 'Connect' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Server, RefreshCw, Inbox, CheckCircle, Plug, Upload } from 'lucide-vue-next'
import { useClusterDialog } from '@/shared/composables/useClusterDialog'
import { useClusterStore } from '@/shared/stores/cluster'
import { k8s } from '@/lib/k8s-sdk'

const { isOpen, close } = useClusterDialog()
const clusterStore = useClusterStore()

// Form state
const connectionMethod = ref('existing')
const selectedCluster = ref<string | null>(null)
const clusterName = ref('')
const kubeconfig = ref('')
const context = ref('')

// UI state
const loading = ref(false)
const loadingDefault = ref(false)
const connecting = ref(false)
const error = ref<string | null>(null)

// Computed
const clusters = computed(() => clusterStore.clusterList)

const canConnect = computed(() => {
  if (connectionMethod.value === 'existing') {
    return selectedCluster.value !== null
  }
  return clusterName.value.trim() !== '' && kubeconfig.value.trim() !== ''
})

// Watch for dialog open to refresh clusters
watch(isOpen, async (open) => {
  if (open) {
    error.value = null
    loading.value = true
    try {
      await clusterStore.loadClusters()
    } catch (e) {
      console.warn('Failed to load clusters:', e)
    } finally {
      loading.value = false
    }
  }
})

function handleOpenChange(open: boolean) {
  if (!open) {
    close()
  }
}

async function loadDefaultKubeconfig() {
  loadingDefault.value = true
  error.value = null
  try {
    const clusterId = await k8s.loadDefaultKubeconfig()
    await clusterStore.loadClusters()
    clusterStore.setActiveCluster(clusterId)
    close()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loadingDefault.value = false
  }
}

async function connect() {
  connecting.value = true
  error.value = null

  try {
    if (connectionMethod.value === 'existing') {
      // Just set the active cluster
      if (selectedCluster.value) {
        clusterStore.setActiveCluster(selectedCluster.value)
        close()
      }
    } else {
      // Add new cluster with kubeconfig
      await clusterStore.addCluster(
        clusterName.value.trim(),
        kubeconfig.value,
        context.value.trim() || clusterName.value.trim()
      )
      close()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    connecting.value = false
  }
}
</script>
