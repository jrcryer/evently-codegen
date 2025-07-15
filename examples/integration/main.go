package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
	fmt.Println("=== AsyncAPI Go Code Generator - Integration Examples ===")
	fmt.Println()

	// Example 1: Generate and compile code
	fmt.Println("=== Example 1: Generate and Compile Code ===")
	demonstrateGenerateAndCompile()
	fmt.Println()

	// Example 2: Integration with Go modules
	fmt.Println("=== Example 2: Integration with Go Modules ===")
	demonstrateGoModuleIntegration()
	fmt.Println()

	// Example 3: Testing generated code
	fmt.Println("=== Example 3: Testing Generated Code ===")
	demonstrateTestingGeneratedCode()
	fmt.Println()

	// Example 4: CI/CD pipeline integration
	fmt.Println("=== Example 4: CI/CD Pipeline Integration ===")
	demonstrateCICDIntegration()
	fmt.Println()

	// Example 5: Code quality validation
	fmt.Println("=== Example 5: Code Quality Validation ===")
	demonstrateCodeQualityValidation()
	fmt.Println()

	fmt.Println("Integration examples completed!")
}

func demonstrateGenerateAndCompile() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "asyncapi-integration-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Working directory: %s\n", tempDir)

	// Sample AsyncAPI specification
	asyncAPISpec := `{
		"asyncapi": "2.6.0",
		"info": {
			"title": "Integration Test API",
			"version": "1.0.0"
		},
		"channels": {
			"user/created": {
				"publish": {
					"message": {
						"payload": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"email": {"type": "string", "format": "email"},
								"name": {"type": "string"},
								"createdAt": {"type": "string", "format": "date-time"}
							},
							"required": ["id", "email", "name"]
						}
					}
				}
			}
		}
	}`

	// Generate Go code
	config := &generator.Config{
		PackageName:     "events",
		OutputDir:       filepath.Join(tempDir, "pkg", "events"),
		IncludeComments: true,
		UsePointers:     true,
	}

	gen := generator.NewGenerator(config)
	err = gen.ParseAndGenerateToFiles([]byte(asyncAPISpec))
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	fmt.Println("✓ Generated Go code successfully")

	// Create a Go module
	err = createGoModule(tempDir, "example.com/integration-test")
	if err != nil {
		log.Printf("Failed to create Go module: %v", err)
		return
	}

	fmt.Println("✓ Created Go module")

	// Create a main.go file that uses the generated types
	mainGoContent := `package main

import (
	"encoding/json"
	"fmt"
	"time"

	"example.com/integration-test/pkg/events"
)

func main() {
	// Create an instance of the generated type
	event := &events.UserCreatedPayload{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated event: %s\n", string(jsonData))

	// Unmarshal back
	var decoded events.UserCreatedPayload
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Decoded event: %+v\n", decoded)
	fmt.Println("Integration test passed!")
}
`

	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGoContent), 0644)
	if err != nil {
		log.Printf("Failed to write main.go: %v", err)
		return
	}

	fmt.Println("✓ Created main.go using generated types")

	// Try to build the project
	err = buildGoProject(tempDir)
	if err != nil {
		log.Printf("Failed to build project: %v", err)
		return
	}

	fmt.Println("✓ Successfully built project with generated types")

	// Try to run the project
	err = runGoProject(tempDir)
	if err != nil {
		log.Printf("Failed to run project: %v", err)
		return
	}

	fmt.Println("✓ Successfully ran project with generated types")
}

