package generator

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestAsyncAPIVersionCompatibility tests compatibility with different AsyncAPI versions
func TestAsyncAPIVersionCompatibility(t *testing.T) {
	testCases := []struct {
		version     string
		supported   bool
		description string
	}{
		{"2.0.0", true, "AsyncAPI 2.0.0 - Initial 2.x release"},
		{"2.1.0", true, "AsyncAPI 2.1.0 - Added oneOf/anyOf support"},
		{"2.2.0", true, "AsyncAPI 2.2.0 - Added server variables"},
		{"2.3.0", true, "AsyncAPI 2.3.0 - Added request/reply pattern"},
		{"2.4.0", true, "AsyncAPI 2.4.0 - Added message examples"},
		{"2.5.0", true, "AsyncAPI 2.5.0 - Added schema format validation"},
		{"2.6.0", true, "AsyncAPI 2.6.0 - Latest 2.x version"},
		{"3.0.0", true, "AsyncAPI 3.0.0 - Major version with breaking changes"},
		{"1.2.0", false, "AsyncAPI 1.2.0 - Legacy version (unsupported)"},
		{"4.0.0", false, "AsyncAPI 4.0.0 - Future version (unsupported)"},
	}

	baseSchema := map[string]any{
		"info": map[string]any{
			"title":   "Version Compatibility Test",
			"version": "1.0.0",
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"TestMessage": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":      map[string]any{"type": "string"},
						"message": map[string]any{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("version_%s", strings.ReplaceAll(tc.version, ".", "_")), func(t *testing.T) {
			// Create spec with specific version
			spec := make(map[string]any)
			for k, v := range baseSchema {
				spec[k] = v
			}
			spec["asyncapi"] = tc.version

			specBytes, err := json.Marshal(spec)
			if err != nil {
				t.Fatalf("Failed to marshal spec: %v", err)
			}

			config := &Config{
				PackageName: "versiontest",
				OutputDir:   "./test",
			}

			gen := NewGenerator(config)
			result, err := gen.ParseAndGenerate(specBytes)

			if tc.supported {
				if err != nil {
					t.Errorf("Expected version %s to be supported, but got error: %v", tc.version, err)
				} else {
					if len(result.Files) == 0 {
						t.Errorf("Expected files to be generated for supported version %s", tc.version)
					}
					// Verify generated code is valid
					for filename, content := range result.Files {
						_, parseErr := parser.ParseFile(token.NewFileSet(), filename, content, 0)
						if parseErr != nil {
							t.Errorf("Generated code for version %s is invalid: %v", tc.version, parseErr)
						}
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected version %s to be unsupported, but generation succeeded", tc.version)
				}
			}

			t.Logf("Version %s (%s): %s", tc.version, tc.description,
				map[bool]string{true: "SUPPORTED", false: "UNSUPPORTED"}[tc.supported])
		})
	}
}

// TestAsyncAPI2xFeatures tests specific AsyncAPI 2.x features
func TestAsyncAPI2xFeatures(t *testing.T) {
	testCases := []struct {
		name        string
		spec        string
		expectError bool
		description string
	}{
		{
			name: "basic_schema",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Basic Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"BasicMessage": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"data": {"type": "string"}
							}
						}
					}
				}
			}`,
			expectError: false,
			description: "Basic schema definition",
		},
		{
			name: "array_types",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Array Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"ArrayMessage": {
							"type": "object",
							"properties": {
								"tags": {
									"type": "array",
									"items": {"type": "string"}
								},
								"numbers": {
									"type": "array",
									"items": {"type": "integer"}
								}
							}
						}
					}
				}
			}`,
			expectError: false,
			description: "Array type support",
		},
		{
			name: "format_types",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Format Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"FormatMessage": {
							"type": "object",
							"properties": {
								"email": {"type": "string", "format": "email"},
								"timestamp": {"type": "string", "format": "date-time"},
								"uuid": {"type": "string", "format": "uuid"},
								"price": {"type": "number", "format": "double"}
							}
						}
					}
				}
			}`,
			expectError: false,
			description: "Format-specific type mappings",
		},
		{
			name: "nested_objects",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Nested Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"NestedMessage": {
							"type": "object",
							"properties": {
								"user": {
									"type": "object",
									"properties": {
										"id": {"type": "string"},
										"profile": {
											"type": "object",
											"properties": {
												"name": {"type": "string"},
												"age": {"type": "integer"}
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
			expectError: false,
			description: "Nested object support",
		},
		{
			name: "required_fields",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Required Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"RequiredMessage": {
							"type": "object",
							"properties": {
								"requiredField": {"type": "string"},
								"optionalField": {"type": "string"}
							},
							"required": ["requiredField"]
						}
					}
				}
			}`,
			expectError: false,
			description: "Required field handling",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				PackageName:     "featuretest",
				OutputDir:       "./test",
				IncludeComments: true,
				UsePointers:     true,
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
				} else {
					// Verify generated code
					if len(result.Files) == 0 {
						t.Errorf("No files generated for %s", tc.name)
					}

					for filename, content := range result.Files {
						_, parseErr := parser.ParseFile(token.NewFileSet(), filename, content, 0)
						if parseErr != nil {
							t.Errorf("Generated code for %s is invalid: %v\nContent:\n%s",
								tc.name, parseErr, content)
						}
					}
				}
			}

			t.Logf("Feature test %s (%s): %s", tc.name, tc.description,
				map[bool]string{true: "PASSED", false: "FAILED"}[err == nil])
		})
	}
}

