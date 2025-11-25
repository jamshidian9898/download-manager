package downloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// MultiPartDownloader handles multi-part downloads
type MultiPartDownloader struct {
	httpDownloader *HTTPDownloader
	maxConnections int
}

// NewMultiPartDownloader creates a new multi-part downloader
func NewMultiPartDownloader(maxConnections int) *MultiPartDownloader {
	return &MultiPartDownloader{
		httpDownloader: NewHTTPDownloader(),
		maxConnections: maxConnections,
	}
}

// PartProgress represents progress information for a single part
type PartProgress struct {
	PartID      int64
	Downloaded  int64
	Speed       int64
	IsCompleted bool
}

// MultiPartProgress represents progress information for the entire download
type MultiPartProgress struct {
	TotalDownloaded int64
	TotalSpeed      int64
	Parts           []PartProgress
	CompletedParts  int
	TotalParts      int
}

// CreateParts creates download parts for multi-part downloading
func (mpd *MultiPartDownloader) CreateParts(download *models.DownloadItem, info *DownloadInfo) ([]*models.DownloadPart, error) {
	if !info.SupportsRange || info.ContentLength <= 0 {
		return nil, fmt.Errorf("server does not support range requests or content length unknown")
	}

	// Use the download's max connections setting
	numParts := download.MaxConnections
	if numParts <= 0 {
		numParts = mpd.maxConnections
	}

	// Don't create more parts than necessary
	minPartSize := int64(1024 * 1024) // 1MB minimum per part
	if info.ContentLength/int64(numParts) < minPartSize {
		numParts = int(info.ContentLength / minPartSize)
		if numParts == 0 {
			numParts = 1
		}
	}

	partSize := info.ContentLength / int64(numParts)
	var parts []*models.DownloadPart

	// Create temporary directory for parts
	tempDir := filepath.Join(filepath.Dir(download.FilePath), ".parts", fmt.Sprintf("download_%d", download.ID))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	for i := 0; i < numParts; i++ {
		startByte := int64(i) * partSize
		endByte := startByte + partSize - 1

		// Last part gets any remaining bytes
		if i == numParts-1 {
			endByte = info.ContentLength - 1
		}

		partPath := filepath.Join(tempDir, fmt.Sprintf("part_%d.tmp", i))
		part := models.NewDownloadPart(download.ID, i, startByte, endByte, partPath)
		parts = append(parts, part)
	}

	return parts, nil
}

// DownloadParts downloads all parts concurrently
func (mpd *MultiPartDownloader) DownloadParts(ctx context.Context, download *models.DownloadItem, parts []*models.DownloadPart, progressCallback func(*MultiPartProgress)) error {
	if len(parts) == 0 {
		return fmt.Errorf("no parts to download")
	}

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Progress tracking
	var progressMutex sync.RWMutex
	progress := &MultiPartProgress{
		Parts:      make([]PartProgress, len(parts)),
		TotalParts: len(parts),
	}

	// Worker pool for downloading parts
	semaphore := make(chan struct{}, download.MaxConnections)
	var wg sync.WaitGroup
	var firstError error
	var errorMutex sync.Mutex

	// Progress update ticker
	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-progressTicker.C:
				if progressCallback != nil {
					progressMutex.RLock()
					progressCopy := *progress
					progressMutex.RUnlock()
					progressCallback(&progressCopy)
				}
			}
		}
	}()

	// Start downloading parts
	for i, part := range parts {
		if part.IsCompleted() {
			progressMutex.Lock()
			progress.CompletedParts++
			progress.Parts[i] = PartProgress{
				PartID:      part.ID,
				Downloaded:  part.Downloaded(),
				Speed:       0,
				IsCompleted: true,
			}
			progressMutex.Unlock()
			continue
		}

		wg.Add(1)
		go func(partIndex int, part *models.DownloadPart) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check if we should stop due to error
			errorMutex.Lock()
			shouldStop := firstError != nil
			errorMutex.Unlock()

			if shouldStop {
				return
			}

			// Start the part
			part.Start()

			// Download the part
			err := mpd.httpDownloader.DownloadPart(ctx, part, download.URL, download.Headers, func(currentByte int64) {
				progressMutex.Lock()
				progress.Parts[partIndex] = PartProgress{
					PartID:      part.ID,
					Downloaded:  part.Downloaded(),
					Speed:       part.Speed,
					IsCompleted: part.IsCompleted(),
				}

				// Recalculate totals
				progress.TotalDownloaded = 0
				progress.TotalSpeed = 0
				progress.CompletedParts = 0

				for _, p := range progress.Parts {
					progress.TotalDownloaded += p.Downloaded
					progress.TotalSpeed += p.Speed
					if p.IsCompleted {
						progress.CompletedParts++
					}
				}
				progressMutex.Unlock()
			})

			if err != nil {
				part.SetError(err.Error())

				errorMutex.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("part %d failed: %w", part.PartNumber, err)
					cancel() // Cancel all other downloads
				}
				errorMutex.Unlock()
			} else {
				part.Complete()
				progressMutex.Lock()
				progress.Parts[partIndex].IsCompleted = true
				progress.CompletedParts++
				progressMutex.Unlock()
			}
		}(i, part)
	}

	// Wait for all parts to complete
	wg.Wait()

	// Check for errors
	if firstError != nil {
		return firstError
	}

	// Final progress update
	if progressCallback != nil {
		progressCallback(progress)
	}

	return nil
}

