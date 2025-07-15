package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2E_CompleteWorkflowWithCompilation tests the complete workflow from AsyncAPI to compiled Go code
func TestE2E_CompleteWorkflowWithCompilation(t *testing.T) {
	testCases := []struct {
		name     string
		spec     string
		expected []string // Expected struct names
	}{
		{
			name: "simple_user_schema",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "User API", "version": "1.0.0"},
				"components": {
					"schemas": {
						"User": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"email": {"type": "string", "format": "email"},
								"name": {"type": "string"},
								"age": {"type": "integer"},
								"isActive": {"type": "boolean"}
							},
							"required": ["id", "email"]
						}
					}
				}
			}`,
			expected: []string{"User"},
		},
		{
			name: "complex_nested_schema",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Order API", "version": "1.0.0"},
				"components": {
					"schemas": {
						"Order": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"amount": {"type": "number"},
								"status": {"type": "string"},
								"items": {
									"type": "array",
									"items": {"type": "string"}
								}
							},
							"required": ["id", "amount"]
						}
					}
				}
			}`,
			expected: []string{"Order"},
		},
		{
			name: "multiple_schemas",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Multi Schema API", "version": "1.0.0"},
				"components": {
					"schemas": {
						"User": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"name": {"type": "string"}
							}
						},
						"Product": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"title": {"type": "string"},
								"price": {"type": "number"}
							}
						},
						"Order": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"userId": {"type": "string"},
								"productId": {"type": "string"}
							}
						}
					}
				}
			}`,
			expected: []string{"User", "Product", "Order"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for this test
			tempDir := t.TempDir()
			outputDir := filepath.Join(tempDir, "generated")

			// Configure generator
			config := &Config{
				PackageName:     "testpkg",
				OutputDir:       outputDir,
				IncludeComments: true,
				UsePointers:     true,
			}

			gen := NewGenerator(config)

			// Parse and generate
			result, err := gen.ParseAndGenerate([]byte(tc.spec))
			if err != nil {
				t.Fatalf("ParseAndGenerate failed: %v", err)
			}

			// Verify no generation errors
			if len(result.Errors) > 0 {
				t.Errorf("Generation completed with errors: %v", result.Errors)
			}

			// Verify files were generated
			if len(result.Files) == 0 {
				t.Fatal("No files were generated")
			}

			// Write files to disk for compilation test
			for filename, content := range result.Files {
				filePath := filepath.Join(outputDir, filename)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", filename, err)
				}
			}

			// Test that generated code compiles
			t.Run("compilation_test", func(t *testing.T) {
				// Create a simple go.mod file
				goModContent := fmt.Sprintf("module %s\n\ngo 1.19\n", config.PackageName)
				goModPath := filepath.Join(outputDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
					t.Fatalf("Failed to write go.mod: %v", err)
				}

				// Try to compile the generated code
				cmd := exec.Command("go", "build", "./...")
				cmd.Dir = outputDir
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("Generated code failed to compile: %v\nOutput: %s", err, string(output))
				}
			})

			// Verify expected structs are present
			t.Run("struct_verification", func(t *testing.T) {
				allFoundStructs := make(map[string]bool)

				for _, content := range result.Files {
					fset := token.NewFileSet()
					file, err := parser.ParseFile(fset, "test.go", content, parser.ParseComments)
					if err != nil {
						t.Fatalf("Failed to parse generated Go code: %v", err)
					}

					ast.Inspect(file, func(n ast.Node) bool {
						if typeSpec, ok := n.(*ast.TypeSpec); ok {
							if _, ok := typeSpec.Type.(*ast.StructType); ok {
								allFoundStructs[typeSpec.Name.Name] = true
							}
						}
						return true
					})
				}

				for _, expectedStruct := range tc.expected {
					if !allFoundStructs[expectedStruct] {
						t.Errorf("Expected struct %s not found in generated code. Found structs: %v", expectedStruct, allFoundStructs)
					}
				}
			})
		})
	}
}

// TestE2E_AsyncAPIVersionCompatibility tests compatibility with different AsyncAPI versions
func TestE2E_AsyncAPIVersionCompatibility(t *testing.T) {
	versions := []string{"2.0.0", "2.1.0", "2.2.0", "2.3.0", "2.4.0", "2.5.0", "2.6.0", "3.0.0"}

	baseSpec := map[string]interface{}{
		"info": map[string]interface{}{
			"title":   "Version Test API",
			"version": "1.0.0",
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"TestMessage": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":      map[string]interface{}{"type": "string"},
						"message": map[string]interface{}{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
		},
	}

	for _, version := range versions {
		t.Run("version_"+strings.ReplaceAll(version, ".", "_"), func(t *testing.T) {
			// Create spec with specific version
			spec := make(map[string]interface{})
			for k, v := range baseSpec {
				spec[k] = v
			}
			spec["asyncapi"] = version

			specBytes, err := json.Marshal(spec)
			if err != nil {
				t.Fatalf("Failed to marshal spec: %v", err)
			}

			config := &Config{
				PackageName:     "versiontest",
				OutputDir:       "./test-output",
				IncludeComments: false,
				UsePointers:     false,
			}

			gen := NewGenerator(config)
			result, err := gen.ParseAndGenerate(specBytes)

			// Check if version is supported
			supportedVersions := gen.GetSupportedVersions()
			isSupported := false
			for _, supported := range supportedVersions {
				if supported == version {
					isSupported = true
					break
				}
			}

			if isSupported {
				if err != nil {
					t.Errorf("Expected version %s to be supported, but got error: %v", version, err)
				}
				if len(result.Files) == 0 {
					t.Errorf("Expected files to be generated for supported version %s", version)
				}
			} else {
				if err == nil {
					t.Errorf("Expected version %s to be unsupported, but generation succeeded", version)
				}
			}
		})
	}
}

// TestE2E_ErrorScenarios tests various error scenarios end-to-end
func TestE2E_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		spec        string
		expectError bool
		errorType   string
	}{
		{
			name:        "invalid_json",
			spec:        `{"asyncapi": "2.6.0", "info": {"title": "Test"} // invalid json`,
			expectError: true,
			errorType:   "parse",
		},
		{
			name: "missing_required_info",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {}
			}`,
			expectError: true,
			errorType:   "validation",
		},
		{
			name: "unsupported_version",
			spec: `{
				"asyncapi": "1.0.0",
				"info": {"title": "Test", "version": "1.0.0"}
			}`,
			expectError: true,
			errorType:   "version",
		},
		{
			name: "invalid_schema_type",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"InvalidType": {
							"type": "invalid_type",
							"properties": {
								"field": {"type": "string"}
							}
						}
					}
				}
			}`,
			expectError: false, // Current implementation handles unknown types gracefully
			errorType:   "",
		},
		{
			name: "circular_reference",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"A": {
							"type": "object",
							"properties": {
								"b": {"$ref": "#/components/schemas/B"}
							}
						},
						"B": {
							"type": "object",
							"properties": {
								"a": {"$ref": "#/components/schemas/A"}
							}
						}
					}
				}
			}`,
			expectError: false, // Should handle circular references gracefully
			errorType:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				PackageName: "errortest",
				OutputDir:   "./test-output",
			}

			gen := NewGenerator(config)
			result, err := gen.ParseAndGenerate([]byte(tc.spec))

			if tc.expectError {
				if err == nil && len(result.Errors) == 0 {
					t.Errorf("Expected error for %s, but generation succeeded", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
			}
		})
	}
}

// TestE2E_LargeSchemaHandling tests handling of large and complex schemas
func TestE2E_LargeSchemaHandling(t *testing.T) {
	// Generate a large schema with many properties
	properties := make(map[string]interface{})
	required := make([]string, 0)

	// Create 100 properties with various types
	for i := 0; i < 100; i++ {
		propName := fmt.Sprintf("field%d", i)
		var propType string
		switch i % 5 {
		case 0:
			propType = "string"
		case 1:
			propType = "integer"
		case 2:
			propType = "number"
		case 3:
			propType = "boolean"
		case 4:
			propType = "array"
		}

		if propType == "array" {
			properties[propName] = map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			}
		} else {
			properties[propName] = map[string]interface{}{
				"type": propType,
			}
		}

		// Make every 10th field required
		if i%10 == 0 {
			required = append(required, propName)
		}
	}

	spec := map[string]interface{}{
		"asyncapi": "2.6.0",
		"info": map[string]interface{}{
			"title":   "Large Schema API",
			"version": "1.0.0",
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"LargeSchema": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		},
	}

	specBytes, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal large spec: %v", err)
	}

	config := &Config{
		PackageName:     "largetest",
		OutputDir:       "./test-output",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := NewGenerator(config)

	// Measure generation time
	start := time.Now()
	result, err := gen.ParseAndGenerate(specBytes)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to generate large schema: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("No files generated for large schema")
	}

	// Verify generation completed in reasonable time (less than 5 seconds)
	if duration > 5*time.Second {
		t.Errorf("Large schema generation took too long: %v", duration)
	}

	// Verify generated code is valid
	for filename, content := range result.Files {
		_, err := parser.ParseFile(token.NewFileSet(), filename, content, 0)
		if err != nil {
			t.Errorf("Generated code for large schema is invalid: %v", err)
		}

		// Verify all fields are present
		fieldCount := strings.Count(content, "Field")
		if fieldCount < 100 {
			t.Errorf("Expected at least 100 fields in generated struct, found %d", fieldCount)
		}
	}

	t.Logf("Large schema generation completed in %v", duration)
}

// TestE2E_FileOperationsWorkflow tests complete file-based workflow
func TestE2E_FileOperationsWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	// Test different file formats
	testCases := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "json_file",
			filename: "test.json",
			content: `{
				"asyncapi": "2.6.0",
				"info": {"title": "JSON Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"JsonMessage": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"data": {"type": "string"}
							}
						}
					}
				}
			}`,
		},
		{
			name:     "yaml_file",
			filename: "test.yaml",
			content: `asyncapi: '2.6.0'
