package storage

import (
	"fmt"
	"sync"

	"github.com/jamshidian/download-manager/backend/pkg/models"
)

// QueueStorage handles persistence of download queues
type QueueStorage struct {
	storage *TransactionalStorage
	queues  map[int64]*models.QueueModel
	nextID  int64
	mutex   sync.RWMutex
}

// QueueData represents the structure saved to JSON
type QueueData struct {
	Queues map[int64]*models.QueueModel `json:"queues"`
	NextID int64                        `json:"next_id"`
}

const queuesFileName = "queues.json"

// NewQueueStorage creates a new queue storage instance
func NewQueueStorage(storage *TransactionalStorage) *QueueStorage {
	qs := &QueueStorage{
		storage: storage,
		queues:  make(map[int64]*models.QueueModel),
		nextID:  1,
	}

	// Load existing data
	qs.load()

	// Create default queue if none exist
	if len(qs.queues) == 0 {
		defaultQueue := models.NewQueueModel("Default Queue")
		qs.Create(defaultQueue)
	}

	return qs
}

// load reads queues from storage
func (qs *QueueStorage) load() error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	if !qs.storage.Exists(queuesFileName) {
		return nil // No existing data
	}

	var data QueueData
	if err := qs.storage.ReadJSON(queuesFileName, &data); err != nil {
		return fmt.Errorf("failed to load queues: %w", err)
	}

	qs.queues = data.Queues
	qs.nextID = data.NextID

	// Ensure we have valid maps
	if qs.queues == nil {
		qs.queues = make(map[int64]*models.QueueModel)
	}

	// Initialize runtime fields for loaded queues
	for _, queue := range qs.queues {
		if queue.ActiveDownloads == nil {
			queue.ActiveDownloads = make(map[int64]bool)
		}
		if queue.DownloadOrder == nil {
			queue.DownloadOrder = make([]int64, 0)
		}
	}

	return nil
}

// save writes queues to storage
func (qs *QueueStorage) save() error {
	data := QueueData{
		Queues: qs.queues,
		NextID: qs.nextID,
	}

	if err := qs.storage.WriteJSON(queuesFileName, data); err != nil {
		return fmt.Errorf("failed to save queues: %w", err)
	}

	return nil
}

// Create adds a new queue
func (qs *QueueStorage) Create(queue *models.QueueModel) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	// Assign ID if not set
	if queue.ID == 0 {
		queue.ID = qs.nextID
		qs.nextID++
	}

	// Initialize runtime fields
	if queue.ActiveDownloads == nil {
		queue.ActiveDownloads = make(map[int64]bool)
	}
	if queue.DownloadOrder == nil {
		queue.DownloadOrder = make([]int64, 0)
	}

	qs.queues[queue.ID] = queue
	return qs.save()
}

// GetByID retrieves a queue by ID
func (qs *QueueStorage) GetByID(id int64) (*models.QueueModel, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	queue, exists := qs.queues[id]
	if !exists {
		return nil, fmt.Errorf("queue with ID %d not found", id)
	}

	return queue, nil
}

// Update updates an existing queue
func (qs *QueueStorage) Update(queue *models.QueueModel) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	if _, exists := qs.queues[queue.ID]; !exists {
		return fmt.Errorf("queue with ID %d not found", queue.ID)
	}

	qs.queues[queue.ID] = queue
	return qs.save()
}

// Delete removes a queue
func (qs *QueueStorage) Delete(id int64) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	if _, exists := qs.queues[id]; !exists {
		return fmt.Errorf("queue with ID %d not found", id)
	}

	delete(qs.queues, id)
	return qs.save()
}

// GetAll returns all queues
func (qs *QueueStorage) GetAll() ([]*models.QueueModel, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	queues := make([]*models.QueueModel, 0, len(qs.queues))
	for _, queue := range qs.queues {
		queues = append(queues, queue)
	}

	return queues, nil
}

// GetByStatus returns queues with a specific status
func (qs *QueueStorage) GetByStatus(status models.QueueStatus) ([]*models.QueueModel, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	var queues []*models.QueueModel
	for _, queue := range qs.queues {
		if queue.Status == status {
			queues = append(queues, queue)
		}
	}

	return queues, nil
}

// GetRunningQueues returns all running queues
func (qs *QueueStorage) GetRunningQueues() ([]*models.QueueModel, error) {
	return qs.GetByStatus(models.QueueStatusRunning)
}

// GetDefaultQueue returns the default queue (first queue or creates one)
func (qs *QueueStorage) GetDefaultQueue() (*models.QueueModel, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	// Return first queue if any exist
	for _, queue := range qs.queues {
		return queue, nil
	}

	// This shouldn't happen as we create a default queue in NewQueueStorage
	return nil, fmt.Errorf("no default queue found")
}

// Count returns the total number of queues
func (qs *QueueStorage) Count() int {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	return len(qs.queues)
}

// GetByName returns a queue by name
func (qs *QueueStorage) GetByName(name string) (*models.QueueModel, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	for _, queue := range qs.queues {
		if queue.Name == name {
			return queue, nil
		}
	}

	return nil, fmt.Errorf("queue with name '%s' not found", name)
}

// AddDownloadToQueue adds a download to a queue's order
func (qs *QueueStorage) AddDownloadToQueue(queueID, downloadID int64) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	queue, exists := qs.queues[queueID]
	if !exists {
		return fmt.Errorf("queue with ID %d not found", queueID)
	}

	queue.AddDownload(downloadID)
	return qs.save()
}

// RemoveDownloadFromQueue removes a download from a queue's order
func (qs *QueueStorage) RemoveDownloadFromQueue(queueID, downloadID int64) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	queue, exists := qs.queues[queueID]
	if !exists {
		return fmt.Errorf("queue with ID %d not found", queueID)
	}

	queue.RemoveDownload(downloadID)
	return qs.save()
}

// SetDownloadActive marks a download as active in a queue
func (qs *QueueStorage) SetDownloadActive(queueID, downloadID int64) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	queue, exists := qs.queues[queueID]
	if !exists {
		return fmt.Errorf("queue with ID %d not found", queueID)
	}

	queue.SetActive(downloadID)
	return qs.save()
}

// SetDownloadInactive marks a download as inactive in a queue
func (qs *QueueStorage) SetDownloadInactive(queueID, downloadID int64) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	queue, exists := qs.queues[queueID]
	if !exists {
		return fmt.Errorf("queue with ID %d not found", queueID)
	}

	queue.SetInactive(downloadID)
	return qs.save()
}

// MoveDownloadInQueue moves a download to a new position in the queue
func (qs *QueueStorage) MoveDownloadInQueue(queueID, downloadID int64, newPosition int) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	queue, exists := qs.queues[queueID]
	if !exists {
		return fmt.Errorf("queue with ID %d not found", queueID)
	}

	queue.MoveDownload(downloadID, newPosition)
	return qs.save()
}

// Backup creates a backup of the queues data
func (qs *QueueStorage) Backup() error {
	return qs.storage.Backup(queuesFileName)
}

// GetNextID returns the next available ID
func (qs *QueueStorage) GetNextID() int64 {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	return qs.nextID
}
