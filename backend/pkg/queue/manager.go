package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jamshidian/download-manager/backend/pkg/downloader"
	"github.com/jamshidian/download-manager/backend/pkg/models"
	"github.com/jamshidian/download-manager/backend/pkg/storage"
)

// Manager handles all download queues and their execution
type Manager struct {
	downloadStorage *storage.DownloadStorage
	partStorage     *storage.PartStorage
	queueStorage    *storage.QueueStorage

	resumeManager   *downloader.ResumeManager
	multiDownloader *downloader.MultiPartDownloader

	activeDownloads map[int64]*ActiveDownload
	schedulers      map[int64]*Scheduler

	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	globalSpeedLimit   int64
	maxGlobalDownloads int

	// Event callbacks
	onDownloadProgress func(downloadID int64, progress *models.Progress)
	onDownloadComplete func(downloadID int64)
	onDownloadFailed   func(downloadID int64, err error)
}

// ActiveDownload represents a currently running download
type ActiveDownload struct {
	Download *models.DownloadItem
	Parts    []*models.DownloadPart
	Context  context.Context
	Cancel   context.CancelFunc
	QueueID  int64
}

// NewManager creates a new queue manager
func NewManager(downloadStorage *storage.DownloadStorage, partStorage *storage.PartStorage, queueStorage *storage.QueueStorage) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		downloadStorage:    downloadStorage,
		partStorage:        partStorage,
		queueStorage:       queueStorage,
		resumeManager:      downloader.NewResumeManager(),
		multiDownloader:    downloader.NewMultiPartDownloader(8),
		activeDownloads:    make(map[int64]*ActiveDownload),
		schedulers:         make(map[int64]*Scheduler),
		ctx:                ctx,
		cancel:             cancel,
		maxGlobalDownloads: 10,
	}
}

// Start initializes and starts all queue schedulers
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Load all queues and start their schedulers
	queues, err := m.queueStorage.GetAll()
	if err != nil {
		return fmt.Errorf("failed to load queues: %w", err)
	}

	for _, queue := range queues {
		scheduler := NewScheduler(queue, m)
		m.schedulers[queue.ID] = scheduler

		// Start scheduler if queue should be running
		if queue.ShouldBeRunning() {
			go scheduler.Start(m.ctx)
		}
	}

	// Start global monitor
	go m.globalMonitor()

	return nil
}

// Stop stops all queue schedulers and active downloads
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Cancel all active downloads
	for _, active := range m.activeDownloads {
		active.Cancel()
	}

	// Stop all schedulers
	for _, scheduler := range m.schedulers {
		scheduler.Stop()
	}

	// Cancel main context
	m.cancel()

	return nil
}

// AddDownload adds a new download to a queue
func (m *Manager) AddDownload(download *models.DownloadItem, queueID int64) error {
	// Set queue ID
	download.QueueID = queueID

	// Save download to storage
	if err := m.downloadStorage.Create(download); err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// Add to queue
	if err := m.queueStorage.AddDownloadToQueue(queueID, download.ID); err != nil {
		return fmt.Errorf("failed to add download to queue: %w", err)
	}

	// Notify scheduler
	m.mutex.RLock()
	if scheduler, exists := m.schedulers[queueID]; exists {
		scheduler.NotifyNewDownload()
	}
	m.mutex.RUnlock()

	return nil
}

// StartDownload starts a specific download
func (m *Manager) StartDownload(downloadID int64) error {
	download, err := m.downloadStorage.GetByID(downloadID)
	if err != nil {
		return fmt.Errorf("download not found: %w", err)
	}

	if !download.CanStart() {
		return fmt.Errorf("download cannot be started in current state: %s", download.Status)
	}

	return m.executeDownload(download)
}

// PauseDownload pauses a specific download
func (m *Manager) PauseDownload(downloadID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	active, exists := m.activeDownloads[downloadID]
	if !exists {
		// Download is not active, just update status
		download, err := m.downloadStorage.GetByID(downloadID)
		if err != nil {
			return fmt.Errorf("download not found: %w", err)
		}

		download.Pause()
		return m.downloadStorage.Update(download)
	}

	// Cancel the active download
	active.Cancel()
	active.Download.Pause()

	// Update storage
	if err := m.downloadStorage.Update(active.Download); err != nil {
		return fmt.Errorf("failed to update download: %w", err)
	}

	// Remove from active downloads
	delete(m.activeDownloads, downloadID)

	// Update queue
	m.queueStorage.SetDownloadInactive(active.QueueID, downloadID)

	return nil
}

