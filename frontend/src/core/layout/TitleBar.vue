<template>
  <div
    class="title-bar flex items-center bg-gray-50/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 h-9"
    style="--wails-draggable: drag"
    @dblclick="handleTitlebarDoubleClick"
  >
    <!-- Cluster Tabs Container -->
    <div class="flex-1 flex items-center h-full min-w-0" :class="isMac ? 'pl-20' : 'pl-4'">
      <ClusterTabs 
        :tabs="tabs" 
        @set-active-tab="setActiveTab"
        @close-tab="closeTab"
        @add-tab="addTab"
      />
    </div>

    <!-- Right Side Actions -->
    <div class="flex items-center h-full pr-4 space-x-2" style="--wails-draggable: no-drag" data-no-drag>
      <!-- Theme Toggle -->
      <ThemeToggle />
      
      <!-- Settings -->
      <button
        @click="openSettings"
        class="settings-btn w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-all duration-200 group"
        title="Settings"
      >
        <Settings :size="16" class="transition-transform duration-200 group-hover:rotate-45" />
      </button>

      <!-- Window Controls -->
      <div v-if="!isMac" class="flex items-center pl-2 border-l border-gray-200/60 dark:border-gray-700/60">
        <button
          @click="minimiseWindow"
          class="w-8 h-8 flex items-center justify-center rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
          title="Minimize"
        >
          <Minus :size="14" />
        </button>
        <button
          @click="toggleMaximise"
          class="w-8 h-8 flex items-center justify-center rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
          title="Maximize or restore"
        >
          <Square :size="12" />
        </button>
        <button
          @click="closeWindow"
          class="w-8 h-8 flex items-center justify-center rounded-md hover:bg-red-500/15 text-gray-500 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
          title="Close"
        >
          <X :size="14" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Settings, Minus, Square, X } from 'lucide-vue-next'
import ClusterTabs from '@/app/frame/ClusterTabs.vue'
import ThemeToggle from '@/app/frame/ThemeToggle.vue'
import type { Tab } from '@/app/frame/ClusterTabs.vue'
import { useClusterDialog } from '@/shared/composables/useClusterDialog'
import { Environment, WindowMinimise, WindowToggleMaximise, Quit } from '@/wailsjs/runtime/runtime'

const { open: openClusterDialog } = useClusterDialog()
const isMac = ref(false)

onMounted(async () => {
  try {
    const env = await Environment()
    isMac.value = env.platform === 'darwin'
  } catch {
    isMac.value = false
  }
})

// Default icon component for tabs
const DefaultIcon = {
  template: `
    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  `
}

// State
const tabs = ref<Tab[]>([
  {
    id: '1',
    title: 'local-cluster',
    icon: DefaultIcon,
    active: true,
    connected: true,
    pinned: false,
    route: '/dashboard'
  }
])

// Tab management methods
const setActiveTab = (tabId: string) => {
  tabs.value.forEach(tab => {
    tab.active = tab.id === tabId
  })
}

const closeTab = (tabId: string) => {
  const tabIndex = tabs.value.findIndex(tab => tab.id === tabId)
  if (tabIndex === -1) return
  
  const wasActive = tabs.value[tabIndex].active
  tabs.value.splice(tabIndex, 1)
  
  // If we closed the active tab, activate another one
  if (wasActive && tabs.value.length > 0) {
    const newActiveIndex = Math.min(tabIndex, tabs.value.length - 1)
    tabs.value[newActiveIndex].active = true
  }
}

const addTab = () => {
  // Open cluster connection dialog instead of adding empty tab
  openClusterDialog()
}

const minimiseWindow = () => {
  WindowMinimise()
}

const toggleMaximise = () => {
  WindowToggleMaximise()
}

const handleTitlebarDoubleClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null
  if (!target) {
    toggleMaximise()
    return
  }

  if (target.closest('button, a, input, select, textarea, [data-no-drag]')) {
    return
  }

  toggleMaximise()
}

const closeWindow = () => {
  Quit()
}

// Settings handler
const openSettings = () => {
  // TODO: Implement settings dialog
  console.log('Open settings')
}

// Expose methods for parent components
defineExpose({
  addTab,
  setActiveTab,
  closeTab,
  tabs
})
</script>

<style scoped>
.title-bar {
  user-select: none;
}

.settings-btn {
  position: relative;
}

.settings-btn::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 10px;
  background: linear-gradient(45deg, transparent, rgba(59, 130, 246, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.settings-btn:hover::before {
  opacity: 1;
}
</style>