// MergeParts combines all downloaded parts into the final file
func (mpd *MultiPartDownloader) MergeParts(download *models.DownloadItem, parts []*models.DownloadPart) error {
	// Create the target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(download.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create the final file
	finalFile, err := os.Create(download.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %w", err)
	}
	defer finalFile.Close()

	// Merge parts in order
	for _, part := range parts {
		if !part.IsCompleted() {
			return fmt.Errorf("part %d is not completed", part.PartNumber)
		}

		partFile, err := os.Open(part.FilePath)
		if err != nil {
			return fmt.Errorf("failed to open part file %s: %w", part.FilePath, err)
		}

		// Copy part content to final file
		buffer := make([]byte, 64*1024) // 64KB buffer
		for {
			n, readErr := partFile.Read(buffer)
			if n > 0 {
				if _, writeErr := finalFile.Write(buffer[:n]); writeErr != nil {
					partFile.Close()
					return fmt.Errorf("failed to write to final file: %w", writeErr)
				}
			}
			if readErr != nil {
				break
			}
		}
		partFile.Close()
	}

	return nil
}

// CleanupParts removes temporary part files and directories
func (mpd *MultiPartDownloader) CleanupParts(download *models.DownloadItem, parts []*models.DownloadPart) error {
	var lastError error

	// Remove individual part files
	for _, part := range parts {
		if err := os.Remove(part.FilePath); err != nil && !os.IsNotExist(err) {
			lastError = err
		}
	}

	// Remove the parts directory
	tempDir := filepath.Join(filepath.Dir(download.FilePath), ".parts", fmt.Sprintf("download_%d", download.ID))
	if err := os.Remove(tempDir); err != nil && !os.IsNotExist(err) {
		lastError = err
	}

	// Try to remove the .parts directory if it's empty
	partsDir := filepath.Join(filepath.Dir(download.FilePath), ".parts")
	os.Remove(partsDir) // Ignore error as it might not be empty

	return lastError
}

// CanUseMultiPart checks if multi-part download is beneficial
func (mpd *MultiPartDownloader) CanUseMultiPart(info *DownloadInfo, minFileSize int64) bool {
	return info.SupportsRange && info.ContentLength > minFileSize
}

// SetMaxConnections sets the maximum number of concurrent connections
func (mpd *MultiPartDownloader) SetMaxConnections(max int) {
	mpd.maxConnections = max
}

// GetHTTPDownloader returns the underlying HTTP downloader
func (mpd *MultiPartDownloader) GetHTTPDownloader() *HTTPDownloader {
	return mpd.httpDownloader
}
