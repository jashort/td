package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Storage handles loading and saving todo lists
type Storage struct {
	filepath string
}

// NewStorage creates a new storage instance with the given file path
func NewStorage(filepath string) *Storage {
	return &Storage{filepath: filepath}
}

// Load loads the todo list from disk
func (s *Storage) Load() (*TodoList, error) {
	// Check if file exists
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		// File doesn't exist, return empty list
		return NewTodoList(), nil
	}

	// Read file
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var list TodoList
	if err := json.Unmarshal(data, &list); err != nil {
		// Try to backup corrupted file
		backupPath := s.filepath + ".backup"
		os.Rename(s.filepath, backupPath)
		return nil, fmt.Errorf("failed to parse JSON (backed up to %s): %w", backupPath, err)
	}

	return &list, nil
}

// Save saves the todo list to disk atomically
func (s *Storage) Save(list *TodoList) error {
	// Ensure directory exists
	dir := filepath.Dir(s.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file
	tmpPath := s.filepath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.filepath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetDefaultPath returns the default storage path (~/.td/todos.json)
func GetDefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".td", "todos.json"), nil
}

// ResolveStoragePath resolves the storage path based on priority:
// 1. CLI flag (if provided)
// 2. Environment variable TD_FILE
// 3. Default path (~/.td/todos.json)
func ResolveStoragePath(flagPath string) (string, error) {
	// 1. Check CLI flag
	if flagPath != "" {
		return flagPath, nil
	}

	// 2. Check environment variable
	if envPath := os.Getenv("TD_FILE"); envPath != "" {
		return envPath, nil
	}

	// 3. Use default
	return GetDefaultPath()
}