info:
  title: YAML Test
  version: '1.0.0'
components:
  schemas:
    YamlMessage:
      type: object
      properties:
        id:
          type: string
        data:
          type: string`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write spec file
			specPath := filepath.Join(tempDir, tc.filename)
			if err := os.WriteFile(specPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write spec file: %v", err)
			}

			// Configure output directory
			outputDir := filepath.Join(tempDir, "output_"+tc.name)
			config := &Config{
				PackageName:     "filetest",
				OutputDir:       outputDir,
				IncludeComments: false,
				UsePointers:     false,
			}

			gen := NewGenerator(config)

			// Test file-to-file workflow
			err := gen.ParseFileAndGenerateToFiles(specPath)
			if err != nil {
				t.Fatalf("ParseFileAndGenerateToFiles failed: %v", err)
			}

			// Verify output directory exists
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				t.Fatalf("Output directory was not created: %s", outputDir)
			}

			// Verify files were created
			files, err := os.ReadDir(outputDir)
			if err != nil {
				t.Fatalf("Failed to read output directory: %v", err)
			}

			if len(files) == 0 {
				t.Fatal("No files were generated")
			}

			// Verify file contents
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				filePath := filepath.Join(outputDir, file.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				// Verify it's valid Go code
				_, err = parser.ParseFile(token.NewFileSet(), file.Name(), content, 0)
				if err != nil {
					t.Errorf("Generated file contains invalid Go code: %v", err)
				}

				// Verify package name
				if !strings.Contains(string(content), "package filetest") {
					t.Errorf("Expected package filetest in generated file")
				}
			}
		})
	}
}

// TestE2E_ConfigurationVariations tests different configuration combinations
func TestE2E_ConfigurationVariations(t *testing.T) {
	baseSpec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Config Test", "version": "1.0.0"},
		"components": {
			"schemas": {
				"ConfigTest": {
					"type": "object",
					"description": "Test configuration options",
					"properties": {
						"requiredField": {"type": "string", "description": "Required field"},
						"optionalField": {"type": "string", "description": "Optional field"},
						"numberField": {"type": "number"},
						"arrayField": {"type": "array", "items": {"type": "string"}}
					},
					"required": ["requiredField"]
				}
			}
		}
	}`

	testCases := []struct {
		name   string
		config *Config
		checks func(t *testing.T, content string)
	}{
		{
			name: "all_options_enabled",
			config: &Config{
				PackageName:     "alloptions",
				OutputDir:       "./test",
				IncludeComments: true,
				UsePointers:     true,
			},
			checks: func(t *testing.T, content string) {
				if !strings.Contains(content, "// Test configuration options") {
					t.Error("Expected struct comment")
				}
				// Check for pointer fields (optional fields should be pointers when UsePointers is true)
				if !strings.Contains(content, "*string") && !strings.Contains(content, "*float64") {
					t.Logf("Generated content: %s", content)
					t.Error("Expected pointer types for optional fields when UsePointers is true")
				}
				if !strings.Contains(content, "// Required field") && !strings.Contains(content, "// Optional field") {
					t.Error("Expected field comments when IncludeComments is true")
				}
			},
		},
		{
			name: "minimal_options",
			config: &Config{
				PackageName:     "minimal",
				OutputDir:       "./test",
				IncludeComments: false,
				UsePointers:     false,
			},
			checks: func(t *testing.T, content string) {
				if strings.Contains(content, "// Test configuration options") {
					t.Error("Unexpected struct comment")
				}
				if strings.Contains(content, "OptionalField *string") {
					t.Error("Unexpected pointer for optional field")
				}
				if strings.Contains(content, "OptionalField string") {
					// This is expected for non-pointer mode
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewGenerator(tc.config)
			result, err := gen.ParseAndGenerate([]byte(baseSpec))
			if err != nil {
				t.Fatalf("Generation failed: %v", err)
			}

			if len(result.Files) == 0 {
				t.Fatal("No files generated")
			}

			for _, content := range result.Files {
				tc.checks(t, content)

				// Verify it's valid Go code
				_, err := parser.ParseFile(token.NewFileSet(), "test.go", content, 0)
				if err != nil {
					t.Errorf("Generated code is invalid: %v", err)
				}
			}
		})
	}
}

