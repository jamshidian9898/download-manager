package models

import "time"

// DownloadStatus represents the current state of a download
type DownloadStatus int

const (
	StatusPending DownloadStatus = iota
	StatusDownloading
	StatusPaused
	StatusCompleted
	StatusFailed
	StatusCanceled
)

func (s DownloadStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusDownloading:
		return "downloading"
	case StatusPaused:
		return "paused"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// QueueStatus represents the state of a download queue
type QueueStatus int

const (
	QueueStatusStopped QueueStatus = iota
	QueueStatusRunning
	QueueStatusPaused
)

func (s QueueStatus) String() string {
	switch s {
	case QueueStatusStopped:
		return "stopped"
	case QueueStatusRunning:
		return "running"
	case QueueStatusPaused:
		return "paused"
	default:
		return "unknown"
	}
}

// Progress represents download progress information
type Progress struct {
	BytesDownloaded int64     `json:"bytes_downloaded"`
	TotalBytes      int64     `json:"total_bytes"`
	Speed           int64     `json:"speed"`      // bytes per second
	ETA             int64     `json:"eta"`        // estimated time remaining in seconds
	Percentage      float64   `json:"percentage"` // completion percentage
	LastUpdate      time.Time `json:"last_update"`
}

// CalculatePercentage calculates the completion percentage
func (p *Progress) CalculatePercentage() {
	if p.TotalBytes > 0 {
		p.Percentage = float64(p.BytesDownloaded) / float64(p.TotalBytes) * 100
	}
}

// CalculateETA calculates estimated time remaining
func (p *Progress) CalculateETA() {
	if p.Speed > 0 {
		remaining := p.TotalBytes - p.BytesDownloaded
		p.ETA = remaining / p.Speed
	}
}
