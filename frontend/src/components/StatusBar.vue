<script setup>
import { computed } from 'vue'
import { Download, Activity, HardDrive } from 'lucide-vue-next'

import { useDownloadStore } from '../stores/downloads.js'
import { useQueueStore } from '../stores/queues.js'
import { useSettingsStore } from '../stores/settings.js'
import { formatters } from '../utils/api.js'

const downloadStore = useDownloadStore()
const queueStore = useQueueStore()
const settingsStore = useSettingsStore()

const totalActiveDownloads = computed(() => {
  return downloadStore.stats.active_downloads || 0
})

const totalSpeed = computed(() => {
  return downloadStore.stats.total_speed || 0
})

const totalDownloads = computed(() => {
  return downloadStore.stats.total_downloads || 0
})

const completedDownloads = computed(() => {
  return downloadStore.stats.completed_downloads || 0
})

const failedDownloads = computed(() => {
  return downloadStore.stats.failed_downloads || 0
})

const speedLimitStatus = computed(() => {
  if (settingsStore.settings.global_speed_limit === 0) {
    return 'Unlimited'
  }
  return formatters.formatSpeed(settingsStore.settings.global_speed_limit)
})

const queueStatus = computed(() => {
  const activeQueues = queueStore.activeQueues.length
  const totalQueues = queueStore.queues.length
  return `${activeQueues}/${totalQueues} queues active`
})
</script>

<template>
  <div class="bg-dark-800 px-6 py-2 flex items-center justify-between text-sm text-gray-400">
    <!-- Left Side - Download Stats -->
    <div class="flex items-center space-x-6">
      <!-- Active Downloads -->
      <div class="flex items-center space-x-2">
        <Download class="w-4 h-4 text-blue-500" />
        <span>{{ totalActiveDownloads }} active</span>
      </div>
      
      <!-- Total Speed -->
      <div class="flex items-center space-x-2">
        <Activity class="w-4 h-4 text-green-500" />
        <span>{{ formatters.formatSpeed(totalSpeed) }}</span>
      </div>
      
      <!-- Queue Status -->
      <div class="flex items-center space-x-2">
        <HardDrive class="w-4 h-4 text-purple-500" />
        <span>{{ queueStatus }}</span>
      </div>
    </div>
    
    <!-- Right Side - System Stats -->
    <div class="flex items-center space-x-6">
      <!-- Download Summary -->
      <div class="flex items-center space-x-4">
        <span class="text-green-500">{{ completedDownloads }} completed</span>
        <span v-if="failedDownloads > 0" class="text-red-500">{{ failedDownloads }} failed</span>
        <span class="text-gray-500">{{ totalDownloads }} total</span>
      </div>
      
      <!-- Speed Limit -->
      <div class="flex items-center space-x-2">
        <span>Speed limit:</span>
        <span :class="settingsStore.isSpeedLimited ? 'text-orange-500' : 'text-green-500'">
          {{ speedLimitStatus }}
        </span>
      </div>
      
      <!-- Version -->
      <div class="text-gray-500">
        v{{ settingsStore.systemInfo.version }}
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Component-specific styles */
</style>
