package generator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSchemaResolver(t *testing.T) {
	baseURI := "https://example.com/schemas/"
	resolver := NewSchemaResolver(baseURI)

	if resolver == nil {
		t.Fatal("NewSchemaResolver returned nil")
	}

	if resolver.baseURI != baseURI {
		t.Errorf("Expected baseURI %s, got %s", baseURI, resolver.baseURI)
	}

	if resolver.cache == nil {
		t.Error("Cache should be initialized")
	}

	if resolver.resolutionStack == nil {
		t.Error("Resolution stack should be initialized")
	}

	if resolver.httpClient == nil {
		t.Error("HTTP client should be initialized")
	}
}

func TestResolveRef_EmptyReference(t *testing.T) {
	resolver := NewSchemaResolver("")

	_, err := resolver.ResolveRef("")

	if err == nil {
		t.Error("Expected error for empty reference")
	}

	if resolverErr, ok := err.(*ResolverError); ok {
		if resolverErr.Reference != "" {
			t.Errorf("Expected empty reference in error, got %s", resolverErr.Reference)
		}
		if resolverErr.Message != "empty reference" {
			t.Errorf("Expected 'empty reference' message, got %s", resolverErr.Message)
		}
	} else {
		t.Errorf("Expected ResolverError, got %T", err)
	}
}

func TestResolveProperty_EmptyReference(t *testing.T) {
	resolver := NewSchemaResolver("")

	_, err := resolver.ResolveProperty("")

	if err == nil {
		t.Error("Expected error for empty reference")
	}

	if resolverErr, ok := err.(*ResolverError); ok {
		if resolverErr.Reference != "" {
			t.Errorf("Expected empty reference in error, got %s", resolverErr.Reference)
		}
		if resolverErr.Message != "empty reference" {
			t.Errorf("Expected 'empty reference' message, got %s", resolverErr.Message)
		}
	} else {
		t.Errorf("Expected ResolverError, got %T", err)
	}
}

func TestCircularReferenceDetection(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Simulate circular reference by adding to resolution stack
	resolver.resolutionStack = []string{"#/components/schemas/A", "#/components/schemas/B"}

	// Try to resolve a reference that's already in the stack
	_, err := resolver.ResolveRef("#/components/schemas/A")

	if err == nil {
		t.Error("Expected circular reference error")
	}

	if circularErr, ok := err.(*CircularReferenceError); ok {
		if circularErr.Reference != "#/components/schemas/A" {
			t.Errorf("Expected reference '#/components/schemas/A', got %s", circularErr.Reference)
		}
		if len(circularErr.Stack) != 2 {
			t.Errorf("Expected stack length 2, got %d", len(circularErr.Stack))
		}
	} else {
		t.Errorf("Expected CircularReferenceError, got %T", err)
	}
}

func TestParseReference(t *testing.T) {
	resolver := NewSchemaResolver("https://example.com/base/")

	tests := []struct {
		name             string
		ref              string
		expectedURI      string
		expectedFragment string
		expectError      bool
	}{
		{
			name:             "absolute URI with fragment",
			ref:              "https://example.com/schema.json#/definitions/User",
			expectedURI:      "https://example.com/schema.json",
			expectedFragment: "/definitions/User",
			expectError:      false,
		},
		{
			name:             "relative URI with fragment",
			ref:              "schema.json#/definitions/User",
			expectedURI:      "https://example.com/base/schema.json",
			expectedFragment: "/definitions/User",
			expectError:      false,
		},
		{
			name:             "fragment only",
			ref:              "#/definitions/User",
			expectedURI:      "",
			expectedFragment: "/definitions/User",
			expectError:      false,
		},
		{
			name:             "URI only",
			ref:              "https://example.com/schema.json",
			expectedURI:      "https://example.com/schema.json",
			expectedFragment: "",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, fragment, err := resolver.parseReference(tt.ref)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if uri != tt.expectedURI {
				t.Errorf("Expected URI %s, got %s", tt.expectedURI, uri)
			}

			if fragment != tt.expectedFragment {
				t.Errorf("Expected fragment %s, got %s", tt.expectedFragment, fragment)
			}
		})
	}
}