// ResumeDownload resumes a paused download
func (m *Manager) ResumeDownload(downloadID int64) error {
	download, err := m.downloadStorage.GetByID(downloadID)
	if err != nil {
		return fmt.Errorf("download not found: %w", err)
	}

	if !download.CanResume() {
		return fmt.Errorf("download cannot be resumed in current state: %s", download.Status)
	}

	return m.executeDownload(download)
}

// CancelDownload cancels a download permanently
func (m *Manager) CancelDownload(downloadID int64) error {
	// Pause first if active
	if err := m.PauseDownload(downloadID); err != nil {
		// Continue even if pause fails
	}

	// Update status to canceled
	download, err := m.downloadStorage.GetByID(downloadID)
	if err != nil {
		return fmt.Errorf("download not found: %w", err)
	}

	download.Cancel()

	// Clean up files
	parts, _ := m.partStorage.GetByDownloadID(downloadID)
	m.resumeManager.CleanupExistingFiles(download, parts)

	// Update storage
	if err := m.downloadStorage.Update(download); err != nil {
		return fmt.Errorf("failed to update download: %w", err)
	}

	// Remove from queue
	m.queueStorage.RemoveDownloadFromQueue(download.QueueID, downloadID)

	return nil
}

// executeDownload executes a download (internal method)
func (m *Manager) executeDownload(download *models.DownloadItem) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already active
	if _, exists := m.activeDownloads[download.ID]; exists {
		return fmt.Errorf("download is already active")
	}

	// Check global download limit
	if len(m.activeDownloads) >= m.maxGlobalDownloads {
		return fmt.Errorf("global download limit reached")
	}

	// Get queue and check its limits
	queue, err := m.queueStorage.GetByID(download.QueueID)
	if err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	if !queue.CanStartNewDownload() {
		return fmt.Errorf("queue cannot start new download")
	}

	// Create download context
	ctx, cancel := context.WithCancel(m.ctx)

	// Get existing parts
	parts, _ := m.partStorage.GetByDownloadID(download.ID)

	// Create active download
	active := &ActiveDownload{
		Download: download,
		Parts:    parts,
		Context:  ctx,
		Cancel:   cancel,
		QueueID:  download.QueueID,
	}

	// Add to active downloads
	m.activeDownloads[download.ID] = active

	// Mark as active in queue
	m.queueStorage.SetDownloadActive(download.QueueID, download.ID)

	// Start download in goroutine
	go m.runDownload(active)

	return nil
}

// runDownload runs the actual download process
func (m *Manager) runDownload(active *ActiveDownload) {
	defer func() {
		m.mutex.Lock()
		delete(m.activeDownloads, active.Download.ID)
		m.queueStorage.SetDownloadInactive(active.QueueID, active.Download.ID)
		m.mutex.Unlock()
	}()

	download := active.Download

	// Progress callback
	progressCallback := func(progress *downloader.MultiPartProgress) {
		if progress != nil {
			download.UpdateProgress(progress.TotalDownloaded, progress.TotalSpeed)
			m.downloadStorage.Update(download)

			if m.onDownloadProgress != nil {
				m.onDownloadProgress(download.ID, &download.Progress)
			}
		}
	}

	// Try to resume the download
	err := m.resumeManager.ResumeDownload(active.Context, download, active.Parts, progressCallback)

	if err != nil {
		download.SetError(err.Error())
		m.downloadStorage.Update(download)

		if m.onDownloadFailed != nil {
			m.onDownloadFailed(download.ID, err)
		}

		// Retry if possible
		if download.CanRetry() {
			// Schedule retry after delay
			go func() {
				time.Sleep(time.Duration(download.RetryCount) * time.Minute)
				m.executeDownload(download)
			}()
		}
	} else {
		// Download completed successfully
		download.Complete()
		m.downloadStorage.Update(download)

		if m.onDownloadComplete != nil {
			m.onDownloadComplete(download.ID)
		}
	}
}

// globalMonitor monitors global state and enforces limits
func (m *Manager) globalMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.enforceGlobalLimits()
			m.updateSchedulers()
		}
	}
}

// enforceGlobalLimits enforces global download limits
func (m *Manager) enforceGlobalLimits() {
	m.mutex.RLock()
	activeCount := len(m.activeDownloads)
	m.mutex.RUnlock()

	// Implement speed limiting if needed
	if m.globalSpeedLimit > 0 && activeCount > 0 {
		_ = m.globalSpeedLimit / int64(activeCount) // perDownloadLimit for future use
		// Apply speed limits to active downloads
		// This would require modifying the downloader to support dynamic speed limits
	}
}

