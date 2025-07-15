package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
	fmt.Println("=== AsyncAPI Go Code Generator - Advanced Features ===")
	fmt.Println()

	// Example 1: Custom Configuration Options
	fmt.Println("=== Example 1: Custom Configuration Options ===")
	demonstrateCustomConfiguration()
	fmt.Println()

	// Example 2: Error Handling Strategies
	fmt.Println("=== Example 2: Error Handling Strategies ===")
	demonstrateErrorHandling()
	fmt.Println()

	// Example 3: Working with Complex Schemas
	fmt.Println("=== Example 3: Working with Complex Schemas ===")
	demonstrateComplexSchemas()
	fmt.Println()

	// Example 4: Performance Considerations
	fmt.Println("=== Example 4: Performance Considerations ===")
	demonstratePerformanceOptimizations()
	fmt.Println()

	// Example 5: Schema Validation and Analysis
	fmt.Println("=== Example 5: Schema Validation and Analysis ===")
	demonstrateSchemaAnalysis()
	fmt.Println()

	fmt.Println("Advanced features demonstration completed!")
}

func demonstrateCustomConfiguration() {
	// Configuration 1: Minimal setup with value types
	config1 := &generator.Config{
		PackageName:     "events",
		OutputDir:       "./generated/minimal",
		IncludeComments: false,
		UsePointers:     false, // Use value types instead of pointers
	}

	// Configuration 2: Full setup with detailed comments
	config2 := &generator.Config{
		PackageName:     "detailedevents",
		OutputDir:       "./generated/detailed",
		IncludeComments: true, // Include all schema descriptions
		UsePointers:     true, // Use pointers for optional fields
	}

	fmt.Printf("Configuration 1 (Minimal):\n")
	fmt.Printf("  Package: %s\n", config1.PackageName)
	fmt.Printf("  Comments: %t\n", config1.IncludeComments)
	fmt.Printf("  Pointers: %t\n", config1.UsePointers)

	fmt.Printf("\nConfiguration 2 (Detailed):\n")
	fmt.Printf("  Package: %s\n", config2.PackageName)
	fmt.Printf("  Comments: %t\n", config2.IncludeComments)
	fmt.Printf("  Pointers: %t\n", config2.UsePointers)

	// Demonstrate configuration validation
	gen1 := generator.NewGenerator(config1)
	if err := gen1.ValidateConfig(); err != nil {
		fmt.Printf("Configuration 1 validation error: %v\n", err)
	} else {
		fmt.Printf("✓ Configuration 1 is valid\n")
	}

	gen2 := generator.NewGenerator(config2)
	if err := gen2.ValidateConfig(); err != nil {
		fmt.Printf("Configuration 2 validation error: %v\n", err)
	} else {
		fmt.Printf("✓ Configuration 2 is valid\n")
	}

	// Show supported AsyncAPI versions
	versions := gen1.GetSupportedVersions()
	fmt.Printf("Supported AsyncAPI versions: %v\n", versions)
}

