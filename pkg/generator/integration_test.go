package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegrationBasicWorkflow tests the complete workflow from AsyncAPI to compiled Go code
func TestIntegrationBasicWorkflow(t *testing.T) {
	// Sample AsyncAPI specification
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "User Events API",
			"version": "1.0.0",
			"description": "API for user-related events"
		},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"description": "User information",
					"properties": {
						"id": {
							"type": "string",
							"description": "Unique user identifier"
						},
						"email": {
							"type": "string",
							"format": "email",
							"description": "User email address"
						},
						"name": {
							"type": "string",
							"description": "User full name"
						},
						"age": {
							"type": "integer",
							"description": "User age"
						},
						"isActive": {
							"type": "boolean",
							"description": "Whether user is active"
						},
						"tags": {
							"type": "array",
							"description": "User tags",
							"items": {
								"type": "string"
							}
						},
						"createdAt": {
							"type": "string",
							"format": "date-time",
							"description": "Account creation timestamp"
						}
					},
					"required": ["id", "email", "name"]
				}
			}
		}
	}`

	// Create generator with test configuration
	config := &Config{
		PackageName:     "testevents",
		OutputDir:       "./test-output",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := NewGenerator(config)

	// Test complete workflow
	result, err := gen.ParseAndGenerate([]byte(asyncAPISpec))
	if err != nil {
		t.Fatalf("ParseAndGenerate failed: %v", err)
	}

	// Verify results
	if len(result.Files) == 0 {
		t.Fatal("Expected generated files, got none")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Generation completed with errors: %v", result.Errors)
	}

	// Verify generated code structure
	for filename, content := range result.Files {
		t.Logf("Generated file: %s", filename)

		// Verify package declaration
		if !strings.Contains(content, "package testevents") {
			t.Errorf("Expected package declaration in %s", filename)
		}

		// Verify struct definition
		if !strings.Contains(content, "type User struct") {
			t.Errorf("Expected User struct definition in %s", filename)
		}

		// Verify required fields are not pointers
		if !strings.Contains(content, "Id string") {
			t.Errorf("Expected non-pointer required field Id in %s", filename)
		}

		// Verify optional fields are pointers (when UsePointers is true)
		if !strings.Contains(content, "Age *int64") {
			t.Errorf("Expected pointer optional field Age in %s", filename)
		}

		// Verify JSON tags
		if !strings.Contains(content, `json:"id"`) {
			t.Errorf("Expected JSON tag for id field in %s", filename)
		}

		// Verify comments are included
		if !strings.Contains(content, "User information") {
			t.Errorf("Expected struct comment in %s", filename)
		}

		// Verify the generated code can be parsed as valid Go
		_, err := parser.ParseFile(token.NewFileSet(), filename, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated code in %s is not valid Go: %v", filename, err)
		}
	}
}

// TestIntegrationFileOperations tests file I/O operations
func TestIntegrationFileOperations(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Sample AsyncAPI specification
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Order Events API",
			"version": "1.0.0"
		},
		"components": {
			"schemas": {
				"Order": {
					"type": "object",
					"properties": {
						"orderId": {"type": "string"},
						"amount": {"type": "number"},
						"items": {
							"type": "array",
							"items": {"type": "string"}
						}
					},
					"required": ["orderId", "amount"]
				}
			}
		}
	}`

	// Write AsyncAPI spec to file
	specFile := filepath.Join(tempDir, "orders.json")
	err := os.WriteFile(specFile, []byte(asyncAPISpec), 0644)
	if err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}

	// Create generator with file output configuration
	outputDir := filepath.Join(tempDir, "generated")
	config := &Config{
		PackageName:     "orders",
		OutputDir:       outputDir,
		IncludeComments: false,
		UsePointers:     false,
	}

	gen := NewGenerator(config)

	// Test file-to-file workflow
	err = gen.ParseFileAndGenerateToFiles(specFile)
	if err != nil {
		t.Fatalf("ParseFileAndGenerateToFiles failed: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Fatalf("Output directory was not created: %s", outputDir)
	}

	// Verify generated files exist
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No files were generated")
	}

	// Verify file content
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(outputDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read generated file %s: %v", file.Name(), err)
		}

		contentStr := string(content)

		// Verify package name
		if !strings.Contains(contentStr, "package orders") {
			t.Errorf("Expected package orders in %s", file.Name())
		}

		// Verify struct definition
		if !strings.Contains(contentStr, "type Order struct") {
			t.Errorf("Expected Order struct in %s", file.Name())
		}

		// Verify the file is valid Go code
		_, err = parser.ParseFile(token.NewFileSet(), file.Name(), content, 0)
		if err != nil {
			t.Errorf("Generated file %s contains invalid Go code: %v", file.Name(), err)
		}
	}
}