// TestE2E_EdgeCases tests various edge cases
func TestE2E_EdgeCases(t *testing.T) {
	testCases := []struct {
		name string
		spec string
	}{
		{
			name: "empty_schema",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Empty", "version": "1.0.0"},
				"components": {
					"schemas": {
						"Empty": {
							"type": "object",
							"properties": {}
						}
					}
				}
			}`,
		},
		{
			name: "schema_with_special_characters",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Special Chars", "version": "1.0.0"},
				"components": {
					"schemas": {
						"SpecialChars": {
							"type": "object",
							"properties": {
								"field-with-dashes": {"type": "string"},
								"field_with_underscores": {"type": "string"},
								"field.with.dots": {"type": "string"},
								"field with spaces": {"type": "string"}
							}
						}
					}
				}
			}`,
		},
		{
			name: "deeply_nested_objects",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Deep Nesting", "version": "1.0.0"},
				"components": {
					"schemas": {
						"DeepNesting": {
							"type": "object",
							"properties": {
								"level1": {
									"type": "object",
									"properties": {
										"level2": {
											"type": "object",
											"properties": {
												"level3": {
													"type": "object",
													"properties": {
														"level4": {
															"type": "object",
															"properties": {
																"deepField": {"type": "string"}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				PackageName:     "edgecase",
				OutputDir:       "./test",
				IncludeComments: true,
				UsePointers:     true,
			}

			gen := NewGenerator(config)
			result, err := gen.ParseAndGenerate([]byte(tc.spec))
			if err != nil {
				t.Fatalf("Failed to handle edge case %s: %v", tc.name, err)
			}

			if len(result.Files) == 0 {
				t.Fatal("No files generated for edge case")
			}

			// Verify generated code is valid
			for filename, content := range result.Files {
				_, err := parser.ParseFile(token.NewFileSet(), filename, content, 0)
				if err != nil {
					t.Errorf("Generated code for edge case %s is invalid: %v\nContent:\n%s", tc.name, err, content)
				}
			}
		})
	}
}
