<template>
  <div class="space-y-4">
    <!-- Summary -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card>
        <CardContent class="pt-4">
          <div class="text-2xl font-bold">{{ pods.length }}</div>
          <p class="text-sm text-muted-foreground">Total Pending</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="space-y-1">
            <div v-for="(count, scheduler) in byScheduler" :key="scheduler" class="flex justify-between text-sm">
              <span class="text-muted-foreground">{{ scheduler }}</span>
              <span class="font-medium">{{ count }}</span>
            </div>
          </div>
          <p class="text-sm text-muted-foreground mt-2">By Scheduler</p>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="pt-4">
          <div class="space-y-1">
            <div v-for="reason in reasons.slice(0, 3)" :key="reason.reason" class="flex justify-between text-sm">
              <span class="text-muted-foreground truncate max-w-[150px]" :title="reason.reason">
                {{ reason.reason?.split(':')[0] || '-' }}
              </span>
              <span class="font-medium">{{ reason.count }}</span>
            </div>
          </div>
          <p class="text-sm text-muted-foreground mt-2">Top Reasons</p>
        </CardContent>
      </Card>
    </div>

    <Alert v-if="showReasonBanner" class="border-amber-200 bg-amber-50 text-amber-900">
      <AlertTriangle class="w-4 h-4" />
      <AlertTitle>Pending reasons unavailable</AlertTitle>
      <AlertDescription>
        Events API is unavailable for this cluster, so scheduling reasons may be missing.
      </AlertDescription>
    </Alert>

    <!-- Filters -->
    <div class="flex flex-wrap items-center gap-4">
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">Scheduler:</span>
        <Select v-model="filterScheduler">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="All schedulers" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem :value="allToken">All</SelectItem>
            <SelectItem v-for="scheduler in schedulers" :key="scheduler" :value="scheduler">
              {{ scheduler }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">Namespace:</span>
        <Input v-model="filterNamespaceInput" class="w-[180px]" placeholder="Filter namespace" />
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">Pool:</span>
        <Select v-model="filterPool">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="All pools" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem :value="allToken">All</SelectItem>
            <SelectItem v-for="pool in pools" :key="pool" :value="pool">
              {{ pool }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button v-if="hasFilters" variant="ghost" size="sm" @click="clearFilters">
        Clear filters
      </Button>
    </div>

    <!-- Reason Chips -->
    <div class="flex flex-wrap items-center gap-2">
      <Button
        size="sm"
        :variant="activeReason ? 'outline' : 'secondary'"
        @click="activeReason = ''"
      >
        All reasons
      </Button>
      <Button
        v-for="reason in reasons"
        :key="reason.reason"
        size="sm"
        :variant="activeReason === reason.reason ? 'secondary' : 'outline'"
        @click="activeReason = reason.reason"
        :title="reason.reason"
      >
        {{ reason.reason }}
      </Button>
    </div>

    <!-- Pending Pods Table -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Clock class="w-5 h-5" />
          Pending Pods
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Namespace</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Scheduler</TableHead>
              <TableHead>GPU Request</TableHead>
              <TableHead>Pool</TableHead>
              <TableHead>Pending Since</TableHead>
              <TableHead>Reason</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <template v-for="group in groupedPods" :key="group.scheduler">
              <TableRow class="bg-muted/40">
                <TableCell colspan="7" class="font-medium">
                  <div class="flex items-center gap-2">
                    <Badge variant="outline" :class="getSchedulerBadgeClass(group.scheduler)">
                      {{ group.scheduler }}
                    </Badge>
                    <span class="text-sm text-muted-foreground">{{ group.count }} pending</span>
                  </div>
                </TableCell>
              </TableRow>
              <TableRow
                v-for="pod in group.pods"
                :key="`${pod.namespace}/${pod.name}`"
                class="cursor-pointer"
                @click="openPodDetails(pod)"
              >
                <TableCell>{{ pod.namespace }}</TableCell>
                <TableCell class="font-medium">{{ pod.name }}</TableCell>
                <TableCell>
                  <Badge variant="outline" :class="getSchedulerBadgeClass(pod.scheduler || 'default-scheduler')">
                    {{ pod.scheduler || 'default-scheduler' }}
                  </Badge>
                </TableCell>
                <TableCell>{{ pod.gpuRequest || '-' }}</TableCell>
                <TableCell>{{ pod.pool || '-' }}</TableCell>
                <TableCell>
                  <span :class="getSinceClass(pod.since)">{{ pod.since || 'Just now' }}</span>
                </TableCell>
                <TableCell class="max-w-[300px]">
                  <span
                    v-if="pod.reason"
                    class="text-sm text-destructive truncate block"
                    :title="pod.reason"
                  >
                    {{ pod.reason }}
                  </span>
                  <span v-else class="text-muted-foreground">-</span>
                </TableCell>
              </TableRow>
            </template>
            <TableRow v-if="filteredPods.length === 0">
              <TableCell colspan="7" class="text-center py-8 text-muted-foreground">
                <div v-if="pods.length === 0">
                  <CheckCircle class="w-8 h-8 mx-auto mb-2 text-green-500" />
                  Scheduler healthy. Last updated {{ formatTimestamp(lastUpdated) }}.
                </div>
                <div v-else>
                  No results match your filters.
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Sheet :open="showDetails" @update:open="showDetails = $event">
      <SheetContent side="right" class="w-[380px] sm:w-[520px]">
        <SheetHeader>
          <SheetTitle>Pod Details</SheetTitle>
          <SheetDescription>
            Pending pod and scheduler context.
          </SheetDescription>
        </SheetHeader>
        <div v-if="selectedPod" class="mt-4 space-y-3 text-sm">
          <div>
            <div class="text-muted-foreground">Pod</div>
            <div class="font-medium">{{ selectedPod.namespace }}/{{ selectedPod.name }}</div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-muted-foreground">Scheduler</div>
              <div>{{ selectedPod.scheduler || 'default-scheduler' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">Pool</div>
              <div>{{ selectedPod.pool || '-' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">GPU Request</div>
              <div>{{ selectedPod.gpuRequest || '-' }}</div>
            </div>
            <div>
              <div class="text-muted-foreground">Pending Since</div>
              <div>{{ selectedPod.since || '-' }}</div>
            </div>
          </div>
          <div>
            <div class="text-muted-foreground">Reason</div>
            <div>{{ selectedPod.reason || 'No reason reported' }}</div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Clock, CheckCircle, AlertTriangle } from 'lucide-vue-next'
import type { PendingPodRow, PendingViewSnapshot } from '../types'

const props = defineProps<{
  data: PendingViewSnapshot
  lastUpdated?: string
}>()

const allToken = '__all__'
const filterScheduler = ref(allToken)
const filterPool = ref(allToken)
const filterNamespaceInput = ref('')
const filterNamespace = ref('')
const activeReason = ref('')
const selectedPod = ref<PendingPodRow | null>(null)
const showDetails = ref(false)
let filterTimer: number | undefined

// Computed properties for null safety
const pods = computed(() => props.data?.pods ?? [])
const byScheduler = computed(() => props.data?.byScheduler ?? {})
const reasons = computed(() => props.data?.reasons ?? [])

const schedulers = computed(() => {
  const set = new Set<string>()
  pods.value.forEach(pod => {
    if (pod.scheduler) set.add(pod.scheduler)
  })
  return Array.from(set)
})

const pools = computed(() => {
  const set = new Set<string>()
  pods.value.forEach(pod => {
    if (pod.pool) set.add(pod.pool)
  })
  return Array.from(set)
})

const filteredPods = computed(() => {
  return pods.value.filter(pod => {
    if (filterScheduler.value !== allToken && pod.scheduler !== filterScheduler.value) return false
    if (filterPool.value !== allToken && pod.pool !== filterPool.value) return false
    if (filterNamespace.value && !pod.namespace.toLowerCase().includes(filterNamespace.value)) return false
    if (activeReason.value && !(pod.reason || '').includes(activeReason.value)) return false
    return true
  })
})

const groupedPods = computed(() => {
  const groups = new Map<string, PendingPodRow[]>()
  filteredPods.value.forEach(pod => {
    const scheduler = pod.scheduler || 'default-scheduler'
    if (!groups.has(scheduler)) {
      groups.set(scheduler, [])
    }
    groups.get(scheduler)!.push(pod)
  })
  return Array.from(groups.entries()).map(([scheduler, pods]) => ({
    scheduler,
    pods,
    count: pods.length
  }))
})

const showReasonBanner = computed(() => {
  if (props.data?.eventsAvailable === false) return true
  return pods.value.length > 0 && reasons.value.length === 0
})

const hasFilters = computed(() => {
  return !!(filterScheduler.value !== allToken || filterPool.value !== allToken || filterNamespaceInput.value || activeReason.value)
})

watch(filterNamespaceInput, (value) => {
  if (filterTimer) {
    window.clearTimeout(filterTimer)
  }
  filterTimer = window.setTimeout(() => {
    filterNamespace.value = value.trim().toLowerCase()
  }, 250)
})

function clearFilters() {
  filterScheduler.value = allToken
  filterPool.value = allToken
  filterNamespaceInput.value = ''
  filterNamespace.value = ''
  activeReason.value = ''
}

function getSinceClass(since?: string): string {
  if (!since) return ''
  // Highlight pods pending for a long time
  if (since.includes('h') || since.includes('d')) {
    return 'text-destructive font-medium'
  }
  return ''
}

function getSchedulerBadgeClass(scheduler: string): string {
  if (scheduler.includes('tensor-fusion')) return 'border-blue-200 bg-blue-100 text-blue-700'
  if (scheduler.includes('default')) return 'border-teal-200 bg-teal-100 text-teal-700'
  if (scheduler.includes('volcano')) return 'border-amber-200 bg-amber-100 text-amber-700'
  return 'border-gray-200 bg-gray-100 text-gray-700'
}

function openPodDetails(pod: PendingPodRow) {
  selectedPod.value = pod
  showDetails.value = true
}

function formatTimestamp(value?: string): string {
  if (!value) return 'unknown'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString()
}
</script>
