<template>
  <div class="space-y-4">
    <!-- Migration Mode Banner -->
    <Card :class="data?.progressiveMigration ? 'border-blue-500' : 'border-green-500'">
      <CardContent class="pt-4">
        <div class="flex items-center gap-4">
          <div
            class="w-12 h-12 rounded-full flex items-center justify-center"
            :class="data?.progressiveMigration ? 'bg-blue-100' : 'bg-green-100'"
          >
            <GitMerge class="w-6 h-6" :class="data?.progressiveMigration ? 'text-blue-600' : 'text-green-600'" />
          </div>
          <div>
            <h3 class="font-semibold">
              {{ data?.progressiveMigration ? 'Progressive Migration Mode' : 'Strict Isolation Mode' }}
            </h3>
            <p class="text-sm text-muted-foreground">
              {{ data?.progressiveMigration
                ? 'GPUs can be shared between TensorFusion and native nvidia-device-plugin'
                : 'GPUs are exclusively managed by their assigned driver'
              }}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- GPU Distribution -->
    <div class="grid grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <PieChart class="w-5 h-5" />
            GPU Distribution by Driver
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <div v-for="(count, driver) in usedByStats" :key="driver" class="space-y-2">
              <div class="flex justify-between text-sm">
                <span class="flex items-center gap-2">
                  <span class="w-3 h-3 rounded-full" :class="getDriverColor(driver as string)"></span>
                  {{ formatDriver(driver as string) }}
                </span>
                <span class="font-medium">{{ count }} GPUs</span>
              </div>
              <div class="h-2 bg-muted rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all"
                  :class="getDriverBgColor(driver as string)"
                  :style="{ width: getPercentage(count) + '%' }"
                ></div>
              </div>
            </div>
            <div v-if="Object.keys(usedByStats).length === 0" class="text-center py-4 text-muted-foreground">
              No GPU data available
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <BarChart class="w-5 h-5" />
            Statistics
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-2 gap-4">
            <div class="text-center p-4 bg-muted/50 rounded-lg">
              <div class="text-3xl font-bold text-blue-600">
                {{ usedByStats['tensor-fusion'] || 0 }}
              </div>
              <p class="text-sm text-muted-foreground">TensorFusion GPUs</p>
            </div>
            <div class="text-center p-4 bg-muted/50 rounded-lg">
              <div class="text-3xl font-bold text-green-600">
                {{ usedByStats['nvidia-device-plugin'] || 0 }}
              </div>
              <p class="text-sm text-muted-foreground">Native NVIDIA GPUs</p>
            </div>
            <div class="text-center p-4 bg-muted/50 rounded-lg">
              <div class="text-3xl font-bold text-gray-600">
                {{ usedByStats['unknown'] || 0 }}
              </div>
              <p class="text-sm text-muted-foreground">Unmanaged GPUs</p>
            </div>
            <div class="text-center p-4 bg-muted/50 rounded-lg">
              <div class="text-3xl font-bold" :class="conflicts.length > 0 ? 'text-destructive' : 'text-green-600'">
                {{ conflicts.length }}
              </div>
              <p class="text-sm text-muted-foreground">Conflicts</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Conflicts Table -->
    <Card v-if="conflicts.length > 0">
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-destructive">
          <AlertTriangle class="w-5 h-5" />
          GPU Usage Conflicts
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Node</TableHead>
              <TableHead>Pod</TableHead>
              <TableHead>Expected</TableHead>
              <TableHead>Actual UsedBy</TableHead>
              <TableHead>Resource Type</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="conflict in conflicts" :key="`${conflict.node}-${conflict.pod?.namespace}/${conflict.pod?.name}`">
              <TableCell class="font-medium">{{ conflict.node }}</TableCell>
              <TableCell>
                <div>{{ conflict.pod?.name }}</div>
                <div class="text-xs text-muted-foreground">{{ conflict.pod?.namespace }}</div>
              </TableCell>
              <TableCell>
                <Badge variant="outline">
                  {{ getExpectedDriver(conflict.resourceType) }}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge variant="destructive">
                  {{ conflict.usedBy }}
                </Badge>
              </TableCell>
              <TableCell>{{ conflict.resourceType }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <!-- No Conflicts -->
    <Card v-else>
      <CardContent class="py-8">
        <div class="text-center">
          <CheckCircle class="w-12 h-12 mx-auto mb-4 text-green-500" />
          <h3 class="font-semibold text-lg">No Conflicts Detected</h3>
          <p class="text-muted-foreground">All GPU assignments are consistent with their resource types</p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { GitMerge, PieChart, BarChart, AlertTriangle, CheckCircle } from 'lucide-vue-next'
import type { MigrationSnapshot } from '../types'
import { UsedByTensorFusion, UsedByNvidiaDevicePlugin } from '../types'

const props = defineProps<{
  data: MigrationSnapshot
}>()

// Computed properties for null safety
const usedByStats = computed(() => props.data?.usedByStats ?? {})
const conflicts = computed(() => props.data?.conflicts ?? [])

const totalGPUs = computed(() => {
  return Object.values(usedByStats.value).reduce((sum, count) => sum + count, 0)
})

function getPercentage(count: number): number {
  if (totalGPUs.value === 0) return 0
  return (count / totalGPUs.value) * 100
}

function formatDriver(driver: string): string {
  switch (driver) {
    case UsedByTensorFusion:
      return 'TensorFusion'
    case UsedByNvidiaDevicePlugin:
      return 'NVIDIA Device Plugin'
    default:
      return driver || 'Unknown'
  }
}

function getDriverColor(driver: string): string {
  switch (driver) {
    case UsedByTensorFusion:
      return 'bg-blue-500'
    case UsedByNvidiaDevicePlugin:
      return 'bg-green-500'
    default:
      return 'bg-gray-500'
  }
}

function getDriverBgColor(driver: string): string {
  switch (driver) {
    case UsedByTensorFusion:
      return 'bg-blue-500'
    case UsedByNvidiaDevicePlugin:
      return 'bg-green-500'
    default:
      return 'bg-gray-400'
  }
}

function getExpectedDriver(resourceType: string): string {
  switch (resourceType) {
    case 'tensor-fusion':
    case 'dra':
      return 'tensor-fusion'
    case 'nvidia.com/gpu':
      return 'nvidia-device-plugin'
    default:
      return 'unknown'
  }
}
</script>
