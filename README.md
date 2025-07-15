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
            profile:
              type: object
              properties:
                firstName:
                  type: string
                lastName:
                  type: string
                age:
                  type: integer
          required: [userId, email]
```

### Generated Go Code

```go
package events

import (
    "time"
)

// UserSignupPayload represents the payload for user/signup channel
type UserSignupPayload struct {
    // Unique identifier for the user
    UserId string `json:"userId"`
    
    // User's email address
    Email string `json:"email"`
    
    Profile *UserSignupPayloadProfile `json:"profile,omitempty"`
}

// UserSignupPayloadProfile represents the profile object
type UserSignupPayloadProfile struct {
    FirstName *string `json:"firstName,omitempty"`
    LastName  *string `json:"lastName,omitempty"`
    Age       *int    `json:"age,omitempty"`
}
```

## Supported AsyncAPI Features

### Schema Types

- ✅ Primitive types (string, number, integer, boolean)
- ✅ Object types with nested properties
- ✅ Array types with typed elements
- ✅ Optional and required fields
- ✅ String formats (date-time, email, etc.)
- ✅ Enumerations
- ✅ Schema references ($ref)

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