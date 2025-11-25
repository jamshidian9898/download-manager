package storage

import (
	"fmt"
	"sync"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// DownloadStorage handles persistence of download items
type DownloadStorage struct {
	storage   *TransactionalStorage
	downloads map[int64]*models.DownloadItem
	nextID    int64
	mutex     sync.RWMutex
}

// DownloadData represents the structure saved to JSON
type DownloadData struct {
	Downloads map[int64]*models.DownloadItem `json:"downloads"`
	NextID    int64                          `json:"next_id"`
}

const downloadsFileName = "downloads.json"

// NewDownloadStorage creates a new download storage instance
func NewDownloadStorage(storage *TransactionalStorage) *DownloadStorage {
	ds := &DownloadStorage{
		storage:   storage,
		downloads: make(map[int64]*models.DownloadItem),
		nextID:    1,
	}

	// Load existing data
	ds.load()
	return ds
}

// load reads downloads from storage
func (ds *DownloadStorage) load() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if !ds.storage.Exists(downloadsFileName) {
		return nil // No existing data
	}

	var data DownloadData
	if err := ds.storage.ReadJSON(downloadsFileName, &data); err != nil {
		return fmt.Errorf("failed to load downloads: %w", err)
	}

	ds.downloads = data.Downloads
	ds.nextID = data.NextID

	// Ensure we have valid maps
	if ds.downloads == nil {
		ds.downloads = make(map[int64]*models.DownloadItem)
	}

	return nil
}

// save writes downloads to storage
func (ds *DownloadStorage) save() error {
	data := DownloadData{
		Downloads: ds.downloads,
		NextID:    ds.nextID,
	}

	if err := ds.storage.WriteJSON(downloadsFileName, data); err != nil {
		return fmt.Errorf("failed to save downloads: %w", err)
	}

	return nil
}

// Create adds a new download item
func (ds *DownloadStorage) Create(download *models.DownloadItem) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	// Assign ID if not set
	if download.ID == 0 {
		download.ID = ds.nextID
		ds.nextID++
	}

	ds.downloads[download.ID] = download
	return ds.save()
}

// GetByID retrieves a download by ID
func (ds *DownloadStorage) GetByID(id int64) (*models.DownloadItem, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	download, exists := ds.downloads[id]
	if !exists {
		return nil, fmt.Errorf("download with ID %d not found", id)
	}

	return download, nil
}

// Update updates an existing download item
func (ds *DownloadStorage) Update(download *models.DownloadItem) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if _, exists := ds.downloads[download.ID]; !exists {
		return fmt.Errorf("download with ID %d not found", download.ID)
	}

	ds.downloads[download.ID] = download
	return ds.save()
}

// Delete removes a download item
func (ds *DownloadStorage) Delete(id int64) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if _, exists := ds.downloads[id]; !exists {
		return fmt.Errorf("download with ID %d not found", id)
	}

	delete(ds.downloads, id)
	return ds.save()
}

// GetAll returns all download items
func (ds *DownloadStorage) GetAll() ([]*models.DownloadItem, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	downloads := make([]*models.DownloadItem, 0, len(ds.downloads))
	for _, download := range ds.downloads {
		downloads = append(downloads, download)
	}

	return downloads, nil
}

// GetByStatus returns downloads with a specific status
func (ds *DownloadStorage) GetByStatus(status models.DownloadStatus) ([]*models.DownloadItem, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	var downloads []*models.DownloadItem
	for _, download := range ds.downloads {
		if download.Status == status {
			downloads = append(downloads, download)
		}
	}

	return downloads, nil
}

// GetByQueue returns downloads in a specific queue
func (ds *DownloadStorage) GetByQueue(queueID int64) ([]*models.DownloadItem, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	var downloads []*models.DownloadItem
	for _, download := range ds.downloads {
		if download.QueueID == queueID {
			downloads = append(downloads, download)
		}
	}

	return downloads, nil
}

// Count returns the total number of downloads
func (ds *DownloadStorage) Count() int {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	return len(ds.downloads)
}

// CountByStatus returns the count of downloads with a specific status
func (ds *DownloadStorage) CountByStatus(status models.DownloadStatus) int {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	count := 0
	for _, download := range ds.downloads {
		if download.Status == status {
			count++
		}
	}

	return count
}

// GetActiveDownloads returns downloads that are currently active (downloading)
func (ds *DownloadStorage) GetActiveDownloads() ([]*models.DownloadItem, error) {
	return ds.GetByStatus(models.StatusDownloading)
}

// GetPendingDownloads returns downloads that are pending
func (ds *DownloadStorage) GetPendingDownloads() ([]*models.DownloadItem, error) {
	return ds.GetByStatus(models.StatusPending)
}

// GetCompletedDownloads returns downloads that are completed
func (ds *DownloadStorage) GetCompletedDownloads() ([]*models.DownloadItem, error) {
	return ds.GetByStatus(models.StatusCompleted)
}

// GetFailedDownloads returns downloads that have failed
func (ds *DownloadStorage) GetFailedDownloads() ([]*models.DownloadItem, error) {
	return ds.GetByStatus(models.StatusFailed)
}

// Backup creates a backup of the downloads data
func (ds *DownloadStorage) Backup() error {
	return ds.storage.Backup(downloadsFileName)
}

// GetNextID returns the next available ID
func (ds *DownloadStorage) GetNextID() int64 {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	return ds.nextID
}
