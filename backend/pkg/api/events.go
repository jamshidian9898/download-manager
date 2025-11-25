package api

import (
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// EventAPI handles real-time events for the frontend
type EventAPI struct {
	ctx         context.Context
	subscribers map[string][]chan interface{}
	mutex       sync.RWMutex
}

// NewEventAPI creates a new event API instance
func NewEventAPI(ctx context.Context) *EventAPI {
	return &EventAPI{
		ctx:         ctx,
		subscribers: make(map[string][]chan interface{}),
	}
}

// Event types
const (
	EventDownloadProgress  = "download:progress"
	EventDownloadStarted   = "download:started"
	EventDownloadCompleted = "download:completed"
	EventDownloadFailed    = "download:failed"
	EventDownloadPaused    = "download:paused"
	EventDownloadResumed   = "download:resumed"
	EventDownloadCanceled  = "download:canceled"
	EventDownloadAdded     = "download:added"
	EventDownloadRemoved   = "download:removed"

	EventQueueStarted = "queue:started"
	EventQueueStopped = "queue:stopped"
	EventQueuePaused  = "queue:paused"
	EventQueueResumed = "queue:resumed"
	EventQueueUpdated = "queue:updated"

	EventSystemStatus = "system:status"
	EventSystemError  = "system:error"
)

// DownloadProgressEvent represents a download progress update
type DownloadProgressEvent struct {
	DownloadID      int64     `json:"download_id"`
	BytesDownloaded int64     `json:"bytes_downloaded"`
	TotalBytes      int64     `json:"total_bytes"`
	Speed           int64     `json:"speed"`
	Percentage      float64   `json:"percentage"`
	ETA             int64     `json:"eta"`
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
}

// DownloadStatusEvent represents a download status change
type DownloadStatusEvent struct {
	DownloadID int64                `json:"download_id"`
	Download   *models.DownloadItem `json:"download"`
	OldStatus  string               `json:"old_status"`
	NewStatus  string               `json:"new_status"`
	Error      string               `json:"error,omitempty"`
	Timestamp  time.Time            `json:"timestamp"`
}

// QueueStatusEvent represents a queue status change
type QueueStatusEvent struct {
	QueueID   int64              `json:"queue_id"`
	Queue     *models.QueueModel `json:"queue"`
	OldStatus string             `json:"old_status"`
	NewStatus string             `json:"new_status"`
	Timestamp time.Time          `json:"timestamp"`
}

// SystemStatusEvent represents system-wide status updates
type SystemStatusEvent struct {
	ActiveDownloads int       `json:"active_downloads"`
	TotalSpeed      int64     `json:"total_speed"`
	MemoryUsage     int64     `json:"memory_usage"`
	Timestamp       time.Time `json:"timestamp"`
}

// EmitDownloadProgress emits a download progress event
func (api *EventAPI) EmitDownloadProgress(downloadID int64, progress *models.Progress) {
	event := &DownloadProgressEvent{
		DownloadID:      downloadID,
		BytesDownloaded: progress.BytesDownloaded,
		TotalBytes:      progress.TotalBytes,
		Speed:           progress.Speed,
		Percentage:      progress.Percentage,
		ETA:             progress.ETA,
		Timestamp:       time.Now(),
	}

	// Emit to Wails frontend
	runtime.EventsEmit(api.ctx, EventDownloadProgress, event)

	// Emit to internal subscribers
	api.emit(EventDownloadProgress, event)
}

// EmitDownloadStatusChange emits a download status change event
func (api *EventAPI) EmitDownloadStatusChange(download *models.DownloadItem, oldStatus, newStatus models.DownloadStatus, err error) {
	event := &DownloadStatusEvent{
		DownloadID: download.ID,
		Download:   download,
		OldStatus:  oldStatus.String(),
		NewStatus:  newStatus.String(),
		Timestamp:  time.Now(),
	}

	if err != nil {
		event.Error = err.Error()
	}

	eventType := ""
	switch newStatus {
	case models.StatusDownloading:
		if oldStatus == models.StatusPaused {
			eventType = EventDownloadResumed
		} else {
			eventType = EventDownloadStarted
		}
	case models.StatusCompleted:
		eventType = EventDownloadCompleted
	case models.StatusFailed:
		eventType = EventDownloadFailed
	case models.StatusPaused:
		eventType = EventDownloadPaused
	case models.StatusCanceled:
		eventType = EventDownloadCanceled
	}

	if eventType != "" {
		runtime.EventsEmit(api.ctx, eventType, event)
		api.emit(eventType, event)
	}
}

// EmitDownloadAdded emits a download added event
func (api *EventAPI) EmitDownloadAdded(download *models.DownloadItem) {
	event := &DownloadStatusEvent{
		DownloadID: download.ID,
		Download:   download,
		NewStatus:  download.Status.String(),
		Timestamp:  time.Now(),
	}

	runtime.EventsEmit(api.ctx, EventDownloadAdded, event)
	api.emit(EventDownloadAdded, event)
}

// EmitDownloadRemoved emits a download removed event
func (api *EventAPI) EmitDownloadRemoved(downloadID int64) {
	event := map[string]interface{}{
		"download_id": downloadID,
		"timestamp":   time.Now(),
	}

	runtime.EventsEmit(api.ctx, EventDownloadRemoved, event)
	api.emit(EventDownloadRemoved, event)
}

// EmitQueueStatusChange emits a queue status change event
func (api *EventAPI) EmitQueueStatusChange(queue *models.QueueModel, oldStatus, newStatus models.QueueStatus) {
	event := &QueueStatusEvent{
		QueueID:   queue.ID,
		Queue:     queue,
		OldStatus: oldStatus.String(),
		NewStatus: newStatus.String(),
		Timestamp: time.Now(),
	}

	eventType := ""
	switch newStatus {
	case models.QueueStatusRunning:
		if oldStatus == models.QueueStatusPaused {
			eventType = EventQueueResumed
		} else {
			eventType = EventQueueStarted
		}
	case models.QueueStatusStopped:
		eventType = EventQueueStopped
	case models.QueueStatusPaused:
		eventType = EventQueuePaused
	}

	if eventType != "" {
		runtime.EventsEmit(api.ctx, eventType, event)
		api.emit(eventType, event)
	}

	// Always emit general queue updated event
	runtime.EventsEmit(api.ctx, EventQueueUpdated, event)
	api.emit(EventQueueUpdated, event)
}

// EmitSystemStatus emits system status updates
func (api *EventAPI) EmitSystemStatus(activeDownloads int, totalSpeed int64, memoryUsage int64) {
	event := &SystemStatusEvent{
		ActiveDownloads: activeDownloads,
		TotalSpeed:      totalSpeed,
		MemoryUsage:     memoryUsage,
		Timestamp:       time.Now(),
	}

	runtime.EventsEmit(api.ctx, EventSystemStatus, event)
	api.emit(EventSystemStatus, event)
}

// EmitSystemError emits system error events
func (api *EventAPI) EmitSystemError(message string, err error) {
	event := map[string]interface{}{
		"message":   message,
		"error":     err.Error(),
		"timestamp": time.Now(),
	}

	runtime.EventsEmit(api.ctx, EventSystemError, event)
	api.emit(EventSystemError, event)
}

// Subscribe subscribes to events (for internal use)
func (api *EventAPI) Subscribe(eventType string) <-chan interface{} {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	ch := make(chan interface{}, 100) // Buffered channel
	api.subscribers[eventType] = append(api.subscribers[eventType], ch)
	return ch
}

// Unsubscribe removes a subscription (for internal use)
func (api *EventAPI) Unsubscribe(eventType string, ch <-chan interface{}) {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	subscribers := api.subscribers[eventType]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			// Remove from slice
			api.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
			close(subscriber)
			break
		}
	}
}