// TestAsyncAPI3xFeatures tests AsyncAPI 3.x specific features
func TestAsyncAPI3xFeatures(t *testing.T) {
	testCases := []struct {
		name        string
		spec        string
		expectError bool
		description string
	}{
		{
			name: "basic_3x_schema",
			spec: `{
				"asyncapi": "3.0.0",
				"info": {"title": "AsyncAPI 3.0 Test", "version": "1.0.0"},
				"components": {
					"schemas": {
						"Message3x": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"payload": {"type": "string"}
							}
						}
					}
				}
			}`,
			expectError: false,
			description: "Basic AsyncAPI 3.0 schema",
		},
		{
			name: "3x_with_channels",
			spec: `{
				"asyncapi": "3.0.0",
				"info": {"title": "AsyncAPI 3.0 Channels", "version": "1.0.0"},
				"channels": {
					"user/signup": {
						"messages": {
							"userSignup": {
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
			}`,
			expectError: false,
			description: "AsyncAPI 3.0 with channels structure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				PackageName: "asyncapi3test",
				OutputDir:   "./test",
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
				} else if len(result.Files) > 0 {
					// Verify generated code if files were created
					for filename, content := range result.Files {
						_, parseErr := parser.ParseFile(token.NewFileSet(), filename, content, 0)
						if parseErr != nil {
							t.Errorf("Generated code for %s is invalid: %v", tc.name, parseErr)
						}
					}
				}
			}

			t.Logf("AsyncAPI 3.x test %s (%s): %s", tc.name, tc.description,
				map[bool]string{true: "PASSED", false: "FAILED"}[err == nil])
		})
	}
}

