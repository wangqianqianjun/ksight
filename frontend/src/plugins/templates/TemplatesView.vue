<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <FileText class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Templates</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled>
            <Filter class="w-4 h-4 mr-2" />
            Filters
          </Button>
          <Button size="sm">
            <Plus class="w-4 h-4 mr-2" />
            New Template
          </Button>
        </div>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-auto">
      <!-- Empty State -->
      <div class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Inbox class="w-8 h-8 text-muted-foreground" />
          </div>
          <h2 class="text-xl font-semibold mb-2">No templates yet</h2>
          <p class="text-muted-foreground mb-6">
            Create reusable YAML templates with parameters and versioning for quick deployments.
          </p>
          <div class="flex items-center justify-center gap-3">
            <Button>
              <Plus class="w-4 h-4 mr-2" />
              Create template
            </Button>
            <Button variant="outline">
              <Download class="w-4 h-4 mr-2" />
              Import
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { FileText, Filter, Plus, Inbox, Download } from 'lucide-vue-next'
</script>