func demonstrateGoModuleIntegration() {
	fmt.Println("Creating example Go module structure...")

	// Create a realistic project structure
	projectStructure := `
project/
├── go.mod
├── go.sum
├── cmd/
│   └── server/
│       └── main.go
├── pkg/
│   ├── events/          # Generated from AsyncAPI
│   ├── handlers/        # Event handlers
│   └── services/        # Business logic
├── api/
│   ├── user-events.yaml
│   └── order-events.yaml
├── scripts/
│   └── generate.sh
└── Makefile
`

	fmt.Println("Recommended project structure:")
	fmt.Println(projectStructure)

	// Show example go.mod content
	goModContent := `module example.com/event-service

go 1.19

require (
	github.com/jrcryer/evently-codegen v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/nats-io/nats.go v1.28.0
)
`

	fmt.Println("Example go.mod:")
	fmt.Println(goModContent)

	// Show example generation script
	generateScript := `#!/bin/bash
# scripts/generate.sh - Generate Go types from AsyncAPI specifications

set -e

echo "Generating Go types from AsyncAPI specifications..."

# Generate user events
evently-codegen -i api/user-events.yaml -o pkg/events/user -p userevents

# Generate order events  
evently-codegen -i api/order-events.yaml -o pkg/events/order -p orderevents

echo "Code generation completed successfully!"

# Format generated code
go fmt ./pkg/events/...

# Run tests to ensure generated code is valid
go test ./pkg/events/...

echo "Generated code validated successfully!"
`

	fmt.Println("Example generation script (scripts/generate.sh):")
	fmt.Println(generateScript)

	// Show example Makefile
	makefileContent := `# Makefile for event service

.PHONY: generate build test clean run

# Generate Go types from AsyncAPI specifications
generate:
	@echo "Generating Go types..."
	@./scripts/generate.sh

# Build the application
build: generate
	@echo "Building application..."
	@go build -o bin/server ./cmd/server

# Run tests
test: generate
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf pkg/events/*/

# Run the application
run: build
	@echo "Starting server..."
	@./bin/server

# Development workflow
dev: clean generate test build
	@echo "Development build completed!"

# CI/CD pipeline
ci: generate test build
	@echo "CI pipeline completed!"
`

	fmt.Println("Example Makefile:")
	fmt.Println(makefileContent)
}

func demonstrateTestingGeneratedCode() {
	fmt.Println("Demonstrating testing strategies for generated code...")

	// Example test file content
	testFileContent := `package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"example.com/event-service/pkg/events/user"
)

func TestUserCreatedPayload_JSONSerialization(t *testing.T) {
	// Test data
	event := &user.UserCreatedPayload{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Test marshaling
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Test unmarshaling
	var decoded user.UserCreatedPayload
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify data integrity
	if decoded.Id != event.Id {
		t.Errorf("Expected Id %s, got %s", event.Id, decoded.Id)
	}
	if decoded.Email != event.Email {
		t.Errorf("Expected Email %s, got %s", event.Email, decoded.Email)
	}
	if decoded.Name != event.Name {
		t.Errorf("Expected Name %s, got %s", event.Name, decoded.Name)
	}
}

func TestUserCreatedPayload_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		event   *user.UserCreatedPayload
		wantErr bool
	}{
		{
			name: "valid event",
			event: &user.UserCreatedPayload{
				Id:    "user-123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			event: &user.UserCreatedPayload{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real scenario, you would validate required fields
			// This is just an example of how to structure tests
			if tt.event.Id == "" && !tt.wantErr {
				t.Error("Expected valid event, but Id is empty")
			}
		})
	}
}

func BenchmarkUserCreatedPayload_Marshal(b *testing.B) {
	event := &user.UserCreatedPayload{
		Id:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(event)
		if err != nil {
			b.Fatal(err)
		}
	}
}
`

	fmt.Println("Example test file (pkg/events/user/user_test.go):")
	fmt.Println(testFileContent)

	// Testing strategies
	fmt.Println("\nTesting Strategies for Generated Code:")
	fmt.Println("1. JSON Serialization/Deserialization Tests")
	fmt.Println("2. Required Field Validation Tests")
	fmt.Println("3. Type Safety Tests")
	fmt.Println("4. Performance Benchmarks")
	fmt.Println("5. Integration Tests with Real Data")
	fmt.Println("6. Schema Compatibility Tests")
}

func demonstrateCICDIntegration() {
	fmt.Println("CI/CD Pipeline Integration Examples...")

	// GitHub Actions workflow
	githubActionsWorkflow := `name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Install AsyncAPI Code Generator
      run: |
        go install github.com/jrcryer/evently-codegen/cmd/evently-codegen@latest
    
    - name: Generate Go types
      run: |
        make generate
    
    - name: Verify generated code
      run: |
        git diff --exit-code || (echo "Generated code is not up to date" && exit 1)
    
    - name: Run tests
      run: |
        go test -v ./...
    
    - name: Build application
      run: |
        go build -v ./...
    
    - name: Run integration tests
      run: |
        make test-integration
`

	fmt.Println("GitHub Actions Workflow (.github/workflows/build.yml):")
	fmt.Println(githubActionsWorkflow)

	// GitLab CI configuration
	gitlabCI := `stages:
  - generate
  - test
  - build

variables:
  GO_VERSION: "1.19"

before_script:
  - apt-get update -qq && apt-get install -y -qq git
  - go version

generate:
  stage: generate
  script:
    - go install github.com/jrcryer/evently-codegen/cmd/evently-codegen@latest
    - make generate
    - git diff --exit-code || (echo "Generated code is not up to date" && exit 1)
  artifacts:
    paths:
      - pkg/events/

test:
  stage: test
  dependencies:
    - generate
  script:
    - go test -v ./...
    - go test -race ./...
    - go test -coverprofile=coverage.out ./...
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

build:
  stage: build
  dependencies:
    - generate
  script:
    - go build -v ./...
  artifacts:
    paths:
      - bin/
`

	fmt.Println("GitLab CI Configuration (.gitlab-ci.yml):")
	fmt.Println(gitlabCI)

	// Docker integration
	dockerfile := `FROM golang:1.19-alpine AS builder

# Install AsyncAPI Code Generator
RUN go install github.com/jrcryer/evently-codegen/cmd/evently-codegen@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Go types
RUN make generate

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
`

	fmt.Println("Dockerfile with code generation:")
	fmt.Println(dockerfile)
}

