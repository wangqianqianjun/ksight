<template>
  <div class="space-y-4">
    <!-- Summary Cards -->
    <div class="grid grid-cols-4 gap-4">
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ summary.totalNodes }}</div>
          <p class="text-sm text-muted-foreground">Total Nodes</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ summary.readyNodes }}</div>
          <p class="text-sm text-muted-foreground">Ready Nodes</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ summary.totalGpus }}</div>
          <p class="text-sm text-muted-foreground">Total GPUs</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="flex items-center gap-2">
            <div v-for="(count, usedBy) in gpusByUsedBy" :key="usedBy" class="flex items-center gap-1">
              <span class="w-3 h-3 rounded-full" :class="getUsedByColor(usedBy as string)"></span>
              <span class="text-sm">{{ count }}</span>
            </div>
          </div>
          <p class="text-sm text-muted-foreground">GPU Distribution</p>
        </CardContent>
      </Card>
    </div>

    <!-- Node Table -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Server class="w-5 h-5" />
          Nodes
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Node</TableHead>
              <TableHead>Zone</TableHead>
              <TableHead>GPU Total</TableHead>
              <TableHead>GPU Used By</TableHead>
              <TableHead>TF Phase</TableHead>
              <TableHead>Pool</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="node in nodes" :key="node.name">
              <TableCell class="font-medium">{{ node.name }}</TableCell>
              <TableCell>{{ node.zone || '-' }}</TableCell>
              <TableCell>{{ node.gpu?.total ?? 0 }}</TableCell>
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
                <Button variant="ghost" size="sm" @click="expandNode = expandNode === node.name ? null : node.name">
                  <ChevronDown class="w-4 h-4" :class="{ 'rotate-180': expandNode === node.name }" />
                </Button>
              </TableCell>
            </TableRow>
            <!-- Expanded GPU Details -->
            <template v-if="expandNode">
              <TableRow v-for="node in nodes.filter(n => n.name === expandNode)" :key="`${node.name}-detail`">
                <TableCell colspan="7" class="bg-muted/30">
                  <div class="p-4">
                    <h4 class="font-medium mb-2">GPU Devices</h4>
                    <div class="grid grid-cols-3 gap-4">
                      <div
                        v-for="device in node.gpu.devices"
                        :key="device.name"
                        class="border rounded-lg p-3 bg-background"
                      >
                        <div class="flex items-center justify-between mb-2">
                          <span class="font-medium">{{ device.name }}</span>
                          <Badge :variant="getUsedByVariant(device.usedBy || 'unknown')">
                            {{ device.usedBy || 'unknown' }}
                          </Badge>
                        </div>
                        <div class="text-sm text-muted-foreground space-y-1">
                          <div v-if="device.model">Model: {{ device.model }}</div>
                          <div v-if="device.capacity?.['tflops']">TFlops: {{ device.capacity['tflops'] }}</div>
                          <div v-if="device.available?.['vram']">VRAM Avail: {{ device.available['vram'] }}</div>
                          <div v-if="device.runningApps?.length">
                            Apps: {{ device.runningApps.map(a => `${a.namespace}/${a.name}`).join(', ') }}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Server, ChevronDown } from 'lucide-vue-next'
import type { NodeViewSnapshot } from '../types'
import { UsedByTensorFusion, UsedByNvidiaDevicePlugin } from '../types'

const props = defineProps<{
  data: NodeViewSnapshot
}>()

const expandNode = ref<string | null>(null)

// Computed properties for null safety
const summary = computed(() => props.data?.summary ?? { totalNodes: 0, readyNodes: 0, totalGpus: 0, gpusByUsedBy: {} })
const nodes = computed(() => props.data?.nodes ?? [])
const gpusByUsedBy = computed(() => summary.value?.gpusByUsedBy ?? {})

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
</script>
