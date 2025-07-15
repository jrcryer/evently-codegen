package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileManager(t *testing.T) {
	fm := NewFileManager()
	if fm == nil {
		t.Fatal("NewFileManager() returned nil")
	}
}

func TestDefaultFileManager_ReadFile(t *testing.T) {
	fm := NewFileManager()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filemanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("successful read", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "test.json")
		testContent := `{"test": "content"}`
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Read the file
		content, err := fm.ReadFile(testFile)
		if err != nil {
			t.Fatalf("ReadFile() failed: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("ReadFile() content = %q, want %q", string(content), testContent)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := fm.ReadFile("")
		if err == nil {
			t.Error("ReadFile() with empty path should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		} else {
			if fileErr.Operation != "read" {
				t.Errorf("FileError.Operation = %q, want %q", fileErr.Operation, "read")
			}
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
		_, err := fm.ReadFile(nonExistentFile)
		if err == nil {
			t.Error("ReadFile() with non-existent file should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		// Create a file with no read permissions
		restrictedFile := filepath.Join(tempDir, "restricted.json")
		err := os.WriteFile(restrictedFile, []byte("test"), 0000)
		if err != nil {
			t.Fatalf("Failed to create restricted file: %v", err)
		}
		defer os.Chmod(restrictedFile, 0644) // Restore permissions for cleanup

		_, err = fm.ReadFile(restrictedFile)
		if err == nil {
			t.Error("ReadFile() with restricted file should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})
}

func TestDefaultFileManager_WriteFile(t *testing.T) {
	fm := NewFileManager()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filemanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("successful write", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "output.go")
		testContent := []byte("package test\n\ntype Test struct {}")

		err := fm.WriteFile(testFile, testContent)
		if err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}

		// Verify file was written correctly
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("Written content = %q, want %q", string(content), string(testContent))
		}
	})

	t.Run("empty path", func(t *testing.T) {
		err := fm.WriteFile("", []byte("test"))
		if err == nil {
			t.Error("WriteFile() with empty path should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		err := fm.WriteFile(testFile, []byte("test"))
		if err == nil {
			t.Error("WriteFile() with non-.go extension should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})

	t.Run("create nested directory", func(t *testing.T) {
		nestedFile := filepath.Join(tempDir, "nested", "deep", "output.go")
		testContent := []byte("package nested")

		err := fm.WriteFile(nestedFile, testContent)
		if err != nil {
			t.Fatalf("WriteFile() with nested path failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(nestedFile); os.IsNotExist(err) {
			t.Error("Nested file was not created")
		}
	})
}

func TestDefaultFileManager_CreateDir(t *testing.T) {
	fm := NewFileManager()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filemanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("successful create", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "newdir")
		err := fm.CreateDir(newDir)
		if err != nil {
			t.Fatalf("CreateDir() failed: %v", err)
		}

		// Verify directory was created
		info, err := os.Stat(newDir)
		if err != nil {
			t.Fatalf("Created directory does not exist: %v", err)
		}

		if !info.IsDir() {
			t.Error("Created path is not a directory")
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		nestedDir := filepath.Join(tempDir, "nested", "deep", "directory")
		err := fm.CreateDir(nestedDir)
		if err != nil {
			t.Fatalf("CreateDir() with nested path failed: %v", err)
		}

		// Verify directory was created
		info, err := os.Stat(nestedDir)
		if err != nil {
			t.Fatalf("Nested directory does not exist: %v", err)
		}

		if !info.IsDir() {
			t.Error("Created nested path is not a directory")
		}
	})

	t.Run("directory already exists", func(t *testing.T) {
		existingDir := filepath.Join(tempDir, "existing")
		err := os.Mkdir(existingDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create existing directory: %v", err)
		}

		// Should not return error if directory already exists
		err = fm.CreateDir(existingDir)
		if err != nil {
			t.Errorf("CreateDir() on existing directory failed: %v", err)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		err := fm.CreateDir("")
		if err == nil {
			t.Error("CreateDir() with empty path should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})

	t.Run("path exists but is file", func(t *testing.T) {
		existingFile := filepath.Join(tempDir, "existingfile")
		err := os.WriteFile(existingFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		err = fm.CreateDir(existingFile)
		if err == nil {
			t.Error("CreateDir() on existing file should return error")
		}

		fileErr := GetFileError(err)
		if fileErr == nil {
			t.Error("Expected FileError")
		}
	})
}

func TestDefaultFileManager_ValidatePath(t *testing.T) {
	fm := NewFileManager()

	t.Run("valid paths", func(t *testing.T) {
		validPaths := []string{
			"test.json",
			"path/to/file.yaml",
			"./relative/path.yml",
			"/absolute/path.go",
		}

		for _, path := range validPaths {
			err := fm.ValidatePath(path)
			if err != nil {
				t.Errorf("ValidatePath(%q) failed: %v", path, err)
			}
		}
	})

	t.Run("invalid paths", func(t *testing.T) {
		invalidPaths := []string{
			"",
			"../parent/file.json",
			"path/../other/file.yaml",
		}

		for _, path := range invalidPaths {
			err := fm.ValidatePath(path)
			if err == nil {
				t.Errorf("ValidatePath(%q) should have failed", path)
			}

			fileErr := GetFileError(err)
			if fileErr == nil {
				t.Errorf("Expected FileError for path %q", path)
			}
		}
	})
}

func TestDefaultFileManager_IsAsyncAPIFile(t *testing.T) {
	fm := NewFileManager()

	t.Run("valid AsyncAPI files", func(t *testing.T) {
		validFiles := []string{
			"asyncapi.json",
			"spec.yaml",
			"api.yml",
			"path/to/asyncapi.JSON",
			"path/to/spec.YAML",
			"path/to/api.YML",
		}

		for _, file := range validFiles {
			if !fm.IsAsyncAPIFile(file) {
				t.Errorf("IsAsyncAPIFile(%q) should return true", file)
			}
		}
	})

	t.Run("invalid AsyncAPI files", func(t *testing.T) {
		invalidFiles := []string{
			"file.txt",
			"file.go",
			"file.xml",
			"file",
			"file.json.bak",
			"file.yaml.old",
		}

		for _, file := range invalidFiles {
			if fm.IsAsyncAPIFile(file) {
				t.Errorf("IsAsyncAPIFile(%q) should return false", file)
			}
		}
	})
}

func TestFileError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		err := &FileError{
			Operation: "read",
			Path:      "/path/to/file.json",
			Message:   "file not found",
		}

		expected := "file read operation failed for path '/path/to/file.json': file not found"
		if err.Error() != expected {
			t.Errorf("FileError.Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("IsFileError", func(t *testing.T) {
		fileErr := &FileError{Operation: "test", Path: "test", Message: "test"}
		if !IsFileError(fileErr) {
			t.Error("IsFileError() should return true for FileError")
		}

		otherErr := os.ErrNotExist
		if IsFileError(otherErr) {
			t.Error("IsFileError() should return false for non-FileError")
		}
	})

	t.Run("GetFileError", func(t *testing.T) {
		fileErr := &FileError{Operation: "test", Path: "test", Message: "test"}
		extracted := GetFileError(fileErr)
		if extracted != fileErr {
			t.Error("GetFileError() should return the same FileError")
		}

		otherErr := os.ErrNotExist
		extracted = GetFileError(otherErr)
		if extracted != nil {
			t.Error("GetFileError() should return nil for non-FileError")
		}
	})
}
