package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Storage defines the interface for flow storage
type Storage interface {
	// SaveFlow saves a flow to storage
	SaveFlow(id string, flow []byte) error
	
	// LoadFlow loads a flow from storage
	LoadFlow(id string) ([]byte, error)
	
	// DeleteFlow deletes a flow from storage
	DeleteFlow(id string) error
	
	// ListFlows lists all flow IDs in storage
	ListFlows() ([]string, error)
}

// FileStorage implements file-based storage for flows
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new FileStorage
func NewFileStorage(baseDir string) (*FileStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	
	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

// SaveFlow saves a flow to a file
func (fs *FileStorage) SaveFlow(id string, flow []byte) error {
	if id == "" {
		return errors.New("flow ID cannot be empty")
	}
	
	// Sanitize ID for use as a filename
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "\\", "_")
	
	filePath := filepath.Join(fs.baseDir, id+".json")
	return ioutil.WriteFile(filePath, flow, 0644)
}

// LoadFlow loads a flow from a file
func (fs *FileStorage) LoadFlow(id string) ([]byte, error) {
	if id == "" {
		return nil, errors.New("flow ID cannot be empty")
	}
	
	// Sanitize ID for use as a filename
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "\\", "_")
	
	filePath := filepath.Join(fs.baseDir, id+".json")
	return ioutil.ReadFile(filePath)
}

// DeleteFlow deletes a flow file
func (fs *FileStorage) DeleteFlow(id string) error {
	if id == "" {
		return errors.New("flow ID cannot be empty")
	}
	
	// Sanitize ID for use as a filename
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "\\", "_")
	
	filePath := filepath.Join(fs.baseDir, id+".json")
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("flow does not exist")
	}
	
	return os.Remove(filePath)
}

// ListFlows lists all flow IDs in the directory
func (fs *FileStorage) ListFlows() ([]string, error) {
	files, err := ioutil.ReadDir(fs.baseDir)
	if err != nil {
		return nil, err
	}
	
	flows := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Remove .json extension
			name := strings.TrimSuffix(file.Name(), ".json")
			flows = append(flows, name)
		}
	}
	
	return flows, nil
}
