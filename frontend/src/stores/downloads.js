import { defineStore } from 'pinia'
import { downloadAPI } from '../utils/api.js'
import { eventManager } from '../utils/events.js'

export const useDownloadStore = defineStore('downloads', {
  state: () => ({
    downloads: [],
    activeDownloads: new Map(),
    selectedDownloads: new Set(),
    loading: false,
    error: null,
    stats: {
      total_downloads: 0,
      active_downloads: 0,
      completed_downloads: 0,
      failed_downloads: 0,
      paused_downloads: 0,
      total_bytes: 0,
      downloaded_bytes: 0,
      total_speed: 0
    },
    filters: {
      status: 'all', // all, downloading, completed, failed, paused
      category: 'all', // all, image, video, audio, document, compressed, apps, other
      search: ''
    },
    sortBy: 'date_added',
    sortOrder: 'desc'
  }),

  getters: {
    filteredDownloads: (state) => {
      let filtered = [...state.downloads]

      // Apply status filter
      if (state.filters.status !== 'all') {
        filtered = filtered.filter(download => download.status === state.filters.status)
      }

      // Apply category filter
      if (state.filters.category !== 'all') {
        filtered = filtered.filter(download => {
          const category = getCategoryFromFilename(download.filename)
          return category === state.filters.category
        })
      }

      // Apply search filter
      if (state.filters.search) {
        const search = state.filters.search.toLowerCase()
        filtered = filtered.filter(download => 
          download.filename.toLowerCase().includes(search) ||
          download.url.toLowerCase().includes(search)
        )
      }

      // Apply sorting
      filtered.sort((a, b) => {
        let aValue, bValue

        switch (state.sortBy) {
          case 'name':
            aValue = a.filename.toLowerCase()
            bValue = b.filename.toLowerCase()
            break
          case 'size':
            aValue = a.total_size || 0
            bValue = b.total_size || 0
            break
          case 'progress':
            aValue = a.progress?.bytes_downloaded || 0
            bValue = b.progress?.bytes_downloaded || 0
            break
          case 'speed':
            aValue = a.progress?.speed || 0
            bValue = b.progress?.speed || 0
            break
          case 'status':
            aValue = a.status
            bValue = b.status
            break
          case 'date_added':
          default:
            aValue = new Date(a.created_at)
            bValue = new Date(b.created_at)
            break
        }

        if (aValue < bValue) return state.sortOrder === 'asc' ? -1 : 1
        if (aValue > bValue) return state.sortOrder === 'asc' ? 1 : -1
        return 0
      })

      return filtered
    },

    downloadsByStatus: (state) => {
      const byStatus = {}
      state.downloads.forEach(download => {
        if (!byStatus[download.status]) {
          byStatus[download.status] = []
        }
        byStatus[download.status].push(download)
      })
      return byStatus
    },

    downloadsByCategory: (state) => {
      const byCategory = {}
      state.downloads.forEach(download => {
        const category = getCategoryFromFilename(download.filename)
        if (!byCategory[category]) {
          byCategory[category] = []
        }
        byCategory[category].push(download)
      })
      return byCategory
    },

    selectedDownloadsList: (state) => {
      return state.downloads.filter(download => 
        state.selectedDownloads.has(download.id)
      )
    }
  },

  actions: {
    async fetchDownloads() {
      this.loading = true
      this.error = null
      
      try {
        this.downloads = await downloadAPI.getAll()
        await this.fetchStats()
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch downloads:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchStats() {
      try {
        this.stats = await downloadAPI.getStats()
      } catch (error) {
        console.error('Failed to fetch download stats:', error)
      }
    },

    async addDownload(downloadRequest) {
      try {
        const response = await downloadAPI.add(downloadRequest)
        if (response.success) {
          this.downloads.unshift(response.download)
          await this.fetchStats()
          return response
        } else {
          throw new Error(response.error)
        }
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async startDownload(downloadId) {
      try {
        await downloadAPI.start(downloadId)
        await this.updateDownloadStatus(downloadId, 'downloading')
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async pauseDownload(downloadId) {
      try {
        await downloadAPI.pause(downloadId)
        await this.updateDownloadStatus(downloadId, 'paused')
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async resumeDownload(downloadId) {
      try {
        await downloadAPI.resume(downloadId)
        await this.updateDownloadStatus(downloadId, 'downloading')
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async cancelDownload(downloadId) {
      try {
        await downloadAPI.cancel(downloadId)
        await this.updateDownloadStatus(downloadId, 'canceled')
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async deleteDownload(downloadId) {
      try {
        await downloadAPI.delete(downloadId)
        this.downloads = this.downloads.filter(d => d.id !== downloadId)
        this.selectedDownloads.delete(downloadId)
        await this.fetchStats()
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async updateDownloadStatus(downloadId, status) {
      const download = this.downloads.find(d => d.id === downloadId)
      if (download) {
        download.status = status
        download.updated_at = new Date().toISOString()
      }
      await this.fetchStats()
    },

    updateDownloadProgress(downloadId, progress) {
      const download = this.downloads.find(d => d.id === downloadId)
      if (download) {
        download.progress = progress
        download.updated_at = new Date().toISOString()
      }
    },

    // Selection management
    selectDownload(downloadId) {
      this.selectedDownloads.add(downloadId)
    },

    deselectDownload(downloadId) {
      this.selectedDownloads.delete(downloadId)
    },

    toggleDownloadSelection(downloadId) {
      if (this.selectedDownloads.has(downloadId)) {
        this.selectedDownloads.delete(downloadId)
      } else {
        this.selectedDownloads.add(downloadId)
      }
    },

    selectAllDownloads() {
      this.filteredDownloads.forEach(download => {
        this.selectedDownloads.add(download.id)
      })
    },

    deselectAllDownloads() {
      this.selectedDownloads.clear()
    },

    // Batch operations
    async startSelectedDownloads() {
      const promises = Array.from(this.selectedDownloads).map(id => 
        this.startDownload(id)
      )
      await Promise.allSettled(promises)
    },

    async pauseSelectedDownloads() {
      const promises = Array.from(this.selectedDownloads).map(id => 
        this.pauseDownload(id)
      )
      await Promise.allSettled(promises)
    },

    async deleteSelectedDownloads() {
      const promises = Array.from(this.selectedDownloads).map(id => 
        this.deleteDownload(id)
      )
      await Promise.allSettled(promises)
      this.selectedDownloads.clear()
    },

    // Filters and sorting
    setStatusFilter(status) {
      this.filters.status = status
    },

    setCategoryFilter(category) {
      this.filters.category = category
    },

    setSearchFilter(search) {
      this.filters.search = search
    },

    setSorting(sortBy, sortOrder = null) {
      if (this.sortBy === sortBy && sortOrder === null) {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc'
      } else {
        this.sortBy = sortBy
        this.sortOrder = sortOrder || 'asc'
      }
    },

    clearFilters() {
      this.filters = {
        status: 'all',
        category: 'all',
        search: ''
      }
    },

    // Event handling
    initializeEventListeners() {
      // Listen for download progress updates
      eventManager.onDownloadProgress((data) => {
        this.updateDownloadProgress(data.download_id, data.progress)
      })

      // Listen for download status changes
      eventManager.onDownloadStatusChange((data) => {
        this.updateDownloadStatus(data.download_id, data.new_status)
        if (data.download) {
          // Update the full download object if provided
          const index = this.downloads.findIndex(d => d.id === data.download_id)
          if (index !== -1) {
            this.downloads[index] = { ...this.downloads[index], ...data.download }
          }
        }
      })

      // Listen for download completion
      eventManager.onDownloadComplete((data) => {
        this.updateDownloadStatus(data.download_id, 'completed')
        this.fetchStats() // Refresh stats
      })

      // Listen for download failures
      eventManager.onDownloadFailed((data) => {
        this.updateDownloadStatus(data.download_id, 'failed')
        this.fetchStats() // Refresh stats
      })

      // Listen for system stats updates
      eventManager.onSystemStats((data) => {
        if (data.download_stats) {
          this.stats = { ...this.stats, ...data.download_stats }
        }
      })
    },

    destroyEventListeners() {
      // Event manager will handle cleanup
      eventManager.destroy()
    }
  }
})

// Helper function to determine file category
function getCategoryFromFilename(filename) {
  const ext = filename.split('.').pop()?.toLowerCase()
  
  const categories = {
    image: ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg', 'tiff', 'ico'],
    video: ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm', 'm4v', '3gp'],
    audio: ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a'],
    document: ['pdf', 'doc', 'docx', 'txt', 'rtf', 'odt', 'xls', 'xlsx', 'ppt', 'pptx'],
    compressed: ['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz'],
    apps: ['exe', 'msi', 'dmg', 'pkg', 'deb', 'rpm', 'appimage']
  }
  
  for (const [category, extensions] of Object.entries(categories)) {
    if (extensions.includes(ext)) {
      return category
    }
  }
  
  return 'other'
}
