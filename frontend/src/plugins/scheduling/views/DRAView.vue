<template>
  <div class="space-y-4">
    <!-- Stats Summary -->
    <div class="grid grid-cols-4 gap-4">
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ claims.length }}</div>
          <p class="text-sm text-muted-foreground">Total Claims</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold text-green-600">{{ data?.stats?.allocated || 0 }}</div>
          <p class="text-sm text-muted-foreground">Allocated</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold text-yellow-600">{{ data?.stats?.pending || 0 }}</div>
          <p class="text-sm text-muted-foreground">Pending</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ data?.stats?.totalDevices || 0 }}</div>
          <p class="text-sm text-muted-foreground">Total Devices</p>
        </CardContent>
      </Card>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <!-- Resource Claims -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <FileText class="w-5 h-5" />
            Resource Claims
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Driver</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Pod</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="claim in claims" :key="`${claim.namespace}/${claim.name}`">
                <TableCell>
                  <div class="font-medium">{{ claim.name }}</div>
                  <div class="text-xs text-muted-foreground">{{ claim.namespace }}</div>
                </TableCell>
                <TableCell>
                  <Badge v-if="claim.driver" variant="outline">
                    {{ formatDriver(claim.driver) }}
                  </Badge>
                  <span v-else class="text-muted-foreground">-</span>
                </TableCell>
                <TableCell>
                  <Badge :variant="claim.allocation ? 'default' : 'secondary'">
                    {{ claim.allocation ? 'Allocated' : 'Pending' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <span v-if="claim.podRef">{{ claim.podRef.name }}</span>
                  <span v-else class="text-muted-foreground">-</span>
                </TableCell>
              </TableRow>
              <TableRow v-if="claims.length === 0">
                <TableCell colspan="4" class="text-center py-8 text-muted-foreground">
                  No resource claims found
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <!-- Resource Slices -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Layers class="w-5 h-5" />
            Resource Slices
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Node</TableHead>
                <TableHead>Driver</TableHead>
                <TableHead>Devices</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="slice in slices" :key="slice.name">
                <TableCell class="font-medium">{{ slice.name }}</TableCell>
                <TableCell>{{ slice.nodeName || '-' }}</TableCell>
                <TableCell>
                  <Badge variant="outline">
                    {{ formatDriver(slice.driver) }}
                  </Badge>
                </TableCell>
                <TableCell>{{ slice.devices }}</TableCell>
              </TableRow>
              <TableRow v-if="slices.length === 0">
                <TableCell colspan="4" class="text-center py-8 text-muted-foreground">
                  No resource slices found
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>

    <!-- DRA Flow Visualization -->
    <Card v-if="claims.length > 0">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <GitBranch class="w-5 h-5" />
          DRA Allocation Flow
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div class="flex items-center justify-center gap-8 py-4">
          <div class="text-center">
            <div class="w-16 h-16 rounded-full bg-blue-100 flex items-center justify-center mx-auto mb-2">
              <FileText class="w-8 h-8 text-blue-600" />
            </div>
            <div class="font-medium">Claims</div>
            <div class="text-2xl">{{ claims.length }}</div>
          </div>
          <ArrowRight class="w-8 h-8 text-muted-foreground" />
          <div class="text-center">
            <div class="w-16 h-16 rounded-full bg-purple-100 flex items-center justify-center mx-auto mb-2">
              <Layers class="w-8 h-8 text-purple-600" />
            </div>
            <div class="font-medium">Slices</div>
            <div class="text-2xl">{{ slices.length }}</div>
          </div>
          <ArrowRight class="w-8 h-8 text-muted-foreground" />
          <div class="text-center">
            <div class="w-16 h-16 rounded-full bg-green-100 flex items-center justify-center mx-auto mb-2">
              <Cpu class="w-8 h-8 text-green-600" />
            </div>
            <div class="font-medium">Devices</div>
            <div class="text-2xl">{{ data?.stats?.totalDevices || 0 }}</div>
          </div>
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
import { FileText, Layers, GitBranch, ArrowRight, Cpu } from 'lucide-vue-next'
import type { DRASnapshot } from '../types'

const props = defineProps<{
  data: DRASnapshot
}>()

// Computed properties to safely access arrays
const claims = computed(() => props.data?.claims ?? [])
const slices = computed(() => props.data?.slices ?? [])

function formatDriver(driver: string): string {
  // Shorten driver name for display
  if (driver?.includes('tensor-fusion')) return 'TensorFusion'
  if (driver?.includes('nvidia')) return 'NVIDIA'
  return driver?.split('.')[0] || driver || '-'
}
</script>
