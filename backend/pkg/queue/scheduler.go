package queue

import (
	"context"
	"sync"
	"time"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// Scheduler handles automatic scheduling and execution of downloads in a queue
type Scheduler struct {
	queue   *models.QueueModel
	manager *Manager

	running bool
	mutex   sync.RWMutex

	stopChan   chan struct{}
	notifyChan chan struct{}
}

// NewScheduler creates a new scheduler for a queue
func NewScheduler(queue *models.QueueModel, manager *Manager) *Scheduler {
	return &Scheduler{
		queue:      queue,
		manager:    manager,
		running:    false,
		stopChan:   make(chan struct{}),
		notifyChan: make(chan struct{}, 1),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}
	s.running = true
	s.mutex.Unlock()

	// Main scheduler loop
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.stop()
			return
		case <-s.stopChan:
			s.stop()
			return
		case <-s.notifyChan:
			s.processQueue()
		case <-ticker.C:
			s.processQueue()
			s.checkSchedule()
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
}

// stop internal stop method
func (s *Scheduler) stop() {
	s.mutex.Lock()
	s.running = false
	s.mutex.Unlock()
}

// IsRunning returns true if the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// NotifyNewDownload notifies the scheduler of a new download
func (s *Scheduler) NotifyNewDownload() {
	select {
	case s.notifyChan <- struct{}{}:
	default:
		// Channel is full, notification will be processed on next tick
	}
}

// processQueue processes the queue and starts downloads as needed
func (s *Scheduler) processQueue() {
	// Get fresh queue state
	queue, err := s.manager.queueStorage.GetByID(s.queue.ID)
	if err != nil {
		return
	}
	s.queue = queue

	// Check if queue should be running
	if !s.queue.ShouldBeRunning() {
		return
	}

	// Start downloads up to the concurrent limit
	for s.queue.CanStartNewDownload() {
		nextDownloadID := s.queue.GetNextDownload()
		if nextDownloadID == 0 {
			break // No more downloads to start
		}

		// Get download details
		download, err := s.manager.downloadStorage.GetByID(nextDownloadID)
		if err != nil {
			// Remove invalid download from queue
			s.queue.RemoveDownload(nextDownloadID)
			s.manager.queueStorage.Update(s.queue)
			continue
		}

		// Check if download can be started
		if !download.CanStart() && !download.CanResume() {
			// Skip this download for now
			continue
		}

		// Try to start the download
		if err := s.manager.executeDownload(download); err != nil {
			// Failed to start, will try again later
			break
		}

		// Mark as active in queue
		s.queue.SetActive(nextDownloadID)
		s.manager.queueStorage.Update(s.queue)
	}
}

// checkSchedule checks if the queue should start/stop based on schedule
func (s *Scheduler) checkSchedule() {
	// Get fresh queue state
	queue, err := s.manager.queueStorage.GetByID(s.queue.ID)
	if err != nil {
		return
	}
	s.queue = queue

	if !s.queue.ScheduleEnabled {
		return
	}

	shouldRun := s.queue.ShouldBeRunning()
	isRunning := s.queue.IsRunning()

	if shouldRun && !isRunning {
		// Start the queue
		s.queue.Start()
		s.manager.queueStorage.Update(s.queue)
	} else if !shouldRun && isRunning {
		// Stop the queue (pause all downloads)
		s.pauseAllDownloads()
		s.queue.Pause()
		s.manager.queueStorage.Update(s.queue)
	}
}

// pauseAllDownloads pauses all active downloads in this queue
func (s *Scheduler) pauseAllDownloads() {
	activeDownloads := s.manager.GetActiveDownloads()

	for downloadID, active := range activeDownloads {
		if active.QueueID == s.queue.ID {
			s.manager.PauseDownload(downloadID)
		}
	}
}

// GetQueueStats returns statistics about the queue
func (s *Scheduler) GetQueueStats() *QueueStats {
	downloads, err := s.manager.downloadStorage.GetByQueue(s.queue.ID)
	if err != nil {
		return &QueueStats{}
	}

	stats := &QueueStats{
		QueueID:     s.queue.ID,
		QueueName:   s.queue.Name,
		TotalCount:  len(downloads),
		IsRunning:   s.queue.IsRunning(),
		ActiveCount: s.queue.ActiveCount(),
	}

	// Count downloads by status
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

		// Calculate total progress
		if download.TotalSize > 0 {
			stats.TotalBytes += download.TotalSize
			stats.DownloadedBytes += download.Progress.BytesDownloaded
		}
	}

	// Calculate overall progress percentage
	if stats.TotalBytes > 0 {
		stats.OverallProgress = float64(stats.DownloadedBytes) / float64(stats.TotalBytes) * 100
	}

	return stats
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

// PriorityScheduler extends the basic scheduler with priority-based scheduling
type PriorityScheduler struct {
	*Scheduler
}

// NewPriorityScheduler creates a new priority-based scheduler
func NewPriorityScheduler(queue *models.QueueModel, manager *Manager) *PriorityScheduler {
	return &PriorityScheduler{
		Scheduler: NewScheduler(queue, manager),
	}
}

// processQueue overrides the basic processQueue to consider priorities
func (ps *PriorityScheduler) processQueue() {
	// Get fresh queue state
	queue, err := ps.manager.queueStorage.GetByID(ps.queue.ID)
	if err != nil {
		return
	}
	ps.queue = queue

	// Check if queue should be running
	if !ps.queue.ShouldBeRunning() {
		return
	}

	// Get all downloads in queue and sort by priority
	downloads, err := ps.manager.downloadStorage.GetByQueue(ps.queue.ID)
	if err != nil {
		return
	}

	// Filter and sort downloads by priority
	var availableDownloads []*models.DownloadItem
	for _, download := range downloads {
		if (download.CanStart() || download.CanResume()) && !ps.queue.IsActive(download.ID) {
			availableDownloads = append(availableDownloads, download)
		}
	}

	// Sort by priority (higher number = higher priority)
	for i := 0; i < len(availableDownloads)-1; i++ {
		for j := i + 1; j < len(availableDownloads); j++ {
			if availableDownloads[i].Priority < availableDownloads[j].Priority {
				availableDownloads[i], availableDownloads[j] = availableDownloads[j], availableDownloads[i]
			}
		}
	}

	// Start downloads up to the concurrent limit
	for ps.queue.CanStartNewDownload() && len(availableDownloads) > 0 {
		download := availableDownloads[0]
		availableDownloads = availableDownloads[1:]

		// Try to start the download
		if err := ps.manager.executeDownload(download); err != nil {
			// Failed to start, will try again later
			break
		}

		// Mark as active in queue
		ps.queue.SetActive(download.ID)
		ps.manager.queueStorage.Update(ps.queue)
	}
}
