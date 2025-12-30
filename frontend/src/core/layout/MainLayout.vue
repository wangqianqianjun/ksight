<template>
  <div class="h-screen flex flex-col bg-background">
    <!-- Custom Title Bar (includes Cluster Tabs) -->
    <TitleBar />

    <div class="flex-1 flex overflow-hidden">
      <!-- Left: Sidebar -->
      <Sidebar class="w-12" />

      <!-- Main Content Area -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <router-view />
      </div>

      <!-- Right: AI Chat (collapsible) -->
      <ChatPanel v-if="showChat" class="w-80" />
    </div>

    <!-- Bottom: Command Tabs -->
    <CommandTabs v-if="showCommandTabs" class="h-64" />

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