func demonstrateCodeQualityValidation() {
	fmt.Println("Code Quality Validation for Generated Code...")

	// Create a temporary directory for validation
	tempDir, err := os.MkdirTemp("", "asyncapi-validation-*")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Generate some sample code
	sampleSpec := `{
		"asyncapi": "2.6.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"channels": {
			"test/event": {
				"publish": {
					"message": {
						"payload": {
							"type": "object",
							"properties": {
								"id": {"type": "string"},
								"name": {"type": "string"}
							}
						}
					}
				}
			}
		}
	}`

	config := &generator.Config{
		PackageName: "validation",
		OutputDir:   filepath.Join(tempDir, "generated"),
	}

	gen := generator.NewGenerator(config)
	err = gen.ParseAndGenerateToFiles([]byte(sampleSpec))
	if err != nil {
		log.Printf("Failed to generate code: %v", err)
		return
	}

	fmt.Println("✓ Generated sample code for validation")

	// Validate Go syntax
	err = validateGoSyntax(filepath.Join(tempDir, "generated"))
	if err != nil {
		log.Printf("Syntax validation failed: %v", err)
		return
	}

	fmt.Println("✓ Go syntax validation passed")

	// Check code formatting
	err = checkGoFormatting(filepath.Join(tempDir, "generated"))
	if err != nil {
		log.Printf("Formatting check failed: %v", err)
		return
	}

	fmt.Println("✓ Go formatting check passed")

	// Quality validation checklist
	fmt.Println("\nCode Quality Validation Checklist:")
	fmt.Println("1. ✓ Go syntax validation")
	fmt.Println("2. ✓ Go formatting (gofmt)")
	fmt.Println("3. □ Linting (golint, golangci-lint)")
	fmt.Println("4. □ Vet analysis (go vet)")
	fmt.Println("5. □ Security scanning (gosec)")
	fmt.Println("6. □ Dependency vulnerability scanning")
	fmt.Println("7. □ Code coverage analysis")
	fmt.Println("8. □ Performance benchmarking")

	// Example quality validation script
	qualityScript := `#!/bin/bash
# scripts/quality-check.sh - Comprehensive code quality validation

set -e

echo "Running code quality checks..."

# 1. Generate code
make generate

# 2. Format check
echo "Checking code formatting..."
if ! gofmt -l pkg/events/ | grep -q .; then
    echo "✓ Code is properly formatted"
else
    echo "✗ Code formatting issues found:"
    gofmt -l pkg/events/
    exit 1
fi

# 3. Vet analysis
echo "Running go vet..."
go vet ./...
echo "✓ go vet passed"

# 4. Linting
echo "Running golangci-lint..."
golangci-lint run ./...
echo "✓ Linting passed"

# 5. Security scanning
echo "Running security scan..."
gosec ./...
echo "✓ Security scan passed"

# 6. Tests with coverage
echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo "✓ Tests passed with coverage report generated"

# 7. Dependency check
echo "Checking for known vulnerabilities..."
go list -json -m all | nancy sleuth
echo "✓ Dependency vulnerability check passed"

echo "All quality checks passed!"
`

	fmt.Println("\nExample Quality Validation Script:")
	fmt.Println(qualityScript)
}

// Helper functions

func createGoModule(dir, moduleName string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = dir
	return cmd.Run()
}

func buildGoProject(dir string) error {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	return cmd.Run()
}

func runGoProject(dir string) error {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run failed: %v, output: %s", err, output)
	}
	fmt.Printf("Program output:\n%s", output)
	return nil
}

func validateGoSyntax(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("syntax error in %s: %v", path, err)
		}

		return nil
	})
}

func checkGoFormatting(dir string) error {
	cmd := exec.Command("gofmt", "-l", dir)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %v", err)
	}

	if len(output) > 0 {
		return fmt.Errorf("files not properly formatted: %s", string(output))
	}

	return nil
}
