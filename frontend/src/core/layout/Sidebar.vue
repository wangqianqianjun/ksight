<template>
  <div class="bg-muted/30 border-r flex flex-col items-center py-2 gap-1">
    <router-link
      v-for="item in sidebarItems"
      :key="item.id"
      :to="`/plugins/${item.pluginId}`"
      class="p-2 rounded-md hover:bg-muted transition-colors relative"
      :class="{ 'bg-primary text-primary-foreground': isActive(item.pluginId) }"
      :title="item.title"
    >
      <component :is="getIcon(item.icon)" class="w-5 h-5" />
      <div
        v-if="item.badge"
        class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-4 h-4 flex items-center justify-center text-xs"
      >
        {{ item.badge }}
      </div>
    </router-link>

    <!-- Command Palette -->
    <div class="mt-auto">
      <Button
        variant="ghost"
        size="icon"
        class="h-8 w-8"
        @click="openCommandPalette"
        title="Command Palette (Cmd+T)"
      >
        <Command class="w-4 h-4" />
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import type { SidebarItem } from '@/core/types'
import { Button } from '@/components/ui/button'
import {
  Layers,
  Activity,
  Server,
  Cpu,
  Box,
  FileText,
  LayoutDashboard,
  Command
} from 'lucide-vue-next'

const route = useRoute()

// Icon mapping
const iconMap: Record<string, any> = {
  'layers': Layers,
  'activity': Activity,
  'server': Server,
  'cpu': Cpu,
  'box': Box,
  'file-text': FileText,
  'chart': LayoutDashboard
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
  { id: 'ops', title: 'Operations', icon: 'activity', order: 2, pluginId: 'operations' },
  { id: 'nodes', title: 'Nodes', icon: 'server', order: 3, pluginId: 'nodes' },
  { id: 'scheduling', title: 'GPU Scheduling', icon: 'cpu', order: 4, pluginId: 'scheduling' },
  { id: 'resources', title: 'Resources', icon: 'box', order: 5, pluginId: 'resources' },
  { id: 'templates', title: 'Templates', icon: 'file-text', order: 6, pluginId: 'templates' },
  { id: 'boards', title: 'Custom Boards', icon: 'chart', order: 7, pluginId: 'boards' }
])

const openCommandPalette = () => {
  // Will implement command palette
  console.log('Open command palette')
}
</script>