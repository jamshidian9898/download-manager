package models

import (
	"time"
)

// QueueModel represents a download queue with scheduling capabilities
type QueueModel struct {
	ID              int64       `json:"id"`
	Name            string      `json:"name"`
	Status          QueueStatus `json:"status"`
	MaxConcurrent   int         `json:"max_concurrent"`           // Maximum concurrent downloads
	SpeedLimit      int64       `json:"speed_limit"`              // Global speed limit for this queue (bytes/sec)
	AutoStart       bool        `json:"auto_start"`               // Automatically start downloads when added
	ScheduleEnabled bool        `json:"schedule_enabled"`         // Enable scheduled start/stop
	ScheduleStart   *time.Time  `json:"schedule_start,omitempty"` // Scheduled start time
	ScheduleStop    *time.Time  `json:"schedule_stop,omitempty"`  // Scheduled stop time
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// Runtime fields (not persisted)
	ActiveDownloads map[int64]bool `json:"-"`              // Currently active download IDs
	DownloadOrder   []int64        `json:"download_order"` // Order of downloads in queue
}

// NewQueueModel creates a new queue with default settings
func NewQueueModel(name string) *QueueModel {
	return &QueueModel{
		Name:            name,
		Status:          QueueStatusStopped,
		MaxConcurrent:   3,
		SpeedLimit:      0, // Unlimited
		AutoStart:       true,
		ScheduleEnabled: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ActiveDownloads: make(map[int64]bool),
		DownloadOrder:   make([]int64, 0),
	}
}

// IsRunning returns true if the queue is currently running
func (q *QueueModel) IsRunning() bool {
	return q.Status == QueueStatusRunning
}

// IsPaused returns true if the queue is paused
func (q *QueueModel) IsPaused() bool {
	return q.Status == QueueStatusPaused
}

// IsStopped returns true if the queue is stopped
func (q *QueueModel) IsStopped() bool {
	return q.Status == QueueStatusStopped
}

// CanStart returns true if the queue can be started
func (q *QueueModel) CanStart() bool {
	return q.Status == QueueStatusStopped || q.Status == QueueStatusPaused
}

// CanPause returns true if the queue can be paused
func (q *QueueModel) CanPause() bool {
	return q.Status == QueueStatusRunning
}

// Start starts the queue
func (q *QueueModel) Start() {
	q.Status = QueueStatusRunning
	q.UpdatedAt = time.Now()
}

// Pause pauses the queue
func (q *QueueModel) Pause() {
	q.Status = QueueStatusPaused
	q.UpdatedAt = time.Now()
}

// Stop stops the queue
func (q *QueueModel) Stop() {
	q.Status = QueueStatusStopped
	q.UpdatedAt = time.Now()
}

// ActiveCount returns the number of currently active downloads
func (q *QueueModel) ActiveCount() int {
	return len(q.ActiveDownloads)
}

// CanStartNewDownload returns true if a new download can be started
func (q *QueueModel) CanStartNewDownload() bool {
	return q.IsRunning() && q.ActiveCount() < q.MaxConcurrent
}

// AddDownload adds a download to the queue order
func (q *QueueModel) AddDownload(downloadID int64) {
	// Check if already in queue
	for _, id := range q.DownloadOrder {
		if id == downloadID {
			return
		}
	}
	q.DownloadOrder = append(q.DownloadOrder, downloadID)
	q.UpdatedAt = time.Now()
}

// RemoveDownload removes a download from the queue order
func (q *QueueModel) RemoveDownload(downloadID int64) {
	for i, id := range q.DownloadOrder {
		if id == downloadID {
			q.DownloadOrder = append(q.DownloadOrder[:i], q.DownloadOrder[i+1:]...)
			break
		}
	}
	delete(q.ActiveDownloads, downloadID)
	q.UpdatedAt = time.Now()
}

// SetActive marks a download as active
func (q *QueueModel) SetActive(downloadID int64) {
	q.ActiveDownloads[downloadID] = true
	q.UpdatedAt = time.Now()
}

// SetInactive marks a download as inactive
func (q *QueueModel) SetInactive(downloadID int64) {
	delete(q.ActiveDownloads, downloadID)
	q.UpdatedAt = time.Now()
}

// IsActive returns true if the download is currently active in this queue
func (q *QueueModel) IsActive(downloadID int64) bool {
	return q.ActiveDownloads[downloadID]
}

// GetNextDownload returns the next download ID to start, or 0 if none available
func (q *QueueModel) GetNextDownload() int64 {
	if !q.CanStartNewDownload() {
		return 0
	}

	for _, downloadID := range q.DownloadOrder {
		if !q.IsActive(downloadID) {
			return downloadID
		}
	}
	return 0
}

// ShouldBeRunning returns true if the queue should be running based on schedule
func (q *QueueModel) ShouldBeRunning() bool {
	if !q.ScheduleEnabled {
		return q.IsRunning()
	}

	now := time.Now()

	// If both start and stop times are set
	if q.ScheduleStart != nil && q.ScheduleStop != nil {
		start := *q.ScheduleStart
		stop := *q.ScheduleStop

		// Handle same-day schedule
		if start.Before(stop) {
			return now.After(start) && now.Before(stop)
		}

		// Handle overnight schedule (start > stop)
		return now.After(start) || now.Before(stop)
	}

	// If only start time is set
	if q.ScheduleStart != nil {
		return now.After(*q.ScheduleStart)
	}

	// If only stop time is set
	if q.ScheduleStop != nil {
		return now.Before(*q.ScheduleStop)
	}

	return q.IsRunning()
}

// MoveDownload moves a download to a new position in the queue
func (q *QueueModel) MoveDownload(downloadID int64, newPosition int) {
	// Remove from current position
	oldIndex := -1
	for i, id := range q.DownloadOrder {
		if id == downloadID {
			oldIndex = i
			break
		}
	}

	if oldIndex == -1 {
		return // Download not found
	}

	// Remove from old position
	q.DownloadOrder = append(q.DownloadOrder[:oldIndex], q.DownloadOrder[oldIndex+1:]...)

	// Insert at new position
	if newPosition >= len(q.DownloadOrder) {
		q.DownloadOrder = append(q.DownloadOrder, downloadID)
	} else {
		q.DownloadOrder = append(q.DownloadOrder[:newPosition], append([]int64{downloadID}, q.DownloadOrder[newPosition:]...)...)
	}

	q.UpdatedAt = time.Now()
}
