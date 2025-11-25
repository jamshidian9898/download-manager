import { defineStore } from 'pinia'
import { systemAPI } from '../utils/api.js'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    settings: {
      default_download_path: '',
      max_concurrent_downloads: 5,
      global_speed_limit: 0,
      auto_start_downloads: true,
      notifications_enabled: true,
      sound_notifications: false,
      minimize_to_tray: true,
      start_minimized: false,
      theme: 'dark',
      language: 'en',
      data_directory: '',
      temp_directory: '',
      connection_timeout: 30,
      retry_attempts: 3,
      retry_delay: 5,
      user_agent: 'AB Download Manager 1.0',
      enable_logging: true,
      log_level: 'info'
    },
    systemInfo: {
      active_downloads: 0,
      total_speed: 0,
      data_directory: '',
      version: '1.0.0'
    },
    loading: false,
    error: null,
    isDirty: false
  }),

  getters: {
    formattedSpeedLimit: (state) => {
      if (state.settings.global_speed_limit === 0) {
        return 'Unlimited'
      }
      return `${Math.round(state.settings.global_speed_limit / 1024)} KB/s`
    },

    isSpeedLimited: (state) => {
      return state.settings.global_speed_limit > 0
    },

    maxConcurrentDownloads: (state) => {
      return state.settings.max_concurrent_downloads
    },

    downloadPath: (state) => {
      return state.settings.default_download_path || '~/Downloads'
    },

    isDarkTheme: (state) => {
      return state.settings.theme === 'dark'
    },

    notificationsEnabled: (state) => {
      return state.settings.notifications_enabled
    }
  },

  actions: {
    async fetchSettings() {
      this.loading = true
      this.error = null
      
      try {
        this.settings = await systemAPI.getSettings()
        this.isDirty = false
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch settings:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchSystemInfo() {
      try {
        this.systemInfo = await systemAPI.getInfo()
      } catch (error) {
        console.error('Failed to fetch system info:', error)
      }
    },

    async saveSettings() {
      this.loading = true
      this.error = null
      
      try {
        await systemAPI.updateSettings(this.settings)
        this.isDirty = false
        return true
      } catch (error) {
        this.error = error.message
        console.error('Failed to save settings:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    updateSetting(key, value) {
      if (this.settings[key] !== value) {
        this.settings[key] = value
        this.isDirty = true
      }
    },

    updateSettings(newSettings) {
      Object.keys(newSettings).forEach(key => {
        if (this.settings.hasOwnProperty(key) && this.settings[key] !== newSettings[key]) {
          this.settings[key] = newSettings[key]
          this.isDirty = true
        }
      })
    },

    resetSettings() {
      this.settings = {
        default_download_path: '',
        max_concurrent_downloads: 5,
        global_speed_limit: 0,
        auto_start_downloads: true,
        notifications_enabled: true,
        sound_notifications: false,
        minimize_to_tray: true,
        start_minimized: false,
        theme: 'dark',
        language: 'en',
        data_directory: '',
        temp_directory: '',
        connection_timeout: 30,
        retry_attempts: 3,
        retry_delay: 5,
        user_agent: 'AB Download Manager 1.0',
        enable_logging: true,
        log_level: 'info'
      }
      this.isDirty = true
    },

    // Specific setting updates
    setDownloadPath(path) {
      this.updateSetting('default_download_path', path)
    },

    setMaxConcurrentDownloads(count) {
      this.updateSetting('max_concurrent_downloads', Math.max(1, Math.min(20, count)))
    },

    setGlobalSpeedLimit(limit) {
      this.updateSetting('global_speed_limit', Math.max(0, limit))
    },

    setTheme(theme) {
      this.updateSetting('theme', theme)
      
      // Apply theme to document
      if (theme === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    },

    toggleNotifications() {
      this.updateSetting('notifications_enabled', !this.settings.notifications_enabled)
    },

    toggleSoundNotifications() {
      this.updateSetting('sound_notifications', !this.settings.sound_notifications)
    },

    toggleMinimizeToTray() {
      this.updateSetting('minimize_to_tray', !this.settings.minimize_to_tray)
    },

    toggleAutoStartDownloads() {
      this.updateSetting('auto_start_downloads', !this.settings.auto_start_downloads)
    },

    // Validation helpers
    validateSettings() {
      const errors = []
      
      if (this.settings.max_concurrent_downloads < 1 || this.settings.max_concurrent_downloads > 20) {
        errors.push('Max concurrent downloads must be between 1 and 20')
      }
      
      if (this.settings.global_speed_limit < 0) {
        errors.push('Speed limit cannot be negative')
      }
      
      if (this.settings.connection_timeout < 5 || this.settings.connection_timeout > 300) {
        errors.push('Connection timeout must be between 5 and 300 seconds')
      }
      
      if (this.settings.retry_attempts < 0 || this.settings.retry_attempts > 10) {
        errors.push('Retry attempts must be between 0 and 10')
      }
      
      return errors
    },

    // Import/Export settings
    exportSettings() {
      return JSON.stringify(this.settings, null, 2)
    },

    importSettings(settingsJson) {
      try {
        const importedSettings = JSON.parse(settingsJson)
        
        // Validate imported settings
        const validKeys = Object.keys(this.settings)
        const filteredSettings = {}
        
        validKeys.forEach(key => {
          if (importedSettings.hasOwnProperty(key)) {
            filteredSettings[key] = importedSettings[key]
          }
        })
        
        this.updateSettings(filteredSettings)
        return true
      } catch (error) {
        this.error = 'Invalid settings format'
        return false
      }
    }
  }
})
