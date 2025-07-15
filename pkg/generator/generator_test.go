package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	// Test with nil config (should use defaults)
	gen := NewGenerator(nil)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	config := gen.GetConfig()
	if config.PackageName != "main" {
		t.Errorf("Expected default package name 'main', got '%s'", config.PackageName)
	}

	if config.OutputDir != "./generated" {
		t.Errorf("Expected default output dir './generated', got '%s'", config.OutputDir)
	}

	if !config.IncludeComments {
		t.Error("Expected IncludeComments to be true by default")
	}

	if !config.UsePointers {
		t.Error("Expected UsePointers to be true by default")
	}

	// Verify components are initialized
	if gen.parser == nil {
		t.Error("Expected parser to be initialized")
	}

	if gen.codegen == nil {
		t.Error("Expected code generator to be initialized")
	}

	if gen.resolver == nil {
		t.Error("Expected resolver to be initialized")
	}
}

func TestNewGeneratorWithConfig(t *testing.T) {
	customConfig := &Config{
		PackageName:     "mypackage",
		OutputDir:       "./custom",
		IncludeComments: false,
		UsePointers:     false,
	}

	gen := NewGenerator(customConfig)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	config := gen.GetConfig()
	if config.PackageName != "mypackage" {
		t.Errorf("Expected package name 'mypackage', got '%s'", config.PackageName)
	}

	if config.OutputDir != "./custom" {
		t.Errorf("Expected output dir './custom', got '%s'", config.OutputDir)
	}

	if config.IncludeComments {
		t.Error("Expected IncludeComments to be false")
	}

	if config.UsePointers {
		t.Error("Expected UsePointers to be false")
	}
}

func TestSetConfig(t *testing.T) {
	gen := NewGenerator(nil)

	newConfig := &Config{
		PackageName: "updated",
		OutputDir:   "./updated",
	}

	gen.SetConfig(newConfig)

	config := gen.GetConfig()
	if config.PackageName != "updated" {
		t.Errorf("Expected updated package name 'updated', got '%s'", config.PackageName)
	}

	// Test with nil config (should not update)
	gen.SetConfig(nil)
	config = gen.GetConfig()
	if config.PackageName != "updated" {
		t.Errorf("Expected package name to remain 'updated', got '%s'", config.PackageName)
	}
}