func demonstrateErrorHandling() {
	config := &generator.Config{
		PackageName: "errortest",
		OutputDir:   "./generated/errors",
	}
	gen := generator.NewGenerator(config)

	// Test 1: Invalid AsyncAPI specification
	fmt.Println("Test 1: Invalid JSON syntax")
	invalidJSON := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Test API"
			// Missing comma - invalid JSON
		}
	}`

	parseResult, err := gen.Parse([]byte(invalidJSON))
	if err != nil {
		fmt.Printf("✓ Correctly caught parse error: %v\n", err)
	} else if len(parseResult.Errors) > 0 {
		fmt.Printf("✓ Parse completed with %d errors:\n", len(parseResult.Errors))
		for i, parseErr := range parseResult.Errors {
			fmt.Printf("  %d. %v\n", i+1, parseErr)
		}
	}

	// Test 2: Missing required fields
	fmt.Println("\nTest 2: Missing required fields")
	incompleteSpec := `{
		"asyncapi": "2.6.0"
	}`

	parseResult, err = gen.Parse([]byte(incompleteSpec))
	if err != nil {
		fmt.Printf("✓ Correctly caught validation error: %v\n", err)
	} else if len(parseResult.Errors) > 0 {
		fmt.Printf("✓ Validation completed with %d errors:\n", len(parseResult.Errors))
		for i, parseErr := range parseResult.Errors {
			fmt.Printf("  %d. %v\n", i+1, parseErr)
		}
	}

	// Test 3: Unsupported AsyncAPI version
	fmt.Println("\nTest 3: Unsupported AsyncAPI version")
	unsupportedVersion := `{
		"asyncapi": "1.0.0",
		"info": {
			"title": "Old API",
			"version": "1.0.0"
		}
	}`

	parseResult, err = gen.Parse([]byte(unsupportedVersion))
	if err != nil {
		fmt.Printf("✓ Correctly caught version error: %v\n", err)
	} else if len(parseResult.Errors) > 0 {
		fmt.Printf("✓ Version check completed with %d warnings:\n", len(parseResult.Errors))
		for i, parseErr := range parseResult.Errors {
			fmt.Printf("  %d. %v\n", i+1, parseErr)
		}
	}

	// Test 4: File not found error
	fmt.Println("\nTest 4: File not found error")
	_, err = gen.ParseFile("nonexistent-file.yaml")
	if err != nil {
		fmt.Printf("✓ Correctly caught file error: %v\n", err)
	}
}

func demonstrateComplexSchemas() {
	// Complex AsyncAPI specification with nested objects, arrays, and references
	complexSpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Complex Schema API",
			"version": "1.0.0"
		},
		"channels": {
			"complex/event": {
				"publish": {
					"message": {
						"payload": {
							"$ref": "#/components/schemas/ComplexEvent"
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"ComplexEvent": {
					"type": "object",
					"description": "A complex event with nested structures",
					"properties": {
						"id": {
							"type": "string",
							"description": "Event identifier"
						},
						"metadata": {
							"type": "object",
							"description": "Event metadata",
							"properties": {
								"source": {"type": "string"},
								"timestamp": {"type": "string", "format": "date-time"},
								"version": {"type": "integer"},
								"tags": {
									"type": "array",
									"items": {"type": "string"}
								}
							}
						},
						"payload": {
							"type": "object",
							"description": "Event payload",
							"properties": {
								"user": {"$ref": "#/components/schemas/User"},
								"actions": {
									"type": "array",
									"items": {"$ref": "#/components/schemas/Action"}
								},
								"context": {
									"type": "object",
									"additionalProperties": true
								}
							}
						}
					},
					"required": ["id", "metadata", "payload"]
				},
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"email": {"type": "string", "format": "email"},
						"profile": {
							"type": "object",
							"properties": {
								"name": {"type": "string"},
								"preferences": {
									"type": "object",
									"properties": {
										"notifications": {"type": "boolean"},
										"theme": {"type": "string", "enum": ["light", "dark"]}
									}
								}
							}
						}
					},
					"required": ["id", "email"]
				},
				"Action": {
					"type": "object",
					"properties": {
						"type": {"type": "string"},
						"timestamp": {"type": "string", "format": "date-time"},
						"data": {
							"type": "object",
							"additionalProperties": true
						}
					},
					"required": ["type", "timestamp"]
				}
			}
		}
	}`

	config := &generator.Config{
		PackageName:     "complex",
		OutputDir:       "./generated/complex",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := generator.NewGenerator(config)

	fmt.Println("Parsing complex AsyncAPI specification...")
	parseResult, err := gen.Parse([]byte(complexSpec))
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	if len(parseResult.Errors) > 0 {
		fmt.Printf("Parse completed with %d warnings\n", len(parseResult.Errors))
	}

	fmt.Printf("Successfully parsed complex specification:\n")
	fmt.Printf("  Title: %s\n", parseResult.Spec.Info.Title)
	fmt.Printf("  Schemas found: %d\n", len(parseResult.Messages))

	// Analyze schema complexity
	for name, schema := range parseResult.Messages {
		fmt.Printf("\nSchema: %s\n", name)
		fmt.Printf("  Type: %s\n", schema.Type)
		fmt.Printf("  Properties: %d\n", len(schema.Properties))
		fmt.Printf("  Required fields: %d\n", len(schema.Required))

		// Count nested levels
		maxDepth := calculateSchemaDepth(schema, 0)
		fmt.Printf("  Max nesting depth: %d\n", maxDepth)
	}

	// Generate code and analyze output
	fmt.Println("\nGenerating Go code...")
	generateResult, err := gen.Generate(parseResult.Messages)
	if err != nil {
		log.Printf("Generation error: %v", err)
		return
	}

	fmt.Printf("Generated %d Go files:\n", len(generateResult.Files))
	for filename, content := range generateResult.Files {
		lines := len(strings.Split(content, "\n"))
		structs := strings.Count(content, "type ")
		fmt.Printf("  %s: %d lines, %d structs\n", filename, lines, structs)
	}
}

