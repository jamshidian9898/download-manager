package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamshidian/download-manager/backend/pkg/downloader"
	"github.com/jamshidian/download-manager/backend/pkg/models"
	"github.com/jamshidian/download-manager/backend/pkg/queue"
)

// DownloadAPI provides download-related API methods for the frontend
type DownloadAPI struct {
	manager        *queue.Manager
	httpDownloader *downloader.HTTPDownloader
}

// NewDownloadAPI creates a new download API instance
func NewDownloadAPI(manager *queue.Manager) *DownloadAPI {
	return &DownloadAPI{
		manager:        manager,
		httpDownloader: downloader.NewHTTPDownloader(),
	}
}

// AddDownloadRequest represents a request to add a new download
type AddDownloadRequest struct {
	URL            string            `json:"url"`
	Filename       string            `json:"filename,omitempty"`
	FilePath       string            `json:"file_path,omitempty"`
	QueueID        int64             `json:"queue_id,omitempty"`
	MaxConnections int               `json:"max_connections,omitempty"`
	SpeedLimit     int64             `json:"speed_limit,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	UserAgent      string            `json:"user_agent,omitempty"`
	Referrer       string            `json:"referrer,omitempty"`
	Priority       int               `json:"priority,omitempty"`
}

// AddDownloadResponse represents the response from adding a download
type AddDownloadResponse struct {
	Success    bool                 `json:"success"`
	Error      string               `json:"error,omitempty"`
	Download   *models.DownloadItem `json:"download,omitempty"`
	DownloadID int64                `json:"download_id,omitempty"`
}

// DownloadInfoResponse represents download information from a URL
type DownloadInfoResponse struct {
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	URL           string `json:"url"`
	Filename      string `json:"filename"`
	ContentLength int64  `json:"content_length"`
	SupportsRange bool   `json:"supports_range"`
	ContentType   string `json:"content_type"`
}

// GetDownloadInfo retrieves information about a download URL
func (api *DownloadAPI) GetDownloadInfo(ctx context.Context, url string, headers map[string]string) *DownloadInfoResponse {
	if headers == nil {
		headers = make(map[string]string)
	}

	info, err := api.httpDownloader.GetDownloadInfo(ctx, url, headers)
	if err != nil {
		return &DownloadInfoResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	return &DownloadInfoResponse{
		Success:       true,
		URL:           info.URL,
		Filename:      info.Filename,
		ContentLength: info.ContentLength,
		SupportsRange: info.SupportsRange,
		ContentType:   info.ContentType,
	}
}

// AddDownload adds a new download to the system
func (api *DownloadAPI) AddDownload(ctx context.Context, req *AddDownloadRequest) *AddDownloadResponse {
	if req.URL == "" {
		return &AddDownloadResponse{
			Success: false,
			Error:   "URL is required",
		}
	}

	// Get download info first
	info, err := api.httpDownloader.GetDownloadInfo(ctx, req.URL, req.Headers)
	if err != nil {
		return &AddDownloadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get download info: %v", err),
		}
	}

	// Determine filename
	filename := req.Filename
	if filename == "" {
		filename = info.Filename
	}
	if filename == "" {
		filename = "download"
	}

	// Determine file path
	filePath := req.FilePath
	if filePath == "" {
		// Use default download path from settings
		// For now, use a default path
		filePath = filepath.Join("downloads", filename)
	}

	// Create download item
	download := models.NewDownloadItem(req.URL, filename, filePath)
	download.TotalSize = info.ContentLength

	// Apply request settings
	if req.MaxConnections > 0 {
		download.MaxConnections = req.MaxConnections
	}
	if req.SpeedLimit > 0 {
		download.SpeedLimit = req.SpeedLimit
	}
	if req.Headers != nil {
		download.Headers = req.Headers
	}
	if req.UserAgent != "" {
		download.UserAgent = req.UserAgent
	}
	if req.Referrer != "" {
		download.Referrer = req.Referrer
	}
	download.Priority = req.Priority

	// Determine queue ID
	queueID := req.QueueID
	if queueID == 0 {
		// Get default queue
		defaultQueue, err := api.manager.GetQueues()
		if err != nil || len(defaultQueue) == 0 {
			return &AddDownloadResponse{
				Success: false,
				Error:   "No queues available",
			}
		}
		queueID = defaultQueue[0].ID
	}

	// Add to manager
	if err := api.manager.AddDownload(download, queueID); err != nil {
		return &AddDownloadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to add download: %v", err),
		}
	}

	return &AddDownloadResponse{
		Success:    true,
		Download:   download,
		DownloadID: download.ID,
	}
}

// GetDownloads returns all downloads
func (api *DownloadAPI) GetDownloads() ([]*models.DownloadItem, error) {
	return api.manager.GetDownloads()
}

// GetDownload returns a specific download by ID
func (api *DownloadAPI) GetDownload(downloadID int64) (*models.DownloadItem, error) {
	return api.manager.GetDownload(downloadID)
}

// GetDownloadsByQueue returns downloads for a specific queue
func (api *DownloadAPI) GetDownloadsByQueue(queueID int64) ([]*models.DownloadItem, error) {
	return api.manager.GetDownloadsByQueue(queueID)
}

// StartDownload starts a specific download
func (api *DownloadAPI) StartDownload(downloadID int64) error {
	return api.manager.StartDownload(downloadID)
}

// PauseDownload pauses a specific download
func (api *DownloadAPI) PauseDownload(downloadID int64) error {
	return api.manager.PauseDownload(downloadID)
}

// ResumeDownload resumes a paused download
func (api *DownloadAPI) ResumeDownload(downloadID int64) error {
	return api.manager.ResumeDownload(downloadID)
}

// CancelDownload cancels a download permanently
func (api *DownloadAPI) CancelDownload(downloadID int64) error {
	return api.manager.CancelDownload(downloadID)
}

// DeleteDownload removes a download from the system
func (api *DownloadAPI) DeleteDownload(downloadID int64) error {
	return api.manager.DeleteDownload(downloadID)
}

// GetActiveDownloads returns currently active downloads
func (api *DownloadAPI) GetActiveDownloads() map[int64]*queue.ActiveDownload {
	return api.manager.GetActiveDownloads()
}

// DownloadStats represents download statistics
type DownloadStats struct {
	TotalDownloads     int   `json:"total_downloads"`
	ActiveDownloads    int   `json:"active_downloads"`
	CompletedDownloads int   `json:"completed_downloads"`
	FailedDownloads    int   `json:"failed_downloads"`
	PausedDownloads    int   `json:"paused_downloads"`
	TotalBytes         int64 `json:"total_bytes"`
	DownloadedBytes    int64 `json:"downloaded_bytes"`
	TotalSpeed         int64 `json:"total_speed"`
}

// GetDownloadStats returns overall download statistics
func (api *DownloadAPI) GetDownloadStats() (*DownloadStats, error) {
	downloads, err := api.manager.GetDownloads()
	if err != nil {
		return nil, err
	}

	stats := &DownloadStats{}
	activeDownloads := api.manager.GetActiveDownloads()

	for _, download := range downloads {
		stats.TotalDownloads++
		stats.TotalBytes += download.TotalSize
		stats.DownloadedBytes += download.Progress.BytesDownloaded

		switch download.Status {
		case models.StatusDownloading:
			stats.ActiveDownloads++
		case models.StatusCompleted:
			stats.CompletedDownloads++
		case models.StatusFailed:
			stats.FailedDownloads++
		case models.StatusPaused:
			stats.PausedDownloads++
		}
	}

	// Calculate total speed from active downloads
	for _, active := range activeDownloads {
		stats.TotalSpeed += active.Download.Progress.Speed
	}

	return stats, nil
}

// UpdateDownloadRequest represents a request to update download settings
type UpdateDownloadRequest struct {
	DownloadID     int64             `json:"download_id"`
	MaxConnections *int              `json:"max_connections,omitempty"`
	SpeedLimit     *int64            `json:"speed_limit,omitempty"`
	Priority       *int              `json:"priority,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	UserAgent      *string           `json:"user_agent,omitempty"`
	Referrer       *string           `json:"referrer,omitempty"`
}

// UpdateDownload updates download settings
func (api *DownloadAPI) UpdateDownload(req *UpdateDownloadRequest) error {
	download, err := api.manager.GetDownload(req.DownloadID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if req.MaxConnections != nil {
		download.MaxConnections = *req.MaxConnections
	}
	if req.SpeedLimit != nil {
		download.SpeedLimit = *req.SpeedLimit
	}
	if req.Priority != nil {
		download.Priority = *req.Priority
	}
	if req.Headers != nil {
		download.Headers = req.Headers
	}
	if req.UserAgent != nil {
		download.UserAgent = *req.UserAgent
	}
	if req.Referrer != nil {
		download.Referrer = *req.Referrer
	}

	// This would need to be implemented in the manager
	// For now, return nil
	return nil
}

// MoveDownloadRequest represents a request to move a download in queue
type MoveDownloadRequest struct {
	DownloadID  int64 `json:"download_id"`
	QueueID     int64 `json:"queue_id"`
	NewPosition int   `json:"new_position"`
}

// MoveDownload moves a download to a new position in the queue
func (api *DownloadAPI) MoveDownload(req *MoveDownloadRequest) error {
	// This would need to be implemented in the queue storage
	// For now, return nil
	return nil
}
