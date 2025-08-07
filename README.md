# AsyncAPI Go Code Generator

A Go library and CLI tool that parses AsyncAPI specification files and generates corresponding Go type definitions. This tool enables developers to automatically generate strongly-typed Go structs from AsyncAPI event schemas, reducing manual coding effort and ensuring consistency between API specifications and Go code.

## Features

- **AsyncAPI Support**: Compatible with AsyncAPI 2.x and 3.x specifications
- **Multiple Formats**: Supports JSON and YAML input formats
- **Go Best Practices**: Generates idiomatic Go code with proper naming conventions
- **Schema Resolution**: Handles external references and complex schema relationships
- **Dual Interface**: Available as both a CLI tool and Go library
- **Comprehensive Testing**: Extensive test coverage with integration and performance tests

## Installation

### Quick Install (Recommended)

Use our installation script to download and install the latest release:

```bash
# Install latest version to /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/jrcryer/evently-codegen/main/scripts/install.sh | bash

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/jrcryer/evently-codegen/main/scripts/install.sh | INSTALL_DIR=~/.local/bin bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/jrcryer/evently-codegen/main/scripts/install.sh | bash -s -- --version v1.0.0
```

### Package Managers

#### Go Install

```bash
# Install latest from source
go install github.com/jrcryer/evently-codegen/cmd/evently-codegen@latest

# Install specific version
go install github.com/jrcryer/evently-codegen/cmd/evently-codegen@v1.0.0
```

### Manual Download