// emit sends events to internal subscribers
func (api *EventAPI) emit(eventType string, data interface{}) {
	api.mutex.RLock()
	subscribers := api.subscribers[eventType]
	api.mutex.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- data:
		default:
			// Channel is full, skip this subscriber
		}
	}
}

// GetEventHistory returns recent events (for debugging/monitoring)
type EventHistoryItem struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

var (
	eventHistory []EventHistoryItem
	historyMutex sync.RWMutex
	maxHistory   = 1000
)

// addToHistory adds an event to the history
func addToHistory(eventType string, data interface{}) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	item := EventHistoryItem{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	eventHistory = append(eventHistory, item)

	// Keep only the last maxHistory items
	if len(eventHistory) > maxHistory {
		eventHistory = eventHistory[len(eventHistory)-maxHistory:]
	}
}

// GetEventHistory returns recent events
func (api *EventAPI) GetEventHistory(limit int) []EventHistoryItem {
	historyMutex.RLock()
	defer historyMutex.RUnlock()

	if limit <= 0 || limit > len(eventHistory) {
		limit = len(eventHistory)
	}

	// Return the last 'limit' items
	start := len(eventHistory) - limit
	result := make([]EventHistoryItem, limit)
	copy(result, eventHistory[start:])

	return result
}

// ClearEventHistory clears the event history
func (api *EventAPI) ClearEventHistory() {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	eventHistory = eventHistory[:0]
}

// StartSystemMonitor starts monitoring system status
func (api *EventAPI) StartSystemMonitor(activeDownloadsFunc func() int, totalSpeedFunc func() int64) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-api.ctx.Done():
				return
			case <-ticker.C:
				activeDownloads := activeDownloadsFunc()
				totalSpeed := totalSpeedFunc()

				// Get memory usage (simplified)
				var memoryUsage int64 = 0 // This would need proper implementation

				api.EmitSystemStatus(activeDownloads, totalSpeed, memoryUsage)
			}
		}
	}()
}
