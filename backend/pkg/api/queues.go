package api

import (
	"fmt"

	"github.com/jamshidian/download-manager/backend/pkg/models"
	"github.com/jamshidian/download-manager/backend/pkg/queue"
)

// QueueAPI provides queue-related API methods for the frontend
type QueueAPI struct {
	manager *queue.Manager
}

// NewQueueAPI creates a new queue API instance
func NewQueueAPI(manager *queue.Manager) *QueueAPI {
	return &QueueAPI{
		manager: manager,
	}
}

// CreateQueueRequest represents a request to create a new queue
type CreateQueueRequest struct {
	Name            string `json:"name"`
	MaxConcurrent   int    `json:"max_concurrent,omitempty"`
	SpeedLimit      int64  `json:"speed_limit,omitempty"`
	AutoStart       *bool  `json:"auto_start,omitempty"`
	ScheduleEnabled *bool  `json:"schedule_enabled,omitempty"`
}

// CreateQueueResponse represents the response from creating a queue
type CreateQueueResponse struct {
	Success bool               `json:"success"`
	Error   string             `json:"error,omitempty"`
	Queue   *models.QueueModel `json:"queue,omitempty"`
}

// CreateQueue creates a new download queue
func (api *QueueAPI) CreateQueue(req *CreateQueueRequest) *CreateQueueResponse {
	if req.Name == "" {
		return &CreateQueueResponse{
			Success: false,
			Error:   "Queue name is required",
		}
	}

	queue, err := api.manager.CreateQueue(req.Name)
	if err != nil {
		return &CreateQueueResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create queue: %v", err),
		}
	}

	// Apply optional settings
	if req.MaxConcurrent > 0 {
		queue.MaxConcurrent = req.MaxConcurrent
	}
	if req.SpeedLimit > 0 {
		queue.SpeedLimit = req.SpeedLimit
	}
	if req.AutoStart != nil {
		queue.AutoStart = *req.AutoStart
	}
	if req.ScheduleEnabled != nil {
		queue.ScheduleEnabled = *req.ScheduleEnabled
	}

	// Update the queue with new settings
	if err := api.manager.UpdateQueue(queue); err != nil {
		return &CreateQueueResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update queue settings: %v", err),
		}
	}

	return &CreateQueueResponse{
		Success: true,
		Queue:   queue,
	}
}

// GetQueues returns all queues
func (api *QueueAPI) GetQueues() ([]*models.QueueModel, error) {
	return api.manager.GetQueues()
}

// GetQueue returns a specific queue by ID
func (api *QueueAPI) GetQueue(queueID int64) (*models.QueueModel, error) {
	return api.manager.GetQueue(queueID)
}

// UpdateQueueRequest represents a request to update queue settings
type UpdateQueueRequest struct {
	QueueID         int64  `json:"queue_id"`
	Name            string `json:"name,omitempty"`
	MaxConcurrent   *int   `json:"max_concurrent,omitempty"`
	SpeedLimit      *int64 `json:"speed_limit,omitempty"`
	AutoStart       *bool  `json:"auto_start,omitempty"`
	ScheduleEnabled *bool  `json:"schedule_enabled,omitempty"`
	ScheduleStart   string `json:"schedule_start,omitempty"` // Time format: "15:04"
	ScheduleStop    string `json:"schedule_stop,omitempty"`  // Time format: "15:04"
}

// UpdateQueue updates queue settings
func (api *QueueAPI) UpdateQueue(req *UpdateQueueRequest) error {
	queue, err := api.manager.GetQueue(req.QueueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		queue.Name = req.Name
	}
	if req.MaxConcurrent != nil {
		if *req.MaxConcurrent < 1 {
			return fmt.Errorf("max concurrent must be at least 1")
		}
		queue.MaxConcurrent = *req.MaxConcurrent
	}
	if req.SpeedLimit != nil {
		if *req.SpeedLimit < 0 {
			return fmt.Errorf("speed limit cannot be negative")
		}
		queue.SpeedLimit = *req.SpeedLimit
	}
	if req.AutoStart != nil {
		queue.AutoStart = *req.AutoStart
	}
	if req.ScheduleEnabled != nil {
		queue.ScheduleEnabled = *req.ScheduleEnabled
	}

	// Handle schedule times
	if req.ScheduleStart != "" {
		// Parse time and set schedule start
		// This would need proper time parsing implementation
	}
	if req.ScheduleStop != "" {
		// Parse time and set schedule stop
		// This would need proper time parsing implementation
	}

	return api.manager.UpdateQueue(queue)
}

// DeleteQueue deletes a queue
func (api *QueueAPI) DeleteQueue(queueID int64) error {
	// Check if queue has downloads
	downloads, err := api.manager.GetDownloadsByQueue(queueID)
	if err != nil {
		return fmt.Errorf("failed to check queue downloads: %w", err)
	}

	if len(downloads) > 0 {
		return fmt.Errorf("cannot delete queue with active downloads")
	}

	return api.manager.DeleteQueue(queueID)
}

