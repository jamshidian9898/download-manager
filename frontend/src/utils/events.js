import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js'

// Event types that the backend emits
export const EVENT_TYPES = {
  DOWNLOAD_PROGRESS: 'download:progress',
  DOWNLOAD_STATUS_CHANGE: 'download:status_change',
  DOWNLOAD_COMPLETE: 'download:complete',
  DOWNLOAD_FAILED: 'download:failed',
  QUEUE_STATUS_CHANGE: 'queue:status_change',
  SYSTEM_STATS: 'system:stats'
}

class EventManager {
  constructor() {
    this.listeners = new Map()
  }

  // Subscribe to an event
  on(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set())
      
      // Set up Wails event listener
      EventsOn(eventType, (data) => {
        const callbacks = this.listeners.get(eventType)
        if (callbacks) {
          callbacks.forEach(callback => {
            try {
              callback(data)
            } catch (error) {
              console.error(`Error in event callback for ${eventType}:`, error)
            }
          })
        }
      })
    }
    
    this.listeners.get(eventType).add(callback)
    
    // Return unsubscribe function
    return () => this.off(eventType, callback)
  }

  // Unsubscribe from an event
  off(eventType, callback) {
    const callbacks = this.listeners.get(eventType)
    if (callbacks) {
      callbacks.delete(callback)
      
      // If no more callbacks, remove Wails listener
      if (callbacks.size === 0) {
        EventsOff(eventType)
        this.listeners.delete(eventType)
      }
    }
  }

  // Subscribe to download progress updates
  onDownloadProgress(callback) {
    return this.on(EVENT_TYPES.DOWNLOAD_PROGRESS, callback)
  }

  // Subscribe to download status changes
  onDownloadStatusChange(callback) {
    return this.on(EVENT_TYPES.DOWNLOAD_STATUS_CHANGE, callback)
  }

  // Subscribe to download completion
  onDownloadComplete(callback) {
    return this.on(EVENT_TYPES.DOWNLOAD_COMPLETE, callback)
  }

  // Subscribe to download failures
  onDownloadFailed(callback) {
    return this.on(EVENT_TYPES.DOWNLOAD_FAILED, callback)
  }

  // Subscribe to queue status changes
  onQueueStatusChange(callback) {
    return this.on(EVENT_TYPES.QUEUE_STATUS_CHANGE, callback)
  }

  // Subscribe to system stats updates
  onSystemStats(callback) {
    return this.on(EVENT_TYPES.SYSTEM_STATS, callback)
  }

  // Clean up all listeners
  destroy() {
    for (const eventType of this.listeners.keys()) {
      EventsOff(eventType)
    }
    this.listeners.clear()
  }
}

// Export singleton instance
export const eventManager = new EventManager()

// Composable for Vue components
export function useEvents() {
  return {
    eventManager,
    EVENT_TYPES,
    
    // Convenience methods
    onDownloadProgress: (callback) => eventManager.onDownloadProgress(callback),
    onDownloadStatusChange: (callback) => eventManager.onDownloadStatusChange(callback),
    onDownloadComplete: (callback) => eventManager.onDownloadComplete(callback),
    onDownloadFailed: (callback) => eventManager.onDownloadFailed(callback),
    onQueueStatusChange: (callback) => eventManager.onQueueStatusChange(callback),
    onSystemStats: (callback) => eventManager.onSystemStats(callback)
  }
}
