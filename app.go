package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/jamshidian/download-manager/backend/pkg/api"
	"github.com/jamshidian/download-manager/backend/pkg/config"
	"github.com/jamshidian/download-manager/backend/pkg/models"
	"github.com/jamshidian/download-manager/backend/pkg/queue"
	"github.com/jamshidian/download-manager/backend/pkg/storage"
)

// App struct
type App struct {
	ctx context.Context

	// Core components
	settings        *config.Settings
	storage         *storage.TransactionalStorage
	downloadStorage *storage.DownloadStorage
	partStorage     *storage.PartStorage
	queueStorage    *storage.QueueStorage
	queueManager    *queue.Manager

	// API layers
	downloadAPI *api.DownloadAPI
	queueAPI    *api.QueueAPI
	eventAPI    *api.EventAPI
}

// NewApp creates a new App application struct
func NewApp() *App {
	log.Println("[DEBUG] Creating new App instance")
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	log.Println("[DEBUG] App startup initiated")
	a.ctx = ctx

	// Initialize settings
	log.Println("[DEBUG] Initializing default settings")
	a.settings = config.NewDefaultSettings()
	a.settings.ApplyDefaults()
	log.Printf("[DEBUG] Settings initialized: DataDirectory=%s, MaxConcurrentDownloads=%d",
		a.settings.DataDirectory, a.settings.MaxConcurrentDownloads)

	// Validate and create directories
	log.Println("[DEBUG] Validating settings")
	if errors := a.settings.Validate(); len(errors) > 0 {
		log.Printf("[WARN] Settings validation errors: %v", errors)
		// Use defaults for invalid settings
	} else {
		log.Println("[DEBUG] Settings validation passed")
	}

	// Initialize storage
	log.Println("[DEBUG] Initializing storage layer")
	if err := a.initializeStorage(); err != nil {
		log.Fatalf("[ERROR] Failed to initialize storage: %v", err)
	}
	log.Println("[DEBUG] Storage layer initialized successfully")

	// Initialize queue manager
	log.Println("[DEBUG] Initializing queue manager")
	if err := a.initializeQueueManager(); err != nil {
		log.Fatalf("[ERROR] Failed to initialize queue manager: %v", err)
	}
	log.Println("[DEBUG] Queue manager initialized successfully")

	// Initialize API layers
	log.Println("[DEBUG] Initializing API layers")
	a.initializeAPIs()
	log.Println("[DEBUG] API layers initialized successfully")

	// Start the queue manager
	log.Println("[DEBUG] Starting queue manager")
	if err := a.queueManager.Start(); err != nil {
		log.Fatalf("[ERROR] Failed to start queue manager: %v", err)
	}
	log.Println("[DEBUG] Queue manager started successfully")

	log.Println("[INFO] Download Manager initialized successfully")
}

// initializeStorage sets up the storage layer
func (a *App) initializeStorage() error {
	log.Printf("[DEBUG] initializeStorage: Creating transactional storage at %s", a.settings.DataDirectory)

	// Create transactional storage
	var err error
	a.storage, err = storage.NewTransactionalStorage(a.settings.DataDirectory)
	if err != nil {
		log.Printf("[ERROR] initializeStorage: Failed to create transactional storage: %v", err)
		return fmt.Errorf("failed to create transactional storage: %w", err)
	}
	log.Println("[DEBUG] initializeStorage: Transactional storage created successfully")

	// Create storage instances
	log.Println("[DEBUG] initializeStorage: Creating storage instances")
	a.downloadStorage = storage.NewDownloadStorage(a.storage)
	a.partStorage = storage.NewPartStorage(a.storage)
	a.queueStorage = storage.NewQueueStorage(a.storage)
	log.Println("[DEBUG] initializeStorage: All storage instances created successfully")

	return nil
}

// initializeQueueManager sets up the queue manager
func (a *App) initializeQueueManager() error {
	log.Println("[DEBUG] initializeQueueManager: Creating queue manager")
	a.queueManager = queue.NewManager(a.downloadStorage, a.partStorage, a.queueStorage)
	log.Println("[DEBUG] initializeQueueManager: Queue manager created")

	// Configure manager with settings
	log.Printf("[DEBUG] initializeQueueManager: Setting global speed limit to %d bytes/s", a.settings.GlobalSpeedLimit)
	a.queueManager.SetGlobalSpeedLimit(a.settings.GlobalSpeedLimit)

	log.Printf("[DEBUG] initializeQueueManager: Setting max concurrent downloads to %d", a.settings.MaxConcurrentDownloads)
	a.queueManager.SetMaxGlobalDownloads(a.settings.MaxConcurrentDownloads)

	log.Println("[DEBUG] initializeQueueManager: Queue manager configured successfully")
	return nil
}

