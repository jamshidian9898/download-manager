<script setup>
import { onMounted, ref } from 'vue'
import { createPinia } from 'pinia'
import { 
  Download, 
  Plus, 
  Settings, 
  Search, 
  Play, 
  Pause, 
  Square,
  MoreHorizontal,
  Filter
} from 'lucide-vue-next'

import Sidebar from './components/Sidebar.vue'
import DownloadList from './components/DownloadList.vue'
import AddDownload from './components/AddDownload.vue'
import DownloadDetails from './components/DownloadDetails.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import StatusBar from './components/StatusBar.vue'

import { useDownloadStore } from './stores/downloads.js'
import { useQueueStore } from './stores/queues.js'
import { useSettingsStore } from './stores/settings.js'

const downloadStore = useDownloadStore()
const queueStore = useQueueStore()
const settingsStore = useSettingsStore()

const showAddDownload = ref(false)
const showDownloadDetails = ref(false)
const showSettings = ref(false)
const selectedDownloadId = ref(null)
const searchQuery = ref('')

onMounted(async () => {
  // Initialize stores
  await Promise.all([
    settingsStore.fetchSettings(),
    settingsStore.fetchSystemInfo(),
    queueStore.fetchQueues(),
    downloadStore.fetchDownloads()
  ])
  
  // Apply theme
  settingsStore.setTheme(settingsStore.settings.theme)
  
  // Initialize event listeners for real-time updates
  downloadStore.initializeEventListeners()
})

const handleAddDownload = () => {
  showAddDownload.value = true
}

const handleShowDetails = (downloadId) => {
  selectedDownloadId.value = downloadId
  showDownloadDetails.value = true
}

const handleShowSettings = () => {
  showSettings.value = true
}

const handleSearch = (query) => {
  downloadStore.setSearchFilter(query)
}

const handleStartAll = async () => {
  await queueStore.startAllQueues()
}

const handlePauseAll = async () => {
  await queueStore.pauseAllQueues()
}

const handleStopAll = async () => {
  await queueStore.stopAllQueues()
}
</script>

<template>
  <div class="flex h-screen bg-dark-900 text-white">
    <!-- Sidebar -->
    <Sidebar 
      class="w-64 flex-shrink-0"
      @add-download="handleAddDownload"
      @show-settings="handleShowSettings"
    />
    
    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Header -->
      <header class="bg-dark-800 border-b border-dark-700 px-6 py-4">
        <div class="flex items-center justify-between">
          <!-- Title and Queue Info -->
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2">
              <Download class="w-6 h-6 text-primary-500" />
              <h1 class="text-xl font-semibold">AB Download Manager</h1>
            </div>
            
            <div v-if="queueStore.currentQueue" class="text-sm text-gray-400">
              Queue: {{ queueStore.currentQueue.name }}
            </div>
          </div>
          
          <!-- Search and Controls -->
          <div class="flex items-center space-x-4">
            <!-- Search -->
            <div class="relative">
              <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                v-model="searchQuery"
                @input="handleSearch(searchQuery)"
                type="text"
                placeholder="Search in the List"
                class="input pl-10 pr-4 py-2 w-64"
              />
            </div>
            
            <!-- Control Buttons -->
            <div class="flex items-center space-x-2">
              <button
                @click="handleStartAll"
                class="btn btn-primary p-2"
                title="Start All"
              >
                <Play class="w-4 h-4" />
              </button>
              
              <button
                @click="handlePauseAll"
                class="btn btn-secondary p-2"
                title="Pause All"
              >
                <Pause class="w-4 h-4" />
              </button>
              
              <button
                @click="handleStopAll"
                class="btn btn-secondary p-2"
                title="Stop All"
              >
                <Square class="w-4 h-4" />
              </button>
              
              <button
                @click="handleAddDownload"
                class="btn btn-primary"
                title="Add Download"
              >
                <Plus class="w-4 h-4 mr-2" />
                Add URI
              </button>
              
              <button class="btn btn-secondary p-2">
                <Filter class="w-4 h-4" />
              </button>
              
              <button class="btn btn-secondary p-2">
                <MoreHorizontal class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </header>
      
      <!-- Download List -->
      <main class="flex-1 overflow-hidden">
        <DownloadList 
          @show-details="handleShowDetails"
          class="h-full"
        />
      </main>
      
      <!-- Status Bar -->
      <StatusBar class="border-t border-dark-700" />
    </div>
    
    <!-- Modals -->
    <AddDownload 
      v-if="showAddDownload"
      @close="showAddDownload = false"
    />
    
    <DownloadDetails
      v-if="showDownloadDetails"
      :download-id="selectedDownloadId"
      @close="showDownloadDetails = false"
    />
    
    <SettingsPanel
      v-if="showSettings"
      @close="showSettings = false"
    />
  </div>
</template>

<style scoped>
/* Component-specific styles */
</style>
