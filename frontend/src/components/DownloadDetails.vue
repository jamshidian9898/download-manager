<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { 
  X, 
  Info, 
  Settings as SettingsIcon,
  Play,
  Pause,
  Square,
  Trash2,
  FolderOpen,
  Copy,
  RefreshCw
} from 'lucide-vue-next'

import { useDownloadStore } from '../stores/downloads.js'
import { formatters } from '../utils/api.js'

const props = defineProps({
  downloadId: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['close'])

const downloadStore = useDownloadStore()

const activeTab = ref('info')
const download = ref(null)
const loading = ref(false)

const tabs = [
  { id: 'info', name: 'Info', icon: Info },
  { id: 'settings', name: 'Settings', icon: SettingsIcon }
]

const progressPercentage = computed(() => {
  if (!download.value?.total_size || download.value.total_size === 0) return 0
  return Math.round((download.value.progress?.bytes_downloaded || 0) / download.value.total_size * 100)
})

const statusColor = computed(() => {
  const colors = {
    'pending': 'text-yellow-500',
    'downloading': 'text-blue-500',
    'paused': 'text-orange-500',
    'completed': 'text-green-500',
    'failed': 'text-red-500',
    'canceled': 'text-gray-500'
  }
  return colors[download.value?.status] || 'text-gray-500'
})

const progressBarColor = computed(() => {
  switch (download.value?.status) {
    case 'downloading':
      return 'bg-blue-500'
    case 'completed':
      return 'bg-green-500'
    case 'paused':
      return 'bg-orange-500'
    case 'failed':
      return 'bg-red-500'
    default:
      return 'bg-gray-500'
  }
})

const estimatedTimeLeft = computed(() => {
  if (download.value?.status !== 'downloading' || !download.value?.progress?.speed) {
    return '--'
  }
  
  const remaining = download.value.total_size - (download.value.progress?.bytes_downloaded || 0)
  const seconds = remaining / download.value.progress.speed
  return formatters.formatDuration(seconds)
})

const canStart = computed(() => {
  return ['pending', 'paused', 'failed'].includes(download.value?.status)
})

const canPause = computed(() => {
  return download.value?.status === 'downloading'
})

const canResume = computed(() => {
  return download.value?.status === 'paused'
})

onMounted(async () => {
  await fetchDownload()
})

watch(() => props.downloadId, async () => {
  await fetchDownload()
})

const fetchDownload = async () => {
  loading.value = true
  try {
    download.value = await downloadStore.getDownload(props.downloadId)
  } catch (error) {
    console.error('Failed to fetch download:', error)
  } finally {
    loading.value = false
  }
}

const handleAction = async (action) => {
  try {
    switch (action) {
      case 'start':
        await downloadStore.startDownload(props.downloadId)
        break
      case 'pause':
        await downloadStore.pauseDownload(props.downloadId)
        break
      case 'resume':
        await downloadStore.resumeDownload(props.downloadId)
        break
      case 'cancel':
        await downloadStore.cancelDownload(props.downloadId)
        break
      case 'delete':
        if (confirm('Are you sure you want to delete this download?')) {
          await downloadStore.deleteDownload(props.downloadId)
          emit('close')
        }
        break
    }
    await fetchDownload()
  } catch (error) {
    console.error('Action failed:', error)
  }
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
  }
}

const openFolder = () => {
  // This would integrate with Wails to open the file location
  console.log('Open folder:', download.value?.file_path)
}

const formatDate = (dateString) => {
  if (!dateString) return '--'
  return new Date(dateString).toLocaleString()
}

// Mock part data for demonstration
const parts = computed(() => {
  if (!download.value || download.value.status === 'pending') return []
  
  // Generate mock parts based on download progress
  const partCount = Math.min(8, download.value.max_connections || 4)
  const totalSize = download.value.total_size || 0
  const downloaded = download.value.progress?.bytes_downloaded || 0
  const partSize = Math.floor(totalSize / partCount)
  
  return Array.from({ length: partCount }, (_, i) => {
    const start = i * partSize
    const end = i === partCount - 1 ? totalSize : (i + 1) * partSize
    const partDownloaded = Math.min(end - start, Math.max(0, downloaded - start))
    
    return {
      id: i + 1,
      status: partDownloaded === (end - start) ? 'Completed' : 
              partDownloaded > 0 ? 'Receiving Data' : 'Waiting',
      downloaded: partDownloaded,
      total: end - start,
      progress: Math.round((partDownloaded / (end - start)) * 100)
    }
  })
})
</script>