func TestParseReference_RelativeWithoutBase(t *testing.T) {
	resolver := NewSchemaResolver("")

	_, _, err := resolver.parseReference("schema.json#/definitions/User")

	if err == nil {
		t.Error("Expected error for relative reference without base URI")
	}
}

func TestIsAbsoluteURI(t *testing.T) {
	resolver := NewSchemaResolver("")

	tests := []struct {
		uri      string
		expected bool
	}{
		{"https://example.com/schema.json", true},
		{"http://example.com/schema.json", true},
		{"file:///path/to/schema.json", true},
		{"schema.json", false},
		{"./schema.json", false},
		{"../schema.json", false},
		{"/absolute/path/schema.json", false}, // This is absolute path but not absolute URI
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			result := resolver.isAbsoluteURI(tt.uri)
			if result != tt.expected {
				t.Errorf("Expected %v for URI %s, got %v", tt.expected, tt.uri, result)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	testData := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	err := os.WriteFile(tmpFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading the file
	data, err := resolver.loadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to load file: %v", err)
	}

	if string(data) != testData {
		t.Errorf("Expected %s, got %s", testData, string(data))
	}
}

func TestLoadFile_FileURL(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	testData := `{"type": "object"}`
	err := os.WriteFile(tmpFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading with file:// URL
	fileURL := "file://" + tmpFile
	data, err := resolver.loadFile(fileURL)
	if err != nil {
		t.Errorf("Failed to load file URL: %v", err)
	}

	if string(data) != testData {
		t.Errorf("Expected %s, got %s", testData, string(data))
	}
}

func TestLoadFile_NonExistent(t *testing.T) {
	resolver := NewSchemaResolver("")

	_, err := resolver.loadFile("/non/existent/file.json")

	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadHTTP(t *testing.T) {
	// Create a test HTTP server
	testData := `{"type": "object", "properties": {"id": {"type": "string"}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	resolver := NewSchemaResolver("")

	data, err := resolver.loadHTTP(server.URL)
	if err != nil {
		t.Errorf("Failed to load HTTP: %v", err)
	}

	if string(data) != testData {
		t.Errorf("Expected %s, got %s", testData, string(data))
	}
}

func TestLoadHTTP_NotFound(t *testing.T) {
	// Create a test HTTP server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resolver := NewSchemaResolver("")

	_, err := resolver.loadHTTP(server.URL)

	if err == nil {
		t.Error("Expected error for HTTP 404")
	}
}

func TestResolveJSONPointer(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Test document
	document := map[string]any{
		"definitions": map[string]any{
			"User": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
			},
		},
		"items": []any{
			"first",
			"second",
			map[string]any{
				"nested": "value",
			},
		},
	}

	tests := []struct {
		name        string
		pointer     string
		expected    any
		expectError bool
	}{
		{
			name:        "root document",
			pointer:     "/",
			expected:    document,
			expectError: false,
		},
		{
			name:        "simple property",
			pointer:     "/definitions",
			expected:    document["definitions"],
			expectError: false,
		},
		{
			name:        "nested property",
			pointer:     "/definitions/User",
			expected:    document["definitions"].(map[string]any)["User"],
			expectError: false,
		},
		{
			name:        "deeply nested property",
			pointer:     "/definitions/User/type",
			expected:    "object",
			expectError: false,
		},
		{
			name:        "array element",
			pointer:     "/items/0",
			expected:    "first",
			expectError: false,
		},
		{
			name:        "nested in array",
			pointer:     "/items/2/nested",
			expected:    "value",
			expectError: false,
		},
		{
			name:        "non-existent property",
			pointer:     "/nonexistent",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid array index",
			pointer:     "/items/invalid",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "array index out of bounds",
			pointer:     "/items/10",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.resolveJSONPointer(document, tt.pointer)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && fmt.Sprintf("%v", result) != fmt.Sprintf("%v", tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResolveJSONPointer_EscapedTokens(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Test document with special characters
	document := map[string]any{
		"a/b":   "slash",
		"c~d":   "tilde",
		"e~f/g": "both",
	}

	tests := []struct {
		name     string
		pointer  string
		expected string
	}{
		{
			name:     "escaped slash",
			pointer:  "/a~1b",
			expected: "slash",
		},
		{
			name:     "escaped tilde",
			pointer:  "/c~0d",
			expected: "tilde",
		},
		{
			name:     "both escaped",
			pointer:  "/e~0f~1g",
			expected: "both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.resolveJSONPointer(document, tt.pointer)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertToMessageSchema(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Test data that should convert to MessageSchema
	data := map[string]any{
		"type":        "object",
		"title":       "User",
		"description": "A user object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"age": map[string]any{
				"type": "integer",
			},
		},
		"required": []any{"name"},
	}

	schema, err := resolver.convertToMessageSchema(data)
	if err != nil {
		t.Errorf("Failed to convert to MessageSchema: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if schema.Title != "User" {
		t.Errorf("Expected title 'User', got %s", schema.Title)
	}

	if schema.Description != "A user object" {
		t.Errorf("Expected description 'A user object', got %s", schema.Description)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required field 'name', got %v", schema.Required)
	}
}

func TestConvertToProperty(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Test data that should convert to Property
	data := map[string]any{
		"type":        "string",
		"description": "A string property",
		"format":      "email",
		"maxLength":   100,
	}

	property, err := resolver.convertToProperty(data)
	if err != nil {
		t.Errorf("Failed to convert to Property: %v", err)
	}

	if property.Type != "string" {
		t.Errorf("Expected type 'string', got %s", property.Type)
	}

	if property.Description != "A string property" {
		t.Errorf("Expected description 'A string property', got %s", property.Description)
	}

	if property.Format != "email" {
		t.Errorf("Expected format 'email', got %s", property.Format)
	}

	if property.MaxLength == nil || *property.MaxLength != 100 {
		t.Errorf("Expected maxLength 100, got %v", property.MaxLength)
	}
}

func TestCaching(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Test that cache is used
	testData := &MessageSchema{
		Type:  "object",
		Title: "Cached Schema",
	}

	// Manually add to cache
	resolver.cache["test-ref"] = testData

	// Resolve should return cached version
	result, err := resolver.ResolveRef("test-ref")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != testData {
		t.Error("Expected cached schema to be returned")
	}
}

func TestClearCache(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Add something to cache
	resolver.cache["test"] = "value"

	if len(resolver.cache) != 1 {
		t.Error("Cache should have one item")
	}

	resolver.ClearCache()

	if len(resolver.cache) != 0 {
		t.Error("Cache should be empty after clearing")
	}
}

func TestSetBaseURI(t *testing.T) {
	resolver := NewSchemaResolver("initial")

	newURI := "https://example.com/new/"
	resolver.SetBaseURI(newURI)

	if resolver.baseURI != newURI {
		t.Errorf("Expected baseURI %s, got %s", newURI, resolver.baseURI)
	}
}

func TestGetCacheSize(t *testing.T) {
	resolver := NewSchemaResolver("")

	if resolver.GetCacheSize() != 0 {
		t.Error("Initial cache size should be 0")
	}

	resolver.cache["test1"] = "value1"
	resolver.cache["test2"] = "value2"

	if resolver.GetCacheSize() != 2 {
		t.Errorf("Expected cache size 2, got %d", resolver.GetCacheSize())
	}
}

func TestLoadDocument_JSONAndYAML(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create temporary files
	tmpDir := t.TempDir()

	// JSON file
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonData := `{"type": "object", "title": "JSON Schema"}`
	err := os.WriteFile(jsonFile, []byte(jsonData), 0644)
	if err != nil {
		t.Fatalf("Failed to create JSON file: %v", err)
	}

	// YAML file
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	yamlData := `type: object
title: YAML Schema`
	err = os.WriteFile(yamlFile, []byte(yamlData), 0644)
	if err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	// Test JSON loading
	jsonDoc, err := resolver.loadDocument(jsonFile)
	if err != nil {
		t.Errorf("Failed to load JSON document: %v", err)
	}

	if jsonMap, ok := jsonDoc.(map[string]any); ok {
		if jsonMap["title"] != "JSON Schema" {
			t.Errorf("Expected title 'JSON Schema', got %v", jsonMap["title"])
		}
	} else {
		t.Error("JSON document should be a map")
	}

	// Test YAML loading
	yamlDoc, err := resolver.loadDocument(yamlFile)
	if err != nil {
		t.Errorf("Failed to load YAML document: %v", err)
	}

	if yamlMap, ok := yamlDoc.(map[string]any); ok {
		if yamlMap["title"] != "YAML Schema" {
			t.Errorf("Expected title 'YAML Schema', got %v", yamlMap["title"])
		}
	} else {
		t.Error("YAML document should be a map")
	}
}

func TestLoadDocument_InvalidFormat(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create temporary file with invalid content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	invalidData := `{invalid json and yaml content`
	err := os.WriteFile(tmpFile, []byte(invalidData), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	_, err = resolver.loadDocument(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid document format")
	}
}

// Integration test for complete reference resolution
func TestIntegrationResolveRef_LocalFile(t *testing.T) {
	// Create a temporary schema file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.json")

	schemaData := map[string]any{
		"definitions": map[string]any{
			"User": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	jsonData, err := json.Marshal(schemaData)
	if err != nil {
		t.Fatalf("Failed to marshal schema data: %v", err)
	}

	err = os.WriteFile(schemaFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create resolver with base URI pointing to the directory
	resolver := NewSchemaResolver("file://" + tmpDir + "/")

	// Resolve reference to the User definition
	ref := "schema.json#/definitions/User"
	schema, err := resolver.ResolveRef(ref)
	if err != nil {
		t.Errorf("Failed to resolve reference: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required field 'name', got %v", schema.Required)
	}
}

// Additional edge case tests

func TestResolveFragment_SimpleFragment(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := map[string]interface{}{
		"User": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	result, err := resolver.resolveFragment(document, "User")
	if err != nil {
		t.Errorf("Failed to resolve simple fragment: %v", err)
	}

	if userSchema, ok := result.(map[string]interface{}); ok {
		if userSchema["type"] != "object" {
			t.Errorf("Expected type 'object', got %v", userSchema["type"])
		}
	} else {
		t.Error("Result should be a map")
	}
}

func TestResolveFragment_EmptyFragment(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := map[string]interface{}{
		"test": "value",
	}

	result, err := resolver.resolveFragment(document, "")
	if err != nil {
		t.Errorf("Failed to resolve empty fragment: %v", err)
	}

	if fmt.Sprintf("%v", result) != fmt.Sprintf("%v", document) {
		t.Error("Empty fragment should return the document itself")
	}
}

func TestResolveFragment_NonExistentSimpleFragment(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := map[string]interface{}{
		"User": "value",
	}

	_, err := resolver.resolveFragment(document, "NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent fragment")
	}
}

func TestResolveFragment_UnsupportedFormat(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := map[string]interface{}{
		"test": "value",
	}

	_, err := resolver.resolveFragment(document, "unsupported-format")
	if err == nil {
		t.Error("Expected error for unsupported fragment format")
	}
}

func TestResolveJSONPointer_ArrayEndMarker(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	_, err := resolver.resolveJSONPointer(document, "/items/-")
	if err == nil {
		t.Error("Expected error for array end marker '-'")
	}
}

func TestResolveJSONPointer_NonObjectArray(t *testing.T) {
	resolver := NewSchemaResolver("")

	document := "not an object or array"

	_, err := resolver.resolveJSONPointer(document, "/test")
	if err == nil {
		t.Error("Expected error when resolving pointer in non-object/array value")
	}
}

func TestConvertToMessageSchema_InvalidData(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create data that can't be marshaled to JSON
	invalidData := make(chan int)

	_, err := resolver.convertToMessageSchema(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

func TestConvertToProperty_InvalidData(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create data that can't be marshaled to JSON
	invalidData := make(chan int)

	_, err := resolver.convertToProperty(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

func TestResolveReference_FragmentOnlyWithoutBase(t *testing.T) {
	resolver := NewSchemaResolver("")

	_, err := resolver.resolveReference("#/definitions/User")
	if err == nil {
		t.Error("Expected error for fragment-only reference without base document")
	}

	if resolverErr, ok := err.(*ResolverError); ok {
		if !strings.Contains(resolverErr.Message, "fragment-only references not supported") {
			t.Errorf("Expected fragment-only error message, got: %s", resolverErr.Message)
		}
	} else {
		t.Errorf("Expected ResolverError, got %T", err)
	}
}

func TestLoadDocument_Caching(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "cached.json")

	testData := `{"cached": true}`
	err := os.WriteFile(tmpFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load document first time
	doc1, err := resolver.loadDocument(tmpFile)
	if err != nil {
		t.Errorf("Failed to load document first time: %v", err)
	}

	// Modify the file
	modifiedData := `{"cached": false}`
	err = os.WriteFile(tmpFile, []byte(modifiedData), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Load document second time - should get cached version
	doc2, err := resolver.loadDocument(tmpFile)
	if err != nil {
		t.Errorf("Failed to load document second time: %v", err)
	}

	// Should be the same (cached) document
	if fmt.Sprintf("%v", doc1) != fmt.Sprintf("%v", doc2) {
		t.Error("Second load should return cached document")
	}

	// Verify it's the original content, not the modified content
	if docMap, ok := doc1.(map[string]interface{}); ok {
		if docMap["cached"] != true {
			t.Error("Cached document should have original content")
		}
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	resolver := NewSchemaResolver("")

	// Verify HTTP client has timeout set
	if resolver.httpClient.Timeout == 0 {
		t.Error("HTTP client should have timeout configured")
	}
}

func TestParseReference_InvalidBaseURI(t *testing.T) {
	resolver := NewSchemaResolver("://invalid-uri")

	_, _, err := resolver.parseReference("relative.json")
	if err == nil {
		t.Error("Expected error for invalid base URI")
	}
}

func TestParseReference_InvalidReferenceURI(t *testing.T) {
	resolver := NewSchemaResolver("https://example.com/")

	// Create a reference with invalid characters
	invalidRef := "ht tp://invalid uri.json"
	_, _, err := resolver.parseReference(invalidRef)
	if err == nil {
		t.Error("Expected error for invalid reference URI")
	}
}

// Test integration with HTTP server that returns different content types
func TestLoadHTTP_ContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        string
		expectError bool
	}{
		{
			name:        "JSON content type",
			contentType: "application/json",
			data:        `{"type": "object"}`,
			expectError: false,
		},
		{
			name:        "YAML content type",
			contentType: "application/yaml",
			data:        "type: object",
			expectError: false,
		},
		{
			name:        "Plain text",
			contentType: "text/plain",
			data:        `{"type": "object"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.data))
			}))
			defer server.Close()

			resolver := NewSchemaResolver("")

			data, err := resolver.loadHTTP(server.URL)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && string(data) != tt.data {
				t.Errorf("Expected %s, got %s", tt.data, string(data))
			}
		})
	}
}