// initializeAPIs sets up the API layers
func (a *App) initializeAPIs() {
	log.Println("[DEBUG] initializeAPIs: Creating API instances")
	a.downloadAPI = api.NewDownloadAPI(a.queueManager)
	a.queueAPI = api.NewQueueAPI(a.queueManager)
	a.eventAPI = api.NewEventAPI(a.ctx)
	log.Println("[DEBUG] initializeAPIs: API instances created")

	// Set up event callbacks
	log.Println("[DEBUG] initializeAPIs: Setting up event callbacks")
	a.queueManager.SetOnDownloadProgress(func(downloadID int64, progress *models.Progress) {
		log.Printf("[DEBUG] Event: Download progress - ID: %d, Progress: %d%%", downloadID, progress.Percentage)
		a.eventAPI.EmitDownloadProgress(downloadID, progress)
	})

	a.queueManager.SetOnDownloadComplete(func(downloadID int64) {
		log.Printf("[DEBUG] Event: Download completed - ID: %d", downloadID)
		if download, err := a.downloadStorage.GetByID(downloadID); err == nil {
			a.eventAPI.EmitDownloadStatusChange(download, models.StatusDownloading, models.StatusCompleted, nil)
		} else {
			log.Printf("[ERROR] Failed to get download %d for completion event: %v", downloadID, err)
		}
	})

	a.queueManager.SetOnDownloadFailed(func(downloadID int64, err error) {
		log.Printf("[DEBUG] Event: Download failed - ID: %d, Error: %v", downloadID, err)
		if download, err := a.downloadStorage.GetByID(downloadID); err == nil {
			a.eventAPI.EmitDownloadStatusChange(download, models.StatusDownloading, models.StatusFailed, err)
		} else {
			log.Printf("[ERROR] Failed to get download %d for failure event: %v", downloadID, err)
		}
	})
	log.Println("[DEBUG] initializeAPIs: Event callbacks configured")

	// Start system monitor
	log.Println("[DEBUG] initializeAPIs: Starting system monitor")
	a.eventAPI.StartSystemMonitor(
		func() int {
			activeCount := len(a.queueManager.GetActiveDownloads())
			log.Printf("[DEBUG] SystemMonitor: Active downloads count: %d", activeCount)
			return activeCount
		},
		func() int64 {
			var totalSpeed int64
			activeDownloads := a.queueManager.GetActiveDownloads()
			for _, active := range activeDownloads {
				totalSpeed += active.Download.Progress.Speed
			}
			log.Printf("[DEBUG] SystemMonitor: Total speed: %d bytes/s", totalSpeed)
			return totalSpeed
		},
	)
	log.Println("[DEBUG] initializeAPIs: System monitor started")
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) {
	log.Println("[INFO] Shutdown: Initiating graceful shutdown of Download Manager...")

	// Stop queue manager
	if a.queueManager != nil {
		log.Println("[DEBUG] Shutdown: Stopping queue manager")
		if err := a.queueManager.Stop(); err != nil {
			log.Printf("[ERROR] Shutdown: Error stopping queue manager: %v", err)
		} else {
			log.Println("[DEBUG] Shutdown: Queue manager stopped successfully")
		}
	} else {
		log.Println("[DEBUG] Shutdown: Queue manager is nil, skipping stop")
	}

	log.Println("[INFO] Shutdown: Download Manager shut down complete")
}

// GetSettings returns current application settings
func (a *App) GetSettings() *config.Settings {
	log.Println("[DEBUG] GetSettings: Returning current application settings")
	return a.settings
}

