package models

import (
	"time"
)

// DownloadItem represents a single download item
type DownloadItem struct {
	ID             int64             `json:"id"`
	URL            string            `json:"url"`
	Filename       string            `json:"filename"`
	FilePath       string            `json:"file_path"`
	TotalSize      int64             `json:"total_size"`
	Status         DownloadStatus    `json:"status"`
	Progress       Progress          `json:"progress"`
	Parts          []int64           `json:"parts"`           // IDs of download parts
	MaxConnections int               `json:"max_connections"` // Number of concurrent connections
	SpeedLimit     int64             `json:"speed_limit"`     // Speed limit in bytes per second (0 = unlimited)
	Headers        map[string]string `json:"headers"`         // Custom HTTP headers
	UserAgent      string            `json:"user_agent"`
	Referrer       string            `json:"referrer"`
	CreatedAt      time.Time         `json:"created_at"`
	StartedAt      *time.Time        `json:"started_at,omitempty"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	LastError      string            `json:"last_error,omitempty"`
	RetryCount     int               `json:"retry_count"`
	MaxRetries     int               `json:"max_retries"`
	QueueID        int64             `json:"queue_id"`
	Priority       int               `json:"priority"` // Higher number = higher priority
}

// NewDownloadItem creates a new download item with default values
func NewDownloadItem(url, filename, filePath string) *DownloadItem {
	return &DownloadItem{
		URL:            url,
		Filename:       filename,
		FilePath:       filePath,
		Status:         StatusPending,
		MaxConnections: 8,
		SpeedLimit:     0,
		Headers:        make(map[string]string),
		UserAgent:      "Download Manager 1.0",
		CreatedAt:      time.Now(),
		RetryCount:     0,
		MaxRetries:     3,
		Priority:       0,
		Progress: Progress{
			BytesDownloaded: 0,
			TotalBytes:      0,
			Speed:           0,
			ETA:             0,
			Percentage:      0,
			LastUpdate:      time.Now(),
		},
	}
}

// IsCompleted returns true if the download is completed
func (d *DownloadItem) IsCompleted() bool {
	return d.Status == StatusCompleted
}

// IsPaused returns true if the download is paused
func (d *DownloadItem) IsPaused() bool {
	return d.Status == StatusPaused
}

// IsActive returns true if the download is currently downloading
func (d *DownloadItem) IsActive() bool {
	return d.Status == StatusDownloading
}

// CanResume returns true if the download can be resumed
func (d *DownloadItem) CanResume() bool {
	return d.Status == StatusPaused || d.Status == StatusFailed
}

// CanStart returns true if the download can be started
func (d *DownloadItem) CanStart() bool {
	return d.Status == StatusPending || d.Status == StatusFailed
}

// UpdateProgress updates the download progress
func (d *DownloadItem) UpdateProgress(bytesDownloaded int64, speed int64) {
	d.Progress.BytesDownloaded = bytesDownloaded
	d.Progress.Speed = speed
	d.Progress.LastUpdate = time.Now()
	d.Progress.CalculatePercentage()
	d.Progress.CalculateETA()
}

// SetError sets the last error and increments retry count
func (d *DownloadItem) SetError(err string) {
	d.LastError = err
	d.RetryCount++
	d.Status = StatusFailed
}

// CanRetry returns true if the download can be retried
func (d *DownloadItem) CanRetry() bool {
	return d.RetryCount < d.MaxRetries && d.Status == StatusFailed
}

// Start marks the download as started
func (d *DownloadItem) Start() {
	d.Status = StatusDownloading
	now := time.Now()
	d.StartedAt = &now
	d.LastError = ""
}

// Complete marks the download as completed
func (d *DownloadItem) Complete() {
	d.Status = StatusCompleted
	now := time.Now()
	d.CompletedAt = &now
	d.Progress.Percentage = 100
}

// Pause marks the download as paused
func (d *DownloadItem) Pause() {
	d.Status = StatusPaused
}

// Cancel marks the download as canceled
func (d *DownloadItem) Cancel() {
	d.Status = StatusCanceled
}