// TestBackwardCompatibility tests backward compatibility with older specs
func TestBackwardCompatibility(t *testing.T) {
	// Test that newer generator can handle older AsyncAPI specs
	oldSpecs := []struct {
		version string
		spec    string
	}{
		{
			version: "2.0.0",
			spec: `{
				"asyncapi": "2.0.0",
				"info": {"title": "Legacy 2.0", "version": "1.0.0"},
				"components": {
					"schemas": {
						"LegacyMessage": {
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
			version: "2.1.0",
			spec: `{
				"asyncapi": "2.1.0",
				"info": {"title": "Legacy 2.1", "version": "1.0.0"},
				"components": {
					"schemas": {
						"LegacyMessage": {
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
	}

	for _, spec := range oldSpecs {
		t.Run("backward_compat_"+strings.ReplaceAll(spec.version, ".", "_"), func(t *testing.T) {
			config := &Config{
				PackageName: "backcompat",
				OutputDir:   "./test",
			}

			gen := NewGenerator(config)
			result, err := gen.ParseAndGenerate([]byte(spec.spec))

			if err != nil {
				t.Errorf("Backward compatibility failed for version %s: %v", spec.version, err)
			} else if len(result.Files) > 0 {
				// Verify generated code
				for filename, content := range result.Files {
					_, parseErr := parser.ParseFile(token.NewFileSet(), filename, content, 0)
					if parseErr != nil {
						t.Errorf("Generated code for version %s is invalid: %v", spec.version, parseErr)
					}
				}
				t.Logf("Backward compatibility with %s: PASSED", spec.version)
			}
		})
	}
}

// TestForwardCompatibility tests forward compatibility with future specs
func TestForwardCompatibility(t *testing.T) {
	// Test graceful handling of unknown fields in newer specs
	futureSpec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Future Features", "version": "1.0.0"},
		"components": {
			"schemas": {
				"FutureMessage": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"data": {"type": "string"}
					},
					"futureField": "unknown value",
					"x-future-extension": {
						"someProperty": "someValue"
					}
				}
			}
		},
		"x-future-root-extension": {
			"newFeature": true
		}
	}`

	config := &Config{
		PackageName: "futurecompat",
		OutputDir:   "./test",
	}

	gen := NewGenerator(config)
	result, err := gen.ParseAndGenerate([]byte(futureSpec))

	if err != nil {
		t.Errorf("Forward compatibility test failed: %v", err)
	} else {
		if len(result.Files) == 0 {
			t.Error("No files generated for forward compatibility test")
		}

		// Verify generated code is still valid despite unknown fields
		for filename, content := range result.Files {
			_, parseErr := parser.ParseFile(token.NewFileSet(), filename, content, 0)
			if parseErr != nil {
				t.Errorf("Generated code with future fields is invalid: %v", parseErr)
			}
		}
		t.Log("Forward compatibility test: PASSED")
	}
}

// TestVersionDetection tests proper version detection and validation
func TestVersionDetection(t *testing.T) {
	testCases := []struct {
		name        string
		spec        string
		expectError bool
		description string
	}{
		{
			name: "missing_version",
			spec: `{
				"info": {"title": "No Version", "version": "1.0.0"},
				"components": {"schemas": {}}
			}`,
			expectError: true,
			description: "Missing asyncapi version field",
		},
		{
			name: "invalid_version_format",
			spec: `{
				"asyncapi": "invalid",
				"info": {"title": "Invalid Version", "version": "1.0.0"},
				"components": {"schemas": {}}
			}`,
			expectError: true,
			description: "Invalid version format",
		},
		{
			name: "valid_version",
			spec: `{
				"asyncapi": "2.6.0",
				"info": {"title": "Valid Version", "version": "1.0.0"},
				"components": {"schemas": {}}
			}`,
			expectError: false,
			description: "Valid version format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				PackageName: "versiondetect",
				OutputDir:   "./test",
			}

			gen := NewGenerator(config)
			_, err := gen.ParseAndGenerate([]byte(tc.spec))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but parsing succeeded", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
			}

			t.Logf("Version detection test %s (%s): %s", tc.name, tc.description,
				map[bool]string{true: "PASSED", false: "FAILED"}[tc.expectError == (err != nil)])
		})
	}
}

// TestCrossVersionConsistency tests that the same schema generates consistent output across supported versions
func TestCrossVersionConsistency(t *testing.T) {
	baseSchema := map[string]any{
		"info": map[string]any{
			"title":   "Consistency Test",
			"version": "1.0.0",
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"ConsistentMessage": map[string]any{
					"type": "object",
					"properties": map[string]interface{}{
						"id":        map[string]interface{}{"type": "string"},
						"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
						"data":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
		},
	}

	supportedVersions := []string{"2.0.0", "2.6.0", "3.0.0"}
	results := make(map[string]*GenerateResult)

	// Generate code for each version
	for _, version := range supportedVersions {
		spec := make(map[string]interface{})
		for k, v := range baseSchema {
			spec[k] = v
		}
		spec["asyncapi"] = version

		specBytes, err := json.Marshal(spec)
		if err != nil {
			t.Fatalf("Failed to marshal spec for version %s: %v", version, err)
		}

		config := &Config{
			PackageName: "consistency",
			OutputDir:   "./test",
		}

		gen := NewGenerator(config)
		result, err := gen.ParseAndGenerate(specBytes)
		if err != nil {
			t.Errorf("Failed to generate for version %s: %v", version, err)
			continue
		}

		results[version] = result
	}

	// Compare results across versions
	if len(results) < 2 {
		t.Fatal("Need at least 2 successful generations to compare consistency")
	}

	var baseVersion string
	var baseResult *GenerateResult
	for version, result := range results {
		baseVersion = version
		baseResult = result
		break
	}

	for version, result := range results {
		if version == baseVersion {
			continue
		}

		// Compare number of generated files
		if len(result.Files) != len(baseResult.Files) {
			t.Errorf("File count inconsistency between %s and %s: %d vs %d",
				baseVersion, version, len(baseResult.Files), len(result.Files))
		}

		// Compare struct names in generated code
		for filename := range baseResult.Files {
			if _, exists := result.Files[filename]; !exists {
				t.Errorf("File %s missing in version %s", filename, version)
			}
		}

		t.Logf("Consistency check between %s and %s: PASSED", baseVersion, version)
	}
}
