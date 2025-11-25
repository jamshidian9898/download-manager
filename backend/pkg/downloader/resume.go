package downloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// ResumeManager handles download resume functionality
type ResumeManager struct {
	httpDownloader      *HTTPDownloader
	multiPartDownloader *MultiPartDownloader
}

// NewResumeManager creates a new resume manager
func NewResumeManager() *ResumeManager {
	return &ResumeManager{
		httpDownloader:      NewHTTPDownloader(),
		multiPartDownloader: NewMultiPartDownloader(8),
	}
}

// ResumeInfo contains information about resumable downloads
type ResumeInfo struct {
	CanResume       bool
	ExistingSize    int64
	MissingParts    []*models.DownloadPart
	CompletedParts  []*models.DownloadPart
	IsMultiPart     bool
	RequiresRestart bool
}

// CheckResumeCapability checks if a download can be resumed
func (rm *ResumeManager) CheckResumeCapability(ctx context.Context, download *models.DownloadItem, parts []*models.DownloadPart) (*ResumeInfo, error) {
	info := &ResumeInfo{
		CanResume:      false,
		ExistingSize:   0,
		MissingParts:   make([]*models.DownloadPart, 0),
		CompletedParts: make([]*models.DownloadPart, 0),
		IsMultiPart:    len(parts) > 0,
	}

	// Check if this is a multi-part download
	if info.IsMultiPart {
		return rm.checkMultiPartResume(ctx, download, parts, info)
	}

	// Single-part download resume check
	return rm.checkSinglePartResume(ctx, download, info)
}

// checkSinglePartResume checks resume capability for single-part downloads
func (rm *ResumeManager) checkSinglePartResume(ctx context.Context, download *models.DownloadItem, info *ResumeInfo) (*ResumeInfo, error) {
	// Check if the file exists
	if stat, err := os.Stat(download.FilePath); err == nil {
		info.ExistingSize = stat.Size()

		// Get download info to check if server supports resume
		downloadInfo, err := rm.httpDownloader.GetDownloadInfo(ctx, download.URL, download.Headers)
		if err != nil {
			return info, fmt.Errorf("failed to get download info: %w", err)
		}

		// Check if we can resume
		if downloadInfo.SupportsRange && info.ExistingSize < downloadInfo.ContentLength {
			info.CanResume = true
		} else if info.ExistingSize >= downloadInfo.ContentLength {
			// File is already complete
			info.CanResume = false
			download.Complete()
		} else {
			// Server doesn't support resume, need to restart
			info.RequiresRestart = true
		}
	}

	return info, nil
}

// checkMultiPartResume checks resume capability for multi-part downloads
func (rm *ResumeManager) checkMultiPartResume(ctx context.Context, download *models.DownloadItem, parts []*models.DownloadPart, info *ResumeInfo) (*ResumeInfo, error) {
	// Check each part's status
	for _, part := range parts {
		if part.IsCompleted() {
			info.CompletedParts = append(info.CompletedParts, part)
			info.ExistingSize += part.Downloaded()
		} else {
			// Check if part file exists and has some data
			if stat, err := os.Stat(part.FilePath); err == nil && stat.Size() > 0 {
				// Update part's current position based on file size
				part.CurrentByte = part.StartByte + stat.Size()
				if part.CurrentByte > part.EndByte {
					part.CurrentByte = part.EndByte
				}
				info.ExistingSize += part.Downloaded()
			}
			info.MissingParts = append(info.MissingParts, part)
		}
	}

	// We can resume if there are any missing parts
	info.CanResume = len(info.MissingParts) > 0

	// Check if all parts are complete
	if len(info.MissingParts) == 0 {
		// All parts are complete, check if final file exists
		if _, err := os.Stat(download.FilePath); os.IsNotExist(err) {
			// Need to merge parts
			if err := rm.multiPartDownloader.MergeParts(download, parts); err != nil {
				return info, fmt.Errorf("failed to merge parts: %w", err)
			}

			// Clean up part files
			rm.multiPartDownloader.CleanupParts(download, parts)

			download.Complete()
		}
		info.CanResume = false
	}

	return info, nil
}

// ResumeDownload resumes a paused or failed download
func (rm *ResumeManager) ResumeDownload(ctx context.Context, download *models.DownloadItem, parts []*models.DownloadPart, progressCallback func(*MultiPartProgress)) error {
	resumeInfo, err := rm.CheckResumeCapability(ctx, download, parts)
	if err != nil {
		return fmt.Errorf("failed to check resume capability: %w", err)
	}

	if !resumeInfo.CanResume {
		if resumeInfo.RequiresRestart {
			return rm.RestartDownload(ctx, download, parts, progressCallback)
		}
		fmt.Errorf("download cannot be resumed")
		return fmt.Errorf("download cannot be resumed")
	}

	// Update download progress based on existing data
	download.Progress.BytesDownloaded = resumeInfo.ExistingSize
	if download.TotalSize > 0 {
		download.Progress.CalculatePercentage()
	}

	// Resume based on download type
	if resumeInfo.IsMultiPart {
		return rm.resumeMultiPartDownload(ctx, download, resumeInfo.MissingParts, progressCallback)
	}

	return rm.resumeSinglePartDownload(ctx, download, progressCallback)
}

