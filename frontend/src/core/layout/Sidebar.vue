<template>
  <div class="bg-muted/30 border-r shadow-sm flex flex-col items-stretch py-6 px-2 gap-4">
    <router-link
      v-for="item in sidebarItems"
      :key="item.id"
      :to="`/plugins/${item.pluginId}`"
      class="flex flex-col items-center justify-center gap-1 rounded-xl px-2 py-3.5 text-[12px] font-medium text-muted-foreground hover:bg-muted/70 hover:text-foreground transition-colors relative border border-transparent"
      :class="{
        'bg-primary/10 text-foreground border-primary/20 shadow-sm': isActive(item.pluginId)
      }"
      :title="item.title"
      :aria-label="item.title"
    >
      <component :is="getIcon(item.icon)" class="w-6 h-6" />
      <span class="text-center leading-tight">{{ item.title }}</span>
      <div
        v-if="item.badge"
        class="absolute top-1 right-2 bg-red-500 text-white rounded-full w-4 h-4 flex items-center justify-center text-[10px]"
      >
        {{ item.badge }}
      </div>
    </router-link>

  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import type { SidebarItem } from '@/core/types'
import {
  Layers,
  Server,
  Cpu
} from 'lucide-vue-next'

const route = useRoute()

// Icon mapping
const iconMap: Record<string, any> = {
  'layers': Layers,
  'server': Server,
  'cpu': Cpu
}

const getIcon = (iconName: string) => {
  return iconMap[iconName] || Box
}

const isActive = (pluginId: string) => {
  return route.path === `/plugins/${pluginId}`
}

// Mock sidebar items - will be loaded from plugin registry
const sidebarItems = ref<SidebarItem[]>([
  { id: 'apps', title: 'Applications', icon: 'layers', order: 1, pluginId: 'applications' },
  { id: 'nodes', title: 'Nodes', icon: 'server', order: 2, pluginId: 'nodes' },
  { id: 'scheduling', title: 'GPU Scheduling', icon: 'cpu', order: 3, pluginId: 'scheduling' }
])
</script>
