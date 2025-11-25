<script setup>
import { ref, computed } from 'vue'
import { 
  X, 
  Settings as SettingsIcon,
  FolderOpen,
  Save,
  RotateCcw
} from 'lucide-vue-next'

import { useSettingsStore } from '../stores/settings.js'

const emit = defineEmits(['close'])

const settingsStore = useSettingsStore()

const activeTab = ref('general')
const saving = ref(false)

const tabs = [
  { id: 'general', name: 'General' },
  { id: 'downloads', name: 'Downloads' },
  { id: 'network', name: 'Network' },
  { id: 'advanced', name: 'Advanced' }
]

const localSettings = ref({ ...settingsStore.settings })

const hasChanges = computed(() => {
  return settingsStore.isDirty
})

const handleSave = async () => {
  saving.value = true
  try {
    await settingsStore.saveSettings()
  } catch (error) {
    console.error('Failed to save settings:', error)
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  if (confirm('Are you sure you want to reset all settings to defaults?')) {
    settingsStore.resetSettings()
    localSettings.value = { ...settingsStore.settings }
  }
}

const selectDownloadPath = () => {
  // This would integrate with Wails file dialog
  console.log('Select download path')
}

const updateSetting = (key, value) => {
  settingsStore.updateSetting(key, value)
  localSettings.value[key] = value
}
</script>

<template>
  <!-- Modal Overlay -->
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-dark-800 rounded-lg shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-dark-700">
        <div class="flex items-center space-x-2">
          <SettingsIcon class="w-6 h-6 text-primary-500" />
          <h2 class="text-xl font-semibold">Settings</h2>
        </div>
        
        <button
          @click="emit('close')"
          class="text-gray-400 hover:text-white transition-colors"
        >
          <X class="w-6 h-6" />
        </button>
      </div>
      
      <div class="flex flex-1 overflow-hidden">
        <!-- Sidebar -->
        <div class="w-64 bg-dark-900 border-r border-dark-700 p-4">
          <nav class="space-y-1">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="[
                'w-full text-left px-3 py-2 rounded-lg transition-colors',
                activeTab === tab.id
                  ? 'bg-primary-600 text-white'
                  : 'text-gray-300 hover:bg-dark-700 hover:text-white'
              ]"
            >
              {{ tab.name }}
            </button>
          </nav>
        </div>
        
        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- General Tab -->
          <div v-if="activeTab === 'general'" class="space-y-6">
            <h3 class="text-lg font-medium">General Settings</h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Theme
                </label>
                <select
                  :value="settingsStore.settings.theme"
                  @change="updateSetting('theme', $event.target.value)"
                  class="input w-full max-w-xs"
                >
                  <option value="dark">Dark</option>
                  <option value="light">Light</option>
                </select>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Language
                </label>
                <select
                  :value="settingsStore.settings.language"
                  @change="updateSetting('language', $event.target.value)"
                  class="input w-full max-w-xs"
                >
                  <option value="en">English</option>
                  <option value="es">Spanish</option>
                  <option value="fr">French</option>
                </select>
              </div>
              
              <div class="space-y-3">
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.notifications_enabled"
                    @change="updateSetting('notifications_enabled', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Enable notifications</span>
                </label>
                
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.sound_notifications"
                    @change="updateSetting('sound_notifications', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Sound notifications</span>
                </label>
                
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.minimize_to_tray"
                    @change="updateSetting('minimize_to_tray', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Minimize to system tray</span>
                </label>
                
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.start_minimized"
                    @change="updateSetting('start_minimized', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Start minimized</span>
                </label>
              </div>
            </div>
          </div>
          
          <!-- Downloads Tab -->
          <div v-if="activeTab === 'downloads'" class="space-y-6">
            <h3 class="text-lg font-medium">Download Settings</h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Default Download Path
                </label>
                <div class="flex space-x-2">
                  <input
                    :value="settingsStore.settings.default_download_path"
                    @input="updateSetting('default_download_path', $event.target.value)"
                    type="text"
                    class="input flex-1"
                  />
                  <button
                    @click="selectDownloadPath"
                    class="btn btn-secondary p-2"
                  >
                    <FolderOpen class="w-4 h-4" />
                  </button>
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Max Concurrent Downloads
                </label>
                <input
                  :value="settingsStore.settings.max_concurrent_downloads"
                  @input="updateSetting('max_concurrent_downloads', parseInt($event.target.value))"
                  type="number"
                  min="1"
                  max="20"
                  class="input w-full max-w-xs"
                />
              </div>
              
              <div>
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.auto_start_downloads"
                    @change="updateSetting('auto_start_downloads', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Auto-start downloads</span>
                </label>
              </div>
            </div>
          </div>
          
          <!-- Network Tab -->
          <div v-if="activeTab === 'network'" class="space-y-6">
            <h3 class="text-lg font-medium">Network Settings</h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Global Speed Limit (KB/s)
                </label>
                <input
                  :value="settingsStore.settings.global_speed_limit ? Math.round(settingsStore.settings.global_speed_limit / 1024) : 0"
                  @input="updateSetting('global_speed_limit', parseInt($event.target.value) * 1024)"
                  type="number"
                  min="0"
                  placeholder="0 = Unlimited"
                  class="input w-full max-w-xs"
                />
                <p class="text-xs text-gray-400 mt-1">0 means unlimited</p>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Connection Timeout (seconds)
                </label>
                <input
                  :value="settingsStore.settings.connection_timeout"
                  @input="updateSetting('connection_timeout', parseInt($event.target.value))"
                  type="number"
                  min="5"
                  max="300"
                  class="input w-full max-w-xs"
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Retry Attempts
                </label>
                <input
                  :value="settingsStore.settings.retry_attempts"
                  @input="updateSetting('retry_attempts', parseInt($event.target.value))"
                  type="number"
                  min="0"
                  max="10"
                  class="input w-full max-w-xs"
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  User Agent
                </label>
                <input
                  :value="settingsStore.settings.user_agent"
                  @input="updateSetting('user_agent', $event.target.value)"
                  type="text"
                  class="input w-full"
                />
              </div>
            </div>
          </div>
          
          <!-- Advanced Tab -->
          <div v-if="activeTab === 'advanced'" class="space-y-6">
            <h3 class="text-lg font-medium">Advanced Settings</h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Data Directory
                </label>
                <input
                  :value="settingsStore.settings.data_directory"
                  @input="updateSetting('data_directory', $event.target.value)"
                  type="text"
                  class="input w-full"
                  readonly
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  Log Level
                </label>
                <select
                  :value="settingsStore.settings.log_level"
                  @change="updateSetting('log_level', $event.target.value)"
                  class="input w-full max-w-xs"
                >
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="warn">Warning</option>
                  <option value="error">Error</option>
                </select>
              </div>
              
              <div>
                <label class="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    :checked="settingsStore.settings.enable_logging"
                    @change="updateSetting('enable_logging', $event.target.checked)"
                    class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-300">Enable logging</span>
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Footer -->
      <div class="flex items-center justify-between p-6 border-t border-dark-700">
        <button
          @click="handleReset"
          class="btn btn-secondary"
        >
          <RotateCcw class="w-4 h-4 mr-2" />
          Reset to Defaults
        </button>
        
        <div class="flex items-center space-x-3">
          <button
            @click="emit('close')"
            class="btn btn-secondary"
          >
            Cancel
          </button>
          
          <button
            @click="handleSave"
            :disabled="!hasChanges || saving"
            class="btn btn-primary"
          >
            <Save class="w-4 h-4 mr-2" />
            {{ saving ? 'Saving...' : 'Save Changes' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Component-specific styles */
</style>