// StartQueue starts a queue
func (api *QueueAPI) StartQueue(queueID int64) error {
	queue, err := api.manager.GetQueue(queueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	if !queue.CanStart() {
		return fmt.Errorf("queue cannot be started in current state")
	}

	queue.Start()
	return api.manager.UpdateQueue(queue)
}

// PauseQueue pauses a queue
func (api *QueueAPI) PauseQueue(queueID int64) error {
	queue, err := api.manager.GetQueue(queueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	if !queue.CanPause() {
		return fmt.Errorf("queue cannot be paused in current state")
	}

	queue.Pause()
	return api.manager.UpdateQueue(queue)
}

// StopQueue stops a queue
func (api *QueueAPI) StopQueue(queueID int64) error {
	queue, err := api.manager.GetQueue(queueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	queue.Stop()
	return api.manager.UpdateQueue(queue)
}

// QueueStats represents statistics for a queue
type QueueStats struct {
	QueueID          int64   `json:"queue_id"`
	QueueName        string  `json:"queue_name"`
	IsRunning        bool    `json:"is_running"`
	TotalCount       int     `json:"total_count"`
	ActiveCount      int     `json:"active_count"`
	PendingCount     int     `json:"pending_count"`
	DownloadingCount int     `json:"downloading_count"`
	CompletedCount   int     `json:"completed_count"`
	FailedCount      int     `json:"failed_count"`
	PausedCount      int     `json:"paused_count"`
	CanceledCount    int     `json:"canceled_count"`
	TotalBytes       int64   `json:"total_bytes"`
	DownloadedBytes  int64   `json:"downloaded_bytes"`
	OverallProgress  float64 `json:"overall_progress"`
}

// QueueStatsResponse represents queue statistics
type QueueStatsResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Stats   *QueueStats `json:"stats,omitempty"`
}

// GetQueueStats returns statistics for a specific queue
func (api *QueueAPI) GetQueueStats(queueID int64) *QueueStatsResponse {
	// This would need to be implemented in the queue manager
	// For now, return basic stats
	queue, err := api.manager.GetQueue(queueID)
	if err != nil {
		return &QueueStatsResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	downloads, err := api.manager.GetDownloadsByQueue(queueID)
	if err != nil {
		return &QueueStatsResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	stats := &QueueStats{
		QueueID:     queue.ID,
		QueueName:   queue.Name,
		IsRunning:   queue.IsRunning(),
		TotalCount:  len(downloads),
		ActiveCount: queue.ActiveCount(),
	}

	// Count by status
	for _, download := range downloads {
		switch download.Status {
		case models.StatusPending:
			stats.PendingCount++
		case models.StatusDownloading:
			stats.DownloadingCount++
		case models.StatusCompleted:
			stats.CompletedCount++
		case models.StatusFailed:
			stats.FailedCount++
		case models.StatusPaused:
			stats.PausedCount++
		case models.StatusCanceled:
			stats.CanceledCount++
		}

		stats.TotalBytes += download.TotalSize
		stats.DownloadedBytes += download.Progress.BytesDownloaded
	}

	if stats.TotalBytes > 0 {
		stats.OverallProgress = float64(stats.DownloadedBytes) / float64(stats.TotalBytes) * 100
	}

	return &QueueStatsResponse{
		Success: true,
		Stats:   stats,
	}
}

// GetAllQueueStats returns statistics for all queues
func (api *QueueAPI) GetAllQueueStats() ([]*QueueStats, error) {
	queues, err := api.manager.GetQueues()
	if err != nil {
		return nil, err
	}

	var allStats []*QueueStats
	for _, queue := range queues {
		statsResp := api.GetQueueStats(queue.ID)
		if statsResp.Success {
			allStats = append(allStats, statsResp.Stats)
		}
	}

	return allStats, nil
}

// MoveDownloadBetweenQueuesRequest represents a request to move download between queues
type MoveDownloadBetweenQueuesRequest struct {
	DownloadID  int64 `json:"download_id"`
	FromQueueID int64 `json:"from_queue_id"`
	ToQueueID   int64 `json:"to_queue_id"`
	NewPosition int   `json:"new_position,omitempty"`
}

// MoveDownloadBetweenQueues moves a download from one queue to another
func (api *QueueAPI) MoveDownloadBetweenQueues(req *MoveDownloadBetweenQueuesRequest) error {
	// Get the download
	download, err := api.manager.GetDownload(req.DownloadID)
	if err != nil {
		return fmt.Errorf("download not found: %w", err)
	}

	// Verify current queue
	if download.QueueID != req.FromQueueID {
		return fmt.Errorf("download is not in the specified source queue")
	}

	// Check if download is active (cannot move active downloads)
	activeDownloads := api.manager.GetActiveDownloads()
	if _, isActive := activeDownloads[req.DownloadID]; isActive {
		return fmt.Errorf("cannot move active download")
	}

	// Verify target queue exists
	_, err = api.manager.GetQueue(req.ToQueueID)
	if err != nil {
		return fmt.Errorf("target queue not found: %w", err)
	}

	// This would need to be implemented in the manager
	// For now, just update the download's queue ID
	download.QueueID = req.ToQueueID

	return nil
}

// QueueOrderRequest represents a request to reorder downloads in a queue
type QueueOrderRequest struct {
	QueueID     int64   `json:"queue_id"`
	DownloadIDs []int64 `json:"download_ids"` // New order of download IDs
}

// ReorderQueue reorders downloads in a queue
func (api *QueueAPI) ReorderQueue(req *QueueOrderRequest) error {
	queue, err := api.manager.GetQueue(req.QueueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	// Verify all downloads belong to this queue
	for _, downloadID := range req.DownloadIDs {
		download, err := api.manager.GetDownload(downloadID)
		if err != nil {
			return fmt.Errorf("download %d not found", downloadID)
		}
		if download.QueueID != req.QueueID {
			return fmt.Errorf("download %d does not belong to queue %d", downloadID, req.QueueID)
		}
	}

	// Update queue order
	queue.DownloadOrder = req.DownloadIDs
	return api.manager.UpdateQueue(queue)
}
