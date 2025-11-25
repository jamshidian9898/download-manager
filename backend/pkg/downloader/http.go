package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// HTTPDownloader handles HTTP downloads with support for range requests
type HTTPDownloader struct {
	client     *http.Client
	userAgent  string
	timeout    time.Duration
	bufferSize int
}

// NewHTTPDownloader creates a new HTTP downloader
func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent:  "Download Manager 1.0",
		timeout:    30 * time.Second,
		bufferSize: 32 * 1024, // 32KB buffer
	}
}

// DownloadInfo contains information about a download URL
type DownloadInfo struct {
	URL           string
	ContentLength int64
	SupportsRange bool
	Filename      string
	ContentType   string
	LastModified  *time.Time
	ETag          string
}

// GetDownloadInfo retrieves information about a download URL
func (d *HTTPDownloader) GetDownloadInfo(ctx context.Context, url string, headers map[string]string) (*DownloadInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", d.userAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	info := &DownloadInfo{
		URL:           url,
		ContentLength: resp.ContentLength,
		SupportsRange: strings.Contains(resp.Header.Get("Accept-Ranges"), "bytes"),
		ContentType:   resp.Header.Get("Content-Type"),
		ETag:          resp.Header.Get("ETag"),
	}

	// Extract filename from Content-Disposition or URL
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if filename := extractFilenameFromContentDisposition(cd); filename != "" {
			info.Filename = filename
		}
	}
	if info.Filename == "" {
		info.Filename = extractFilenameFromURL(url)
	}

	// Parse Last-Modified header
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(time.RFC1123, lm); err == nil {
			info.LastModified = &t
		}
	}

	return info, nil
}

// DownloadPart downloads a specific part of a file
func (d *HTTPDownloader) DownloadPart(ctx context.Context, part *models.DownloadPart, url string, headers map[string]string, progressCallback func(int64)) error {
	// Create request with range header
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", d.userAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set range header for partial download
	rangeHeader := fmt.Sprintf("bytes=%d-%d", part.CurrentByte, part.EndByte)
	req.Header.Set("Range", rangeHeader)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for partial content response
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Create or open the part file
	if err := os.MkdirAll(filepath.Dir(part.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(part.FilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to the current position if resuming
	if part.CurrentByte > part.StartByte {
		if _, err := file.Seek(part.CurrentByte-part.StartByte, 0); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	}

	// Download with progress tracking
	buffer := make([]byte, d.bufferSize)
	lastProgressUpdate := time.Now()
	bytesAtLastUpdate := part.CurrentByte

	for part.CurrentByte < part.EndByte {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate how much to read (don't exceed part boundary)
		remaining := part.EndByte - part.CurrentByte + 1
		readSize := int64(len(buffer))
		if remaining < readSize {
			readSize = remaining
		}

		n, err := resp.Body.Read(buffer[:readSize])
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			part.CurrentByte += int64(n)

			// Update progress periodically
			now := time.Now()
			if now.Sub(lastProgressUpdate) >= time.Second {
				bytesInSecond := part.CurrentByte - bytesAtLastUpdate
				part.Speed = bytesInSecond

				if progressCallback != nil {
					progressCallback(part.CurrentByte)
				}

				lastProgressUpdate = now
				bytesAtLastUpdate = part.CurrentByte
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// Final progress update
	if progressCallback != nil {
		progressCallback(part.CurrentByte)
	}

	return nil
}

// DownloadSingle downloads a file in a single connection (no multi-part)
func (d *HTTPDownloader) DownloadSingle(ctx context.Context, download *models.DownloadItem, progressCallback func(int64, int64)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", download.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", download.UserAgent)
	if download.Referrer != "" {
		req.Header.Set("Referer", download.Referrer)
	}
	for key, value := range download.Headers {
		req.Header.Set(key, value)
	}

	// Add resume support if file exists
	var resumeOffset int64
	if _, err := os.Stat(download.FilePath); err == nil {
		if stat, err := os.Stat(download.FilePath); err == nil {
			resumeOffset = stat.Size()
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
		}
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(download.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file for writing
	var file *os.File
	if resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent {
		file, err = os.OpenFile(download.FilePath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(download.FilePath)
		resumeOffset = 0
	}
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Update total size if available
	if download.TotalSize == 0 && resp.ContentLength > 0 {
		download.TotalSize = resp.ContentLength + resumeOffset
	}

	// Download with progress tracking
	buffer := make([]byte, d.bufferSize)
	downloaded := resumeOffset
	lastProgressUpdate := time.Now()
	bytesAtLastUpdate := downloaded

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			downloaded += int64(n)

			// Update progress periodically
			now := time.Now()
			if now.Sub(lastProgressUpdate) >= time.Second {
				bytesInSecond := downloaded - bytesAtLastUpdate

				if progressCallback != nil {
					progressCallback(downloaded, bytesInSecond)
				}

				lastProgressUpdate = now
				bytesAtLastUpdate = downloaded
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// Final progress update
	if progressCallback != nil {
		progressCallback(downloaded, 0)
	}

	return nil
}

// extractFilenameFromContentDisposition extracts filename from Content-Disposition header
func extractFilenameFromContentDisposition(cd string) string {
	// Simple extraction - in production, you might want more robust parsing
	parts := strings.Split(cd, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.Trim(filename, "\"")
			return filename
		}
	}
	return ""
}

// extractFilenameFromURL extracts filename from URL path
func extractFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// Remove query parameters
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		if filename != "" {
			return filename
		}
	}
	return "download"
}

// SetUserAgent sets the user agent for requests
func (d *HTTPDownloader) SetUserAgent(userAgent string) {
	d.userAgent = userAgent
}

// SetTimeout sets the timeout for requests
func (d *HTTPDownloader) SetTimeout(timeout time.Duration) {
	d.timeout = timeout
	d.client.Timeout = timeout
}

// SetBufferSize sets the buffer size for downloads
func (d *HTTPDownloader) SetBufferSize(size int) {
	d.bufferSize = size
}