Download pre-built binaries from the [releases page](https://github.com/jrcryer/evently-codegen/releases):

```bash
# Linux amd64
wget https://github.com/jrcryer/evently-codegen/releases/latest/download/evently-codegen-linux-amd64.tar.gz
tar -xzf evently-codegen-linux-amd64.tar.gz
sudo mv evently-codegen-linux-amd64 /usr/local/bin/evently-codegen

# macOS amd64
wget https://github.com/jrcryer/evently-codegen/releases/latest/download/evently-codegen-darwin-amd64.tar.gz
tar -xzf evently-codegen-darwin-amd64.tar.gz
sudo mv evently-codegen-darwin-amd64 /usr/local/bin/evently-codegen

# Windows amd64
# Download evently-codegen-windows-amd64.zip and extract to desired location
```

### Docker

```bash
# Run directly
docker run --rm -v $(pwd):/workspace ghcr.io/jrcryer/evently-codegen:latest \
  -i /workspace/api.yaml -o /workspace/generated -p events

# Pull image
docker pull ghcr.io/jrcryer/evently-codegen:latest

# Build from source
docker build -t evently-codegen .
```

### From Source

```bash
# Clone the repository
git clone https://github.com/jrcryer/evently-codegen.git
cd asyncapi-go-codegen

# Build using Makefile
make build

# Or build manually
go build -o evently-codegen ./cmd/evently-codegen

# Install globally
make install
```

### As a Go Module

```bash
go get github.com/jrcryer/evently-codegen/pkg/generator
```

### Verify Installation

```bash
# Check version
evently-codegen --version

# Show help
evently-codegen --help

# Test with sample file
evently-codegen -i testdata/user-service.yaml -o /tmp/test -p test
```

## Quick Start

### CLI Usage

Generate Go types from an AsyncAPI specification:

```bash
# Basic usage
evently-codegen -i api.yaml -o ./generated -p events

# With verbose output
evently-codegen -i asyncapi.json -o ./types -p myapi -v

# Show help
evently-codegen -h
```

### Library Usage

```go
package main

import (
    "log"
    "github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
    // Create generator with configuration
    config := &generator.Config{
        PackageName:     "events",
        OutputDir:       "./generated",
        IncludeComments: true,
        UsePointers:     true,
    }
    
    gen := generator.NewGenerator(config)
    
    // Parse and generate from file
    err := gen.ParseFileAndGenerateToFiles("api.yaml")
    if err != nil {
        log.Fatalf("Generation failed: %v", err)
    }
    
    log.Println("Go types generated successfully!")
}
```

## CLI Reference

### Command Line Options

| Flag | Long Form | Description | Default |
|------|-----------|-------------|---------|
| `-i` | `--input` | Path to AsyncAPI specification file (required) | - |
| `-o` | `--output` | Output directory for generated Go files | `./generated` |
| `-p` | `--package` | Package name for generated Go code | `main` |
| `-v` | `--verbose` | Enable verbose output | `false` |
| `-h` | `--help` | Show help information | - |

### Examples

```bash
# Generate types for a user service API
evently-codegen -i user-service.yaml -o ./pkg/events -p userevents

# Process a complex e-commerce API with verbose output
evently-codegen --input ecommerce-api.json --output ./internal/types --package commerce --verbose

# Generate with default settings
evently-codegen -i api.yml
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Parse error |
| 4 | Generation error |
| 5 | Validation error |
| 6 | File error |

## Library API

### Core Types

```go
// Generator provides the main API for code generation
type Generator struct {
    // ...
}

// Config holds generation configuration
type Config struct {
    PackageName     string  // Go package name
    OutputDir       string  // Output directory
    IncludeComments bool    // Include schema descriptions as comments
    UsePointers     bool    // Use pointers for optional fields
}

// ParseResult contains parsed AsyncAPI data
type ParseResult struct {
    Spec     *AsyncAPISpec
    Messages map[string]*MessageSchema
    Errors   []error
}

// GenerateResult contains generated Go code
type GenerateResult struct {
    Files   map[string]string  // filename -> content
    Errors  []error
}
```

### Main Methods

```go
// Create a new generator
func NewGenerator(config *Config) *Generator

// Parse AsyncAPI specification from bytes
func (g *Generator) Parse(data []byte) (*ParseResult, error)

// Parse AsyncAPI specification from file
func (g *Generator) ParseFile(filePath string) (*ParseResult, error)

// Generate Go code from parsed messages
func (g *Generator) Generate(messages map[string]*MessageSchema) (*GenerateResult, error)

// Convenience method: parse and generate in one call
func (g *Generator) ParseAndGenerate(data []byte) (*GenerateResult, error)

// Convenience method: parse file and generate to files
func (g *Generator) ParseFileAndGenerateToFiles(filePath string) error
```

## JSON Validation

The generated Go structs include built-in validation methods that allow you to validate JSON data against the original AsyncAPI schema at runtime. This ensures data integrity and helps catch schema violations early.

### Validation Features

- **Schema Validation**: Validates JSON data against AsyncAPI schema constraints
- **Type Validation**: Ensures correct data types (string, number, boolean, array, object)
- **Constraint Validation**: Enforces string length, numeric ranges, array size limits
- **Enum Validation**: Validates enum values with type safety
- **Required Field Validation**: Checks for missing required properties
- **EventBridge Support**: Special validation for AWS EventBridge event structures

### Generated Validation Methods

Every generated struct includes two validation methods:

```go
// Validate validates the struct instance against its schema
func (s *UserSignupPayload) Validate() *ValidationResult

// ValidateJSON validates raw JSON data against the schema
func (s *UserSignupPayload) ValidateJSON(jsonData []byte) *ValidationResult
```

### Basic Usage Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/jrcryer/evently-codegen/pkg/generator"
)

func main() {
    // Create a struct instance
    user := &UserSignupPayload{
        UserId: "user123",
        Email:  "user@example.com",
        Profile: &UserSignupPayloadProfile{
            FirstName: stringPtr("John"),
            LastName:  stringPtr("Doe"),
            Age:       intPtr(30),
        },
    }
    
    // Validate the struct instance
    result := user.Validate()
    if !result.Valid {
        log.Printf("Validation failed: %v", result.Errors)
        return
    }
    
    fmt.Println("Struct validation passed!")
    
    // Validate JSON data
    jsonData := []byte(`{
        "userId": "user456",
        "email": "jane@example.com",
        "profile": {
            "firstName": "Jane",
            "age": 25
        }
    }`)
    
    result = user.ValidateJSON(jsonData)
    if !result.Valid {
        log.Printf("JSON validation failed: %v", result.Errors)
        return
    }
    
    fmt.Println("JSON validation passed!")
}

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int { return &i }
```

### Validation Error Handling

The `ValidationResult` type provides detailed error information:

```go
type ValidationResult struct {
    Valid  bool               `json:"valid"`
    Errors []*ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

Example error handling:

```go
result := user.ValidateJSON(invalidJSON)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Field '%s': %s\n", err.Field, err.Message)
    }
}
```

### Enum Validation

Generated enum types include validation methods:

```go
// For an enum field like "status" with values ["active", "inactive", "pending"]
type UserStatusEnum string

