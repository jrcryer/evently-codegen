# AsyncAPI Go Code Generator Examples

This directory contains comprehensive examples demonstrating how to use the AsyncAPI Go Code Generator both as a CLI tool and as a Go library.

## Examples Overview

### 1. Basic Usage (`basic_usage/`)
Demonstrates the fundamental usage of the library API:
- Creating a generator with configuration
- Parsing AsyncAPI specifications from strings
- Generating Go code in memory
- Working with parse and generation results

### 2. File Operations (`file_operations/`)
Shows how to work with files:
- Reading AsyncAPI specifications from files
- Writing generated Go code to files
- Directory management and file I/O operations
- Error handling for file operations

### 3. CLI Examples (`cli_examples/`)
Demonstrates CLI usage patterns:
- Basic command-line usage
- Advanced CLI options and configurations
- Batch processing multiple files
- Integration with build scripts

### 4. Advanced Features (`advanced_features/`)
Showcases advanced library features:
- Custom type mapping
- Schema reference resolution
- Error handling strategies
- Performance optimization techniques

### 5. Validation Usage (`validation_usage/`)
Comprehensive validation functionality demonstration:
- Basic struct validation using generated `Validate()` methods
- JSON validation using generated `ValidateJSON()` methods
- Constraint validation (string length, numeric ranges, patterns)
- Enum validation with type-safe constants
- EventBridge-specific validation for AWS events
- Error handling and categorization
- Custom validator configuration (strict vs permissive modes)

### 6. Integration Examples (`integration/`)
Real-world integration scenarios:
- Using with Go modules
- Integration with build systems
- CI/CD pipeline integration
- Testing generated code

## Running the Examples

Each example directory contains a `main.go` file that can be run directly:

```bash
# Run basic usage example
cd examples/basic_usage
go run main.go

# Run file operations example
cd examples/file_operations
go run main.go

# Run CLI examples
cd examples/cli_examples
./run_examples.sh

# Run advanced features example
cd examples/advanced_features
go run main.go

# Run validation usage example
cd examples/validation_usage
go run main.go

# Run integration examples
cd examples/integration
go run main.go
```

## Sample AsyncAPI Files

The examples use sample AsyncAPI specifications located in the `../testdata/` directory:

- `user-service.yaml` - User management API with complex nested schemas
- `ecommerce-api.json` - E-commerce API with orders, payments, and inventory

## Generated Code Examples

Each example includes comments showing what the generated Go code looks like, helping you understand the transformation from AsyncAPI schemas to Go types.

## Validation Features

The generated Go structs include built-in validation methods that provide runtime data validation:

### Generated Validation Methods

Every generated struct includes two validation methods:

```go
// Validate validates the struct instance against its schema
func (s *UserStruct) Validate() *ValidationResult

// ValidateJSON validates raw JSON data against the schema  
func (s *UserStruct) ValidateJSON(jsonData []byte) *ValidationResult
```

### Validation Capabilities

- **Type Validation**: Ensures correct data types (string, number, boolean, array, object)
- **Constraint Validation**: Enforces string length, numeric ranges, patterns, etc.
- **Enum Validation**: Validates enum values with generated type-safe constants
- **Required Field Validation**: Checks for missing required properties
- **EventBridge Support**: Special validation for AWS EventBridge event structures
- **Flexible Configuration**: Strict vs permissive validation modes

### Example Usage

```go
// Validate a struct instance
result := user.Validate()
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Field '%s': %s\n", err.Field, err.Message)
    }
}

// Validate JSON data
jsonData := []byte(`{"name": "John", "age": 30}`)
result = user.ValidateJSON(jsonData)
```

See the `validation_usage` example for comprehensive validation demonstrations.

## Prerequisites

- Go 1.19 or later
- AsyncAPI Go Code Generator installed or built from source

## Tips for Using Examples

1. **Start with Basic Usage**: Begin with the `basic_usage` example to understand core concepts
2. **Explore File Operations**: Move to `file_operations` to see real-world file handling
3. **Try CLI Examples**: Use `cli_examples` to understand command-line usage
4. **Learn Validation**: Check `validation_usage` to understand JSON validation capabilities
5. **Advanced Features**: Explore `advanced_features` for complex scenarios
6. **Integration Patterns**: Check `integration` for production usage patterns

## Customizing Examples

Feel free to modify the examples to experiment with different:
- AsyncAPI specifications
- Generator configurations
- Output formats and destinations
- Error handling approaches
- Validation scenarios and constraints
- EventBridge event structures

## Contributing Examples

If you have useful examples or patterns, please consider contributing them to help other users!