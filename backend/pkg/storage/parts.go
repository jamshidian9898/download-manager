package storage

import (
	"fmt"
	"sync"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// PartStorage handles persistence of download parts
type PartStorage struct {
	storage *TransactionalStorage
	parts   map[int64]*models.DownloadPart
	nextID  int64
	mutex   sync.RWMutex
}

// PartData represents the structure saved to JSON
type PartData struct {
	Parts  map[int64]*models.DownloadPart `json:"parts"`
	NextID int64                          `json:"next_id"`
}

const partsFileName = "parts.json"

// NewPartStorage creates a new part storage instance
func NewPartStorage(storage *TransactionalStorage) *PartStorage {
	ps := &PartStorage{
		storage: storage,
		parts:   make(map[int64]*models.DownloadPart),
		nextID:  1,
	}

	// Load existing data
	ps.load()
	return ps
}

// load reads parts from storage
func (ps *PartStorage) load() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if !ps.storage.Exists(partsFileName) {
		return nil // No existing data
	}

	var data PartData
	if err := ps.storage.ReadJSON(partsFileName, &data); err != nil {
		return fmt.Errorf("failed to load parts: %w", err)
	}

	ps.parts = data.Parts
	ps.nextID = data.NextID

	// Ensure we have valid maps
	if ps.parts == nil {
		ps.parts = make(map[int64]*models.DownloadPart)
	}

	return nil
}

// save writes parts to storage
func (ps *PartStorage) save() error {
	data := PartData{
		Parts:  ps.parts,
		NextID: ps.nextID,
	}

	if err := ps.storage.WriteJSON(partsFileName, data); err != nil {
		return fmt.Errorf("failed to save parts: %w", err)
	}

	return nil
}

// Create adds a new download part
func (ps *PartStorage) Create(part *models.DownloadPart) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Assign ID if not set
	if part.ID == 0 {
		part.ID = ps.nextID
		ps.nextID++
	}

	ps.parts[part.ID] = part
	return ps.save()
}

// GetByID retrieves a part by ID
func (ps *PartStorage) GetByID(id int64) (*models.DownloadPart, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	part, exists := ps.parts[id]
	if !exists {
		return nil, fmt.Errorf("part with ID %d not found", id)
	}

	return part, nil
}

// Update updates an existing download part
func (ps *PartStorage) Update(part *models.DownloadPart) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if _, exists := ps.parts[part.ID]; !exists {
		return fmt.Errorf("part with ID %d not found", part.ID)
	}

	ps.parts[part.ID] = part
	return ps.save()
}

// Delete removes a download part
func (ps *PartStorage) Delete(id int64) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if _, exists := ps.parts[id]; !exists {
		return fmt.Errorf("part with ID %d not found", id)
	}

	delete(ps.parts, id)
	return ps.save()
}

// GetByDownloadID returns all parts for a specific download
func (ps *PartStorage) GetByDownloadID(downloadID int64) ([]*models.DownloadPart, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	var parts []*models.DownloadPart
	for _, part := range ps.parts {
		if part.DownloadID == downloadID {
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// GetByStatus returns parts with a specific status
func (ps *PartStorage) GetByStatus(status models.DownloadStatus) ([]*models.DownloadPart, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	var parts []*models.DownloadPart
	for _, part := range ps.parts {
		if part.Status == status {
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// GetAll returns all download parts
func (ps *PartStorage) GetAll() ([]*models.DownloadPart, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	parts := make([]*models.DownloadPart, 0, len(ps.parts))
	for _, part := range ps.parts {
		parts = append(parts, part)
	}

	return parts, nil
}

// DeleteByDownloadID removes all parts for a specific download
func (ps *PartStorage) DeleteByDownloadID(downloadID int64) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for id, part := range ps.parts {
		if part.DownloadID == downloadID {
			delete(ps.parts, id)
		}
	}

	return ps.save()
}

// Count returns the total number of parts
func (ps *PartStorage) Count() int {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	return len(ps.parts)
}

// CountByDownloadID returns the count of parts for a specific download
func (ps *PartStorage) CountByDownloadID(downloadID int64) int {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	count := 0
	for _, part := range ps.parts {
		if part.DownloadID == downloadID {
			count++
		}
	}

	return count
}

// GetActiveParts returns parts that are currently downloading
func (ps *PartStorage) GetActiveParts() ([]*models.DownloadPart, error) {
	return ps.GetByStatus(models.StatusDownloading)
}

// GetCompletedParts returns parts that are completed
func (ps *PartStorage) GetCompletedParts() ([]*models.DownloadPart, error) {
	return ps.GetByStatus(models.StatusCompleted)
}

// GetFailedParts returns parts that have failed
func (ps *PartStorage) GetFailedParts() ([]*models.DownloadPart, error) {
	return ps.GetByStatus(models.StatusFailed)
}

// CreateMultiple creates multiple parts in a single transaction
func (ps *PartStorage) CreateMultiple(parts []*models.DownloadPart) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, part := range parts {
		// Assign ID if not set
		if part.ID == 0 {
			part.ID = ps.nextID
			ps.nextID++
		}
		ps.parts[part.ID] = part
	}

	return ps.save()
}

// UpdateMultiple updates multiple parts in a single transaction
func (ps *PartStorage) UpdateMultiple(parts []*models.DownloadPart) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, part := range parts {
		if _, exists := ps.parts[part.ID]; !exists {
			return fmt.Errorf("part with ID %d not found", part.ID)
		}
		ps.parts[part.ID] = part
	}

	return ps.save()
}

// Backup creates a backup of the parts data
func (ps *PartStorage) Backup() error {
	return ps.storage.Backup(partsFileName)
}

// GetNextID returns the next available ID
func (ps *PartStorage) GetNextID() int64 {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	return ps.nextID
}