func TestParse(t *testing.T) {
	gen := NewGenerator(nil)

	// Test with valid AsyncAPI JSON
	validAsyncAPI := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"channels": {
			"user/signup": {
				"publish": {
					"message": {
						"payload": {
							"type": "object",
							"properties": {
								"userId": {"type": "string"},
								"email": {"type": "string"}
							}
						}
					}
				}
			}
		}
	}`

	result, err := gen.Parse([]byte(validAsyncAPI))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected parse result, got nil")
	}

	if result.Spec == nil {
		t.Error("Expected spec to be parsed")
	}

	if len(result.Messages) == 0 {
		t.Error("Expected messages to be extracted")
	}

	// Test with invalid JSON
	invalidJSON := `{"invalid": json}`
	_, err = gen.Parse([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with empty data
	_, err = gen.Parse([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestGenerate(t *testing.T) {
	gen := NewGenerator(&Config{
		PackageName: "testpkg",
		OutputDir:   "./test",
	})

	// Create test message schema
	messages := map[string]*MessageSchema{
		"User": {
			Type: "object",
			Properties: map[string]*Property{
				"id": {
					Type: "string",
				},
				"name": {
					Type: "string",
				},
			},
			Required: []string{"id"},
		},
	}

	result, err := gen.Generate(messages)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected generate result, got nil")
	}

	if len(result.Files) == 0 {
		t.Error("Expected generated files")
	}

	// Check if generated file contains expected content
	for filename, content := range result.Files {
		if !strings.Contains(content, "package testpkg") {
			t.Errorf("Expected package declaration in %s", filename)
		}
		if !strings.Contains(content, "type User struct") {
			t.Errorf("Expected User struct in %s", filename)
		}
	}
}

func TestParseAndGenerate(t *testing.T) {
	gen := NewGenerator(&Config{
		PackageName: "testpkg",
	})

	validAsyncAPI := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"name": {"type": "string"}
					}
				}
			}
		}
	}`

	result, err := gen.ParseAndGenerate([]byte(validAsyncAPI))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Files) == 0 {
		t.Error("Expected generated files")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorField  string
	}{

		{
			name: "empty package name",
			config: &Config{
				PackageName: "",
				OutputDir:   "./test",
			},
			expectError: true,
			errorField:  "config.PackageName",
		},
		{
			name: "empty output dir",
			config: &Config{
				PackageName: "test",
				OutputDir:   "",
			},
			expectError: true,
			errorField:  "config.OutputDir",
		},
		{
			name: "invalid package name",
			config: &Config{
				PackageName: "123invalid",
				OutputDir:   "./test",
			},
			expectError: true,
			errorField:  "config.PackageName",
		},
		{
			name: "valid config",
			config: &Config{
				PackageName: "validpkg",
				OutputDir:   "./test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.config)
			err := gen.ValidateConfig()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				} else if validationErr, ok := err.(*ValidationError); ok {
					if validationErr.Field != tt.errorField {
						t.Errorf("Expected error field '%s', got '%s'", tt.errorField, validationErr.Field)
					}
				} else {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestGetSupportedVersions(t *testing.T) {
	gen := NewGenerator(nil)
	versions := gen.GetSupportedVersions()

	if len(versions) == 0 {
		t.Error("Expected supported versions, got empty slice")
	}

	// Check for expected versions
	expectedVersions := []string{"2.6.0", "3.0.0"}
	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version '%s' to be supported", expected)
		}
	}
}

func TestParseFile(t *testing.T) {
	gen := NewGenerator(nil)

	// Create a temporary AsyncAPI file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	validAsyncAPI := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		}
	}`

	err := os.WriteFile(testFile, []byte(validAsyncAPI), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test parsing valid file
	result, err := gen.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected parse result, got nil")
	}

	// Test parsing non-existent file
	_, err = gen.ParseFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGenerateToFiles(t *testing.T) {
	tempDir := t.TempDir()
	gen := NewGenerator(&Config{
		PackageName: "testpkg",
		OutputDir:   tempDir,
	})

	messages := map[string]*MessageSchema{
		"User": {
			Type: "object",
			Properties: map[string]*Property{
				"id": {Type: "string"},
			},
		},
	}

	err := gen.GenerateToFiles(messages)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check if file was created
	expectedFile := filepath.Join(tempDir, "user.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", expectedFile)
	}

	// Read and verify file content
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package testpkg") {
		t.Error("Expected package declaration in generated file")
	}
}

func TestParseFileAndGenerateToFiles(t *testing.T) {
	tempDir := t.TempDir()
	gen := NewGenerator(&Config{
		PackageName: "testpkg",
		OutputDir:   tempDir,
	})

	// Create test AsyncAPI file
	testFile := filepath.Join(tempDir, "test.json")
	validAsyncAPI := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"}
					}
				}
			}
		}
	}`

	err := os.WriteFile(testFile, []byte(validAsyncAPI), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test complete workflow
	err = gen.ParseFileAndGenerateToFiles(testFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check if output file was created
	expectedFile := filepath.Join(tempDir, "user.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", expectedFile)
	}
}

func TestGeneratorIsValidGoIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"validName", true},
		{"ValidName", true},
		{"_validName", true},
		{"123invalid", false},
		{"valid123", true},
		{"package", false}, // Go keyword
		{"func", false},    // Go keyword
		{"myPackage", true},
		{"my-package", false}, // Contains hyphen
		{"my.package", false}, // Contains dot
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := IsValidGoIdentifier(test.input)
			if result != test.expected {
				t.Errorf("IsValidGoIdentifier(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}