// resumeSinglePartDownload resumes a single-part download
func (rm *ResumeManager) resumeSinglePartDownload(ctx context.Context, download *models.DownloadItem, progressCallback func(*MultiPartProgress)) error {
	download.Start()

	err := rm.httpDownloader.DownloadSingle(ctx, download, func(downloaded, speed int64) {
		download.UpdateProgress(downloaded, speed)

		if progressCallback != nil {
			progress := &MultiPartProgress{
				TotalDownloaded: downloaded,
				TotalSpeed:      speed,
				CompletedParts:  0,
				TotalParts:      1,
			}
			progressCallback(progress)
		}
	})

	if err != nil {
		download.SetError(err.Error())
		return err
	}

	download.Complete()
	return nil
}

// resumeMultiPartDownload resumes a multi-part download
func (rm *ResumeManager) resumeMultiPartDownload(ctx context.Context, download *models.DownloadItem, missingParts []*models.DownloadPart, progressCallback func(*MultiPartProgress)) error {
	download.Start()

	// Download only the missing parts
	err := rm.multiPartDownloader.DownloadParts(ctx, download, missingParts, progressCallback)
	if err != nil {
		download.SetError(err.Error())
		return err
	}

	// Get all parts (completed + newly downloaded)
	allParts := make([]*models.DownloadPart, 0)
	// This would typically come from storage, but for now we'll assume parts are passed in correctly

	// Merge all parts into final file
	if err := rm.multiPartDownloader.MergeParts(download, allParts); err != nil {
		download.SetError(fmt.Sprintf("failed to merge parts: %v", err))
		return err
	}

	// Clean up temporary files
	rm.multiPartDownloader.CleanupParts(download, allParts)

	download.Complete()
	return nil
}

// RestartDownload completely restarts a download (removes existing files)
func (rm *ResumeManager) RestartDownload(ctx context.Context, download *models.DownloadItem, parts []*models.DownloadPart, progressCallback func(*MultiPartProgress)) error {
	// Remove existing files
	if err := rm.CleanupExistingFiles(download, parts); err != nil {
		return fmt.Errorf("failed to cleanup existing files: %w", err)
	}

	// Reset download progress
	download.Progress.BytesDownloaded = 0
	download.Progress.Speed = 0
	download.Progress.Percentage = 0
	download.RetryCount = 0
	download.LastError = ""

	// Reset parts if multi-part
	for _, part := range parts {
		part.CurrentByte = part.StartByte
		part.Status = models.StatusPending
		part.RetryCount = 0
		part.LastError = ""
	}

	// Start fresh download
	if len(parts) > 0 {
		return rm.resumeMultiPartDownload(ctx, download, parts, progressCallback)
	}

	return rm.resumeSinglePartDownload(ctx, download, progressCallback)
}

// CleanupExistingFiles removes existing download files and parts
func (rm *ResumeManager) CleanupExistingFiles(download *models.DownloadItem, parts []*models.DownloadPart) error {
	var lastError error

	// Remove main download file
	if err := os.Remove(download.FilePath); err != nil && !os.IsNotExist(err) {
		lastError = err
	}

	// Remove part files
	for _, part := range parts {
		if err := os.Remove(part.FilePath); err != nil && !os.IsNotExist(err) {
			lastError = err
		}
	}

	// Remove parts directory
	if len(parts) > 0 {
		tempDir := filepath.Join(filepath.Dir(download.FilePath), ".parts", fmt.Sprintf("download_%d", download.ID))
		if err := os.RemoveAll(tempDir); err != nil && !os.IsNotExist(err) {
			lastError = err
		}
	}

	return lastError
}

// ValidateDownload validates that a completed download is intact
func (rm *ResumeManager) ValidateDownload(download *models.DownloadItem) error {
	stat, err := os.Stat(download.FilePath)
	if err != nil {
		return fmt.Errorf("download file not found: %w", err)
	}

	// Check file size matches expected size
	if download.TotalSize > 0 && stat.Size() != download.TotalSize {
		return fmt.Errorf("file size mismatch: expected %d, got %d", download.TotalSize, stat.Size())
	}

	return nil
}
