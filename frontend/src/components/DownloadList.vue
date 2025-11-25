<script setup>
import { computed, ref } from 'vue'
import { 
  ChevronDown, 
  ChevronUp,
  Play,
  Pause,
  Square,
  Trash2,
  Info,
  MoreHorizontal
} from 'lucide-vue-next'

import DownloadItem from './DownloadItem.vue'
import { useDownloadStore } from '../stores/downloads.js'
import { formatters } from '../utils/api.js'

const emit = defineEmits(['show-details'])

const downloadStore = useDownloadStore()

const sortableColumns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'size', label: 'Size', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'progress', label: 'Downloaded', sortable: true },
  { key: 'speed', label: 'Speed', sortable: true },
  { key: 'date_added', label: 'Time Left', sortable: true },
  { key: 'date_added', label: 'Date Added', sortable: true }
]

const handleSort = (column) => {
  if (column.sortable) {
    downloadStore.setSorting(column.key)
  }
}

const handleSelectAll = () => {
  if (downloadStore.selectedDownloads.size === downloadStore.filteredDownloads.length) {
    downloadStore.deselectAllDownloads()
  } else {
    downloadStore.selectAllDownloads()
  }
}

const handleBulkAction = async (action) => {
  switch (action) {
    case 'start':
      await downloadStore.startSelectedDownloads()
      break
    case 'pause':
      await downloadStore.pauseSelectedDownloads()
      break
    case 'delete':
      if (confirm('Are you sure you want to delete the selected downloads?')) {
        await downloadStore.deleteSelectedDownloads()
      }
      break
  }
}

const isAllSelected = computed(() => {
  return downloadStore.filteredDownloads.length > 0 && 
         downloadStore.selectedDownloads.size === downloadStore.filteredDownloads.length
})

const isPartiallySelected = computed(() => {
  return downloadStore.selectedDownloads.size > 0 && 
         downloadStore.selectedDownloads.size < downloadStore.filteredDownloads.length
})

const getSortIcon = (columnKey) => {
  if (downloadStore.sortBy !== columnKey) return null
  return downloadStore.sortOrder === 'asc' ? ChevronUp : ChevronDown
}
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Bulk Actions Bar -->
    <div 
      v-if="downloadStore.selectedDownloads.size > 0"
      class="bg-primary-600 px-4 py-2 flex items-center justify-between"
    >
      <span class="text-sm font-medium">
        {{ downloadStore.selectedDownloads.size }} item(s) selected
      </span>
      
      <div class="flex items-center space-x-2">
        <button
          @click="handleBulkAction('start')"
          class="btn btn-secondary btn-sm"
          title="Start Selected"
        >
          <Play class="w-4 h-4" />
        </button>
        
        <button
          @click="handleBulkAction('pause')"
          class="btn btn-secondary btn-sm"
          title="Pause Selected"
        >
          <Pause class="w-4 h-4" />
        </button>
        
        <button
          @click="handleBulkAction('delete')"
          class="btn btn-danger btn-sm"
          title="Delete Selected"
        >
          <Trash2 class="w-4 h-4" />
        </button>
      </div>
    </div>
    
    <!-- Table Header -->
    <div class="bg-dark-800 border-b border-dark-700 px-4 py-3">
      <div class="grid grid-cols-12 gap-4 items-center text-sm font-medium text-gray-400">
        <!-- Checkbox -->
        <div class="col-span-1">
          <input
            type="checkbox"
            :checked="isAllSelected"
            :indeterminate="isPartiallySelected"
            @change="handleSelectAll"
            class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
          />
        </div>
        
        <!-- Name -->
        <div class="col-span-4">
          <button
            @click="handleSort({ key: 'name', sortable: true })"
            class="flex items-center space-x-1 hover:text-white transition-colors"
          >
            <span>Name</span>
            <component 
              :is="getSortIcon('name')" 
              v-if="getSortIcon('name')"
              class="w-4 h-4"
            />
          </button>
        </div>
        
        <!-- Size -->
        <div class="col-span-1">
          <button
            @click="handleSort({ key: 'size', sortable: true })"
            class="flex items-center space-x-1 hover:text-white transition-colors"
          >
            <span>Size</span>
            <component 
              :is="getSortIcon('size')" 
              v-if="getSortIcon('size')"
              class="w-4 h-4"
            />
          </button>
        </div>
        
        <!-- Status -->
        <div class="col-span-1">
          <button
            @click="handleSort({ key: 'status', sortable: true })"
            class="flex items-center space-x-1 hover:text-white transition-colors"
          >
            <span>Status</span>
            <component 
              :is="getSortIcon('status')" 
              v-if="getSortIcon('status')"
              class="w-4 h-4"
            />
          </button>
        </div>
        
        <!-- Downloaded -->
        <div class="col-span-2">
          <span>Downloaded</span>
        </div>
        
        <!-- Speed -->
        <div class="col-span-1">
          <button
            @click="handleSort({ key: 'speed', sortable: true })"
            class="flex items-center space-x-1 hover:text-white transition-colors"
          >
            <span>Speed</span>
            <component 
              :is="getSortIcon('speed')" 
              v-if="getSortIcon('speed')"
              class="w-4 h-4"
            />
          </button>
        </div>
        
        <!-- Time Left -->
        <div class="col-span-1">
          <span>Time Left</span>
        </div>
        
        <!-- Date Added -->
        <div class="col-span-1">
          <button
            @click="handleSort({ key: 'date_added', sortable: true })"
            class="flex items-center space-x-1 hover:text-white transition-colors"
          >
            <span>Date Added</span>
            <component 
              :is="getSortIcon('date_added')" 
              v-if="getSortIcon('date_added')"
              class="w-4 h-4"
            />
          </button>
        </div>
      </div>
    </div>
    
    <!-- Download Items -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="downloadStore.loading" class="flex items-center justify-center h-32">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
      </div>
      
      <div v-else-if="downloadStore.error" class="flex items-center justify-center h-32 text-red-500">
        Error: {{ downloadStore.error }}
      </div>
      
      <div v-else-if="downloadStore.filteredDownloads.length === 0" class="flex items-center justify-center h-32 text-gray-500">
        No downloads found
      </div>
      
      <div v-else>
        <DownloadItem
          v-for="download in downloadStore.filteredDownloads"
          :key="download.id"
          :download="download"
          :selected="downloadStore.selectedDownloads.has(download.id)"
          @select="downloadStore.toggleDownloadSelection(download.id)"
          @show-details="emit('show-details', download.id)"
          @start="downloadStore.startDownload(download.id)"
          @pause="downloadStore.pauseDownload(download.id)"
          @resume="downloadStore.resumeDownload(download.id)"
          @cancel="downloadStore.cancelDownload(download.id)"
          @delete="downloadStore.deleteDownload(download.id)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom checkbox styles */
input[type="checkbox"]:indeterminate {
  background-image: url("data:image/svg+xml,%3csvg viewBox='0 0 16 16' fill='white' xmlns='http://www.w3.org/2000/svg'%3e%3cpath d='M4 8h8'/%3e%3c/svg%3e");
}
</style>
