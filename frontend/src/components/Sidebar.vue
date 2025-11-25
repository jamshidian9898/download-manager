<script setup>
import { computed } from 'vue'
import { 
  Download, 
  Plus, 
  Settings, 
  Image, 
  Music, 
  Video, 
  FileText, 
  Archive, 
  Smartphone,
  File,
  CheckCircle,
  XCircle,
  Pause,
  Clock,
  StopCircle
} from 'lucide-vue-next'

import { useDownloadStore } from '../stores/downloads.js'
import { useQueueStore } from '../stores/queues.js'

const emit = defineEmits(['add-download', 'show-settings'])

const downloadStore = useDownloadStore()
const queueStore = useQueueStore()

const categories = [
  { id: 'all', name: 'All', icon: Download, count: computed(() => downloadStore.downloads.length) },
  { id: 'image', name: 'Image', icon: Image, count: computed(() => downloadStore.downloadsByCategory.image?.length || 0) },
  { id: 'music', name: 'Music', icon: Music, count: computed(() => downloadStore.downloadsByCategory.audio?.length || 0) },
  { id: 'video', name: 'Video', icon: Video, count: computed(() => downloadStore.downloadsByCategory.video?.length || 0) },
  { id: 'document', name: 'Document', icon: FileText, count: computed(() => downloadStore.downloadsByCategory.document?.length || 0) },
  { id: 'compressed', name: 'Compressed', icon: Archive, count: computed(() => downloadStore.downloadsByCategory.compressed?.length || 0) },
  { id: 'apps', name: 'Apps', icon: Smartphone, count: computed(() => downloadStore.downloadsByCategory.apps?.length || 0) },
  { id: 'other', name: 'Other', icon: File, count: computed(() => downloadStore.downloadsByCategory.other?.length || 0) }
]

const statusFilters = [
  { id: 'finished', name: 'Finished', icon: CheckCircle, count: computed(() => downloadStore.downloadsByStatus.completed?.length || 0), color: 'text-green-500' },
  { id: 'unfinished', name: 'Unfinished', icon: Clock, count: computed(() => (downloadStore.downloadsByStatus.pending?.length || 0) + (downloadStore.downloadsByStatus.downloading?.length || 0) + (downloadStore.downloadsByStatus.paused?.length || 0)), color: 'text-yellow-500' }
]

const handleCategoryClick = (categoryId) => {
  downloadStore.setCategoryFilter(categoryId)
}

const handleStatusClick = (statusId) => {
  if (statusId === 'finished') {
    downloadStore.setStatusFilter('completed')
  } else if (statusId === 'unfinished') {
    downloadStore.setStatusFilter('all')
    // Could be enhanced to show only unfinished items
  }
}
</script>

<template>
  <aside class="bg-dark-800 border-r border-dark-700 flex flex-col">
    <!-- Header -->
    <div class="p-4 border-b border-dark-700">
      <div class="flex items-center space-x-2 mb-4">
        <Download class="w-6 h-6 text-primary-500" />
        <span class="font-semibold text-lg">AB Download Manager</span>
      </div>
      
      <!-- Add Download Button -->
      <button
        @click="emit('add-download')"
        class="w-full btn btn-primary flex items-center justify-center space-x-2"
      >
        <Plus class="w-4 h-4" />
        <span>Add URI</span>
      </button>
    </div>
    
    <!-- Categories -->
    <div class="flex-1 overflow-y-auto">
      <div class="p-4">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">
          Categories
        </h3>
        
        <nav class="space-y-1">
          <button
            v-for="category in categories"
            :key="category.id"
            @click="handleCategoryClick(category.id)"
            :class="[
              'w-full flex items-center justify-between px-3 py-2 text-sm rounded-lg transition-colors',
              downloadStore.filters.category === category.id
                ? 'bg-primary-600 text-white'
                : 'text-gray-300 hover:bg-dark-700 hover:text-white'
            ]"
          >
            <div class="flex items-center space-x-3">
              <component :is="category.icon" class="w-4 h-4" />
              <span>{{ category.name }}</span>
            </div>
            <span class="text-xs bg-dark-600 px-2 py-1 rounded-full">
              {{ category.count.value }}
            </span>
          </button>
        </nav>
      </div>
      
      <!-- Status Filters -->
      <div class="p-4 border-t border-dark-700">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">
          Status
        </h3>
        
        <nav class="space-y-1">
          <button
            v-for="status in statusFilters"
            :key="status.id"
            @click="handleStatusClick(status.id)"
            class="w-full flex items-center justify-between px-3 py-2 text-sm rounded-lg transition-colors text-gray-300 hover:bg-dark-700 hover:text-white"
          >
            <div class="flex items-center space-x-3">
              <component :is="status.icon" :class="['w-4 h-4', status.color]" />
              <span>{{ status.name }}</span>
            </div>
            <span class="text-xs bg-dark-600 px-2 py-1 rounded-full">
              {{ status.count.value }}
            </span>
          </button>
        </nav>
      </div>
      
      <!-- Queue Info -->
      <div class="p-4 border-t border-dark-700" v-if="queueStore.currentQueue">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">
          Current Queue
        </h3>
        
        <div class="bg-dark-700 rounded-lg p-3">
          <div class="flex items-center justify-between mb-2">
            <span class="font-medium">{{ queueStore.currentQueue.name }}</span>
            <span :class="[
              'w-2 h-2 rounded-full',
              queueStore.currentQueue.is_running ? 'bg-green-500' : 'bg-gray-500'
            ]"></span>
          </div>
          
          <div class="text-xs text-gray-400 space-y-1">
            <div class="flex justify-between">
              <span>Active:</span>
              <span>{{ queueStore.getQueueActiveCount(queueStore.currentQueue.id) }}</span>
            </div>
            <div class="flex justify-between">
              <span>Total:</span>
              <span>{{ queueStore.getQueueDownloadCount(queueStore.currentQueue.id) }}</span>
            </div>
            <div class="flex justify-between">
              <span>Max Concurrent:</span>
              <span>{{ queueStore.currentQueue.max_concurrent }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Footer -->
    <div class="p-4 border-t border-dark-700">
      <button
        @click="emit('show-settings')"
        class="w-full flex items-center space-x-3 px-3 py-2 text-sm text-gray-300 hover:bg-dark-700 hover:text-white rounded-lg transition-colors"
      >
        <Settings class="w-4 h-4" />
        <span>Settings</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
/* Component-specific styles */
</style>
