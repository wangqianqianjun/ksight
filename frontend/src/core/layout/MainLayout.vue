<template>
  <div class="h-screen flex bg-background">
    <!-- Left: Sidebar (full height) -->
    <Sidebar class="w-20 flex-shrink-0" />

    <!-- Right: Title Bar + Main Content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Custom Title Bar (includes Cluster Tabs) -->
      <TitleBar />

      <!-- Main Content Area -->
      <div class="flex-1 flex overflow-hidden">
        <div class="flex-1 flex flex-col overflow-hidden px-6 pt-2">
          <router-view />
        </div>

        <!-- Right: AI Chat (collapsible) -->
        <ChatPanel v-if="showChat" class="w-80 flex-shrink-0" />
      </div>

      <!-- Bottom: Command Tabs -->
      <CommandTabs v-if="showCommandTabs" class="h-64" />
    </div>

    <!-- Cluster Connection Dialog -->
    <ClusterConnectionDialog />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import TitleBar from './TitleBar.vue'
import Sidebar from './Sidebar.vue'
import ChatPanel from './ChatPanel.vue'
import CommandTabs from './CommandTabs.vue'
import ClusterConnectionDialog from '@/shared/components/ClusterConnectionDialog.vue'

// Layout state
const showChat = ref(false)
const showCommandTabs = ref(false)
</script>
