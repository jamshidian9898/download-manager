<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  X, 
  Download, 
  FolderOpen, 
  RefreshCw,
  Settings as SettingsIcon,
  Copy
} from 'lucide-vue-next'

import { useDownloadStore } from '../stores/downloads.js'
import { useQueueStore } from '../stores/queues.js'
import { useSettingsStore } from '../stores/settings.js'
import { downloadAPI, formatters } from '../utils/api.js'

const emit = defineEmits(['close'])

const downloadStore = useDownloadStore()
const queueStore = useQueueStore()
const settingsStore = useSettingsStore()

const url = ref('')
const filename = ref('')
const filePath = ref('')
const selectedQueueId = ref(null)
const maxConnections = ref(8)
const speedLimit = ref(0)
const headers = ref({})
const userAgent = ref('')
const referrer = ref('')
const priority = ref(0)

const loading = ref(false)
const error = ref('')
const downloadInfo = ref(null)
const showAdvanced = ref(false)

const isValidUrl = computed(() => {
  try {
    new URL(url.value)
    return true
  } catch {
    return false
  }
})

const canSubmit = computed(() => {
  return url.value.trim() && !loading.value && isValidUrl.value
})

onMounted(() => {
  // Set default values
  if (queueStore.queues.length > 0) {
    selectedQueueId.value = queueStore.currentQueue?.id || queueStore.queues[0].id
  }
  
  filePath.value = settingsStore.downloadPath
  userAgent.value = settingsStore.settings.user_agent || 'AB Download Manager 1.0'
})