// updateSchedulers checks and updates scheduler states
func (m *Manager) updateSchedulers() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for queueID, scheduler := range m.schedulers {
		queue, err := m.queueStorage.GetByID(queueID)
		if err != nil {
			continue
		}

		// Check if scheduler should be running based on schedule
		shouldRun := queue.ShouldBeRunning()
		isRunning := scheduler.IsRunning()

		if shouldRun && !isRunning {
			go scheduler.Start(m.ctx)
		} else if !shouldRun && isRunning {
			scheduler.Stop()
		}
	}
}

// Event callback setters
func (m *Manager) SetOnDownloadProgress(callback func(downloadID int64, progress *models.Progress)) {
	m.onDownloadProgress = callback
}

func (m *Manager) SetOnDownloadComplete(callback func(downloadID int64)) {
	m.onDownloadComplete = callback
}

func (m *Manager) SetOnDownloadFailed(callback func(downloadID int64, err error)) {
	m.onDownloadFailed = callback
}

// Configuration methods
func (m *Manager) SetGlobalSpeedLimit(limit int64) {
	m.globalSpeedLimit = limit
}

func (m *Manager) SetMaxGlobalDownloads(max int) {
	m.maxGlobalDownloads = max
}

// Queue management methods
func (m *Manager) CreateQueue(name string) (*models.QueueModel, error) {
	queue := models.NewQueueModel(name)
	if err := m.queueStorage.Create(queue); err != nil {
		return nil, err
	}

	// Create scheduler for new queue
	m.mutex.Lock()
	scheduler := NewScheduler(queue, m)
	m.schedulers[queue.ID] = scheduler
	m.mutex.Unlock()

	return queue, nil
}

func (m *Manager) GetQueues() ([]*models.QueueModel, error) {
	return m.queueStorage.GetAll()
}

func (m *Manager) GetQueue(queueID int64) (*models.QueueModel, error) {
	return m.queueStorage.GetByID(queueID)
}

func (m *Manager) UpdateQueue(queue *models.QueueModel) error {
	return m.queueStorage.Update(queue)
}

func (m *Manager) DeleteQueue(queueID int64) error {
	// Stop scheduler first
	m.mutex.Lock()
	if scheduler, exists := m.schedulers[queueID]; exists {
		scheduler.Stop()
		delete(m.schedulers, queueID)
	}
	m.mutex.Unlock()

	return m.queueStorage.Delete(queueID)
}

// Download management methods
func (m *Manager) GetDownloads() ([]*models.DownloadItem, error) {
	return m.downloadStorage.GetAll()
}

func (m *Manager) GetDownload(downloadID int64) (*models.DownloadItem, error) {
	return m.downloadStorage.GetByID(downloadID)
}

func (m *Manager) GetDownloadsByQueue(queueID int64) ([]*models.DownloadItem, error) {
	return m.downloadStorage.GetByQueue(queueID)
}

func (m *Manager) GetActiveDownloads() map[int64]*ActiveDownload {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	active := make(map[int64]*ActiveDownload)
	for k, v := range m.activeDownloads {
		active[k] = v
	}
	return active
}

// GetDownloadStorage returns the download storage instance
func (m *Manager) GetDownloadStorage() *storage.DownloadStorage {
	return m.downloadStorage
}

// GetPartStorage returns the part storage instance
func (m *Manager) GetPartStorage() *storage.PartStorage {
	return m.partStorage
}

// GetQueueStorage returns the queue storage instance
func (m *Manager) GetQueueStorage() *storage.QueueStorage {
	return m.queueStorage
}

// DeleteDownload completely removes a download from the system
func (m *Manager) DeleteDownload(downloadID int64) error {
	// Get download info before deletion
	download, err := m.downloadStorage.GetByID(downloadID)
	if err != nil {
		return fmt.Errorf("download not found: %w", err)
	}

	// Cancel first if active (this also cleans up files)
	if err := m.CancelDownload(downloadID); err != nil {
		// Continue with deletion even if cancel fails
		// The download might not be active
	}

	// Delete all associated parts
	if err := m.partStorage.DeleteByDownloadID(downloadID); err != nil {
		// Continue even if part deletion fails
	}

	// Remove download from queue
	if err := m.queueStorage.RemoveDownloadFromQueue(download.QueueID, downloadID); err != nil {
		// Continue even if queue removal fails
	}

	// Delete the download record from storage
	if err := m.downloadStorage.Delete(downloadID); err != nil {
		return fmt.Errorf("failed to delete download from storage: %w", err)
	}

	return nil
}
