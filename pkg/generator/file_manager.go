package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultFileManager implements the FileManager interface
type DefaultFileManager struct{}

// NewFileManager creates a new DefaultFileManager instance
func NewFileManager() *DefaultFileManager {
	return &DefaultFileManager{}
}

// ReadFile reads a file from the specified path and returns its contents
func (fm *DefaultFileManager) ReadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, &FileError{
			Operation: "read",
			Path:      path,
			Message:   "file path cannot be empty",
		}
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return nil, &FileError{
			Operation: "read",
			Path:      cleanPath,
			Message:   "file does not exist",
		}
	}

	// Read the file
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, &FileError{
			Operation: "read",
			Path:      cleanPath,
			Message:   fmt.Sprintf("failed to read file: %v", err),
		}
	}

	return data, nil
}

// WriteFile writes content to the specified file path
func (fm *DefaultFileManager) WriteFile(path string, content []byte) error {
	if path == "" {
		return &FileError{
			Operation: "write",
			Path:      path,
			Message:   "file path cannot be empty",
		}
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Validate file extension for Go files
	if !strings.HasSuffix(cleanPath, ".go") {
		return &FileError{
			Operation: "write",
			Path:      cleanPath,
			Message:   "file must have .go extension",
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(cleanPath)
	if err := fm.CreateDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	err := os.WriteFile(cleanPath, content, 0644)
	if err != nil {
		return &FileError{
			Operation: "write",
			Path:      cleanPath,
			Message:   fmt.Sprintf("failed to write file: %v", err),
		}
	}

	return nil
}

// CreateDir creates a directory and all necessary parent directories
func (fm *DefaultFileManager) CreateDir(path string) error {
	if path == "" {
		return &FileError{
			Operation: "mkdir",
			Path:      path,
			Message:   "directory path cannot be empty",
		}
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check if path already exists
	if info, err := os.Stat(cleanPath); err == nil {
		if !info.IsDir() {
			return &FileError{
				Operation: "mkdir",
				Path:      cleanPath,
				Message:   "path exists but is not a directory",
			}
		}
		// Directory already exists, nothing to do
		return nil
	}

	// Create directory with all parent directories
	err := os.MkdirAll(cleanPath, 0755)
	if err != nil {
		return &FileError{
			Operation: "mkdir",
			Path:      cleanPath,
			Message:   fmt.Sprintf("failed to create directory: %v", err),
		}
	}

	return nil
}

// ValidatePath validates a file or directory path
func (fm *DefaultFileManager) ValidatePath(path string) error {
	if path == "" {
		return &FileError{
			Operation: "validate",
			Path:      path,
			Message:   "path cannot be empty",
		}
	}

	// Check for invalid characters or patterns before cleaning
	if strings.Contains(path, "..") {
		return &FileError{
			Operation: "validate",
			Path:      path,
			Message:   "path cannot contain '..' for security reasons",
		}
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check if path is absolute when it shouldn't be (for relative paths)
	if filepath.IsAbs(cleanPath) && !strings.HasPrefix(cleanPath, "/") {
		// Allow absolute paths on Unix-like systems
		return nil
	}

	return nil
}

// IsAsyncAPIFile checks if the file has a valid AsyncAPI file extension
func (fm *DefaultFileManager) IsAsyncAPIFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"
}

// FileError represents file operation errors
type FileError struct {
	Operation string
	Path      string
	Message   string
}

func (e *FileError) Error() string {
	return fmt.Sprintf("file %s operation failed for path '%s': %s", e.Operation, e.Path, e.Message)
}

// IsFileError checks if an error is a FileError
func IsFileError(err error) bool {
	_, ok := err.(*FileError)
	return ok
}

// GetFileError extracts FileError from an error, returns nil if not a FileError
func GetFileError(err error) *FileError {
	if fileErr, ok := err.(*FileError); ok {
		return fileErr
	}
	return nil
}