const fetchDownloadInfo = async () => {
  if (!isValidUrl.value) return
  
  loading.value = true
  error.value = ''
  
  try {
    const response = await downloadAPI.getInfo(url.value, headers.value)
    
    if (response.success) {
      downloadInfo.value = response
      filename.value = response.filename || ''
      
      // Update file path with detected filename
      if (response.filename && filePath.value) {
        const basePath = filePath.value.replace(/\/[^\/]*$/, '')
        filePath.value = `${basePath}/${response.filename}`
      }
    } else {
      error.value = response.error || 'Failed to get download info'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleUrlChange = () => {
  downloadInfo.value = null
  filename.value = ''
  error.value = ''
  
  if (isValidUrl.value) {
    // Auto-fetch info after a short delay
    setTimeout(fetchDownloadInfo, 500)
  }
}

const selectDownloadPath = () => {
  // This would integrate with Wails file dialog
  // For now, just show a placeholder
  console.log('Select download path')
}

const addHeader = () => {
  const key = prompt('Header name:')
  const value = prompt('Header value:')
  
  if (key && value) {
    headers.value[key] = value
  }
}

const removeHeader = (key) => {
  delete headers.value[key]
}

const handleSubmit = async () => {
  if (!canSubmit.value) return
  
  loading.value = true
  error.value = ''
  
  try {
    const request = {
      url: url.value.trim(),
      filename: filename.value.trim(),
      file_path: filePath.value.trim(),
      queue_id: selectedQueueId.value,
      max_connections: maxConnections.value,
      speed_limit: speedLimit.value > 0 ? speedLimit.value * 1024 : 0, // Convert KB/s to bytes/s
      headers: Object.keys(headers.value).length > 0 ? headers.value : undefined,
      user_agent: userAgent.value.trim(),
      referrer: referrer.value.trim(),
      priority: priority.value
    }
    
    const response = await downloadStore.addDownload(request)
    
    if (response.success) {
      emit('close')
    } else {
      error.value = response.error || 'Failed to add download'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const pasteFromClipboard = async () => {
  try {
    const text = await navigator.clipboard.readText()
    if (text) {
      url.value = text.trim()
      handleUrlChange()
    }
  } catch (err) {
    console.error('Failed to read clipboard:', err)
  }
}
</script>

<template>
  <!-- Modal Overlay -->
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-dark-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-dark-700">
        <div class="flex items-center space-x-2">
          <Download class="w-6 h-6 text-primary-500" />
          <h2 class="text-xl font-semibold">Add download</h2>
        </div>
        
        <button
          @click="emit('close')"
          class="text-gray-400 hover:text-white transition-colors"
        >
          <X class="w-6 h-6" />
        </button>
      </div>
      
      <!-- Content -->
      <div class="p-6 space-y-6">
        <!-- URL Input -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">
            URL *
          </label>
          <div class="flex space-x-2">
            <input
              v-model="url"
              @input="handleUrlChange"
              type="url"
              placeholder="https://download-installer.cdn.mozilla.net/pub/firefox/releases/129.0.2/"
              class="input flex-1"
              :class="{ 'border-red-500': error && !isValidUrl }"
            />
            <button
              @click="pasteFromClipboard"
              class="btn btn-secondary p-2"
              title="Paste from clipboard"
            >
              <Copy class="w-4 h-4" />
            </button>
            <button
              @click="fetchDownloadInfo"
              :disabled="!isValidUrl || loading"
              class="btn btn-secondary p-2"
              title="Refresh info"
            >
              <RefreshCw :class="['w-4 h-4', loading ? 'animate-spin' : '']" />
            </button>
          </div>
        </div>
        
        <!-- Download Info -->
        <div v-if="downloadInfo" class="bg-dark-700 rounded-lg p-4">
          <div class="flex items-center space-x-2 mb-2">
            <Download class="w-4 h-4 text-green-500" />
            <span class="text-sm font-medium text-green-500">Download Info Detected</span>
          </div>
          
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-gray-400">Size:</span>
              <span class="ml-2 text-white">{{ formatters.formatBytes(downloadInfo.content_length) }}</span>
            </div>
            <div>
              <span class="text-gray-400">Resume Support:</span>
              <span class="ml-2" :class="downloadInfo.supports_range ? 'text-green-500' : 'text-red-500'">
                {{ downloadInfo.supports_range ? 'Yes' : 'No' }}
              </span>
            </div>
            <div class="col-span-2">
              <span class="text-gray-400">Content Type:</span>
              <span class="ml-2 text-white">{{ downloadInfo.content_type || 'Unknown' }}</span>
            </div>
          </div>
        </div>
        
        <!-- Error Display -->
        <div v-if="error" class="bg-red-600/20 border border-red-600 rounded-lg p-3">
          <p class="text-red-400 text-sm">{{ error }}</p>
        </div>
        
        <!-- File Details -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Filename
            </label>
            <input
              v-model="filename"
              type="text"
              placeholder="Firefox Installer.exe"
              class="input w-full"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Queue
            </label>
            <select
              v-model="selectedQueueId"
              class="input w-full"
            >
              <option
                v-for="queue in queueStore.queues"
                :key="queue.id"
                :value="queue.id"
              >
                {{ queue.name }}
              </option>
            </select>
          </div>
        </div>
        
        <!-- Download Path -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">
            Download Path
          </label>
          <div class="flex space-x-2">
            <input
              v-model="filePath"
              type="text"
              placeholder="C:\Users\amir\Downloads\ABDM"
              class="input flex-1"
            />
            <button
              @click="selectDownloadPath"
              class="btn btn-secondary p-2"
              title="Browse"
            >
              <FolderOpen class="w-4 h-4" />
            </button>
          </div>
        </div>
        
        <!-- Advanced Settings Toggle -->
        <div>
          <button
            @click="showAdvanced = !showAdvanced"
            class="flex items-center space-x-2 text-primary-500 hover:text-primary-400 transition-colors"
          >
            <SettingsIcon class="w-4 h-4" />
            <span>Advanced Settings</span>
          </button>
        </div>
        
        <!-- Advanced Settings -->
        <div v-if="showAdvanced" class="space-y-4 border-t border-dark-700 pt-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">
                Max Connections
              </label>
              <input
                v-model.number="maxConnections"
                type="number"
                min="1"
                max="16"
                class="input w-full"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">
                Speed Limit (KB/s)
              </label>
              <input
                v-model.number="speedLimit"
                type="number"
                min="0"
                placeholder="0 = Unlimited"
                class="input w-full"
              />
            </div>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              User Agent
            </label>
            <input
              v-model="userAgent"
              type="text"
              class="input w-full"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Referrer
            </label>
            <input
              v-model="referrer"
              type="text"
              placeholder="Optional"
              class="input w-full"
            />
          </div>
          
          <!-- Custom Headers -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-sm font-medium text-gray-300">
                Custom Headers
              </label>
              <button
                @click="addHeader"
                class="btn btn-secondary btn-sm"
              >
                Add Header
              </button>
            </div>
            
            <div v-if="Object.keys(headers).length > 0" class="space-y-2">
              <div
                v-for="(value, key) in headers"
                :key="key"
                class="flex items-center space-x-2 bg-dark-700 rounded p-2"
              >
                <span class="text-sm text-gray-300 flex-1">{{ key }}: {{ value }}</span>
                <button
                  @click="removeHeader(key)"
                  class="text-red-400 hover:text-red-300"
                >
                  <X class="w-4 h-4" />
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
          Cancel
        </button>
        
        <button
          @click="handleSubmit"
          :disabled="!canSubmit"
          class="btn btn-primary"
        >
          <RefreshCw v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
          <Download v-else class="w-4 h-4 mr-2" />
          {{ loading ? 'Adding...' : 'Download' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Component-specific styles */
</style>