const (
    UserStatusActive   UserStatusEnum = "active"
    UserStatusInactive UserStatusEnum = "inactive"
    UserStatusPending  UserStatusEnum = "pending"
)

// IsValid returns true if the enum value is valid
func (e UserStatusEnum) IsValid() bool {
    switch e {
    case UserStatusActive, UserStatusInactive, UserStatusPending:
        return true
    default:
        return false
    }
}

// UserStatusEnumValues returns all valid values
func UserStatusEnumValues() []UserStatusEnum {
    return []UserStatusEnum{
        UserStatusActive,
        UserStatusInactive,
        UserStatusPending,
    }
}
```

### EventBridge Validation

For AWS EventBridge events, use the specialized EventBridge validator:

```go
import "github.com/jrcryer/evently-codegen/pkg/generator"

func validateEventBridgeEvent() {
    validator := generator.NewEventBridgeValidator()
    
    // Validate EventBridge event structure
    eventJSON := []byte(`{
        "version": "0",
        "id": "12345678-1234-1234-1234-123456789012",
        "detail-type": "User Signup",
        "source": "com.example.userservice",
        "account": "123456789012",
        "time": "2023-01-01T12:00:00Z",
        "region": "us-east-1",
        "detail": {
            "userId": "user123",
            "email": "user@example.com"
        }
    }`)
    
    result := validator.ValidateEventBridgeEvent(eventJSON)
    if !result.Valid {
        log.Printf("EventBridge validation failed: %v", result.Errors)
        return
    }
    
    // Validate with detail schema
    detailSchema := &generator.MessageSchema{
        Type: "object",
        Properties: map[string]*generator.Property{
            "userId": {Type: "string"},
            "email":  {Type: "string"},
        },
        Required: []string{"userId", "email"},
    }
    
    result = validator.ValidateEventBridgeEventWithSchema(eventJSON, detailSchema)
    if !result.Valid {
        log.Printf("Detail validation failed: %v", result.Errors)
        return
    }
    
    fmt.Println("EventBridge event validation passed!")
}
```

### Validation Configuration

Create validators with different configurations:

```go
// Strict mode - rejects additional properties
strictValidator := generator.NewValidator(true)

// Permissive mode - allows additional properties (default)
permissiveValidator := generator.NewValidator(false)

// Use custom validator
result := strictValidator.ValidateJSON(jsonData, schema)
```

## Generated Code Examples

### Input AsyncAPI Schema

```yaml
asyncapi: '2.6.0'
info:
  title: User Service API
  version: '1.0.0'
channels:
  user/signup:
    publish:
      message:
        payload:
          type: object
          properties:
            userId:
              type: string
              description: Unique identifier for the user
            email:
              type: string
              format: email
              description: User's email address
            status:
              type: string
              enum: ["active", "inactive", "pending"]
              description: User status
            profile:
              type: object
              properties:
                firstName:
                  type: string
                lastName:
                  type: string
                age:
                  type: integer
                  minimum: 0
                  maximum: 150
          required: [userId, email]
```

### Generated Go Code

```go
package events

// UserSignupPayload represents the payload for user/signup channel
type UserSignupPayload struct {
    // Unique identifier for the user
    UserId string `json:"userId"`
    
    // User's email address
    Email string `json:"email"`
    
    // User status
    Status UserSignupPayloadStatusEnum `json:"status,omitempty"`
    
    Profile *UserSignupPayloadProfile `json:"profile,omitempty"`
}

// UserSignupPayloadProfile represents the profile object
type UserSignupPayloadProfile struct {
    FirstName *string `json:"firstName,omitempty"`
    LastName  *string `json:"lastName,omitempty"`
    Age       *int    `json:"age,omitempty"`
}

// UserSignupPayloadStatusEnum represents the status enum
type UserSignupPayloadStatusEnum string

const (
    UserSignupPayloadStatusActive   UserSignupPayloadStatusEnum = "active"
    UserSignupPayloadStatusInactive UserSignupPayloadStatusEnum = "inactive"
    UserSignupPayloadStatusPending  UserSignupPayloadStatusEnum = "pending"
)