// TestIntegrationComplexSchema tests generation with complex nested schemas
func TestIntegrationComplexSchema(t *testing.T) {
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Complex Schema API",
			"version": "1.0.0"
		},
		"components": {
			"schemas": {
				"Product": {
					"type": "object",
					"description": "Product information",
					"properties": {
						"id": {"type": "string"},
						"name": {"type": "string"},
						"price": {"type": "number"},
						"category": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"name": {"type": "string"},
								"parentId": {"type": "string"}
							}
						},
						"variants": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"sku": {"type": "string"},
									"color": {"type": "string"},
									"size": {"type": "string"},
									"stock": {"type": "integer"}
								}
							}
						},
						"metadata": {
							"type": "object",
							"additionalProperties": true
						}
					},
					"required": ["id", "name", "price"]
				}
			}
		}
	}`

	config := &Config{
		PackageName:     "products",
		OutputDir:       "./test-output",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := NewGenerator(config)

	// Parse and generate
	result, err := gen.ParseAndGenerate([]byte(asyncAPISpec))
	if err != nil {
		t.Fatalf("Failed to parse and generate complex schema: %v", err)
	}

	// Verify generation succeeded
	if len(result.Files) == 0 {
		t.Fatal("No files generated for complex schema")
	}

	// Verify generated code
	for filename, content := range result.Files {
		// Verify the code is syntactically correct
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
		if err != nil {
			t.Fatalf("Generated code is not valid Go: %v\nContent:\n%s", err, content)
		}

		// Verify struct definitions exist
		hasProductStruct := false
		ast.Inspect(file, func(n ast.Node) bool {
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				if typeSpec.Name.Name == "Product" {
					hasProductStruct = true
				}
			}
			return true
		})

		if !hasProductStruct {
			t.Errorf("Expected Product struct definition in %s", filename)
		}

		// Verify array and object field types
		if !strings.Contains(content, "[]") {
			t.Errorf("Expected array type in generated code for %s", filename)
		}
	}
}

// TestIntegrationErrorHandling tests error handling scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	gen := NewGenerator(&Config{
		PackageName: "test",
		OutputDir:   "./test-output",
	})

	// Test invalid JSON
	_, err := gen.Parse([]byte(`{"invalid": json}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test unsupported AsyncAPI version
	invalidVersionSpec := `{
		"asyncapi": "1.0.0",
		"info": {"title": "Test", "version": "1.0.0"}
	}`
	_, err = gen.Parse([]byte(invalidVersionSpec))
	if err == nil {
		t.Error("Expected error for unsupported AsyncAPI version")
	}

	// Test missing required fields
	missingFieldsSpec := `{
		"asyncapi": "2.6.0",
		"info": {"title": ""}
	}`
	_, err = gen.Parse([]byte(missingFieldsSpec))
	if err == nil {
		t.Error("Expected error for missing required fields")
	}

	// Test invalid configuration
	invalidGen := NewGenerator(&Config{
		PackageName: "", // Invalid empty package name
		OutputDir:   "./test",
	})
	err = invalidGen.ValidateConfig()
	if err == nil {
		t.Error("Expected validation error for empty package name")
	}
}

