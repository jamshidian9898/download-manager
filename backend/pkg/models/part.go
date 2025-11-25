package models

import (
	"time"
)

// DownloadPart represents a chunk of a download for multi-part downloading
type DownloadPart struct {
	ID          int64          `json:"id"`
	DownloadID  int64          `json:"download_id"`
	PartNumber  int            `json:"part_number"`
	StartByte   int64          `json:"start_byte"`
	EndByte     int64          `json:"end_byte"`
	CurrentByte int64          `json:"current_byte"`
	Status      DownloadStatus `json:"status"`
	FilePath    string         `json:"file_path"` // Temporary file path for this part
	Speed       int64          `json:"speed"`     // Current download speed in bytes/sec
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	LastError   string         `json:"last_error,omitempty"`
	RetryCount  int            `json:"retry_count"`
	MaxRetries  int            `json:"max_retries"`
}

// NewDownloadPart creates a new download part
func NewDownloadPart(downloadID int64, partNumber int, startByte, endByte int64, filePath string) *DownloadPart {
	return &DownloadPart{
		DownloadID:  downloadID,
		PartNumber:  partNumber,
		StartByte:   startByte,
		EndByte:     endByte,
		CurrentByte: startByte,
		Status:      StatusPending,
		FilePath:    filePath,
		CreatedAt:   time.Now(),
		RetryCount:  0,
		MaxRetries:  3,
	}
}

// Size returns the total size of this part in bytes
func (p *DownloadPart) Size() int64 {
	return p.EndByte - p.StartByte + 1
}

// Downloaded returns the number of bytes downloaded for this part
func (p *DownloadPart) Downloaded() int64 {
	return p.CurrentByte - p.StartByte
}

// Remaining returns the number of bytes remaining for this part
func (p *DownloadPart) Remaining() int64 {
	return p.EndByte - p.CurrentByte
}

// Progress returns the completion percentage of this part
func (p *DownloadPart) Progress() float64 {
	if p.Size() == 0 {
		return 0
	}
	return float64(p.Downloaded()) / float64(p.Size()) * 100
}

// IsCompleted returns true if this part is completed
func (p *DownloadPart) IsCompleted() bool {
	return p.Status == StatusCompleted || p.CurrentByte >= p.EndByte
}

// IsPaused returns true if this part is paused
func (p *DownloadPart) IsPaused() bool {
	return p.Status == StatusPaused
}

// IsActive returns true if this part is currently downloading
func (p *DownloadPart) IsActive() bool {
	return p.Status == StatusDownloading
}

// CanResume returns true if this part can be resumed
func (p *DownloadPart) CanResume() bool {
	return p.Status == StatusPaused || p.Status == StatusFailed
}

// CanStart returns true if this part can be started
func (p *DownloadPart) CanStart() bool {
	return p.Status == StatusPending || p.Status == StatusFailed
}

// Start marks the part as started
func (p *DownloadPart) Start() {
	p.Status = StatusDownloading
	now := time.Now()
	p.StartedAt = &now
	p.LastError = ""
}

// Complete marks the part as completed
func (p *DownloadPart) Complete() {
	p.Status = StatusCompleted
	now := time.Now()
	p.CompletedAt = &now
	p.CurrentByte = p.EndByte
}

// Pause marks the part as paused
func (p *DownloadPart) Pause() {
	p.Status = StatusPaused
}

// SetError sets the last error and increments retry count
func (p *DownloadPart) SetError(err string) {
	p.LastError = err
	p.RetryCount++
	p.Status = StatusFailed
}

// CanRetry returns true if this part can be retried
func (p *DownloadPart) CanRetry() bool {
	return p.RetryCount < p.MaxRetries && p.Status == StatusFailed
}

// UpdateProgress updates the current byte position and speed
func (p *DownloadPart) UpdateProgress(currentByte int64, speed int64) {
	p.CurrentByte = currentByte
	p.Speed = speed

	// Auto-complete if we've reached the end
	if p.CurrentByte >= p.EndByte && p.Status == StatusDownloading {
		p.Complete()
	}
}