func calculateSchemaDepth(schema *generator.MessageSchema, currentDepth int) int {
	if schema == nil || schema.Properties == nil {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, prop := range schema.Properties {
		if prop.Type == "object" && prop.Properties != nil {
			// Create a temporary MessageSchema to recurse
			tempSchema := &generator.MessageSchema{
				Type:       prop.Type,
				Properties: prop.Properties,
			}
			depth := calculateSchemaDepth(tempSchema, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

func demonstratePerformanceOptimizations() {
	// Create a large AsyncAPI specification for performance testing
	fmt.Println("Creating large AsyncAPI specification for performance testing...")

	// Generate a spec with many schemas
	largeSpec := generateLargeAsyncAPISpec(100) // 100 schemas

	config := &generator.Config{
		PackageName: "performance",
		OutputDir:   "./generated/performance",
	}

	gen := generator.NewGenerator(config)

	// Measure parsing performance
	fmt.Println("Measuring parsing performance...")
	parseResult, err := gen.Parse([]byte(largeSpec))
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	fmt.Printf("Parsed specification with %d schemas\n", len(parseResult.Messages))

	// Measure generation performance
	fmt.Println("Measuring generation performance...")
	generateResult, err := gen.Generate(parseResult.Messages)
	if err != nil {
		log.Printf("Generation error: %v", err)
		return
	}

	fmt.Printf("Generated %d Go files\n", len(generateResult.Files))

	// Calculate total lines of generated code
	totalLines := 0
	for _, content := range generateResult.Files {
		totalLines += len(strings.Split(content, "\n"))
	}
	fmt.Printf("Total lines of generated code: %d\n", totalLines)

	// Performance tips
	fmt.Println("\nPerformance Tips:")
	fmt.Println("1. Use streaming for very large specifications")
	fmt.Println("2. Enable caching for repeated schema resolution")
	fmt.Println("3. Process schemas in parallel when possible")
	fmt.Println("4. Use minimal configuration for faster generation")
}

func generateLargeAsyncAPISpec(numSchemas int) string {
	var sb strings.Builder

	sb.WriteString(`{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Large API",
			"version": "1.0.0"
		},
		"channels": {`)

	// Generate channels
	for i := 0; i < numSchemas; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`
			"schema%d/event": {
				"publish": {
					"message": {
						"payload": {
							"$ref": "#/components/schemas/Schema%d"
						}
					}
				}
			}`, i, i))
	}

	sb.WriteString(`
		},
		"components": {
			"schemas": {`)

	// Generate schemas
	for i := 0; i < numSchemas; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`
				"Schema%d": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"name": {"type": "string"},
						"value": {"type": "number"},
						"active": {"type": "boolean"},
						"tags": {
							"type": "array",
							"items": {"type": "string"}
						}
					},
					"required": ["id", "name"]
				}`, i))
	}

	sb.WriteString(`
			}
		}
	}`)

	return sb.String()
}

func demonstrateSchemaAnalysis() {
	// Load a real AsyncAPI specification for analysis
	specPath := "../../testdata/user-service.yaml"

	config := &generator.Config{
		PackageName: "analysis",
		OutputDir:   "./generated/analysis",
	}

	gen := generator.NewGenerator(config)

	fmt.Printf("Analyzing AsyncAPI specification: %s\n", specPath)

	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		fmt.Printf("Specification file not found: %s\n", specPath)
		return
	}

	parseResult, err := gen.ParseFile(specPath)
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	// Analyze the specification
	fmt.Printf("\n=== Specification Analysis ===\n")
	fmt.Printf("Title: %s\n", parseResult.Spec.Info.Title)
	fmt.Printf("Version: %s\n", parseResult.Spec.Info.Version)
	fmt.Printf("Description: %s\n", parseResult.Spec.Info.Description)

	fmt.Printf("\n=== Schema Analysis ===\n")
	fmt.Printf("Total schemas: %d\n", len(parseResult.Messages))

	// Analyze each schema
	for name, schema := range parseResult.Messages {
		fmt.Printf("\nSchema: %s\n", name)
		fmt.Printf("  Description: %s\n", schema.Description)
		fmt.Printf("  Type: %s\n", schema.Type)
		fmt.Printf("  Properties: %d\n", len(schema.Properties))
		fmt.Printf("  Required fields: %d (%v)\n", len(schema.Required), schema.Required)

		// Analyze property types
		typeCount := make(map[string]int)
		for _, prop := range schema.Properties {
			typeCount[prop.Type]++
		}

		fmt.Printf("  Property types:\n")
		for propType, count := range typeCount {
			fmt.Printf("    %s: %d\n", propType, count)
		}

		// Check for complex nested structures
		hasNested := false
		for _, prop := range schema.Properties {
			if prop.Type == "object" || prop.Type == "array" {
				hasNested = true
				break
			}
		}
		fmt.Printf("  Has nested structures: %t\n", hasNested)
	}

	// Generate and analyze output
	fmt.Printf("\n=== Generation Analysis ===\n")
	generateResult, err := gen.Generate(parseResult.Messages)
	if err != nil {
		log.Printf("Generation error: %v", err)
		return
	}

	fmt.Printf("Generated files: %d\n", len(generateResult.Files))

	totalLines := 0
	totalStructs := 0
	for filename, content := range generateResult.Files {
		lines := len(strings.Split(content, "\n"))
		structs := strings.Count(content, "type ")
		totalLines += lines
		totalStructs += structs

		fmt.Printf("  %s: %d lines, %d structs\n", filename, lines, structs)
	}

	fmt.Printf("\nTotals:\n")
	fmt.Printf("  Lines of code: %d\n", totalLines)
	fmt.Printf("  Struct definitions: %d\n", totalStructs)
	fmt.Printf("  Average lines per file: %.1f\n", float64(totalLines)/float64(len(generateResult.Files)))
}
