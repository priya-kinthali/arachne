package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// StorageBackend defines the interface for different storage backends
type StorageBackend interface {
	Save(ctx context.Context, data []ScrapedData) error
	Load(ctx context.Context) ([]ScrapedData, error)
	Close() error
}

// JSONStorage implements StorageBackend for JSON file storage
type JSONStorage struct {
	filename string
}

// NewJSONStorage creates a new JSON storage backend
func NewJSONStorage(filename string) *JSONStorage {
	return &JSONStorage{filename: filename}
}

// Save saves scraped data to JSON file
func (j *JSONStorage) Save(ctx context.Context, data []ScrapedData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return os.WriteFile(j.filename, jsonData, 0644)
}

// Load loads scraped data from JSON file
func (j *JSONStorage) Load(ctx context.Context) ([]ScrapedData, error) {
	data, err := os.ReadFile(j.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []ScrapedData{}, nil
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var results []ScrapedData
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return results, nil
}

// Close implements StorageBackend interface
func (j *JSONStorage) Close() error {
	return nil
}

// MemoryStorage implements StorageBackend for in-memory storage (useful for testing)
type MemoryStorage struct {
	data []ScrapedData
}

// NewMemoryStorage creates a new in-memory storage backend
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make([]ScrapedData, 0)}
}

// Save saves scraped data to memory
func (m *MemoryStorage) Save(ctx context.Context, data []ScrapedData) error {
	m.data = append(m.data, data...)
	return nil
}

// Load loads scraped data from memory
func (m *MemoryStorage) Load(ctx context.Context) ([]ScrapedData, error) {
	return m.data, nil
}

// Close implements StorageBackend interface
func (m *MemoryStorage) Close() error {
	return nil
}

// StorageManager manages storage operations
type StorageManager struct {
	backend StorageBackend
}

// NewStorageManager creates a new storage manager
func NewStorageManager(backend StorageBackend) *StorageManager {
	return &StorageManager{backend: backend}
}

// SaveResults saves scraping results
func (sm *StorageManager) SaveResults(ctx context.Context, results []ScrapedData) error {
	return sm.backend.Save(ctx, results)
}

// LoadResults loads previously saved results
func (sm *StorageManager) LoadResults(ctx context.Context) ([]ScrapedData, error) {
	return sm.backend.Load(ctx)
}

// Close closes the storage backend
func (sm *StorageManager) Close() error {
	return sm.backend.Close()
}
