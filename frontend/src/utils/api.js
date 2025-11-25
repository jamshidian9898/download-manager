import { 
  GetDownloadInfo, 
  AddDownload, 
  GetDownloads, 
  GetDownload,
  StartDownload,
  PauseDownload,
  ResumeDownload,
  CancelDownload,
  DeleteDownload,
  GetDownloadStats,
  CreateQueue,
  GetQueues,
  GetQueue,
  UpdateQueue,
  DeleteQueue,
  StartQueue,
  PauseQueue,
  StopQueue,
  GetQueueStats,
  GetAllQueueStats,
  GetSystemInfo,
  GetSettings,
  UpdateSettings
} from '../../wailsjs/go/main/App.js'

// Download API wrapper
export const downloadAPI = {
  async getInfo(url, headers = {}) {
    try {
      return await GetDownloadInfo(url, headers)
    } catch (error) {
      console.error('Failed to get download info:', error)
      throw error
    }
  },

  async add(request) {
    try {
      return await AddDownload(request)
    } catch (error) {
      console.error('Failed to add download:', error)
      throw error
    }
  },

  async getAll() {
    try {
      return await GetDownloads()
    } catch (error) {
      console.error('Failed to get downloads:', error)
      throw error
    }
  },

  async get(downloadId) {
    try {
      return await GetDownload(downloadId)
    } catch (error) {
      console.error('Failed to get download:', error)
      throw error
    }
  },

  async start(downloadId) {
    try {
      return await StartDownload(downloadId)
    } catch (error) {
      console.error('Failed to start download:', error)
      throw error
    }
  },

  async pause(downloadId) {
    try {
      return await PauseDownload(downloadId)
    } catch (error) {
      console.error('Failed to pause download:', error)
      throw error
    }
  },

  async resume(downloadId) {
    try {
      return await ResumeDownload(downloadId)
    } catch (error) {
      console.error('Failed to resume download:', error)
      throw error
    }
  },

  async cancel(downloadId) {
    try {
      return await CancelDownload(downloadId)
    } catch (error) {
      console.error('Failed to cancel download:', error)
      throw error
    }
  },

  async delete(downloadId) {
    try {
      return await DeleteDownload(downloadId)
    } catch (error) {
      console.error('Failed to delete download:', error)
      throw error
    }
  },

  async getStats() {
    try {
      return await GetDownloadStats()
    } catch (error) {
      console.error('Failed to get download stats:', error)
      throw error
    }
  }
}

// Queue API wrapper
export const queueAPI = {
  async create(request) {
    try {
      return await CreateQueue(request)
    } catch (error) {
      console.error('Failed to create queue:', error)
      throw error
    }
  },

  async getAll() {
    try {
      return await GetQueues()
    } catch (error) {
      console.error('Failed to get queues:', error)
      throw error
    }
  },

  async get(queueId) {
    try {
      return await GetQueue(queueId)
    } catch (error) {
      console.error('Failed to get queue:', error)
      throw error
    }
  },

  async update(request) {
    try {
      return await UpdateQueue(request)
    } catch (error) {
      console.error('Failed to update queue:', error)
      throw error
    }
  },

  async delete(queueId) {
    try {
      return await DeleteQueue(queueId)
    } catch (error) {
      console.error('Failed to delete queue:', error)
      throw error
    }
  },

  async start(queueId) {
    try {
      return await StartQueue(queueId)
    } catch (error) {
      console.error('Failed to start queue:', error)
      throw error
    }
  },

  async pause(queueId) {
    try {
      return await PauseQueue(queueId)
    } catch (error) {
      console.error('Failed to pause queue:', error)
      throw error
    }
  },

  async stop(queueId) {
    try {
      return await StopQueue(queueId)
    } catch (error) {
      console.error('Failed to stop queue:', error)
      throw error
    }
  },

  async getStats(queueId) {
    try {
      return await GetQueueStats(queueId)
    } catch (error) {
      console.error('Failed to get queue stats:', error)
      throw error
    }
  },

  async getAllStats() {
    try {
      return await GetAllQueueStats()
    } catch (error) {
      console.error('Failed to get all queue stats:', error)
      throw error
    }
  }
}

// System API wrapper
export const systemAPI = {
  async getInfo() {
    try {
      return await GetSystemInfo()
    } catch (error) {
      console.error('Failed to get system info:', error)
      throw error
    }
  },

  async getSettings() {
    try {
      return await GetSettings()
    } catch (error) {
      console.error('Failed to get settings:', error)
      throw error
    }
  },

  async updateSettings(settings) {
    try {
      return await UpdateSettings(settings)
    } catch (error) {
      console.error('Failed to update settings:', error)
      throw error
    }
  }
}

// Helper functions
export const formatters = {
  formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
  },

  formatSpeed(bytesPerSecond) {
    return this.formatBytes(bytesPerSecond) + '/s'
  },

  formatDuration(seconds) {
    if (!seconds || seconds < 0) return '--'
    
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    
    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`
    } else {
      return `${secs}s`
    }
  },

  formatProgress(downloaded, total) {
    if (!total || total === 0) return 0
    return Math.round((downloaded / total) * 100)
  },

  getStatusColor(status) {
    const colors = {
      'pending': 'text-yellow-500',
      'downloading': 'text-blue-500',
      'paused': 'text-orange-500',
      'completed': 'text-green-500',
      'failed': 'text-red-500',
      'canceled': 'text-gray-500'
    }
    return colors[status] || 'text-gray-500'
  },

  getStatusIcon(status) {
    const icons = {
      'pending': 'Clock',
      'downloading': 'Download',
      'paused': 'Pause',
      'completed': 'CheckCircle',
      'failed': 'XCircle',
      'canceled': 'StopCircle'
    }
    return icons[status] || 'Circle'
  }
}