// IsValid returns true if the enum value is valid
func (e UserSignupPayloadStatusEnum) IsValid() bool {
    switch e {
    case UserSignupPayloadStatusActive, UserSignupPayloadStatusInactive, UserSignupPayloadStatusPending:
        return true
    default:
        return false
    }
}

// Validate validates the UserSignupPayload struct against its schema
func (s *UserSignupPayload) Validate() *ValidationResult {
    validator := NewValidator(false)
    schema := &Property{
        Type: "object",
        Properties: map[string]*Property{
            "userId": {Type: "string"},
            "email":  {Type: "string"},
            "status": {Type: "string", Enum: []interface{}{"active", "inactive", "pending"}},
            "profile": {Type: "object"},
        },
        Required: []string{"userId", "email"},
    }
    
    data := map[string]any{
        "userId":  s.UserId,
        "email":   s.Email,
        "status":  s.Status,
        "profile": s.Profile,
    }
    
    return validator.ValidateValue(data, schema, "")
}

// ValidateJSON validates raw JSON data against the UserSignupPayload schema
func (s *UserSignupPayload) ValidateJSON(jsonData []byte) *ValidationResult {
    validator := NewValidator(false)
    schema := &MessageSchema{
        Type: "object",
        Properties: map[string]*Property{
            "userId": {Type: "string"},
            "email":  {Type: "string"},
            "status": {Type: "string", Enum: []interface{}{"active", "inactive", "pending"}},
            "profile": {Type: "object"},
        },
        Required: []string{"userId", "email"},
    }
    
    return validator.ValidateJSON(jsonData, schema)
}
```

## Supported AsyncAPI Features

### Schema Types

- ✅ Primitive types (string, number, integer, boolean)
- ✅ Object types with nested properties
- ✅ Array types with typed elements
- ✅ Optional and required fields
- ✅ String formats (date-time, email, etc.)
- ✅ Enumerations with type-safe constants
- ✅ Schema references ($ref)

### Validation Features

- ✅ JSON schema validation against AsyncAPI specifications
- ✅ Type validation (string, number, boolean, array, object)
- ✅ Constraint validation (min/max length, numeric ranges, patterns)
- ✅ Enum validation with generated type-safe constants
- ✅ Required field validation
- ✅ EventBridge event structure validation
- ✅ Configurable strict/permissive validation modes

### AsyncAPI Versions

- ✅ AsyncAPI 2.6.0
- ✅ AsyncAPI 3.0.0
- ⚠️ Earlier versions (limited support)

### Input Formats

- ✅ JSON (.json)
- ✅ YAML (.yaml, .yml)

## Project Structure

```
.
├── cmd/
│   └── evently-codegen/         # CLI application entry point
├── pkg/
│   └── generator/                # Core library package
│       ├── types.go              # Data structures and types
│       ├── interfaces.go         # Core interfaces
│       ├── errors.go             # Error types
│       ├── generator.go          # Main Generator API
│       ├── parser.go             # AsyncAPI parser
│       ├── code_generator.go     # Go code generation
│       ├── type_mapper.go        # Type mapping logic
│       ├── resolver.go           # Schema reference resolver
│       ├── file_manager.go       # File I/O operations
│       └── *_test.go             # Unit and integration tests
├── internal/
│   └── cli/                      # CLI-specific internal code
│       ├── config.go             # CLI configuration
│       └── *_test.go             # CLI tests
├── examples/                     # Usage examples
│   ├── basic_usage/              # Basic library usage
│   └── file_operations/          # File-based operations
├── testdata/                     # Test AsyncAPI specifications
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── Makefile                      # Build and development tasks
└── README.md                     # This file
```

## Development

### Building

```bash
# Build the CLI tool
make build

# Build for all platforms
make build-all

# Install locally
make install
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run performance tests
make test-performance
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security checks
make security
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run the test suite (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add comprehensive tests for new features
- Update documentation for API changes
- Ensure backward compatibility when possible
- Use semantic versioning for releases

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://github.com/jrcryer/evently-codegen/wiki)
- 🐛 [Issue Tracker](https://github.com/jrcryer/evently-codegen/issues)
- 💬 [Discussions](https://github.com/jrcryer/evently-codegen/discussions)

## Acknowledgments

- [AsyncAPI Initiative](https://www.asyncapi.com/) for the specification
- Go community for excellent tooling and libraries
- Contributors and users of this project