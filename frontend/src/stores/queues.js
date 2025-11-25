import { defineStore } from 'pinia'
import { queueAPI } from '../utils/api.js'

export const useQueueStore = defineStore('queues', {
  state: () => ({
    queues: [],
    currentQueue: null,
    loading: false,
    error: null,
    stats: new Map() // queueId -> stats
  }),

  getters: {
    activeQueues: (state) => {
      return state.queues.filter(queue => queue.is_running)
    },

    queueById: (state) => {
      return (queueId) => state.queues.find(queue => queue.id === queueId)
    },

    queueStats: (state) => {
      return (queueId) => state.stats.get(queueId)
    },

    totalActiveDownloads: (state) => {
      return Array.from(state.stats.values()).reduce((total, stats) => {
        return total + (stats?.active_count || 0)
      }, 0)
    },

    totalSpeed: (state) => {
      return Array.from(state.stats.values()).reduce((total, stats) => {
        return total + (stats?.total_speed || 0)
      }, 0)
    }
  },

  actions: {
    async fetchQueues() {
      this.loading = true
      this.error = null
      
      try {
        this.queues = await queueAPI.getAll()
        
        // Set default queue if none selected
        if (!this.currentQueue && this.queues.length > 0) {
          this.currentQueue = this.queues[0]
        }
        
        await this.fetchAllStats()
      } catch (error) {
        this.error = error.message
        console.error('Failed to fetch queues:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchAllStats() {
      try {
        const allStats = await queueAPI.getAllStats()
        
        // Clear existing stats
        this.stats.clear()
        
        // Update stats map
        allStats.forEach(stats => {
          this.stats.set(stats.queue_id, stats)
        })
      } catch (error) {
        console.error('Failed to fetch queue stats:', error)
      }
    },

    async fetchQueueStats(queueId) {
      try {
        const response = await queueAPI.getStats(queueId)
        if (response.success) {
          this.stats.set(queueId, response.stats)
        }
      } catch (error) {
        console.error('Failed to fetch queue stats:', error)
      }
    },

    async createQueue(queueRequest) {
      try {
        const response = await queueAPI.create(queueRequest)
        if (response.success) {
          this.queues.push(response.queue)
          
          // Set as current queue if it's the first one
          if (this.queues.length === 1) {
            this.currentQueue = response.queue
          }
          
          return response
        } else {
          throw new Error(response.error)
        }
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async updateQueue(queueRequest) {
      try {
        await queueAPI.update(queueRequest)
        
        // Update local queue data
        const queueIndex = this.queues.findIndex(q => q.id === queueRequest.queue_id)
        if (queueIndex !== -1) {
          // Fetch updated queue data
          const updatedQueue = await queueAPI.get(queueRequest.queue_id)
          this.queues[queueIndex] = updatedQueue
          
          // Update current queue if it's the one being updated
          if (this.currentQueue?.id === queueRequest.queue_id) {
            this.currentQueue = updatedQueue
          }
        }
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async deleteQueue(queueId) {
      try {
        await queueAPI.delete(queueId)
        
        // Remove from local state
        this.queues = this.queues.filter(q => q.id !== queueId)
        this.stats.delete(queueId)
        
        // Update current queue if deleted
        if (this.currentQueue?.id === queueId) {
          this.currentQueue = this.queues.length > 0 ? this.queues[0] : null
        }
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async startQueue(queueId) {
      try {
        await queueAPI.start(queueId)
        await this.updateQueueStatus(queueId, true)
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async pauseQueue(queueId) {
      try {
        await queueAPI.pause(queueId)
        await this.updateQueueStatus(queueId, false)
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async stopQueue(queueId) {
      try {
        await queueAPI.stop(queueId)
        await this.updateQueueStatus(queueId, false)
      } catch (error) {
        this.error = error.message
        throw error
      }
    },

    async updateQueueStatus(queueId, isRunning) {
      const queue = this.queues.find(q => q.id === queueId)
      if (queue) {
        queue.is_running = isRunning
        queue.updated_at = new Date().toISOString()
      }
      
      // Refresh stats
      await this.fetchQueueStats(queueId)
    },

    setCurrentQueue(queue) {
      this.currentQueue = queue
    },

    // Batch operations for all queues
    async startAllQueues() {
      const promises = this.queues.map(queue => this.startQueue(queue.id))
      await Promise.allSettled(promises)
    },

    async pauseAllQueues() {
      const promises = this.queues.map(queue => this.pauseQueue(queue.id))
      await Promise.allSettled(promises)
    },

    async stopAllQueues() {
      const promises = this.queues.map(queue => this.stopQueue(queue.id))
      await Promise.allSettled(promises)
    },

    // Helper methods
    getQueueDownloadCount(queueId) {
      const stats = this.stats.get(queueId)
      return stats?.total_count || 0
    },

    getQueueActiveCount(queueId) {
      const stats = this.stats.get(queueId)
      return stats?.active_count || 0
    },

    getQueueProgress(queueId) {
      const stats = this.stats.get(queueId)
      return stats?.overall_progress || 0
    },

    isQueueRunning(queueId) {
      const queue = this.queues.find(q => q.id === queueId)
      return queue?.is_running || false
    },

    canStartQueue(queueId) {
      const queue = this.queues.find(q => q.id === queueId)
      const stats = this.stats.get(queueId)
      return queue && !queue.is_running && (stats?.pending_count > 0 || stats?.paused_count > 0)
    },

    canPauseQueue(queueId) {
      const queue = this.queues.find(q => q.id === queueId)
      const stats = this.stats.get(queueId)
      return queue && queue.is_running && stats?.active_count > 0
    }
  }
})