// TestIntegrationVersionSupport tests support for different AsyncAPI versions
func TestIntegrationVersionSupport(t *testing.T) {
	gen := NewGenerator(&Config{
		PackageName: "versiontest",
		OutputDir:   "./test-output",
	})

	versions := []string{"2.0.0", "2.6.0", "3.0.0"}

	for _, version := range versions {
		t.Run("version_"+version, func(t *testing.T) {
			spec := `{
				"asyncapi": "` + version + `",
				"info": {
					"title": "Version Test API",
					"version": "1.0.0"
				},
				"components": {
					"schemas": {
						"TestMessage": {
							"type": "object",
							"properties": {
								"id": {"type": "string"}
							}
						}
					}
				}
			}`

			result, err := gen.ParseAndGenerate([]byte(spec))
			if err != nil {
				t.Fatalf("Failed to parse AsyncAPI version %s: %v", version, err)
			}

			if len(result.Files) == 0 {
				t.Errorf("No files generated for AsyncAPI version %s", version)
			}

			// Verify generated code is valid
			for filename, content := range result.Files {
				_, err := parser.ParseFile(token.NewFileSet(), filename, content, 0)
				if err != nil {
					t.Errorf("Invalid Go code generated for version %s: %v", version, err)
				}
			}
		})
	}

	// Test supported versions list
	supportedVersions := gen.GetSupportedVersions()
	if len(supportedVersions) == 0 {
		t.Error("Expected non-empty list of supported versions")
	}

	for _, version := range versions {
		found := false
		for _, supported := range supportedVersions {
			if supported == version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Version %s should be in supported versions list", version)
		}
	}
}

// TestIntegrationConfigurationOptions tests different configuration options
func TestIntegrationConfigurationOptions(t *testing.T) {
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Config Test", "version": "1.0.0"},
		"components": {
			"schemas": {
				"TestStruct": {
					"type": "object",
					"description": "Test structure",
					"properties": {
						"required_field": {"type": "string"},
						"optional_field": {"type": "string"}
					},
					"required": ["required_field"]
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
			name: "with_comments",
			config: &Config{
				PackageName:     "withcomments",
				OutputDir:       "./test",
				IncludeComments: true,
				UsePointers:     false,
			},
			checks: func(t *testing.T, content string) {
				if !strings.Contains(content, "// Test structure") {
					t.Error("Expected struct comment when IncludeComments is true")
				}
			},
		},
		{
			name: "without_comments",
			config: &Config{
				PackageName:     "withoutcomments",
				OutputDir:       "./test",
				IncludeComments: false,
				UsePointers:     false,
			},
			checks: func(t *testing.T, content string) {
				if strings.Contains(content, "// Test structure") {
					t.Error("Unexpected struct comment when IncludeComments is false")
				}
			},
		},
		{
			name: "with_pointers",
			config: &Config{
				PackageName:     "withpointers",
				OutputDir:       "./test",
				IncludeComments: false,
				UsePointers:     true,
			},
			checks: func(t *testing.T, content string) {
				if !strings.Contains(content, "OptionalField *string") {
					t.Error("Expected pointer type for optional field when UsePointers is true")
				}
			},
		},
		{
			name: "without_pointers",
			config: &Config{
				PackageName:     "withoutpointers",
				OutputDir:       "./test",
				IncludeComments: false,
				UsePointers:     false,
			},
			checks: func(t *testing.T, content string) {
				if strings.Contains(content, "OptionalField *string") {
					t.Error("Unexpected pointer type for optional field when UsePointers is false")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewGenerator(tc.config)
			result, err := gen.ParseAndGenerate([]byte(asyncAPISpec))
			if err != nil {
				t.Fatalf("Failed to generate with config %s: %v", tc.name, err)
			}

			if len(result.Files) == 0 {
				t.Fatal("No files generated")
			}

			for _, content := range result.Files {
				tc.checks(t, content)
			}
		})
	}
}
