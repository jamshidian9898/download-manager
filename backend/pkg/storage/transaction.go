package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TransactionalStorage provides atomic file operations for JSON data
type TransactionalStorage struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewTransactionalStorage creates a new transactional storage instance
func NewTransactionalStorage(dataDir string) (*TransactionalStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &TransactionalStorage{
		dataDir: dataDir,
	}, nil
}

// ReadJSON reads and unmarshals JSON data from a file
func (ts *TransactionalStorage) ReadJSON(filename string, data interface{}) error {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	filePath := filepath.Join(ts.dataDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if err := json.Unmarshal(fileData, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	return nil
}

// WriteJSON marshals and writes JSON data to a file atomically
func (ts *TransactionalStorage) WriteJSON(filename string, data interface{}) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	filePath := filepath.Join(ts.dataDir, filename)
	tempPath := filePath + ".tmp"

	// Marshal data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file first
	if err := ioutil.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to target file
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temporary file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (ts *TransactionalStorage) Exists(filename string) bool {
	filePath := filepath.Join(ts.dataDir, filename)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Delete removes a file
func (ts *TransactionalStorage) Delete(filename string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	filePath := filepath.Join(ts.dataDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}
	return nil
}

// Backup creates a backup of a file with timestamp
func (ts *TransactionalStorage) Backup(filename string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	filePath := filepath.Join(ts.dataDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // Nothing to backup
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(ts.dataDir, fmt.Sprintf("%s.backup.%s", filename, timestamp))

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// ListFiles returns a list of all files in the data directory
func (ts *TransactionalStorage) ListFiles() ([]string, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	files, err := ioutil.ReadDir(ts.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// GetDataDir returns the data directory path
func (ts *TransactionalStorage) GetDataDir() string {
	return ts.dataDir
}
