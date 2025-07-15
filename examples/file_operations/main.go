package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
	// Create a sample AsyncAPI file
	sampleAsyncAPI := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "E-commerce Events API",
			"version": "1.2.0",
			"description": "Event-driven API for e-commerce operations"
		},
		"components": {
			"schemas": {
				"Order": {
					"type": "object",
					"description": "Customer order information",
					"properties": {
						"orderId": {
							"type": "string",
							"description": "Unique order identifier"
						},
						"customerId": {
							"type": "string",
							"description": "Customer identifier"
						},
						"items": {
							"type": "array",
							"description": "List of ordered items",
							"items": {
								"type": "object",
								"properties": {
									"productId": {"type": "string"},
									"quantity": {"type": "integer"},
									"price": {"type": "number"}
								}
							}
						},
						"totalAmount": {
							"type": "number",
							"description": "Total order amount"
						},
						"status": {
							"type": "string",
							"enum": ["pending", "confirmed", "shipped", "delivered"],
							"description": "Order status"
						},
						"createdAt": {
							"type": "string",
							"format": "date-time",
							"description": "Order creation timestamp"
						}
					},
					"required": ["orderId", "customerId", "totalAmount", "status"]
				},
				"Customer": {
					"type": "object",
					"description": "Customer information",
					"properties": {
						"customerId": {
							"type": "string",
							"description": "Unique customer identifier"
						},
						"email": {
							"type": "string",
							"format": "email",
							"description": "Customer email address"
						},
						"firstName": {
							"type": "string",
							"description": "Customer first name"
						},
						"lastName": {
							"type": "string",
							"description": "Customer last name"
						},
						"address": {
							"type": "object",
							"description": "Customer address",
							"properties": {
								"street": {"type": "string"},
								"city": {"type": "string"},
								"state": {"type": "string"},
								"zipCode": {"type": "string"},
								"country": {"type": "string"}
							}
						}
					},
					"required": ["customerId", "email", "firstName", "lastName"]
				}
			}
		}
	}`

	// Create temporary directory for this example
	tempDir, err := os.MkdirTemp("", "asyncapi-example-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	fmt.Printf("Working in temporary directory: %s\n", tempDir)

	// Write the AsyncAPI spec to a file
	specFile := filepath.Join(tempDir, "ecommerce-api.json")
	if err := os.WriteFile(specFile, []byte(sampleAsyncAPI), 0644); err != nil {
		log.Fatalf("Failed to write AsyncAPI spec file: %v", err)
	}

	// Create generator with configuration for file operations
	config := &generator.Config{
		PackageName:     "ecommerce",
		OutputDir:       filepath.Join(tempDir, "generated"),
		IncludeComments: true,
		UsePointers:     false, // Use value types instead of pointers
	}

	gen := generator.NewGenerator(config)

	// Example 1: Parse from file and generate to files
	fmt.Println("=== Example 1: Parse file and generate to files ===")

	err = gen.ParseFileAndGenerateToFiles(specFile)
	if err != nil {
		log.Fatalf("Failed to parse file and generate: %v", err)
	}

	fmt.Println("Successfully generated Go files from AsyncAPI specification")

	// List generated files
	outputDir := config.OutputDir
	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Fatalf("Failed to read output directory: %v", err)
	}

	fmt.Printf("Generated files in %s:\n", outputDir)
	for _, file := range files {
		if !file.IsDir() {
			fmt.Printf("  - %s\n", file.Name())

			// Show content of each generated file
			filePath := filepath.Join(outputDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("    Error reading file: %v\n", err)
				continue
			}

			fmt.Printf("    Content preview (first 200 chars):\n")
			preview := string(content)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("    %s\n\n", preview)
		}
	}

	// Example 2: Parse file and work with results programmatically
	fmt.Println("=== Example 2: Parse file and work with results programmatically ===")

	parseResult, err := gen.ParseFile(specFile)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	fmt.Printf("Parsed specification: %s\n", parseResult.Spec.Info.Title)
	fmt.Printf("API Version: %s\n", parseResult.Spec.Info.Version)
	fmt.Printf("Description: %s\n", parseResult.Spec.Info.Description)

	fmt.Printf("\nFound %d schemas:\n", len(parseResult.Messages))
	for name, schema := range parseResult.Messages {
		fmt.Printf("  - %s: %s (type: %s)\n", name, schema.Description, schema.Type)
		if len(schema.Properties) > 0 {
			fmt.Printf("    Properties: %d\n", len(schema.Properties))
			for propName, prop := range schema.Properties {
				fmt.Printf("      - %s: %s\n", propName, prop.Type)
			}
		}
	}

	// Example 3: Generate code and work with it in memory
	fmt.Println("\n=== Example 3: Generate code in memory ===")

	generateResult, err := gen.Generate(parseResult.Messages)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	fmt.Printf("Generated %d Go files in memory:\n", len(generateResult.Files))
	for filename, content := range generateResult.Files {
		fmt.Printf("\n--- %s ---\n", filename)
		lines := len(strings.Split(content, "\n"))
		fmt.Printf("Lines of code: %d\n", lines)

		// Count structs in the generated code
		structCount := strings.Count(content, "type ") - strings.Count(content, "type (")
		fmt.Printf("Struct definitions: %d\n", structCount)
	}

	// Example 4: Show supported AsyncAPI versions
	fmt.Println("\n=== Example 4: Supported AsyncAPI versions ===")
	versions := gen.GetSupportedVersions()
	fmt.Printf("Supported AsyncAPI versions: %v\n", versions)

	fmt.Println("\nExample completed successfully!")
}