// UpdateSettings updates application settings
func (a *App) UpdateSettings(newSettings *config.Settings) error {
	log.Println("[DEBUG] UpdateSettings: Validating new settings")
	if errors := newSettings.Validate(); len(errors) > 0 {
		log.Printf("[ERROR] UpdateSettings: Settings validation failed: %v", errors)
		return fmt.Errorf("settings validation failed: %v", errors)
	}
	log.Println("[DEBUG] UpdateSettings: Settings validation passed")

	a.settings = newSettings
	log.Println("[DEBUG] UpdateSettings: Settings updated")

	// Apply settings to queue manager
	if a.queueManager != nil {
		log.Printf("[DEBUG] UpdateSettings: Applying global speed limit: %d bytes/s", a.settings.GlobalSpeedLimit)
		a.queueManager.SetGlobalSpeedLimit(a.settings.GlobalSpeedLimit)

		log.Printf("[DEBUG] UpdateSettings: Applying max concurrent downloads: %d", a.settings.MaxConcurrentDownloads)
		a.queueManager.SetMaxGlobalDownloads(a.settings.MaxConcurrentDownloads)

		log.Println("[DEBUG] UpdateSettings: Settings applied to queue manager")
	} else {
		log.Println("[DEBUG] UpdateSettings: Queue manager is nil, skipping settings application")
	}

	log.Println("[DEBUG] UpdateSettings: Settings update completed successfully")
	return nil
}

// =============================================================================
// Download API Methods (exposed to frontend)
// =============================================================================

// GetDownloadInfo gets information about a download URL
func (a *App) GetDownloadInfo(url string, headers map[string]string) *api.DownloadInfoResponse {
	log.Printf("[DEBUG] GetDownloadInfo: Getting info for URL: %s", url)
	response := a.downloadAPI.GetDownloadInfo(a.ctx, url, headers)
	log.Printf("[DEBUG] GetDownloadInfo: Response success: %t", response.Success)
	return response
}

// AddDownload adds a new download
func (a *App) AddDownload(req *api.AddDownloadRequest) *api.AddDownloadResponse {
	log.Printf("[DEBUG] AddDownload: Adding download for URL: %s", req.URL)

	// Set default file path if not provided
	if req.FilePath == "" {
		log.Println("[DEBUG] AddDownload: No file path provided, generating default")
		filename := req.Filename
		if filename == "" {
			log.Println("[DEBUG] AddDownload: No filename provided, getting from URL info")
			// Get filename from URL info
			info := a.downloadAPI.GetDownloadInfo(a.ctx, req.URL, req.Headers)
			if info.Success && info.Filename != "" {
				filename = info.Filename
				log.Printf("[DEBUG] AddDownload: Got filename from URL: %s", filename)
			} else {
				filename = "download"
				log.Println("[DEBUG] AddDownload: Using default filename: download")
			}
		}
		req.FilePath = filepath.Join(a.settings.DefaultDownloadPath, filename)
		log.Printf("[DEBUG] AddDownload: Generated file path: %s", req.FilePath)
	}

	response := a.downloadAPI.AddDownload(a.ctx, req)
	log.Printf("[DEBUG] AddDownload: Response success: %t, ID: %d", response.Success, response.DownloadID)
	return response
}

// GetDownloads returns all downloads
func (a *App) GetDownloads() ([]*models.DownloadItem, error) {
	log.Println("[DEBUG] GetDownloads: Fetching all downloads")
	downloads, err := a.downloadAPI.GetDownloads()
	if err != nil {
		log.Printf("[ERROR] GetDownloads: Failed to get downloads: %v", err)
	} else {
		log.Printf("[DEBUG] GetDownloads: Retrieved %d downloads", len(downloads))
	}
	return downloads, err
}

// GetDownload returns a specific download
func (a *App) GetDownload(downloadID int64) (*models.DownloadItem, error) {
	log.Printf("[DEBUG] GetDownload: Fetching download ID: %d", downloadID)
	download, err := a.downloadAPI.GetDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] GetDownload: Failed to get download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] GetDownload: Retrieved download %d: %s", downloadID, download.Filename)
	}
	return download, err
}

// StartDownload starts a download
func (a *App) StartDownload(downloadID int64) error {
	log.Printf("[DEBUG] StartDownload: Starting download ID: %d", downloadID)
	err := a.downloadAPI.StartDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] StartDownload: Failed to start download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] StartDownload: Successfully started download %d", downloadID)
	}
	return err
}

// PauseDownload pauses a download
func (a *App) PauseDownload(downloadID int64) error {
	log.Printf("[DEBUG] PauseDownload: Pausing download ID: %d", downloadID)
	err := a.downloadAPI.PauseDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] PauseDownload: Failed to pause download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] PauseDownload: Successfully paused download %d", downloadID)
	}
	return err
}

