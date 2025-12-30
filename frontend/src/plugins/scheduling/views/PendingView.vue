<template>
  <div class="space-y-4">
    <!-- Summary -->
    <div class="grid grid-cols-3 gap-4">
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

    <!-- Filters -->
    <div class="flex items-center gap-4">
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">Scheduler:</span>
        <Select v-model="filterScheduler">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="All schedulers" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All</SelectItem>
            <SelectItem v-for="scheduler in schedulers" :key="scheduler" :value="scheduler">
              {{ scheduler }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">Pool:</span>
        <Select v-model="filterPool">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="All pools" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All</SelectItem>
            <SelectItem v-for="pool in pools" :key="pool" :value="pool">
              {{ pool }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
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
            <TableRow v-for="pod in filteredPods" :key="`${pod.namespace}/${pod.name}`">
              <TableCell>{{ pod.namespace }}</TableCell>
              <TableCell class="font-medium">{{ pod.name }}</TableCell>
              <TableCell>
                <Badge :variant="pod.scheduler === 'tensor-fusion-scheduler' ? 'default' : 'secondary'">
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
            <TableRow v-if="filteredPods.length === 0">
              <TableCell colspan="7" class="text-center py-8 text-muted-foreground">
                <CheckCircle class="w-8 h-8 mx-auto mb-2 text-green-500" />
                No pending pods
              </TableCell>
            </TableRow>
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Clock, CheckCircle } from 'lucide-vue-next'
import type { PendingViewSnapshot } from '../types'

const props = defineProps<{
  data: PendingViewSnapshot
}>()

const filterScheduler = ref('')
const filterPool = ref('')

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
    if (filterScheduler.value && pod.scheduler !== filterScheduler.value) return false
    if (filterPool.value && pod.pool !== filterPool.value) return false
    return true
  })
})

function getSinceClass(since?: string): string {
  if (!since) return ''
  // Highlight pods pending for a long time
  if (since.includes('h') || since.includes('d')) {
    return 'text-destructive font-medium'
  }
  return ''
}
</script>
