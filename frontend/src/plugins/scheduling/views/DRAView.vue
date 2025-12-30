<template>
  <div class="space-y-4">
    <Card v-if="isNotAvailable">
      <CardContent class="py-8 text-center">
        <div class="w-12 h-12 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
          <AlertTriangle class="w-6 h-6 text-amber-600" />
        </div>
        <h3 class="font-semibold">DRA not available</h3>
        <p class="text-muted-foreground text-sm mt-1 max-w-md mx-auto">
          {{ data?.message || 'Dynamic Resource Allocation (DRA) APIs are not available on this cluster. DRA requires Kubernetes 1.26+ and the DynamicResourceAllocation feature gate to be enabled.' }}
        </p>
        <p class="text-xs text-muted-foreground mt-3">
          No ResourceClaims or ResourceSlices were found.
        </p>
      </CardContent>
    </Card>

    <template v-else>
      <!-- Stats Summary -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent class="pt-4">
            <div class="text-2xl font-bold text-amber-600">{{ stats.unallocated }}</div>
            <p class="text-sm text-muted-foreground">Unallocated</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-4">
            <div class="text-2xl font-bold text-blue-600">{{ stats.allocating }}</div>
            <p class="text-sm text-muted-foreground">Allocating</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-4">
            <div class="text-2xl font-bold text-green-600">{{ stats.allocated }}</div>
            <p class="text-sm text-muted-foreground">Allocated</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-4">
            <div class="text-2xl font-bold">{{ stats.totalDevices }}</div>
            <p class="text-sm text-muted-foreground">Total Devices</p>
          </CardContent>
        </Card>
      </div>

      <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
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
                  <TableHead>Class</TableHead>
                  <TableHead>Driver</TableHead>
                  <TableHead>Allocation</TableHead>
                  <TableHead>Pod</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="claim in claims"
                  :key="`${claim.namespace}/${claim.name}`"
                  class="cursor-pointer"
                  :class="selectedClaimKey === `${claim.namespace}/${claim.name}` ? 'bg-muted/50' : ''"
                  @click="selectedClaim = claim"
                >
                  <TableCell>
                    <div class="font-medium">{{ claim.name }}</div>
                    <div class="text-xs text-muted-foreground">{{ claim.namespace }}</div>
                  </TableCell>
                  <TableCell>{{ claim.resourceClass || '-' }}</TableCell>
                  <TableCell>
                    <Badge v-if="claim.driver" variant="outline">
                      {{ formatDriver(claim.driver) }}
                    </Badge>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell class="max-w-[180px]">
                    <span class="truncate block" :title="formatAllocation(claim.allocation)">
                      {{ formatAllocation(claim.allocation) }}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span v-if="claim.podRef">{{ claim.podRef.name }}</span>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                </TableRow>
                <TableRow v-if="claims.length === 0">
                  <TableCell colspan="5" class="text-center py-8 text-muted-foreground">
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
              <div class="text-2xl">{{ stats.totalDevices }}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Selected Claim Chain -->
      <Card v-if="selectedClaim">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Link class="w-5 h-5" />
            Claim Chain: {{ selectedClaim.name }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="rounded-md border p-3">
              <div class="text-xs text-muted-foreground">Claim</div>
              <div class="font-medium">{{ selectedClaim.namespace }}/{{ selectedClaim.name }}</div>
              <div class="text-sm text-muted-foreground">Class: {{ selectedClaim.resourceClass || '-' }}</div>
              <div class="text-sm text-muted-foreground">Driver: {{ formatDriver(selectedClaim.driver || '-') }}</div>
            </div>
            <div class="rounded-md border p-3">
              <div class="text-xs text-muted-foreground">Slices</div>
              <div class="space-y-1">
                <div v-for="slice in linkedSlices" :key="slice" class="text-sm">
                  {{ slice }}
                </div>
                <div v-if="linkedSlices.length === 0" class="text-sm text-muted-foreground">
                  No slice mapping
                </div>
              </div>
            </div>
            <div class="rounded-md border p-3">
              <div class="text-xs text-muted-foreground">Devices</div>
              <div class="space-y-1">
                <div v-for="device in linkedDevices" :key="device" class="text-sm">
                  {{ device }}
                </div>
                <div v-if="linkedDevices.length === 0" class="text-sm text-muted-foreground">
                  No device refs
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { FileText, Layers, GitBranch, ArrowRight, Cpu, AlertTriangle, Link } from 'lucide-vue-next'
import type { ClaimRow, DRASnapshot } from '../types'

const props = defineProps<{
  data: DRASnapshot
}>()

const selectedClaim = ref<ClaimRow | null>(null)

// DRA is not available if explicitly marked or if all data is empty/zero
const isNotAvailable = computed(() => {
  if (props.data?.status === 'not_available') return true
  // Also show not available if no claims, no slices, and all stats are zero
  const hasClaims = (props.data?.claims?.length ?? 0) > 0
  const hasSlices = (props.data?.slices?.length ?? 0) > 0
  const hasDevices = (props.data?.stats?.totalDevices ?? 0) > 0
  return !hasClaims && !hasSlices && !hasDevices && !props.data?.status
})
const claims = computed(() => props.data?.claims ?? [])
const slices = computed(() => props.data?.slices ?? [])
const stats = computed(() => ({
  unallocated: props.data?.stats?.unallocated ?? 0,
  allocating: props.data?.stats?.allocating ?? 0,
  allocated: props.data?.stats?.allocated ?? 0,
  totalDevices: props.data?.stats?.totalDevices ?? 0
}))

const selectedClaimKey = computed(() => {
  if (!selectedClaim.value) return ''
  return `${selectedClaim.value.namespace}/${selectedClaim.value.name}`
})

const linkedSlices = computed(() => {
  if (!selectedClaim.value) return []
  return extractAllocationSlices(selectedClaim.value.allocation)
})

const linkedDevices = computed(() => {
  if (!selectedClaim.value) return []
  return extractAllocationDevices(selectedClaim.value.allocation)
})

function formatDriver(driver: string): string {
  if (!driver) return '-'
  if (driver.includes('tensor-fusion')) return 'TensorFusion'
  if (driver.includes('nvidia')) return 'NVIDIA'
  return driver.split('.')[0] || driver
}

function formatAllocation(allocation?: Record<string, unknown>): string {
  if (!allocation) return 'Pending'
  const slices = extractAllocationSlices(allocation)
  const devices = extractAllocationDevices(allocation)
  const details: string[] = []
  if (slices.length) details.push(`Slices ${slices.join(', ')}`)
  if (devices.length) details.push(`Devices ${devices.join(', ')}`)
  if (!details.length) return 'Allocated'
  return details.join(' | ')
}

function extractAllocationSlices(allocation?: Record<string, unknown>): string[] {
  if (!allocation || typeof allocation !== 'object') return []
  const candidates = ['slice', 'slices', 'resourceSlices']
  for (const key of candidates) {
    const value = allocation[key as keyof typeof allocation]
    const extracted = extractStringList(value)
    if (extracted.length) return extracted
  }
  return []
}

function extractAllocationDevices(allocation?: Record<string, unknown>): string[] {
  if (!allocation || typeof allocation !== 'object') return []
  const candidates = ['devices', 'deviceRefs', 'deviceIds']
  for (const key of candidates) {
    const value = allocation[key as keyof typeof allocation]
    const extracted = extractStringList(value)
    if (extracted.length) return extracted
  }
  return []
}

function extractStringList(value: unknown): string[] {
  if (!value) return []
  if (Array.isArray(value)) {
    return value
      .map(item => {
        if (typeof item === 'string') return item
        if (typeof item === 'object' && item) {
          const name = (item as { name?: string }).name
          const device = (item as { device?: string }).device
          return name || device || ''
        }
        return ''
      })
      .filter(Boolean)
  }
  if (typeof value === 'string') {
    return [value]
  }
  return []
}
</script>
