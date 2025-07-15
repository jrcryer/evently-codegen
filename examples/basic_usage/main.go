package main

import (
	"fmt"
	"log"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
	// Example AsyncAPI specification
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "User Service API",
			"version": "1.0.0",
			"description": "API for managing user events"
		},
		"channels": {
			"user/signup": {
				"publish": {
					"message": {
						"payload": {
							"type": "object",
							"properties": {
								"userId": {
									"type": "string",
									"description": "Unique identifier for the user"
								},
								"email": {
									"type": "string",
									"format": "email",
									"description": "User's email address"
								},
								"name": {
									"type": "string",
									"description": "User's full name"
								},
								"createdAt": {
									"type": "string",
									"format": "date-time",
									"description": "Account creation timestamp"
								}
							},
							"required": ["userId", "email"]
						}
					}
				}
			}
		}
	}`

	// Create a new generator with custom configuration
	config := &generator.Config{
		PackageName:     "events",
		OutputDir:       "./generated",
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := generator.NewGenerator(config)

	// Validate the configuration
	if err := gen.ValidateConfig(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Parse the AsyncAPI specification
	parseResult, err := gen.Parse([]byte(asyncAPISpec))
	if err != nil {
		log.Fatalf("Failed to parse AsyncAPI spec: %v", err)
	}

	// Check for parsing errors
	if len(parseResult.Errors) > 0 {
		fmt.Printf("Parsing completed with %d warnings:\n", len(parseResult.Errors))
		for _, parseErr := range parseResult.Errors {
			fmt.Printf("  - %v\n", parseErr)
		}
	}

	fmt.Printf("Successfully parsed AsyncAPI spec: %s v%s\n",
		parseResult.Spec.Info.Title, parseResult.Spec.Info.Version)
	fmt.Printf("Found %d message schemas\n", len(parseResult.Messages))

	// Generate Go code from the parsed messages
	generateResult, err := gen.Generate(parseResult.Messages)
	if err != nil {
		log.Fatalf("Failed to generate Go code: %v", err)
	}

	// Check for generation errors
	if len(generateResult.Errors) > 0 {
		fmt.Printf("Generation completed with %d warnings:\n", len(generateResult.Errors))
		for _, genErr := range generateResult.Errors {
			fmt.Printf("  - %v\n", genErr)
		}
	}

	// Display generated files
	fmt.Printf("Generated %d Go files:\n", len(generateResult.Files))
	for filename, content := range generateResult.Files {
		fmt.Printf("\n=== %s ===\n", filename)
		fmt.Println(content)
	}

	// Alternative: Use the convenience method to parse and generate in one step
	fmt.Println("\n=== Using ParseAndGenerate convenience method ===")
	result, err := gen.ParseAndGenerate([]byte(asyncAPISpec))
	if err != nil {
		log.Fatalf("Failed to parse and generate: %v", err)
	}

	fmt.Printf("Generated %d files using convenience method\n", len(result.Files))
}