// ResumeDownload resumes a download
func (a *App) ResumeDownload(downloadID int64) error {
	log.Printf("[DEBUG] ResumeDownload: Resuming download ID: %d", downloadID)
	err := a.downloadAPI.ResumeDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] ResumeDownload: Failed to resume download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] ResumeDownload: Successfully resumed download %d", downloadID)
	}
	return err
}

// CancelDownload cancels a download
func (a *App) CancelDownload(downloadID int64) error {
	log.Printf("[DEBUG] CancelDownload: Canceling download ID: %d", downloadID)
	err := a.downloadAPI.CancelDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] CancelDownload: Failed to cancel download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] CancelDownload: Successfully canceled download %d", downloadID)
	}
	return err
}

// DeleteDownload deletes a download
func (a *App) DeleteDownload(downloadID int64) error {
	log.Printf("[DEBUG] DeleteDownload: Deleting download ID: %d", downloadID)
	err := a.downloadAPI.DeleteDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] DeleteDownload: Failed to delete download %d: %v", downloadID, err)
	} else {
		log.Printf("[DEBUG] DeleteDownload: Successfully deleted download %d", downloadID)
	}
	return err
}

// GetDownloadStats returns download statistics
func (a *App) GetDownloadStats() (*api.DownloadStats, error) {
	log.Println("[DEBUG] GetDownloadStats: Fetching download statistics")
	stats, err := a.downloadAPI.GetDownloadStats()
	if err != nil {
		log.Printf("[ERROR] GetDownloadStats: Failed to get stats: %v", err)
	} else {
		log.Printf("[DEBUG] GetDownloadStats: Retrieved stats - Total: %d, Active: %d, Completed: %d",
			stats.TotalDownloads, stats.ActiveDownloads, stats.CompletedDownloads)
	}
	return stats, err
}

// =============================================================================
// Queue API Methods (exposed to frontend)
// =============================================================================

// CreateQueue creates a new queue
func (a *App) CreateQueue(req *api.CreateQueueRequest) *api.CreateQueueResponse {
	log.Printf("[DEBUG] CreateQueue: Creating queue with name: %s", req.Name)
	response := a.queueAPI.CreateQueue(req)
	log.Printf("[DEBUG] CreateQueue: Response success: %t, ID: %d", response.Success, response.Queue.ID)
	return response
}

// GetQueues returns all queues
func (a *App) GetQueues() ([]*models.QueueModel, error) {
	log.Println("[DEBUG] GetQueues: Fetching all queues")
	queues, err := a.queueAPI.GetQueues()
	if err != nil {
		log.Printf("[ERROR] GetQueues: Failed to get queues: %v", err)
	} else {
		log.Printf("[DEBUG] GetQueues: Retrieved %d queues", len(queues))
	}
	return queues, err
}

// GetQueue returns a specific queue
func (a *App) GetQueue(queueID int64) (*models.QueueModel, error) {
	log.Printf("[DEBUG] GetQueue: Fetching queue ID: %d", queueID)
	queue, err := a.queueAPI.GetQueue(queueID)
	if err != nil {
		log.Printf("[ERROR] GetQueue: Failed to get queue %d: %v", queueID, err)
	} else {
		log.Printf("[DEBUG] GetQueue: Retrieved queue %d: %s", queueID, queue.Name)
	}
	return queue, err
}

// UpdateQueue updates queue settings
func (a *App) UpdateQueue(req *api.UpdateQueueRequest) error {
	log.Printf("[DEBUG] UpdateQueue: Updating queue ID: %d", req.QueueID)
	err := a.queueAPI.UpdateQueue(req)
	if err != nil {
		log.Printf("[ERROR] UpdateQueue: Failed to update queue %d: %v", req.QueueID, err)
	} else {
		log.Printf("[DEBUG] UpdateQueue: Successfully updated queue %d", req.QueueID)
	}
	return err
}

// DeleteQueue deletes a queue
func (a *App) DeleteQueue(queueID int64) error {
	log.Printf("[DEBUG] DeleteQueue: Deleting queue ID: %d", queueID)
	err := a.queueAPI.DeleteQueue(queueID)
	if err != nil {
		log.Printf("[ERROR] DeleteQueue: Failed to delete queue %d: %v", queueID, err)
	} else {
		log.Printf("[DEBUG] DeleteQueue: Successfully deleted queue %d", queueID)
	}
	return err
}