<template>
  <!-- Modal Overlay -->
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-dark-800 rounded-lg shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-dark-700">
        <div class="flex items-center space-x-3 min-w-0 flex-1">
          <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center flex-shrink-0">
            <span class="text-white font-bold text-sm">{{ progressPercentage }}%</span>
          </div>
          <div class="min-w-0 flex-1">
            <h2 class="text-lg font-semibold truncate" :title="download?.filename">
              {{ download?.filename || 'Loading...' }}
            </h2>
            <p :class="['text-sm capitalize', statusColor]">
              {{ download?.status || 'Loading' }}
            </p>
          </div>
        </div>
        
        <button
          @click="emit('close')"
          class="text-gray-400 hover:text-white transition-colors ml-4"
        >
          <X class="w-6 h-6" />
        </button>
      </div>
      
      <!-- Progress Bar -->
      <div v-if="download" class="px-6 py-4 bg-dark-900">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center space-x-4 text-sm">
            <span class="text-white">
              {{ formatters.formatBytes(download.progress?.bytes_downloaded || 0) }}
            </span>
            <span class="text-gray-400">
              {{ formatters.formatSpeed(download.progress?.speed || 0) }}
            </span>
            <span class="text-gray-400">
              {{ estimatedTimeLeft }} left
            </span>
          </div>
          <span class="text-sm text-gray-400">
            {{ progressPercentage }}%
          </span>
        </div>
        
        <div class="w-full bg-dark-600 rounded-full h-3">
          <div 
            :class="['h-3 rounded-full transition-all duration-300', progressBarColor]"
            :style="{ width: `${progressPercentage}%` }"
          ></div>
        </div>
      </div>
      
      <!-- Action Buttons -->
      <div class="px-6 py-4 border-b border-dark-700 flex items-center justify-between">
        <div class="flex items-center space-x-2">
          <button
            v-if="canStart"
            @click="handleAction('start')"
            class="btn btn-primary"
          >
            <Play class="w-4 h-4 mr-2" />
            Start
          </button>
          
          <button
            v-if="canPause"
            @click="handleAction('pause')"
            class="btn btn-secondary"
          >
            <Pause class="w-4 h-4 mr-2" />
            Pause
          </button>
          
          <button
            v-if="canResume"
            @click="handleAction('resume')"
            class="btn btn-primary"
          >
            <Play class="w-4 h-4 mr-2" />
            Resume
          </button>
          
          <button
            @click="handleAction('cancel')"
            class="btn btn-secondary"
          >
            <Square class="w-4 h-4 mr-2" />
            Cancel
          </button>
        </div>
        
        <div class="flex items-center space-x-2">
          <button
            @click="openFolder"
            class="btn btn-secondary"
            title="Open Folder"
          >
            <FolderOpen class="w-4 h-4" />
          </button>
          
          <button
            @click="handleAction('delete')"
            class="btn btn-danger"
            title="Delete"
          >
            <Trash2 class="w-4 h-4" />
          </button>
          
          <button
            @click="fetchDownload"
            class="btn btn-secondary"
            title="Refresh"
          >
            <RefreshCw class="w-4 h-4" />
          </button>
        </div>
      </div>
      
      <!-- Tabs -->
      <div class="border-b border-dark-700">
        <nav class="flex space-x-8 px-6">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="[
              'flex items-center space-x-2 py-4 border-b-2 font-medium text-sm transition-colors',
              activeTab === tab.id
                ? 'border-primary-500 text-primary-500'
                : 'border-transparent text-gray-400 hover:text-gray-300'
            ]"
          >
            <component :is="tab.icon" class="w-4 h-4" />
            <span>{{ tab.name }}</span>
          </button>
        </nav>
      </div>
      
      <!-- Tab Content -->
      <div class="flex-1 overflow-y-auto p-6">
        <!-- Info Tab -->
        <div v-if="activeTab === 'info'" class="space-y-6">
          <!-- Basic Info -->
          <div>
            <h3 class="text-lg font-medium mb-4">Download Information</h3>
            
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div class="space-y-3">
                <div>
                  <label class="text-gray-400">Name:</label>
                  <div class="flex items-center space-x-2 mt-1">
                    <span class="text-white">{{ download?.filename || '--' }}</span>
                    <button
                      @click="copyToClipboard(download?.filename)"
                      class="text-gray-400 hover:text-white"
                    >
                      <Copy class="w-3 h-3" />
                    </button>
                  </div>
                </div>
                
                <div>
                  <label class="text-gray-400">Status:</label>
                  <p :class="['mt-1 capitalize', statusColor]">{{ download?.status || '--' }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Size:</label>
                  <p class="text-white mt-1">{{ formatters.formatBytes(download?.total_size || 0) }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Downloaded:</label>
                  <p class="text-white mt-1">{{ formatters.formatBytes(download?.progress?.bytes_downloaded || 0) }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Speed:</label>
                  <p class="text-white mt-1">{{ formatters.formatSpeed(download?.progress?.speed || 0) }}</p>
                </div>
              </div>
              
              <div class="space-y-3">
                <div>
                  <label class="text-gray-400">Remaining Time:</label>
                  <p class="text-white mt-1">{{ estimatedTimeLeft }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Resume Support:</label>
                  <p class="text-white mt-1">Yes</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Created:</label>
                  <p class="text-white mt-1">{{ formatDate(download?.created_at) }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Updated:</label>
                  <p class="text-white mt-1">{{ formatDate(download?.updated_at) }}</p>
                </div>
                
                <div>
                  <label class="text-gray-400">Connections:</label>
                  <p class="text-white mt-1">{{ download?.max_connections || '--' }}</p>
                </div>
              </div>
            </div>
          </div>
          
          <!-- URL -->
          <div>
            <label class="text-gray-400 text-sm">URL:</label>
            <div class="flex items-center space-x-2 mt-1">
              <input
                :value="download?.url"
                readonly
                class="input flex-1 text-sm"
              />
              <button
                @click="copyToClipboard(download?.url)"
                class="btn btn-secondary p-2"
                title="Copy URL"
              >
                <Copy class="w-4 h-4" />
              </button>
            </div>
          </div>
          
          <!-- File Path -->
          <div>
            <label class="text-gray-400 text-sm">File Path:</label>
            <div class="flex items-center space-x-2 mt-1">
              <input
                :value="download?.file_path"
                readonly
                class="input flex-1 text-sm"
              />
              <button
                @click="copyToClipboard(download?.file_path)"
                class="btn btn-secondary p-2"
                title="Copy Path"
              >
                <Copy class="w-4 h-4" />
              </button>
            </div>
          </div>
          
          <!-- Part Info -->
          <div v-if="parts.length > 0">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-medium">Part Info</h3>
              <button class="text-primary-500 hover:text-primary-400 text-sm">
                ▼ Collapse
              </button>
            </div>
            
            <div class="bg-dark-700 rounded-lg overflow-hidden">
              <div class="grid grid-cols-4 gap-4 p-3 bg-dark-600 text-sm font-medium text-gray-300">
                <span>#</span>
                <span>Status</span>
                <span>Downloaded</span>
                <span>Total</span>
              </div>
              
              <div class="divide-y divide-dark-600">
                <div
                  v-for="part in parts"
                  :key="part.id"
                  class="grid grid-cols-4 gap-4 p-3 text-sm"
                >
                  <span class="text-gray-300">{{ part.id }}</span>
                  <span :class="[
                    part.status === 'Completed' ? 'text-green-500' :
                    part.status === 'Receiving Data' ? 'text-blue-500' : 'text-yellow-500'
                  ]">
                    {{ part.status }}
                  </span>
                  <span class="text-white">{{ formatters.formatBytes(part.downloaded) }}</span>
                  <span class="text-white">{{ formatters.formatBytes(part.total) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Settings Tab -->
        <div v-if="activeTab === 'settings'" class="space-y-6">
          <h3 class="text-lg font-medium">Download Settings</h3>
          
          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">
                Max Connections
              </label>
              <input
                :value="download?.max_connections || 8"
                type="number"
                min="1"
                max="16"
                class="input w-full"
                readonly
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">
                Speed Limit (KB/s)
              </label>
              <input
                :value="download?.speed_limit ? Math.round(download.speed_limit / 1024) : 0"
                type="number"
                min="0"
                class="input w-full"
                readonly
              />
            </div>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              User Agent
            </label>
            <input
              :value="download?.user_agent || 'AB Download Manager 1.0'"
              type="text"
              class="input w-full"
              readonly
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Referrer
            </label>
            <input
              :value="download?.referrer || ''"
              type="text"
              placeholder="None"
              class="input w-full"
              readonly
            />
          </div>
          
          <!-- Custom Headers -->
          <div v-if="download?.headers && Object.keys(download.headers).length > 0">
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Custom Headers
            </label>
            <div class="space-y-2">
              <div
                v-for="(value, key) in download.headers"
                :key="key"
                class="flex items-center justify-between bg-dark-700 rounded p-3"
              >
                <span class="text-sm text-gray-300">{{ key }}: {{ value }}</span>
                <button
                  @click="copyToClipboard(`${key}: ${value}`)"
                  class="text-gray-400 hover:text-white"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 p-6 border-t border-dark-700">
        <button
          @click="emit('close')"
          class="btn btn-secondary"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Component-specific styles */
</style>
