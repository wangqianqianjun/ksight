<template>
  <div class="space-y-6">
    <!-- Filters -->
    <Card>
      <CardContent class="pt-5 pb-5 space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground">
          <span>Showing {{ filteredNodes.length }} of {{ nodes.length }} nodes</span>
          <div v-if="Object.keys(gpusByUsedBy).length" class="flex flex-wrap items-center gap-2">
            <span class="uppercase text-[10px] tracking-wide">Used by</span>
            <div v-for="(count, usedBy) in gpusByUsedBy" :key="usedBy" class="flex items-center gap-1">
              <span class="w-2 h-2 rounded-full" :class="getUsedByColor(usedBy as string)"></span>
              <span>{{ usedBy }} {{ count }}</span>
            </div>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <Input
            v-model="filterNameInput"
            class="w-[220px]"
            placeholder="Filter node name"
          />
          <Select v-model="filterZone">
            <SelectTrigger class="w-[180px]">
              <SelectValue placeholder="All zones" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="allToken">All zones</SelectItem>
              <SelectItem v-for="zone in zones" :key="zone" :value="zone">
                {{ zone }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Select v-model="filterUsedBy">
            <SelectTrigger class="w-[200px]">
              <SelectValue placeholder="All usedBy" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="allToken">All usedBy</SelectItem>
              <SelectItem v-for="usedBy in usedByOptions" :key="usedBy" :value="usedBy">
                {{ usedBy }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Select v-model="filterPool">
            <SelectTrigger class="w-[180px]">
              <SelectValue placeholder="All pools" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="allToken">All pools</SelectItem>
              <SelectItem v-for="pool in pools" :key="pool" :value="pool">
                {{ pool }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Button v-if="hasFilters" variant="ghost" size="sm" @click="clearFilters">
            Clear filters
          </Button>
          <ToggleGroup v-model="viewMode" type="single" variant="outline" size="sm" class="ml-auto">
            <ToggleGroupItem value="table" aria-label="Table view">
              <List class="w-4 h-4" />
            </ToggleGroupItem>
            <ToggleGroupItem value="cards" aria-label="Card view">
              <LayoutGrid class="w-4 h-4" />
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
      </CardContent>
    </Card>

    <!-- Node Table -->
    <Card v-if="viewMode === 'table'">
      <CardContent class="pt-5 pb-5">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Node</TableHead>
              <TableHead>Zone</TableHead>
              <TableHead>GPU Total</TableHead>
              <TableHead>Allocatable / Used</TableHead>
              <TableHead>GPU Used By</TableHead>
              <TableHead>TF Phase</TableHead>
              <TableHead>Pool</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="node in filteredNodes"
              :key="node.name"
              class="cursor-pointer"
              @click="openNodeDetails(node)"
            >
              <TableCell class="font-medium">{{ node.name }}</TableCell>
              <TableCell>{{ node.zone || '-' }}</TableCell>
              <TableCell>{{ node.gpu?.total ?? 0 }}</TableCell>
              <TableCell>
                <span>{{ getAllocatable(node) ?? '-' }}</span>
                <span class="text-muted-foreground"> / </span>
                <span>{{ getUsedCount(node) }}</span>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <template v-for="(count, usedBy) in (node.gpu?.usedBy ?? {})" :key="usedBy">
                    <Badge :variant="getUsedByVariant(usedBy as string)">
                      {{ usedBy }}: {{ count }}
                    </Badge>
                  </template>
                  <span v-if="!node.gpu?.usedBy || Object.keys(node.gpu.usedBy).length === 0" class="text-muted-foreground">-</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge v-if="node.tensorFusion?.phase" :variant="getPhaseVariant(node.tensorFusion.phase)">
                  {{ node.tensorFusion.phase }}
                </Badge>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell>{{ node.tensorFusion?.pool || '-' }}</TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <Button variant="ghost" size="sm" title="Details">
                    <Server class="w-4 h-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    title="GPU metrics integration is coming in a future release. Will show real-time GPU utilization, memory usage, and temperature."
                    disabled
                    class="opacity-50"
                  >
                    <BarChart3 class="w-4 h-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    title="GPU devices"
                    @click.stop="expandNode = expandNode === node.name ? null : node.name"
                  >
                    <ChevronDown class="w-4 h-4" :class="{ 'rotate-180': expandNode === node.name }" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <!-- Expanded GPU Details -->
            <template v-if="expandNode">
              <TableRow
                v-for="node in filteredNodes.filter(n => n.name === expandNode)"
                :key="`${node.name}-detail`"
              >
                <TableCell colspan="8" class="bg-muted/30">
                  <div class="p-4">
                    <div class="flex items-center justify-between mb-3">
                      <h4 class="font-medium flex items-center gap-2">
                        <Server class="w-4 h-4" />
                        GPU Devices on {{ node.name }}
                      </h4>
                      <Button variant="ghost" size="sm" @click="expandNode = null" title="Close">
                        <X class="w-4 h-4" />
                      </Button>
                    </div>
                    <div v-if="node.gpu?.devices?.length" class="grid grid-cols-3 gap-4">
                      <div
                        v-for="device in node.gpu.devices"
                        :key="device.name"
                        class="border rounded-lg p-3 bg-background"
                      >
                        <div class="flex items-center justify-between mb-2">
                          <span class="font-medium" :title="device.name">{{ device.name }}</span>
                          <Badge :variant="getUsedByVariant(device.usedBy || 'unknown')">
                            {{ device.usedBy || 'unknown' }}
                          </Badge>
                        </div>
                        <div class="text-sm text-muted-foreground space-y-1">
                          <div v-if="device.model">Model: {{ device.model }}</div>
                          <div v-if="device.capacity?.['tflops']">TFlops: {{ device.capacity['tflops'] }}</div>
                          <div v-if="device.available?.['vram']">VRAM Avail: {{ device.available['vram'] }}</div>
                          <div v-if="device.runningApps?.length">
                            Apps: {{ getUniqueApps(device.runningApps) }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-else class="text-sm text-muted-foreground py-4 text-center">
                      No GPU devices found on this node.
                    </div>
                  </div>
                </TableCell>
              </TableRow>
            </template>
            <TableRow v-if="filteredNodes.length === 0">
              <TableCell colspan="8" class="text-center py-8 text-muted-foreground">
                No nodes match the current filters.
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <!-- Card View -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <Card
        v-for="node in filteredNodes"
        :key="node.name"
      >
        <CardHeader class="cursor-pointer" @click="openNodeDetails(node)">
          <CardTitle class="flex items-center justify-between gap-2">
            <span class="truncate">{{ node.name }}</span>
            <Badge v-if="node.tensorFusion?.phase" :variant="getPhaseVariant(node.tensorFusion.phase)">
              {{ node.tensorFusion.phase }}
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-3 cursor-pointer" @click="openNodeDetails(node)">
          <div class="flex items-center justify-between text-sm text-muted-foreground">
            <span>{{ node.zone || 'unknown zone' }}</span>
            <span>{{ node.tensorFusion?.pool || 'no pool' }}</span>
          </div>
          <div class="flex flex-wrap gap-2">
            <Badge
              v-for="(count, usedBy) in (node.gpu?.usedBy ?? {})"
              :key="usedBy"
              :variant="getUsedByVariant(usedBy as string)"
            >
              {{ usedBy }} {{ count }}
            </Badge>
            <span v-if="!node.gpu?.usedBy || Object.keys(node.gpu.usedBy).length === 0" class="text-muted-foreground">
              No usage data
            </span>
          </div>
          <div class="text-xs text-muted-foreground">
            Allocatable {{ getAllocatable(node) ?? '-' }} / Used {{ getUsedCount(node) }}
          </div>
          <div class="flex flex-wrap gap-2">
            <Badge
              v-for="device in node.gpu?.devices ?? []"
              :key="`${node.name}-${device.name}`"
              variant="outline"
              :title="device.model ? `${device.name} - ${device.model}` : device.name"
            >
              {{ device.name }}
            </Badge>
          </div>
        </CardContent>
        <CardFooter class="flex items-center gap-2 pt-3 border-t">
          <Button variant="ghost" size="sm" title="View details" @click="openNodeDetails(node)">
            <Server class="w-4 h-4 mr-1" />
            Details
          </Button>
          <Button
            variant="ghost"
            size="sm"
            title="GPU metrics integration is coming in a future release. Will show real-time GPU utilization, memory usage, and temperature."
            disabled
            class="opacity-50 cursor-not-allowed"
          >
            <BarChart3 class="w-4 h-4 mr-1" />
            Metrics
          </Button>
          <Button
            variant="ghost"
            size="sm"
            :title="expandNode === node.name ? 'Hide GPU devices' : 'Show GPU devices'"
            @click="expandNode = expandNode === node.name ? null : node.name"
          >
            <ChevronDown class="w-4 h-4 mr-1" :class="{ 'rotate-180': expandNode === node.name }" />
            GPUs
          </Button>
        </CardFooter>
        <!-- Expanded GPU Details for Card View -->
        <div v-if="expandNode === node.name" class="px-4 pb-4 border-t bg-muted/30">
          <div class="pt-3">
            <h4 class="font-medium mb-2 flex items-center justify-between">
              <span>GPU Devices</span>
              <Button variant="ghost" size="sm" @click="expandNode = null" title="Close">
                <X class="w-4 h-4" />
              </Button>
            </h4>
            <div class="grid grid-cols-1 gap-2">
              <div
                v-for="device in node.gpu?.devices ?? []"
                :key="device.name"
                class="border rounded-lg p-2 bg-background text-sm"
              >
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium">{{ device.name }}</span>
                  <Badge :variant="getUsedByVariant(device.usedBy || 'unknown')" class="text-xs">
                    {{ device.usedBy || 'unknown' }}
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground space-y-0.5">
                  <div v-if="device.model">Model: {{ device.model }}</div>
                  <div v-if="device.runningApps?.length">
                    Apps: {{ getUniqueApps(device.runningApps) }}
                  </div>
                </div>
              </div>
              <div v-if="!node.gpu?.devices?.length" class="text-sm text-muted-foreground py-2">
                No GPU devices found on this node.
              </div>
            </div>
          </div>
        </div>
      </Card>
      <Card v-if="filteredNodes.length === 0">
        <CardContent class="py-8 text-center text-muted-foreground">
          No nodes match the current filters.
        </CardContent>
      </Card>
    </div>

    <Sheet :open="showDetails" @update:open="showDetails = $event">
      <SheetContent side="right" class="w-[420px] sm:w-[640px]">
        <SheetHeader>
          <SheetTitle class="flex items-center gap-2">
            <Server class="w-5 h-5" />
            {{ selectedNode?.name || 'Node details' }}
          </SheetTitle>
          <SheetDescription>
            GPU usage, TensorFusion state, and workload hints.
          </SheetDescription>
        </SheetHeader>
        <div v-if="selectedNode" class="mt-4 space-y-4">
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div>
              <div class="text-muted-foreground">Zone</div>
              <div>{{ selectedNode.zone || '-' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">Pool</div>
              <div>{{ selectedNode.tensorFusion?.pool || '-' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">TF Phase</div>
              <div>{{ selectedNode.tensorFusion?.phase || '-' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">Allocatable / Used</div>
              <div>{{ getAllocatable(selectedNode) ?? '-' }} / {{ getUsedCount(selectedNode) }}</div>
            </div>
          </div>
          <div>
            <h4 class="font-medium mb-2">GPU Devices</h4>
            <div class="grid grid-cols-2 gap-2">
              <div
                v-for="device in selectedNode.gpu?.devices ?? []"
                :key="`${selectedNode.name}-${device.name}`"
                class="rounded-md border p-2 text-sm"
              >
                <div class="flex items-center justify-between">
                  <span class="font-medium">{{ device.name }}</span>
                  <Badge :variant="getUsedByVariant(device.usedBy || 'unknown')">
                    {{ device.usedBy || 'unknown' }}
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground mt-1 space-y-1">
                  <div v-if="device.model">Model: {{ device.model }}</div>
                  <div v-if="device.capacity?.['tflops']">TFlops: {{ device.capacity['tflops'] }}</div>
                  <div v-if="device.available?.['vram']">VRAM Avail: {{ device.available['vram'] }}</div>
                </div>
              </div>
            </div>
          </div>
          <div v-if="selectedNode.pods?.length">
            <h4 class="font-medium mb-2">Pods</h4>
            <div class="flex flex-wrap gap-2">
              <Badge
                v-for="pod in selectedNode.pods"
                :key="`${pod.namespace}/${pod.name}`"
                variant="outline"
              >
                {{ pod.namespace }}/{{ pod.name }}
              </Badge>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Server, ChevronDown, LayoutGrid, List, BarChart3, X } from 'lucide-vue-next'
import type { NodeRow, NodeViewSnapshot } from '../types'
import { UsedByTensorFusion, UsedByNvidiaDevicePlugin } from '../types'

const props = defineProps<{
  data: NodeViewSnapshot
}>()

const expandNode = ref<string | null>(null)
const viewMode = ref<'table' | 'cards'>('table')
const selectedNode = ref<NodeRow | null>(null)
const showDetails = ref(false)

const filterNameInput = ref('')
const filterName = ref('')
const allToken = '__all__'
const filterZone = ref(allToken)
const filterUsedBy = ref(allToken)
const filterPool = ref(allToken)
let filterTimer: number | undefined

// Computed properties for null safety
const summary = computed(() => props.data?.summary ?? { totalNodes: 0, readyNodes: 0, totalGpus: 0, gpusByUsedBy: {} })
const nodes = computed(() => props.data?.nodes ?? [])
const gpusByUsedBy = computed(() => summary.value?.gpusByUsedBy ?? {})

const zones = computed(() => {
  const set = new Set<string>()
  nodes.value.forEach(node => {
    if (node.zone) set.add(node.zone)
  })
  return Array.from(set)
})

const pools = computed(() => {
  const set = new Set<string>()
  nodes.value.forEach(node => {
    const pool = node.tensorFusion?.pool
    if (pool) set.add(pool)
  })
  return Array.from(set)
})

const usedByOptions = computed(() => {
  const set = new Set<string>()
  nodes.value.forEach(node => {
    Object.keys(node.gpu?.usedBy ?? {}).forEach(key => set.add(key))
    node.gpu?.devices?.forEach(device => {
      if (device.usedBy) set.add(device.usedBy)
    })
  })
  return Array.from(set)
})

const filteredNodes = computed(() => {
  return nodes.value.filter(node => {
    if (filterName.value && !node.name.toLowerCase().includes(filterName.value)) return false
    if (filterZone.value !== allToken && node.zone !== filterZone.value) return false
    if (filterPool.value !== allToken && node.tensorFusion?.pool !== filterPool.value) return false
    if (filterUsedBy.value !== allToken) {
      const usedBy = filterUsedBy.value
      const matchesSummary = node.gpu?.usedBy?.[usedBy]
      const matchesDevices = node.gpu?.devices?.some(device => device.usedBy === usedBy)
      if (!matchesSummary && !matchesDevices) return false
    }
    return true
  })
})

const hasFilters = computed(() => {
  return !!(filterNameInput.value || filterZone.value !== allToken || filterUsedBy.value !== allToken || filterPool.value !== allToken)
})

watch(filterNameInput, (value) => {
  if (filterTimer) {
    window.clearTimeout(filterTimer)
  }
  filterTimer = window.setTimeout(() => {
    filterName.value = value.trim().toLowerCase()
  }, 250)
})

watch(filteredNodes, (next) => {
  if (expandNode.value && !next.some(node => node.name === expandNode.value)) {
    expandNode.value = null
  }
})

function clearFilters() {
  filterNameInput.value = ''
  filterName.value = ''
  filterZone.value = allToken
  filterUsedBy.value = allToken
  filterPool.value = allToken
}

function openNodeDetails(node: NodeRow) {
  selectedNode.value = node
  showDetails.value = true
}

function getUsedCount(node: NodeRow): number {
  return Object.values(node.gpu?.usedBy ?? {}).reduce((sum, count) => sum + count, 0)
}

function getAllocatable(node: NodeRow): string | number | null {
  const allocatable = node.allocatable?.['nvidia.com/gpu'] ?? node.capacity?.['nvidia.com/gpu']
  return allocatable ?? null
}

function getUsedByColor(usedBy: string): string {
  switch (usedBy) {
    case UsedByTensorFusion:
      return 'bg-blue-500'
    case UsedByNvidiaDevicePlugin:
      return 'bg-green-500'
    default:
      return 'bg-gray-500'
  }
}

function getUsedByVariant(usedBy: string): 'default' | 'secondary' | 'outline' {
  switch (usedBy) {
    case UsedByTensorFusion:
      return 'default'
    case UsedByNvidiaDevicePlugin:
      return 'secondary'
    default:
      return 'outline'
  }
}

function getPhaseVariant(phase: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (phase.toLowerCase()) {
    case 'running':
      return 'default'
    case 'pending':
      return 'secondary'
    case 'migrating':
      return 'outline'
    default:
      return 'outline'
  }
}

// Deduplicate running apps and format as string
function getUniqueApps(apps: Array<{ namespace: string; name: string; count?: number }>): string {
  const seen = new Map<string, number>()
  for (const app of apps) {
    const key = `${app.namespace}/${app.name}`
    seen.set(key, (seen.get(key) ?? 0) + (app.count ?? 1))
  }
  return Array.from(seen.entries())
    .map(([key, count]) => count > 1 ? `${key} (${count})` : key)
    .join(', ')
}
</script>