// StartQueue starts a queue
func (a *App) StartQueue(queueID int64) error {
	log.Printf("[DEBUG] StartQueue: Starting queue ID: %d", queueID)
	err := a.queueAPI.StartQueue(queueID)
	if err != nil {
		log.Printf("[ERROR] StartQueue: Failed to start queue %d: %v", queueID, err)
	} else {
		log.Printf("[DEBUG] StartQueue: Successfully started queue %d", queueID)
	}
	return err
}

// PauseQueue pauses a queue
func (a *App) PauseQueue(queueID int64) error {
	log.Printf("[DEBUG] PauseQueue: Pausing queue ID: %d", queueID)
	err := a.queueAPI.PauseQueue(queueID)
	if err != nil {
		log.Printf("[ERROR] PauseQueue: Failed to pause queue %d: %v", queueID, err)
	} else {
		log.Printf("[DEBUG] PauseQueue: Successfully paused queue %d", queueID)
	}
	return err
}

// StopQueue stops a queue
func (a *App) StopQueue(queueID int64) error {
	log.Printf("[DEBUG] StopQueue: Stopping queue ID: %d", queueID)
	err := a.queueAPI.StopQueue(queueID)
	if err != nil {
		log.Printf("[ERROR] StopQueue: Failed to stop queue %d: %v", queueID, err)
	} else {
		log.Printf("[DEBUG] StopQueue: Successfully stopped queue %d", queueID)
	}
	return err
}

// GetQueueStats returns queue statistics
func (a *App) GetQueueStats(queueID int64) *api.QueueStatsResponse {
	log.Printf("[DEBUG] GetQueueStats: Fetching stats for queue ID: %d", queueID)
	stats := a.queueAPI.GetQueueStats(queueID)
	log.Printf("[DEBUG] GetQueueStats: Retrieved stats for queue %d - Success: %t", queueID, stats.Success)
	return stats
}

// GetAllQueueStats returns statistics for all queues
func (a *App) GetAllQueueStats() ([]*api.QueueStats, error) {
	log.Println("[DEBUG] GetAllQueueStats: Fetching statistics for all queues")
	stats, err := a.queueAPI.GetAllQueueStats()
	if err != nil {
		log.Printf("[ERROR] GetAllQueueStats: Failed to get queue stats: %v", err)
	} else {
		log.Printf("[DEBUG] GetAllQueueStats: Retrieved stats for %d queues", len(stats))
	}
	return stats, err
}

// =============================================================================
// System Methods
// =============================================================================

// GetSystemInfo returns system information
func (a *App) GetSystemInfo() map[string]interface{} {
	log.Println("[DEBUG] GetSystemInfo: Collecting system information")
	activeDownloads := a.queueManager.GetActiveDownloads()

	var totalSpeed int64
	for _, active := range activeDownloads {
		totalSpeed += active.Download.Progress.Speed
	}

	systemInfo := map[string]interface{}{
		"active_downloads": len(activeDownloads),
		"total_speed":      totalSpeed,
		"data_directory":   a.settings.DataDirectory,
		"version":          "1.0.0",
	}

	log.Printf("[DEBUG] GetSystemInfo: Active downloads: %d, Total speed: %d bytes/s",
		len(activeDownloads), totalSpeed)
	return systemInfo
}

// OpenDownloadFolder opens the download folder in file manager
func (a *App) OpenDownloadFolder(downloadID int64) error {
	log.Printf("[DEBUG] OpenDownloadFolder: Opening folder for download ID: %d", downloadID)
	download, err := a.downloadAPI.GetDownload(downloadID)
	if err != nil {
		log.Printf("[ERROR] OpenDownloadFolder: Failed to get download %d: %v", downloadID, err)
		return err
	}

	folderPath := filepath.Dir(download.FilePath)
	log.Printf("[DEBUG] OpenDownloadFolder: Opening folder: %s", folderPath)

	// This would need OS-specific implementation
	// For now, just return the path
	log.Printf("[INFO] OpenDownloadFolder: Would open folder for download: %s", folderPath)
	return nil
}

// Greet returns a greeting for the given name (legacy method)
func (a *App) Greet(name string) string {
	log.Printf("[DEBUG] Greet: Greeting user: %s", name)
	greeting := fmt.Sprintf("Hello %s, Welcome to Download Manager!", name)
	log.Printf("[DEBUG] Greet: Generated greeting: %s", greeting)
	return greeting
}
